package config

import (
	"flag"
	"fmt"
	"os"
	"path"

	log "github.com/alecthomas/log4go"

	"github.com/hashicorp/consul/api"
	yaml "gopkg.in/yaml.v2"
)

const (
	// ConsulPath consul路径
	ConsulPath = "service/ops/golang/automation"
)

var (
	// consulHTTPAddr consul地址
	consulHTTPAddr string
)

func init() {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	flag.StringVar(&consulHTTPAddr, "consul", addr, "Consul HTTP addr")
}

// Config 配置
type Config struct {
	Server    `yaml:"server"`
	Web       `yaml:"web"`
	Cloudwise `yaml:"cloudwise"`
	DNSPod    `yaml:"dnspod"`
	MySQL     `yaml:"mysql"`
	Domain
}

// Server 服务参数配置
type Server struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	LoopInterval int    `yaml:"loopinterval"`
	TTL          int    `yaml:"ttl"`
}

// Web 相关
type Web struct {
	Open         bool   `yaml:"open"`
	Addr         string `yaml:"addr"`
	StatusURL    string `yaml:"statusURL"`
	CloudwiseURL string `yaml:"cloudwiseURL"`
	DNSPodURL    string `yaml:"dnspodURL"`
}

// Cloudwise 监控宝相关
type Cloudwise struct {
	Token string `yaml:"token"`
}

// DNSPod dnspod
type DNSPod struct {
	Token         string `yaml:"token"`
	CallBackToken string `yaml:"callback_token"`
}

// MySQL mysql
type MySQL struct {
	Addr   string `yaml:"addr"`
	DB     string `yaml:"db"`
	User   string `yaml:"user"`
	Passwd string `yaml:"passwd"`
}

// Domain 域名
type Domain struct {
	A
	CNAME
}

// A 记录，域名重要等级：1，2，3
type A struct {
	Line   map[string][]string
	Record map[string][]string
}

// CNAME Config中CNAME记录
type CNAME struct {
	ALI, WS, TX map[string][]string
}

// NewConfig 新的配置
func NewConfig() *Config {
	return &Config{}
}

func InitConfig() (conf *Config, err error) {
	conf = NewConfig()

	log.Info("Consul HTTP Addr: %s", consulHTTPAddr)
	consul, err := NewConsul(consulHTTPAddr)
	if err != nil {
		return conf, err
	}

	serverConfig, err := consul.Read("config")
	if err != nil {
		return conf, err
	}

	if err = yaml.Unmarshal(serverConfig, &conf); err != nil {
		return conf, err
	}

	conf.Domain.A.Line, err = consul.ReadMapList("domain/A/line")
	if err != nil {
		return conf, err
	}

	conf.Domain.A.Record, err = consul.ReadMapList("domain/A/record")
	if err != nil {
		return conf, err
	}

	log.Info("Conf: %v", conf)

	return conf, nil
}

// Consul 结构
type Consul struct {
	Clt *api.Client
}

// NewConsul 新Consul
func NewConsul(addr string) (consul *Consul, err error) {
	clt, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		return consul, err
	}

	return &Consul{Clt: clt}, nil
}

// Read 读取consul
func (c *Consul) Read(key string) (b []byte, err error) {
	kv := c.Clt.KV()
	paths := path.Join(ConsulPath, key)
	pair, _, err := kv.Get(paths, nil)
	if err != nil {
		return b, err
	}

	if pair == nil {
		return b, fmt.Errorf("key(%s) is not exist", key)
	}

	return pair.Value, nil
}

// ReadMapList 读取mapList
func (c *Consul) ReadMapList(key string) (m map[string][]string, err error) {
	b, err := c.Read(key)
	if err != nil {
		return m, err
	}

	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return m, err
	}

	return m, nil
}
