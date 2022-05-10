package main

import (
	"flag"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"log"
)

func main() {
	log.Println("Starting microservice")

	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(&cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("EventSourcing")
	appLogger.Infof("CFG: %+v", cfg)
}
