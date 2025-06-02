package processloader

import (
	"context"
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/process"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	dir string,
	pool *pgxpool.Pool,
) {
	store := store.NewStore(pool)
	readerSvc := service.NewConfigReader(dir, store)

	validationSvc := service.NewValidationService()

	senderSvc := service.NewSenderService()

	process.Process(procSpawnFn, ctx, readerSvc, func(path string) error {
		process, err := readerSvc.ParseConfigFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}

		if err := validationSvc.Validate(process); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		if err := senderSvc.Send(process); err != nil {
			return fmt.Errorf("webapi send error: %w", err)
		}

		return nil
	})
}
