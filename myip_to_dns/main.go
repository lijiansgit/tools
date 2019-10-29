package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/robfig/cron"
)

const (
	SOHU_URL = "http://pv.sohu.com/cityjson"
	NET_URL  = "http://www.net.cn/static/customercare/yourip.asp"
	RegionID = "cn-hangzhou"
)

var (
	myRecord        *alidns.Record
	accessKeyId     string
	accessKeySecret string
	domainName      string
	recordName      string
	cronFormat      string
)

func init() {
	myRecord = &alidns.Record{}
	flag.StringVar(&accessKeyId, "ak", "", "aliyun accessKeyId")
	flag.StringVar(&accessKeySecret, "aks", "", "aliyun accessKeySecret")
	flag.StringVar(&domainName, "dn", "lijian.site", "domain name")
	flag.StringVar(&recordName, "rn", "home", "record name")
	flag.StringVar(&cronFormat, "cron", "*/5 * * * * *", "cron format")
}

func main() {
	flag.Parse()

	c := cron.New()
	c.AddFunc(cronFormat, run)
	log.Println("start cron:", cronFormat)
	c.Start()

	select {}
}

func run() {
	log.Println("run cron")
	ip, err := compareIP()
	if err != nil {
		log.Println("ERROR:", err)
	}

	if ip == "" {
		log.Println("WARN: ip is null!")
		return
	}

	if err := aliyunUpdateDomainRecord(ip); err != nil {
		log.Println("ERROR:", err)
	}
}

func compareIP() (ip string, err error) {
	sIP, err := getMyIP(SOHU_URL)
	if err != nil {
		return ip, err
	}

	nIP, err := getMyIP(NET_URL)
	if err != nil {
		return ip, err
	}

	// sIP, nIP = "127.0.0.1", "127.0.0.1"
	if sIP != nIP {
		return ip, fmt.Errorf("compare ip err: %s/%s", sIP, nIP)
	}

	return sIP, nil
}

func getMyIP(ipURL string) (ip string, err error) {
	// fmt.Println("111111111")
	log.Println("Get URL:", ipURL)
	resp, err := http.Get(ipURL)
	if err != nil {
		return ip, err
	}
	// fmt.Println("222222222")

	defer resp.Body.Close()

	bodyByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ip, err
	}

	reg, err := regexp.Compile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	if err != nil {
		return ip, err
	}

	ip = reg.FindString(string(bodyByte))
	log.Println(fmt.Sprintf("'%s' return myip: %s", ipURL, ip))

	return ip, nil
}

func aliyunUpdateDomainRecord(ip string) (err error) {
	client, err := alidns.NewClientWithAccessKey(RegionID, accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}

	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domainName
	request.KeyWord = recordName
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		return err
	}

	for _, record := range response.DomainRecords.Record {
		if record.RR == recordName {
			myRecord.RecordId = record.RecordId
			myRecord.RR = recordName
			myRecord.Value = record.Value
			myRecord.Type = record.Type
		}
	}

	if myRecord.Value == "" {
		return fmt.Errorf("record: %s is not found", recordName)
	}

	if myRecord.Value == ip {
		log.Println("record is exist, not need update:", ip)
		return nil
	}

	req := alidns.CreateUpdateDomainRecordRequest()
	req.Scheme = "https"
	req.RecordId, req.Value = myRecord.RecordId, ip
	req.RR, req.Type = myRecord.RR, myRecord.Type
	if _, err = client.UpdateDomainRecord(req); err != nil {
		return err
	}

	log.Println(fmt.Sprintf("%s.%s value: %s update ok!",
		recordName, domainName, ip))
	return nil
}
