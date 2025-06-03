package main

import (
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/config"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service/component"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/rabbitmq"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/webapi"
)

func main() {
	appController := lifecycle.NewController()
	appCtx, procSpawnFn := appController.Start()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed reading configuration: %w", err))
	}

	srv := webapi.NewServer(appCtx)

	var tlsConfig *rabbitmq.TLSConfig
	if cfg.AppEnv != "local" {
		tlsConfig = &rabbitmq.TLSConfig{
			CAFile:   cfg.RabbitMQCAFile,
			CertFile: cfg.RabbitMQCertFile,
			KeyFile:  cfg.RabbitMQKeyFile,
		}
	}
	rmqClient, err := rabbitmq.NewClient(cfg.RabbitMQConnURL, cfg.RabbitMQQueue, tlsConfig)
	if err != nil {
		panic(fmt.Errorf("failed to connect to RabbitMQ: %w", err))
	}

	producer.Process(procSpawnFn, appCtx, srv, rmqClient)

	rmqConn := component.NewRabbitMQChecker(rmqClient.Connection())
	healthCheckService := service.NewHealthCheckService(rmqConn)
	healthcheck.Process(procSpawnFn, appCtx, srv, healthCheckService)

	webapi.Start(procSpawnFn, srv, cfg.APIPort)

	appController.Wait()
}
