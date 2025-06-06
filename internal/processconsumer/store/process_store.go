package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ProcessRunsTable = "process_runs"
	ProcessLogsTable = "process_logs"
)

type ProcessDBStore struct {
	pool *pgxpool.Pool
}

func NewProcessDBStore(pool *pgxpool.Pool) *ProcessDBStore {
	return &ProcessDBStore{pool: pool}
}

func (s *ProcessDBStore) InsertProcess(ctx context.Context, run model.ProcessRun) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, definition, status, started_at)
		VALUES ($1, $2, $3, $4)
	`, ProcessRunsTable)

	_, err := s.pool.Exec(ctx, query, run.ID, run.Definition, run.Status, run.StartedAt)
	return err
}

func (s *ProcessDBStore) UpdateProcessStatus(ctx context.Context, id uuid.UUID, status model.ProcessStatus) error {
	query := fmt.Sprintf(`
		UPDATE %s SET status = $1, ended_at = $2 WHERE id = $3
	`, ProcessRunsTable)
	
	completedAt := time.Now()
	_, err := s.pool.Exec(ctx, query, status, &completedAt, id)
	return err
}

func (s *ProcessDBStore) GetProcessByID(ctx context.Context, id uuid.UUID) (model.ProcessRun, error) {
	var (
		run     model.ProcessRun
		endedAt *time.Time
	)

	query := fmt.Sprintf(`
		SELECT id, definition, status, started_at, ended_at
		FROM %s WHERE id = $1
	`, ProcessRunsTable)

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&run.ID, &run.Definition, &run.Status, &run.StartedAt, &endedAt,
	)
	if err != nil {
		return model.ProcessRun{}, err
	}

	run.EndedAt = endedAt
	return run, nil
}

// Currenly lists all process, not only with status "running". The goal here is to have some processes to list.
func (s *ProcessDBStore) ListRunningProcesses(ctx context.Context) ([]model.ProcessRun, error) {
	query := fmt.Sprintf(`
		SELECT id, definition, status, started_at, ended_at
		FROM %s ORDER BY started_at DESC
	`, ProcessRunsTable)

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.ProcessRun
	for rows.Next() {
		var run model.ProcessRun
		var endedAt *time.Time

		if err := rows.Scan(&run.ID, &run.Definition, &run.Status, &run.StartedAt, &endedAt); err != nil {
			return nil, err
		}

		run.EndedAt = endedAt
		results = append(results, run)
	}

	return results, nil
}

func (s *ProcessDBStore) AppendProcessLog(ctx context.Context, processID uuid.UUID, log string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (process_id, log) VALUES ($1, $2)
	`, ProcessLogsTable)

	_, err := s.pool.Exec(ctx, query, processID, log)
	return err
}

func (s *ProcessDBStore) GetProcessLogs(ctx context.Context, processID uuid.UUID) ([]model.ProcessLog, error) {
	query := fmt.Sprintf(`
		SELECT id, process_id, log, created_at FROM %s
		WHERE process_id = $1 ORDER BY created_at ASC
	`, ProcessLogsTable)

	rows, err := s.pool.Query(ctx, query, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.ProcessLog
	for rows.Next() {
		var l model.ProcessLog
		if err := rows.Scan(&l.ID, &l.ProcessID, &l.Log, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
