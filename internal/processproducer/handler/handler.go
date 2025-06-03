package handler

import (
	"context"
	"net/http"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const successfullyAddedProcess = "successfully added process"

//counterfeiter:generate . Publisher
type Publisher interface {
	Publish(context.Context, model.Message) error
	Close() error
}

func RegisterHandlers(ctx context.Context, srv *echo.Echo, publisher Publisher) {
	if srv != nil {
		srv.POST("/startProcess", handleNewProcess(ctx, publisher))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleNewProcess(ctx context.Context, publisher Publisher) echo.HandlerFunc {
	validatorSvc := validator.NewProcessValidator()
	
	return func(c echo.Context) error {
		var process model.ProcessDefinition
		if err := c.Bind(&process); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		if err := validatorSvc.Validate(process); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		message := model.Message{
			UUID:         uuid.New(),
			ProcessDefinition: process,
		}

		if err := publisher.Publish(ctx, message); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": successfullyAddedProcess,
		})
	}
}
