package main

import (
	"flag"
	"fmt"
	"strings"
)

var (
	ips        string
	consulAddr string
	ali        *ALI
	err        error
)

const (
	stepOne   = "Step1:"
	stepTwo   = "Step2:"
	stepThree = "Step3:"
	stepFour  = "Step4:"
	line      = "-------------------------------"
	noEcs     = "无"
)

func init() {
	flag.StringVar(&ips, "ip", "172.28.245.140", "172.28.245.140,172.28.245.141")
	flag.StringVar(&consulAddr, "cAddr", "http://consul.qtt6.cn,http://op-consul.qutoutiao.net",
		"http://consul.qtt6.cn")
}

func main() {
	flag.Parse()

	fmt.Printf("IP: "+ipFormat, ips)
	ipList := strings.Split(ips, ",")

	fmt.Println(line)
	fmt.Println(stepOne)
	fmt.Println("检查SLB:")
	ali, err = NewALI()
	if err != nil {
		panic(err)
	}

	if err = ali.CheckEcsInSlb(ipList); err != nil {
		panic(err)
	}

	fmt.Println(line)
	fmt.Println(stepTwo)
	fmt.Println("检查VTM:")
	if err = CheckVTM(ipList, vtm1Name); err != nil {
		panic(err)
	}

	if err = CheckVTM(ipList, vtm2Name); err != nil {
		panic(err)
	}

	fmt.Println(line)
	fmt.Println(stepThree)
	fmt.Println("检查Jenkins:")
	if err = CheckJenkins(ipList); err != nil {
		panic(err)
	}

	fmt.Println(line)
	fmt.Println(stepFour)
	fmt.Println("检查Consul:")
	consulList := strings.Split(consulAddr, ",")
	for _, addr := range consulList {
		fmt.Println("addr:", addr)
		if err = CheckConsul(addr, ipList); err != nil {
			panic(err)
		}
	}
}
