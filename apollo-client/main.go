package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ChengjinWu/gojson"
	"github.com/zouyx/agollo"
)

var (
	appID      string
	cluster    string
	ip         string
	namespaces string
	configPath string
)

func init() {
	flag.StringVar(&appID, "a", "ops", "apollo appid")
	flag.StringVar(&cluster, "c", "default", "apollo cluster")
	flag.StringVar(&ip, "i", "http://127.0.0.1:8080", "apollo ip")
	flag.StringVar(&namespaces, "n", "application.properties", "apollo namespaces, 可指定多个，用逗号分割")
	flag.StringVar(&configPath, "conf", "config", "apollo config download path")
}

func main() {
	flag.Parse()

	log.Printf("appID: %s, cluster: %s, ip: %s, namespaces: %s, configPath: %s \n",
		appID, cluster, ip, namespaces, configPath)

	if err := os.RemoveAll(configPath); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		panic(err)
	}

	start()
}

func start() {
	config := func() (*agollo.AppConfig, error) {
		return &agollo.AppConfig{
			AppId:          appID,
			Cluster:        cluster,
			Ip:             ip,
			NamespaceName:  namespaces,
			IsBackupConfig: false,
		}, nil
	}
	agollo.InitCustomConfig(config)
	agollo.SetLogger(&DefaultLogger{})

	err := agollo.Start()
	if err != nil {
		panic(err)
	}

	for _, namespace := range strings.Split(namespaces, ",") {
		cache := agollo.GetConfigCache(namespace)
		cache.Range(func(key, value interface{}) bool {
			content := value.([]byte)
			checkFormat(namespace, content)
			writeFile(fmt.Sprintf("%s", key), namespace, content)

			return true
		})
	}
}

func writeFile(key, namespace string, content []byte) {
	// 判断.properties格式，为key=value存储
	var data string
	ext := path.Ext(namespace)
	if ext == "" || ext == ".properties" {
		data = fmt.Sprintf("%s = %s\n", key, string(content))
	} else {
		data = string(content)
	}

	if ext == "" {
		namespace = fmt.Sprintf("%s.properties", namespace)
	}
	namespaceConfig := fmt.Sprintf("%s/%s", configPath, namespace)
	f, err := os.OpenFile(namespaceConfig, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0744)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	if _, err = f.WriteString(data); err != nil {
		panic(err)
	}
}

func checkFormat(namespace string, content []byte) {
	var err error

	switch path.Ext(namespace) {
	case ".json":
		err = gojson.CheckValid(content)
	case ".yml", ".yaml":
		res := make(map[string]interface{})
		err = yaml.Unmarshal(content, res)
	default:
		err = nil
	}

	if err != nil {
		log.Printf("%s content: \n%s \n", namespace, string(content))
		log.Printf("%s format is err: %v \n", namespace, err)
		os.Exit(11)
	}
}

type DefaultLogger struct {
}

func (this *DefaultLogger) Debugf(format string, params ...interface{}) {
	this.Debug(format, params)
}

func (this *DefaultLogger) Infof(format string, params ...interface{}) {
	this.Debug(format, params)
}

func (this *DefaultLogger) Warnf(format string, params ...interface{}) error {
	this.Debug(format, params)
	return nil
}

func (this *DefaultLogger) Errorf(format string, params ...interface{}) error {
	this.Debug(format, params)
	return nil
}

func (this *DefaultLogger) Debug(v ...interface{}) {
	log.Printf("agollo %v", v)
}
func (this *DefaultLogger) Info(v ...interface{}) {
	this.Debug(v)
}

func (this *DefaultLogger) Warn(v ...interface{}) error {
	this.Debug(v)
	return nil
}

func (this *DefaultLogger) Error(v ...interface{}) error {
	this.Debug(v)
	return nil
}
