package main

import (
	"strconv"
	"sync"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"

	"github.com/prometheus/client_golang/prometheus"
)

type goCollector struct {
	requestsDesc        *prometheus.Desc
	slbServerStatusList []*SlbServerStatus
	lock                sync.Mutex
}

// NewGoCollector returns a collector which exports metrics about the current
// go process.
func NewGoCollector() *goCollector {
	return &goCollector{
		slbServerStatusList: make([]*SlbServerStatus, 0),
		requestsDesc: prometheus.NewDesc(
			"slb_backend_server_health_status",
			"https://help.aliyun.com/document_detail/27635.html?spm=a2c4g.11186623.6.661.38195a8ayapTRQ",
			[]string{"slb_id", "slb_name", "listener_port", "server_id", "server_port", "server_health_status"},
			nil,
		),
	}
}

// Describe returns all descriptions of the collector.
func (c *goCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.requestsDesc
}

// Collect returns the current state of all metrics of the collector.
func (c *goCollector) Collect(ch chan<- prometheus.Metric) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.slbServerStatusList) == 0 {
		return
	}

	for _, v := range c.slbServerStatusList {
		ch <- prometheus.MustNewConstMetric(c.requestsDesc,
			prometheus.CounterValue, 1, v.ID, v.Name, strconv.Itoa(v.ListenerPort),
			v.ServerID, strconv.Itoa(v.ServerPort), v.ServerHealthStatus)
	}

	//c.slbServerStatusList = make([]*SlbServerStatus, 0)
}

func (c *goCollector) Add(Id, Name string, server slb.BackendServer) {
	c.lock.Lock()
	defer c.lock.Unlock()

	s := &SlbServerStatus{
		ID:                 Id,
		Name:               Name,
		ListenerPort:       server.ListenerPort,
		ServerID:           server.ServerId,
		ServerPort:         server.Port,
		ServerHealthStatus: server.ServerHealthStatus,
	}

	c.slbServerStatusList = append(c.slbServerStatusList, s)
}

type SlbServerStatus struct {
	ID                 string
	Name               string
	ListenerPort       int
	ServerID           string
	ServerPort         int
	ServerHealthStatus string
}
