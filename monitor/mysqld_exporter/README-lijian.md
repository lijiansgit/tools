## 支持同时连接N个MySQL
## MySQL 列表从阿里云接口拉取
## 自动生成Prometheus监控主配置文件
## 监控Target自动生产
## 密码等隐私信息全部在password.go 中，默认都为空，编译之前请手动填写

## depend:
    go get -u -v github.com/prometheus/client_golang/prometheus
    go get -u -v github.com/go-sql-driver/mysql
    go get -u -v github.com/prometheus
    go get -u -v gopkg.in/alecthomas/kingpin.v2
    go get -u -v gopkg.in/ini.v1

## build:
    go build