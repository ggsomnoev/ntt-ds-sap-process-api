package handler

import (
	"context"
	"net/http"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const successfullyAddedProcess = "successfully added process"

type StartProcessRequest struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters"`
}

//counterfeiter:generate . ProcessDefinitionStore
type ProcessDefinitionStore interface {
	GetProcessPathByName(context.Context, string) (string, error)
}

//counterfeiter:generate . Reader
type Reader interface {
	ParseConfigFile(string) (model.ProcessDefinition, error)
	GetProcessNameFromFile(string) (string, error)
	ApplyTemplatingToTasks([]model.Task, map[string]string) ([]model.Task, error)
}

//counterfeiter:generate . Publisher
type Publisher interface {
	Publish(context.Context, model.Message) error
	Close() error
}

//counterfeiter:generate . Validator
type Validator interface {
	Validate(model.ProcessDefinition) error
}

func RegisterHandlers(ctx context.Context, srv *echo.Echo, publisher Publisher, reader Reader, store ProcessDefinitionStore, validator Validator) {
	if srv != nil {
		srv.POST("/startProcess", handleNewProcess(ctx, publisher, reader, store, validator))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleNewProcess(ctx context.Context, publisher Publisher, reader Reader, store ProcessDefinitionStore, validator Validator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req StartProcessRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		processPath, err := store.GetProcessPathByName(ctx, req.Name)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Process definition template not found",
				"error":   err.Error(),
			})
		}

		processDef, err := reader.ParseConfigFile(processPath)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Could not parse process definition template",
				"error":   err.Error(),
			})
		}

		tasks, err := reader.ApplyTemplatingToTasks(processDef.Tasks, req.Parameters)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Failed to apply parameters",
				"error":   err.Error(),
			})
		}

		processDef.Tasks = tasks

		if err := validator.Validate(processDef); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		message := model.Message{
			UUID:              uuid.New(),
			ProcessDefinition: processDef,
		}

		if err := publisher.Publish(ctx, message); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": successfullyAddedProcess,
		})
	}
}
