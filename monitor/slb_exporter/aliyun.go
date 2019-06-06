package main

import (
	"fmt"
	"sync"
	"time"

	log "github.com/alecthomas/log4go"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

type ALI struct {
	slbClient  *slb.Client
	slbList    []slb.LoadBalancer
	slbRequest *slb.DescribeLoadBalancersRequest
	wg         *sync.WaitGroup
}

func NewALI() (a *ALI, err error) {
	a = new(ALI)
	a.slbClient, err = slb.NewClientWithAccessKey(
		AliYunRegionID,
		AliYunAccessKeyID,
		AliYunAccessKeySecret)
	if err != nil {
		return a, err
	}

	a.slbRequest = slb.CreateDescribeLoadBalancersRequest()

	return a, nil
}

func (a *ALI) GetSlb() (err error) {
	log.Info("GetSlb START")

	response, err := a.slbClient.DescribeLoadBalancers(a.slbRequest)
	if err != nil {
		return err
	}

	if !response.IsSuccess() {
		return fmt.Errorf("%s", response.String())
	}

	a.slbList = response.LoadBalancers.LoadBalancer

	a.wg = new(sync.WaitGroup)

	for _, lb := range a.slbList {
		log.Debug("SLB: %v", lb)

		a.wg.Add(1)

		go a.GetSlbHealthStatus(lb.LoadBalancerId, lb.LoadBalancerName)

		time.Sleep(10 * time.Millisecond)
	}

	a.wg.Wait()

	log.Info("GetSlb END")
	return nil
}

func (a *ALI) GetSlbHealthStatus(id, name string) {
	defer a.wg.Done()

	//id = "lb-2zewf78xyvlpec31ptwtt"
	log.Info("SLB ID: %s", id)
	req := slb.CreateDescribeHealthStatusRequest()
	req.LoadBalancerId = id
	resp, err := a.slbClient.DescribeHealthStatus(req)
	if err != nil {
		log.Error("a.slbClient.DescribeHealthStatus(%v)", err)
	}

	if !resp.IsSuccess() {
		log.Error("%s", resp.String())
	}

	for _, v := range resp.BackendServers.BackendServer {
		collector.Add(id, name, v)
	}

	//log.Info(resp.RequestId)
}
