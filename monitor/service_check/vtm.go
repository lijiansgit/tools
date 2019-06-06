package main

import (
	"fmt"

	"github.com/lijiansgit/go/libs"
)

const (
	confPath      = "/data/git.qutoutiao.net/xuyi/"
	vtm2RulesPath = confPath + "vtm2-conf/rules/"
	vtm2PoolsPath = confPath + "vtm2-conf/pools/"
	vtm2Address   = "59.110.123.172"
	vtm2LoginURL  = "https://vtm2.qutoutiao.net:9090"
	vtm1PoolsPath = confPath + "vtm1-conf/pools/"
	vtm1RulesPath = confPath + "vtm1-conf/rules/"
	vtm1Address   = "59.110.123.172"
	//vtm1->高防cname->高防ip->slb->vtm1->server
)

type VTM struct {
	host      string
	rulesPath string
	poolsPath string
}

func NewVTM(host, address string) *VTM {
	v := new(VTM)
	v.host = host
	if address == vtm2Address {
		v.rulesPath = vtm2RulesPath
		v.poolsPath = vtm2PoolsPath
		//} else {
		//	v.rulesPath = vtm1RulesPath
		//	v.poolsPath = vtm1PoolsPath
	}

	return v
}

func (v *VTM) printRule() {
	cmd := fmt.Sprintf("grep -rl %s %s |xargs cat", v.host, v.rulesPath)
	res, err := libs.Cmd(cmd)
	if err != nil {
		panic(err)
	}

	fmt.Printf("<font color='red'>Rule name:</font> %s \n", v.host)
	fmt.Println(res)
}

func (v *VTM) printPool() {
	cmd := fmt.Sprintf("cat %s%s", v.poolsPath, v.host)
	res, err := libs.Cmd(cmd)
	if err != nil {
		panic(err)
	}

	fmt.Printf("<font color='red'>Pool name:</font> %s \n", v.host)
	fmt.Println(res)
}

func printVtmConf(host, address string) {
	if address != vtm2Address {
		return
	}

	fmt.Println(stepFour)
	fmt.Printf("<font color='red'>VTM LoginURL:</font> %s \n", vtm2LoginURL)
	v := NewVTM(host, address)
	v.printRule()
	v.printPool()
}
