package process

import (
	"context"
	"net/http"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/labstack/echo/v4"
)

func RegisterHandlers(ctx context.Context, srv *echo.Echo, health Service) {
	if srv != nil {
		srv.GET("/healthz", handleHealthCheck(health))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleHealthCheck(h Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		status, ok := h.Status()
		code := http.StatusOK
		if !ok {
			code = http.StatusServiceUnavailable
		}
		return c.JSON(code, status)
	}
}