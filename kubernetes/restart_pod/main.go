package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/lijiansgit/go/libs"
)

var (
	nameSpace string
	pods      []string
	pod       string
	timeStr   string
	wg        *sync.WaitGroup
)

func init() {
	flag.StringVar(&nameSpace, "n", "test", "命名空间名称，默认为test")
	flag.StringVar(&timeStr, "t", "1s", "多长时间启动一个重启线程，默认为1秒，eg: 10ms, 10s")
}

func main() {
	flag.Parse()

	timeDur, err := time.ParseDuration(timeStr)
	if err != nil {
		panic(err)
	}

	startTime := time.Now()
	log.Println("namesapce: ", nameSpace)
	getPodsCmd := fmt.Sprintf("kubectl get pods -n %s -o=name", nameSpace)
	res, err := libs.Cmd(getPodsCmd)
	if err != nil {
		log.Panicln(err)
	}

	pods = strings.Split(res, "\n")
	if len(pods) <= 1 {
		fmt.Println("命名空间不存在：", nameSpace)
		return
	}

	wg = new(sync.WaitGroup)

	for _, pod = range pods {
		if pod == "" {
			continue
		}

		wg.Add(1)
		go restartPod(pod, nameSpace)

		time.Sleep(timeDur)
	}

	wg.Wait()

	log.Println("Run Time: ", time.Since(startTime))
}

func restartPod(pod, nameSpace string) {
	defer wg.Done()

	log.Printf("restartPod: %s START", pod)
	restartPodCmd := fmt.Sprintf("kubectl get %s -n %s -o=yaml |kubectl replace --force -f -", pod, nameSpace)
	_, err := libs.Cmd(restartPodCmd)
	if err != nil {
		log.Printf("restartPod: %s ERR", pod)
		return
	}

	log.Printf("restartPod: %s OK", pod)
}
