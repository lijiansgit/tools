package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/lijiansgit/go/libs"

	"github.com/lijiansgit/go/libs/nets"
)

const (
	vtm1Name        = "vtm1"
	vtm2Name        = "vtm2"
	confPath        = "/data/git.sb.net/xuyi/"
	vtm1confPath    = confPath + "vtm1-conf"
	vtm2confPath    = confPath + "vtm2-conf"
	vtm1LoginURL    = "https://vtm1.sb.net:9090"
	vtm2LoginURL    = "https://vtm2.sb.net:9090"
	vtmLoginFormat  = "<font color='red'>VTM LoginURL:</font> %s \n"
	ipFormat        = "<font color='red'>%s</font> \n"
	jenkinsHost     = "172.16.70.177"
	jenkinsSshUser  = "work"
	jenkinsSshPort  = "22"
	jenkinsSshKey   = "/root/.ssh/id_rsa"
	jenkinsJobsPath = "/data/jenkins/jobs/"
	jenkinsConfXml  = "config.xml"
)

func CheckVTM(ipList []string, vtm string) (err error) {
	var (
		n           int
		vtmConfPath string
	)
	if vtm == vtm1Name {
		fmt.Println(vtm1Name)
		vtmConfPath = vtm1confPath
	}

	if vtm == vtm2Name {
		fmt.Println(vtm2Name)
		vtmConfPath = vtm2confPath
	}

	for _, ip := range ipList {
		cmd := fmt.Sprintf("grep -rl %s %s", ip, vtmConfPath)
		res, err := libs.Cmd(cmd, confPath)
		if err != nil {
			// 未查找到
			if strings.Contains(err.Error(), "exit") {
				continue
			}

			return err
		} else {
			n = n + 1
		}

		res = strings.Replace(res, confPath, "", -1)
		if n != 0 {
			if vtm == vtm1Name {
				fmt.Printf(ipFormat, ip)
				fmt.Printf(vtmLoginFormat, vtm1LoginURL)
				fmt.Println(res)
			}
			if vtm == vtm2Name {
				fmt.Printf(ipFormat, ip)
				fmt.Printf(vtmLoginFormat, vtm2LoginURL)
				fmt.Println(res)
			}
		}
	}

	if n == 0 {
		fmt.Println(noEcs)
	}

	return nil
}

func CheckJenkins(ipList []string) (err error) {
	var n int
	ssh := nets.NewSSH()
	ssh.SetPrivateKey(jenkinsSshKey, jenkinsSshUser)
	ssh.SetHostPort(jenkinsHost, jenkinsSshPort)
	ssh.SetTimeout(2 * time.Second)
	for _, ip := range ipList {
		cmd := fmt.Sprintf("find %s -name %s -print0 |xargs -0 grep -rl %s",
			jenkinsJobsPath, jenkinsConfXml, ip)
		res, err := ssh.Cmd(cmd)
		if err != nil {
			// 未查找到
			if strings.Contains(err.Error(), "exit") {
				continue
			}

			return err
		} else {
			n = n + 1
		}

		res = strings.Replace(res, jenkinsConfXml, "", -1)
		if n != 0 {
			fmt.Printf(ipFormat, ip)
			fmt.Println(strings.Replace(res, jenkinsJobsPath, "", -1))
		}
	}

	if n == 0 {
		fmt.Println(noEcs)
	}

	return nil
}

func CheckConsul(addr string, ipList []string) (err error) {
	var (
		n         int
		ipService map[string]string
	)
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	res, err := client.Agent().Services()
	if err != nil {
		return err
	}

	ipService = make(map[string]string)
	for _, ip := range ipList {
		for _, svc := range res {
			if ip == svc.Address {
				n = n + 1
				ipService[ip] = ipService[ip] +
					fmt.Sprintf("Svc: %s, Tags: %v\n", svc.Service, svc.Tags)
			}
		}
	}

	if n == 0 {
		fmt.Println(noEcs)
		return nil
	}

	for ip, svc := range ipService {
		fmt.Printf(ipFormat, ip)
		fmt.Println(svc)
	}

	return nil
}
