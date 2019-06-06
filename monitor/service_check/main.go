package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	//proxyAddr string
	webURL string
	ali    *ALI
	err    error
)

const (
	stepOne   = "Step1:"
	stepTwo   = "Step2:"
	stepThree = "Step3:"
	stepFour  = "Step4:"
	line      = "-------------------------------"
)

func init() {
	//flag.StringVar(&proxyAddr, "x", "no", "proxy address: ip:port")
	flag.StringVar(&webURL, "w", "http://test-soft.1sapp.com/ping", "host: http://test.com/ping")
}

func main() {
	flag.Parse()

	ali, err = NewALI()
	if err != nil {
		panic(err)
	}

	if err = ali.GetSlb(); err != nil {
		panic(err)
	}

	u, err := url.Parse(webURL)
	if err != nil {
		panic(err)
	}

	ips, err := net.LookupHost(u.Host)
	if err != nil {
		panic(err)
	}

	fmt.Println(stepOne)
	fmt.Printf("域名: %s \n\n", u.Host)

	fmt.Println(stepTwo)
	fmt.Printf("解析IP地址: %v \n\n", ips)

	fmt.Println(stepThree)
	for _, ip := range ips {
		fmt.Printf("负载均衡SLB: %s \n", ip)
		ipp := fmt.Sprintf("%s:%s", ip, slbListenPortDefault)
		if err := Check(ipp); err != nil {
			panic(err)
		}

		slbBackendCheck(ip)
		fmt.Println(line)

		printVtmConf(u.Host, ip)
	}
}

func slbBackendCheck(ip string) {
	err = ali.GetSlbBackend(ip)
	if err != nil {
		panic(err)
	}

	for _, ipp := range ali.slbBackendServerIpps {
		fmt.Printf("Server: %s \n", ipp)
		if err := Check(ipp); err != nil {
			fmt.Println(err)
		}
	}
}

func Check(ipp string) (err error) {
	proxyURL := fmt.Sprintf("http://%s", ipp)
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		Proxy: http.ProxyURL(proxy),
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 2,
	}

	resp, err := httpClient.Get(webURL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("GET Status: %s, Content: %s \n", resp.Status, string(body))
	fmt.Println("")
	return nil
}
