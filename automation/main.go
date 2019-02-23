package main

import (
	"time"

	log "github.com/alecthomas/log4go"
	"github.com/lijiansgit/tools/automation/config"
	"github.com/lijiansgit/tools/automation/handler"
	"github.com/lijiansgit/tools/automation/models"
	domain "github.com/lijiansgit/tools/automation/proto/domain"
	micro "github.com/micro/go-micro"
)

func main() {
	Conf, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	log.LoadConfiguration("./log.xml")
	defer log.Close()

	// init models
	model, err := models.NewModel(Conf)
	if err != nil {
		panic(err)
	}

	// start web
	if Conf.Web.Open {
		web := handler.NewWeb(Conf, model)
		go web.Start()
	}

	// New Service
	service := micro.NewService()

	// Initialise service
	service.Init(
		micro.RegisterTTL(time.Second*time.Duration(Conf.Server.TTL)),
		micro.RegisterInterval(time.Second*time.Duration(Conf.Server.LoopInterval)),
	)

	// Register Handler
	domain.RegisterDomainHandler(service.Server(), handler.NewDomain(Conf, model))

	// Run service
	if err := service.Run(); err != nil {
		panic(err)
	}
}
