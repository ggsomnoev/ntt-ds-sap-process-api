package store

import (
	"context"
	"fmt"
	"time"
)

func (s *Store) DeleteProcessedFile(ctx context.Context, filename string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE filename = $1`, ProcessedFilesTable)

	_, err := s.pool.Exec(ctx, query, filename)
	if err != nil {
		return fmt.Errorf("failed to delete message from DB: %w", err)
	}

	return nil
}

func (s *Store) GetCompletedAtByFilename(ctx context.Context, filename string) (time.Time, error) {
	var completedAt time.Time
	query := fmt.Sprintf(`SELECT completed_at FROM %s WHERE filename = $1`, ProcessedFilesTable)
	err := s.pool.QueryRow(ctx, query, filename).Scan(&completedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get completed_at for filename - %s: %w", filename, err)
	}
	return completedAt, nil
}
