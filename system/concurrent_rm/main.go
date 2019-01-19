package main

import (
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dir     string
	timeStr string
	// ExecCount 执行总计
	ExecCount int64
	// ErrCount 执行失败总计
	ErrCount int64
	wg       *sync.WaitGroup
)

func init() {
	flag.StringVar(&dir, "d", "", "rm dir name")
	flag.StringVar(&timeStr, "t", "1s", "每隔多长时间启动一个删除协程，默认为1秒，eg: 1us, 1ms")
}

func main() {
	flag.Parse()

	timeDur, err := time.ParseDuration(timeStr)
	if err != nil {
		panic(err)
	}

	if dir == "" {
		println("Usage: -h/--help")
		return
	}

	startTime := time.Now()
	go result()

	files, err := openDir(dir)
	if err != nil {
		panic(err)
	}

	wg = new(sync.WaitGroup)
	for _, file := range files {
		wg.Add(1)

		go removeFile(file)

		time.Sleep(timeDur)
	}

	wg.Wait()

	err = os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}

	log.Printf("ExecCount:%d ErrCount:%d \n", ExecCount, ErrCount)
	log.Println("Run Time: ", time.Since(startTime))
}

func removeFile(file string) {
	defer wg.Done()

	err := os.Remove(file)
	if err != nil {
		log.Printf("Exec err(%v) \n", err)
		atomic.AddInt64(&ErrCount, 1)
	}

	atomic.AddInt64(&ExecCount, 1)
}

func openDir(dir string) (res []string, err error) {
	err = filepath.Walk(dir, func(src string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		rel, err := filepath.Rel(dir, src)
		if err != nil {
			return err
		}

		file := path.Join(dir, rel)

		if !f.IsDir() {
			res = append(res, file)
		}
		return nil
	})

	return res, nil
}

func result() {
	var (
		lastTimes int64
		diff      int64
		nowCount  int64
		errCount  int64
		timer     = int64(1)
	)

	for {
		nowCount = atomic.LoadInt64(&ExecCount)
		diff = nowCount - lastTimes
		lastTimes = nowCount
		errCount = atomic.LoadInt64(&ErrCount)
		log.Printf("ExecCount:%d ErrCount:%d exec/s:%d\n", nowCount, errCount, diff/timer)

		time.Sleep(time.Duration(timer) * time.Second)
	}
}
