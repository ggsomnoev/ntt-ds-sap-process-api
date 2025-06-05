package main

import (
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/config"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service/component"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/webapi"
)

func main() {
	appController := lifecycle.NewController()
	appCtx, procSpawnFn := appController.Start()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed reading configuration: %w", err))
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

	srv := webapi.NewServer(appCtx)

	processloader.Process(procSpawnFn, appCtx, cfg.ProcessCfgDir, pool)

	dbComp := component.NewDBChecker(pool)
	healthCheckService := service.NewHealthCheckService(dbComp)
	healthcheck.Process(procSpawnFn, appCtx, srv, healthCheckService)

	webapi.Start(procSpawnFn, srv, cfg.APIPort)

	appController.Wait()
}
