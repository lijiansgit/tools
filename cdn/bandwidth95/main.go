package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	log "github.com/lijiansgit/go/libs/log4go"

	"github.com/lijiansgit/go/libs/db"
)

const (
	// TIMEFORMAT 时间格式
	TIMEFORMAT = "2006-01-02 15:04:05"
	// TIMES 纳秒
	TIMES = int64(1000000000)
)

var (
	addr         string
	user         string
	password     string
	dbName       string
	table        string
	streamFile   string
	timeStr      string
	startTime    string
	endTime      string
	streams      []string
	streamCounts map[string][]int
	influx       *db.InfluxDB
	wg           *sync.WaitGroup
	lock         *sync.RWMutex
	logLevel     int
	logAll       bool
)

func init() {
	flag.StringVar(&addr, "addr", "http://127.0.0.1:8086", "influxdb addr")
	flag.StringVar(&user, "user", "", "influxdb user")
	flag.StringVar(&password, "password", "", "influxdb user's password")
	flag.StringVar(&dbName, "db", "stream_stats", "influxdb database")
	flag.StringVar(&table, "table", "stream_tx", "influxdb table")
	flag.StringVar(&streamFile, "streamFile", "./stream", "只统计文件里流ID的带宽信息，文件为空则统计全部")
	flag.StringVar(&timeStr, "t", "1s", "每隔多长时间启动一个协程，默认为1秒，eg: 1us, 1ms")
	flag.StringVar(&startTime, "st", "2018-06-01 00:00:00", "统计开始时间")
	flag.StringVar(&endTime, "et", "2018-07-01 00:00:00", "统计结束时间")
	flag.IntVar(&logLevel, "logLevel", 0, "日志等级")
	flag.BoolVar(&logAll, "logAll", false, "是否只打印总统计")
	// init
	wg = new(sync.WaitGroup)
	lock = new(sync.RWMutex)
	streamCounts = make(map[string][]int)
}

func main() {
	flag.Parse()

	if logLevel == 0 {
		log.AddFilter("stdout", log.DEBUG, log.NewConsoleLogWriter())
	} else {
		log.AddFilter("stdout", log.INFO, log.NewConsoleLogWriter())
	}
	defer log.Close()

	timeDur, err := time.ParseDuration(timeStr)
	if err != nil {
		panic(err)
	}

	t, err := time.ParseInLocation(TIMEFORMAT, startTime, time.Local)
	if err != nil {
		panic(err)
	}

	startTS := t.UnixNano()
	t, err = time.ParseInLocation(TIMEFORMAT, endTime, time.Local)
	if err != nil {
		panic(err)
	}

	endTS := t.UnixNano()
	log.Info("table: %s, start: %s(%d), end: %s(%d)",
		table, startTime, startTS, endTime, endTS)

	runStart := time.Now()
	influx = db.NewInfluxDB(addr, user, password, dbName, table)
	streams, err = openFile(streamFile)
	if err != nil {
		panic(err)
	}

	log.Debug("read streams: %v", streams)

	for startTS < endTS {
		wg.Add(1)

		go stats(startTS)

		time.Sleep(timeDur)
		startTS += int64(5 * time.Minute)
	}

	wg.Wait()

	for k, v := range streamCounts {
		sort.Ints(v)
		band95 := len(v) * 94 / 100
		if !logAll {
			log.Info("stream: %s, band95: %d", k, v[band95])
		}

		if logAll && k == "all" {
			log.Info("stream: %s, band95: %d", k, v[band95])
		}
	}

	log.Debug("streamCounts: %v", streamCounts)
	log.Info("Run Time: %v", time.Since(runStart))
}

func stats(st int64) {
	defer wg.Done()
	et := st + int64(5*time.Minute)
	sqlFormat := "select * from %s where time >= %d and time <= %d"
	sql := fmt.Sprintf(sqlFormat, table, st, et)
	log.Debug(sql)
	res, err := influx.Query(sql)
	if err != nil {
		log.Error("influx query err(%v)", err)
		return
	}

	log.Debug(res)

	if len(res) == 0 {
		log.Warn("No data: %s", sql)
		return
	}

	series := res[0].Series
	if len(series) == 0 {
		log.Warn("No values: %v", series)
		return
	}

	values := series[0].Values
	bandCount := 0
	for _, v := range values {
		stream := fmt.Sprintf("%v", v[3])
		bands := fmt.Sprintf("%v", v[1])
		band, err := strconv.Atoi(bands)
		if err != nil {
			log.Error("strconv.Atoi() err(%v)", err)
		}

		if !ListExists(streams, stream) {
			continue
		}

		lock.Lock()
		streamCounts[stream] = append(streamCounts[stream], band)
		lock.Unlock()

		bandCount += band
	}

	if bandCount != 0 {
		log.Info("dataTime: %s, bandCount: %d",
			time.Unix(st/TIMES, 0).Format(TIMEFORMAT), bandCount)

		lock.Lock()
		streamCounts["all"] = append(streamCounts["all"], bandCount)
		lock.Unlock()
	}
}

func openFile(fileName string) (res []string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return res, err
	}

	defer f.Close()

	br := bufio.NewReader(f)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		line := string(a)
		if line != "" {
			res = append(res, line)
		}
	}

	return res, nil
}

// ListExists 验证value是否存在于list中
func ListExists(List []string, value string) bool {
	for _, v := range List {
		if v == value {
			return true
		}
	}

	return false
}
