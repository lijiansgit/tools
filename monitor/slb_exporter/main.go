package main

import (
	"flag"
	"net/http"
	"time"

	log "github.com/alecthomas/log4go"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	collector  = NewGoCollector()
	ali        *ALI
	err        error
	debug      int
	listenAddr string
	timeStr    string
)

func init() {
	prometheus.MustRegister(collector)
	flag.IntVar(&debug, "debug", 1, "log level info:1, debug:0")
	flag.StringVar(&listenAddr, "addr", "127.0.0.1:9100", "http listen addr")
	flag.StringVar(&timeStr, "t", "2m", "每隔多久获取一次阿里云数据，eg: 2m, 5m")
}

func main() {
	flag.Parse()

	if debug == 1 {
		log.AddFilter("stdout", log.INFO, log.NewConsoleLogWriter())
	} else {
		log.AddFilter("stdout", log.DEBUG, log.NewConsoleLogWriter())
	}

	defer log.Close()

	log.Info("log level: %d", debug)

	ali, err = NewALI()
	if err != nil {
		panic(err)
	}

	timeDur, err := time.ParseDuration(timeStr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			collector.slbServerStatusList = make([]*SlbServerStatus, 0)
			if err = ali.GetSlb(); err != nil {
				log.Error("ali.GetSlb(%v)", err)
			}

			time.Sleep(timeDur)
		}
	}()

	http.Handle("/metrics", prometheus.Handler())
	log.Info("http start, listen: %s", listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		panic(err)
	}
}
