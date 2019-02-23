package main

import (
	_ "github.com/micro/go-plugins/broker/kafka"
	_ "github.com/micro/go-plugins/client/grpc"
	_ "github.com/micro/go-plugins/registry/zookeeper"
	_ "github.com/micro/go-plugins/server/grpc"
	_ "github.com/micro/go-plugins/transport/grpc"
)
