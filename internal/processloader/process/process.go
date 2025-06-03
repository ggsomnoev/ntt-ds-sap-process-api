package process

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
)

const pollInterval = 60 * time.Second

type Service interface {
	TryProcessConfigs(ctx context.Context) error
}

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	processLoader Service,
) {
	procSpawnFn(func(ctx context.Context) error {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				err := processLoader.TryProcessConfigs(ctx)
				if err != nil {
					return fmt.Errorf("failed to process config files: %w", err)
				}
			}
		}
	}, "Process Config Loader")
}
