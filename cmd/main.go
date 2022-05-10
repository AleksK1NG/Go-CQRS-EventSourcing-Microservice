package main

import (
	"flag"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/server"
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

	appLogger := logger.NewAppLogger(cfg.Logger.LogLevel, cfg.Logger.DevMode, cfg.Logger.Encoder)
	appLogger.InitLogger()
	appLogger.Named(server.GetMicroserviceName(*cfg))
	appLogger.Infof("CFG: %+v", cfg)
	appLogger.Fatal(server.NewServer(appLogger, *cfg).Run())
}
