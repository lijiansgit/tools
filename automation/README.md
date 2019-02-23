# Automation Service

This is the Automation service


## Getting Started

- [Configuration](#configuration)
- [Dependencies](#dependencies)
- [Usage](#usage)
- [功能说明](#功能说明)

## Configuration

- FQDN: go.micro.srv.automation
- Type: srv
- Alias: automation
- consul: SERVICE/OPS/GOLANG/AUTOMATION

## Dependencies

Micro services depend on service discovery. The default is consul.

```
# install consul
brew install consul

# run consul
consul agent -dev
```

## Usage

A Makefile is included for convenience

Build the binary

```
make build
```

Run the service
```
./automation
```

Build a docker image
```
make build-image
```

## 功能说明

### 根据dnspod和监控宝报警回调通知实现自动暂停和开启dns解析的微服务

回调

- 收到dnspod报警回调，监控项目信息写入MySQL->dnspod_data。status字段：Down为故障，Ok为恢复
- 收到监控宝报警回调，监控项目信息写入MySQL->cloudwise_data。msg_status字段：1为故障，2为恢复

自动暂停解析的前提条件和步骤
1. 收到dnspod故障通知
2. 启动暂停程序，暂停程序会循环10次检查是否符合以下要求，每次休眠一分钟，直到成功或者失败
*  dnspod 线路是否和consul配置中的一致
*  查询MySQL，status 和 msg_status 字段是否同时为故障状态
*  dnspod线路开启负载均衡数量在2条以上
*  如果线路已经是暂停状态，退出循环
3. 暂停成功后会把当前时间添加到dnspod解析的备注中，并把操作记录写入到MySQL->domain_op

自动开启解析的前提条件和步骤
1. 收到监控宝报警恢复通知
2. 自动开启程序，开启程序会循环10次检查是否符合以下要求，每次休眠一分钟，直到成功或者失败
*  根据报警信息中的子域名和域名从MySQL->domain_op查询到最新的故障的记录信息，得到要开启的记录值
*  dnspod 线路是否和consul配置中的一致
*  查询MySQL，msg_status 字段是否为恢复状态
*  如果线路已经是开启状态，退出循环
3. 开启成功后会把当前时间添加到dnspod解析的备注中，并把操作记录写入到MySQL->domain_op