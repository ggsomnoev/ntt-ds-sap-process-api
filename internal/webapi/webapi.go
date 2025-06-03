package webapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"

	"github.com/labstack/echo/v4"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
)

func NewServer(ctx context.Context) *echo.Echo {
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true

	e.Use(JSONLoggerMiddleware)

	return e
}

func Start(procSpawnFn lifecycle.ProcessSpawnFunc, srv *echo.Echo, apiPort string) {
	startServer(procSpawnFn, srv, apiPort)
	stopServer(procSpawnFn, srv)
}

func startServer(procSpawnFn lifecycle.ProcessSpawnFunc, e *echo.Echo, apiPort string) {
	procSpawnFn(func(ctx context.Context) error {
		logger.GetLogger().Info(fmt.Sprintf("starting the WebAPI server on port %s", apiPort))

		err := e.Start(fmt.Sprintf(":%s", apiPort))
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start the webAPI server: %w", err)
		}

		return nil
	}, "WebAPI Starter")
}

// With graceful shut down.
func stopServer(procSpawnFn lifecycle.ProcessSpawnFunc, e *echo.Echo) {
	procSpawnFn(func(ctx context.Context) error {
		<-ctx.Done()
		logger.GetLogger().Info("stopping the WebAPI server due to app exit")

		ctxGrace, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		err := e.Shutdown(ctxGrace)
		if err != nil {
			return fmt.Errorf("failed to shutdown the webAPI server: %w", err)
		}

		return nil
	}, "WebAPI Stopper")
}