package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//go:generate counterfeiter . ProcessStore
type ProcessStore interface {
	ListRunningProcesses(context.Context) ([]model.ProcessRun, error)
	GetProcessByID(context.Context, uuid.UUID) (model.ProcessRun, error)
	UpdateProcessStatus(context.Context, uuid.UUID, model.ProcessStatus) error
	GetProcessLogs(context.Context, uuid.UUID) ([]model.ProcessLog, error)
}

func RegisterHandlers(ctx context.Context, srv *echo.Echo, store ProcessStore) {
	if srv != nil {
		srv.GET("/listProcesses", handleListProcesses(ctx, store))
		srv.GET("/listProcess/:id", handleGetProcess(ctx, store))
		srv.POST("/stopProcess/:id", handleStopProcess(ctx, store))
		srv.GET("/processlog/:id", handleGetProcessLogs(ctx, store))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleListProcesses(ctx context.Context, store ProcessStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		processes, err := store.ListRunningProcesses(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list processes")
		}
		return c.JSON(http.StatusOK, processes)
	}
}

func handleGetProcess(ctx context.Context, store ProcessStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid process ID")
		}

		process, err := store.GetProcessByID(ctx, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Process not found")
		}
		return c.JSON(http.StatusOK, process)
	}
}

func handleStopProcess(ctx context.Context, store ProcessStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid process ID")
		}

		if err := store.UpdateProcessStatus(ctx, id, model.StatusStopped); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to stop process")
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": fmt.Sprintf("process with id - %d - successfully stopped!", id),
		})
	}
}

func handleGetProcessLogs(ctx context.Context, store ProcessStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid process ID")
		}

		logs, err := store.GetProcessLogs(ctx, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Logs not found")
		}
		return c.JSON(http.StatusOK, logs)
	}
}
