package pgtx

import (
	"context"
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg/txctx"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Atomically(
	ctx context.Context,
	pool *pgxpool.Pool,
	iso pgx.TxIsoLevel,
	cb func(ctx context.Context, tx pgx.Tx) error,
) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			if errRb := tx.Rollback(ctx); errRb != nil {
				logger.GetLogger().Errorf("could not rollback transaction: %v", err)
			}
		}
	}()

	ctxWithTx := txctx.WithTx(ctx, tx)

	if err = cb(ctxWithTx, tx); err != nil {
		return fmt.Errorf("tx callback failed: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}