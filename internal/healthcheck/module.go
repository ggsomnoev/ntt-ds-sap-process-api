package healthcheck

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/process"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/service/component"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	srv *echo.Echo,
	pool *pgxpool.Pool,
) {

	dbComp := component.NewDBChecker(pool)
	healthCheckService := service.NewHealthCheckService(dbComp)

	process.Process(procSpawnFn, ctx, healthCheckService)

	process.RegisterHandlers(ctx, srv, healthCheckService)
}
