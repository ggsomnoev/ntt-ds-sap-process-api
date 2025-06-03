package producer

import (
	"context"
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processproducer/handler"
	"github.com/labstack/echo/v4"
)

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	srv *echo.Echo,
	publisher handler.Publisher,
) {
	procSpawnFn(func(ctx context.Context) error {
		handler.RegisterHandlers(ctx, srv, publisher)

		<-ctx.Done()
		logger.GetLogger().Info("closing the RabbitMQ connection due to app exit")

		if publisher != nil {
			err := publisher.Close()
			if err != nil {
				return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
			}
		}

		return nil
	}, "Publisher")
}