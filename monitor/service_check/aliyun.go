package main

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

const (
	slbListenPortDefault = "80"
)

type ALI struct {
	slbList              []slb.LoadBalancer
	slbClient            *slb.Client
	slbId                string
	slbBackendServerPort int
	slbBackendServerIpps []string
	slbRequest           *slb.DescribeLoadBalancersRequest
	ecsClient            *ecs.Client
	ecsInstanceId        string
}

func NewALI() (a *ALI, err error) {
	a = new(ALI)
	a.slbClient, err = slb.NewClientWithAccessKey(
		AliYunRegionID,        // Your Region ID
		AliYunAccessKeyID,     // Your AccessKey ID
		AliYunAccessKeySecret) // Your AccessKey Secret
	if err != nil {
		return a, err
	}

	a.slbRequest = slb.CreateDescribeLoadBalancersRequest()

	a.ecsClient, err = ecs.NewClientWithAccessKey(
		AliYunRegionID,
		AliYunAccessKeyID,
		AliYunAccessKeySecret)
	if err != nil {
		return a, err
	}

	return a, nil
}

func (a *ALI) GetSlb() (err error) {
	response, err := a.slbClient.DescribeLoadBalancers(a.slbRequest)
	if err != nil {
		return err
	}

	if !response.IsSuccess() {
		return fmt.Errorf("%s", response.String())
	}

	a.slbList = response.LoadBalancers.LoadBalancer
	//fmt.Println("req id: ", response.RequestId)
	return nil
}

func (a *ALI) GetSlbBackend(ip string) (err error) {
	for _, slb := range a.slbList {
		if ip == slb.Address {
			a.slbId = slb.LoadBalancerId
		}
	}

	req := slb.CreateDescribeLoadBalancerTCPListenerAttributeRequest()
	req.RegionId = AliYunRegionID
	req.LoadBalancerId = a.slbId
	req.ListenerPort = slbListenPortDefault
	resp, err := a.slbClient.DescribeLoadBalancerTCPListenerAttribute(req)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("%s", resp.String())
	}

	a.slbBackendServerPort = resp.BackendServerPort

	request := slb.CreateDescribeLoadBalancerAttributeRequest()
	request.LoadBalancerId = a.slbId
	response, err := a.slbClient.DescribeLoadBalancerAttribute(request)
	if err != nil {
		return err
	}

	if !response.IsSuccess() {
		return fmt.Errorf("%s", response.String())
	}

	for _, b := range response.BackendServers.BackendServer {
		a.ecsInstanceId = b.ServerId
		if err = a.GetBackendServers(); err != nil {
			return err
		}
	}

	return nil
}

func (a *ALI) GetBackendServers() (err error) {
	request := ecs.CreateDescribeInstanceAttributeRequest()
	request.RegionId = AliYunRegionID
	request.InstanceId = a.ecsInstanceId
	response, err := a.ecsClient.DescribeInstanceAttribute(request)
	if err != nil {
		return err
	}

	if !response.IsSuccess() {
		return fmt.Errorf("%s", response.String())
	}

	for _, ip := range response.VpcAttributes.PrivateIpAddress.IpAddress {
		ipp := fmt.Sprintf("%s:%d", ip, a.slbBackendServerPort)
		a.slbBackendServerIpps = append(a.slbBackendServerIpps, ipp)
	}

	return nil
}
