package process

import (
	"context"
	"errors"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
)

const pollInterval = 10 * time.Second

type Service interface {
	Status() (map[string]string, bool)
}

// The LRP and K8s liveness probe will handle app restart. Not really needed but nice to have for stuck services.
func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	healthChecker Service,
) {
	procSpawnFn(func(ctx context.Context) error {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				_, healthy := healthChecker.Status()
				if !healthy {
					return errors.New("health check failed; triggering shutdown")
				}
			}
		}
	}, "Health Checker")
}
