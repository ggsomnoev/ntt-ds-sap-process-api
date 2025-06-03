package healthcheck

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/healthcheck/process"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/labstack/echo/v4"
)

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	srv *echo.Echo,
	healthCheckService process.Service,
) {
	process.Process(procSpawnFn, ctx, healthCheckService)

	process.RegisterHandlers(ctx, srv, healthCheckService)
}
