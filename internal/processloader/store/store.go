package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ProcessedDefinitionsTable = "process_definitions"

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) SaveProcessDefinitionMeta(ctx context.Context, name, path string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, path)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET path = EXCLUDED.path
	`, ProcessedDefinitionsTable)

	_, err := s.pool.Exec(ctx, query, name, path)
	if err != nil {
		return fmt.Errorf("failed to insert process metadata: %w", err)
	}
	return nil
}

func (s *Store) GetProcessPathByName(ctx context.Context, name string) (string, error) {
	query := fmt.Sprintf(`
		SELECT path FROM %s
		WHERE name = $1
	`, ProcessedDefinitionsTable)
	var path string
	err := s.pool.QueryRow(ctx, query, name).Scan(&path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("process definition not found: %s", name)
		}
		return "", fmt.Errorf("failed to fetch path for process %s: %w", name, err)
	}
	return path, nil
}
