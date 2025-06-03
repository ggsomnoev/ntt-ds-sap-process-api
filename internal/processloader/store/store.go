package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg/pgtx"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg/txctx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ProcessedFilesTable = "processed_files"

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) AddProcessedFile(ctx context.Context, filename string) error {
	tx, err := txctx.GetTx(ctx)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (filename) VALUES ($1)
	`, ProcessedFilesTable)

	_, err = tx.Exec(ctx, query, filename)
	return err
}

func (s *Store) MarkCompleted(ctx context.Context, filename string) error {
	tx, err := txctx.GetTx(ctx)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET completed_at = $1 WHERE filename = $2`, ProcessedFilesTable)
	_, err = tx.Exec(ctx, query, time.Now().UTC(), filename)
	return err
}

func (s *Store) FileExists(ctx context.Context, filename string) (bool, error) {
	tx, err := txctx.GetTx(ctx)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE filename = $1)`, ProcessedFilesTable)
	var exists bool
	err = tx.QueryRow(ctx, query, filename).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query failed: %w", err)
	}

	return exists, nil
}

func (s *Store) RunInAtomically(ctx context.Context, cb func(ctx context.Context) error) error {
	err := pgtx.Atomically(ctx, s.pool, pgx.Serializable, func(ctx context.Context, tx pgx.Tx) error {
		ctxWithTx := txctx.WithTx(ctx, tx)

		if err := cb(ctxWithTx); err != nil {
			return fmt.Errorf("callback failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}
