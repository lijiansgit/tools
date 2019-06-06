package main

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

const (
	defaultPageSize = 30
)

type ALI struct {
	slbList    []slb.LoadBalancer
	slbClient  *slb.Client
	slbIds     []string
	slbRequest *slb.DescribeLoadBalancersRequest
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

func (a *ALI) CheckEcsInSlb(ipList []string) (err error) {
	var n int
	for _, ip := range ipList {
		a.slbRequest.ServerIntranetAddress = ip
		resp, err := a.slbClient.DescribeLoadBalancers(a.slbRequest)
		if err != nil {
			return err
		}

		if !resp.IsSuccess() {
			return fmt.Errorf("%s", resp.String())
		}

		if len(resp.LoadBalancers.LoadBalancer) != 0 {
			fmt.Printf(ipFormat, ip)
			n = n + 1
		}

		for _, slb := range resp.LoadBalancers.LoadBalancer {
			fmt.Printf("ID: %s, Name: %s \n", slb.LoadBalancerId, slb.LoadBalancerName)
		}
	}

	if n == 0 {
		fmt.Println(noEcs)
	}

	return nil
}
