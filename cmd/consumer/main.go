package main

import (
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/config"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service/component"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg"
	consumer "github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer"
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

	dbCfg := pg.PoolConfig{
		MinConns:          cfg.DBMinConns,
		MaxConns:          cfg.DBMaxConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheck,
	}
	pool, err := pg.InitPool(appCtx, cfg.DBConnectionURL, dbCfg)
	if err != nil {
		panic(fmt.Errorf("failed initializing DB pool: %w", err))
	}
	defer pool.Close()

	consumer.Process(procSpawnFn, appCtx, srv, pool, rmqClient)

	dbComp := component.NewDBChecker(pool)
	rmqConn := component.NewRabbitMQChecker(rmqClient.Connection())
	healthCheckService := service.NewHealthCheckService(rmqConn, dbComp)
	healthcheck.Process(procSpawnFn, appCtx, srv, healthCheckService)

	var webapiTLS *webapi.TLSConfig
	if cfg.AppEnv != "local" {
		webapiTLS = &webapi.TLSConfig{
			CertFile: cfg.WebAPICertFile,
			KeyFile:  cfg.WebAPIKeyFile,
		}
	}
	webapi.Start(procSpawnFn, srv, cfg.APIPort, webapiTLS)

	appController.Wait()
}
