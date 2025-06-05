package consumer

import (
	"context"
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/handler"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/executor"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Consumer interface {
	Consume(context.Context, func(context.Context, model.Message) error) error
	Close() error
}

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	srv *echo.Echo,
	pool *pgxpool.Pool,
	consumer Consumer,
) {
	procSpawnFn(func(ctx context.Context) error {
		processStore := store.NewProcessDBStore(pool)
		messageStore := store.NewStore(pool)

		handler.RegisterHandlers(ctx, srv, processStore)

		taskHandlers := map[model.ClassType]service.Executor{
			model.LocalCmd: executor.NewLocalCmdService(),
			model.SshCmd:   executor.NewSSHCmdExecutor(),
			model.ScpCmd:   executor.NewLocalCmdService(),
		}

		processHandlerSvc := service.NewService(messageStore, processStore, taskHandlers)
		err := consumer.Consume(ctx, processHandlerSvc.Run)
		if err != nil {
			return fmt.Errorf("consume failed: %w", err)
		}

		<-ctx.Done()
		logger.GetLogger().Info("closing the RabbitMQ connection due to app exit")

		if err := consumer.Close(); err != nil {
			return fmt.Errorf("failed to close consumer: %w", err)
		}

		return nil
	}, "Consumer")
}
