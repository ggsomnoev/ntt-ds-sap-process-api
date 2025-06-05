package processloader

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/process"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/service/reader"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processloader/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	dir string,
	pool *pgxpool.Pool,
) (*store.Store, *reader.ConfigReader) {
	store := store.NewStore(pool)
	reader := reader.NewConfigReader(dir)
	processLoaderSvc := service.NewService(store, reader)

	process.Process(procSpawnFn, ctx, processLoaderSvc)

	return store, reader
}
