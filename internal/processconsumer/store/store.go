package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg/pgtx"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/pg/txctx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ProcessEventsTable = "process_events"

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) AddMessage(ctx context.Context, m model.Message) error {
	tx, err := txctx.GetTx(ctx)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (uuid, payload, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO NOTHING
	`, ProcessEventsTable)

	_, err = tx.Exec(ctx, query, m.UUID, m, time.Now().UTC())
	return err
}

func (s *Store) MarkCompleted(ctx context.Context, uuid uuid.UUID) error {
	tx, err := txctx.GetTx(ctx)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET completed_at = $1 WHERE uuid = $2`, ProcessEventsTable)
	_, err = tx.Exec(ctx, query, time.Now().UTC(), uuid)
	return err
}

func (s *Store) MessageExists(ctx context.Context, id uuid.UUID) (bool, error) {
	tx, err := txctx.GetTx(ctx)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE uuid = $1)`, ProcessEventsTable)
	var exists bool
	err = tx.QueryRow(ctx, query, id).Scan(&exists)
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