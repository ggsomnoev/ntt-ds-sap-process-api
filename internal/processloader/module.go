package processloader

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/process"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/reader"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/sender"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/validator"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	dir string,
	pool *pgxpool.Pool,
	webAPIAddress string,
) {
	store := store.NewStore(pool)
	reader := reader.NewConfigReader(dir)
	validator := validator.NewProcessValidator()
	sender := sender.NewHTTPProcessSender(webAPIAddress)
	processLoaderSvc := service.NewService(store, reader, validator, sender)

	process.Process(procSpawnFn, ctx, processLoaderSvc)
}
