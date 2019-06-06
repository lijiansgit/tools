package main

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/prometheus/common/log"
	"golang.org/x/exp/errors/fmt"
)

func GetRDS(page int) (total int, RDSList []rds.DBInstance, err error) {
	// Create an ECS client
	rdsClient, err := rds.NewClientWithAccessKey(
		AliYunRegionID,             // Your Region ID
		AliYunAccessKeyID,         // Your AccessKey ID
		AliYunAccessKeySecret)     // Your AccessKey Secret
	if err != nil {
		// Handle exceptions
		return total, RDSList, err
	}
	// Create an API request and set parameters
	request := rds.CreateDescribeDBInstancesRequest()
	// Set the request.PageSize to "10"
	request.PageSize = requests.NewInteger(30)
	request.PageNumber = requests.NewInteger(page)
	// Initiate the request and handle exceptions
	response, err := rdsClient.DescribeDBInstances(request)
	if err != nil {
		return total, RDSList, err
	}

	if ! response.IsSuccess() {
		return total, RDSList, fmt.Errorf("%s", response.String())
	}

	total = response.TotalRecordCount
	RDSList = response.Items.DBInstance
	return  total, RDSList, nil
}

func GetRDSList() (l []string) {
	total, _, err := GetRDS(1)
	if err != nil {
		log.Errorf("GETRDS err(%v)", err)
	}

	totalPage := total / 30 + 1
	for page := 1;page <= totalPage; page++ {
		_, RDSList, err := GetRDS(page)
		if err != nil {
			log.Errorf("GETRDS err(%v)", err)
		}

		for _,v := range RDSList {
			//name := fmt.Sprintf("%s:%s", v.ZoneId, v.DBInstanceId)
			l = append(l, v.DBInstanceId)
		}
	}

	return l
}
