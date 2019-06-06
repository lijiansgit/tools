package main

import (
	"github.com/prometheus/common/log"
	"golang.org/x/exp/errors/fmt"
	"strings"
	"os"
)

const (
	//CONFNAME="/data/app/prometheus/prometheus.yml"
)

func CreateFile() {
	f, err := os.Create(*prometheusConfPath)
	if err != nil {
		log.Errorf("os.Create err(%v)", err)
	}

	defer f.Close()

	var str string
	str = "global:\n  scrape_interval: 10s\n  evaluation_interval: 10s\n  external_labels:\n"
	str = str + "    monitor: 'qtt-thanos-prometheus-db'\n"
	str = str + "alerting:\n  alertmanagers:\n  - static_configs:\n"
	str = str + "    - targets: ['monitor-gecailong.qtt.com:9093']\n"
	str = str + "rule_files: ['rules/*.rules']\n"
	str = str + "scrape_configs:\n  - job_name: 'prometheus'\n"
	str = str + "    static_configs:\n"
	str = str + fmt.Sprintf("    - targets: ['%s:9090']\n", strings.Split(*listenAddress, ":")[0])

	_ , err = f.WriteString(str)
	if err != nil {
		log.Errorf("WriteFile err(%v)", err)
	}
}

func WriteFile(name string) {
	f, err := os.OpenFile(*prometheusConfPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Errorf("os.OpenFile err(%v)", err)
	}

	defer f.Close()

	var str string
	str = str + fmt.Sprintf("  - job_name: '%s'\n", name)
	str = str + fmt.Sprintf("    metrics_path: '/metrics/%s'\n", name)
	str = str + fmt.Sprintf("    static_configs:\n")
	str = str + fmt.Sprintf("    - targets: ['%s']\n", *listenAddress)
	_ , err = f.WriteString(str)
	if err != nil {
		log.Errorf("WriteFile err(%v)", err)
	}
}
