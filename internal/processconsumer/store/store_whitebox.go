package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/google/uuid"
)

func (s *Store) DeleteMessageByUUID(ctx context.Context, uuid uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE uuid = $1`, ProcessEventsTable)

	_, err := s.pool.Exec(ctx, query, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete message from DB: %w", err)
	}

	return nil
}

func (s *Store) GetMessageByUUID(ctx context.Context, uuid uuid.UUID) (model.Message, error) {
	query := fmt.Sprintf(`SELECT payload FROM %s WHERE uuid = $1`, ProcessEventsTable)

	var m model.Message
	err := s.pool.QueryRow(ctx, query, uuid).Scan(&m)
	if err != nil {
		return model.Message{}, fmt.Errorf("failed to get message by uuid: %w", err)
	}
	return m, nil
}

func (s *Store) GetCompletedAtByUUID(ctx context.Context, uuid uuid.UUID) (time.Time, error) {
	var completedAt time.Time
	query := fmt.Sprintf(`SELECT completed_at FROM %s WHERE uuid = $1`, ProcessEventsTable)
	err := s.pool.QueryRow(ctx, query, uuid).Scan(&completedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get completed_at for UUID %s: %w", uuid, err)
	}
	return completedAt, nil
}