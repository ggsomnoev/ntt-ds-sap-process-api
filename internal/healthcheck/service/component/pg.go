package component

import (
	"context"
	"fmt"
	"time"
)

const timeout = 2 * time.Second

type DBConnector interface {
	Ping(ctx context.Context) error
}

type DBChecker struct {
	db DBConnector
}

func NewDBChecker(db DBConnector) *DBChecker {
	return &DBChecker{db: db}
}

func (d *DBChecker) Name() string {
	return "database"
}

func (d *DBChecker) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := d.db.Ping(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}
	return nil
}
