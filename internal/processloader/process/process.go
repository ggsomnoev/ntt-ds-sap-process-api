package process

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/lifecycle"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
)

const pollInterval = 15 * time.Second

type Service interface {
	TryProcessConfigs(func(string) error) error
	ParseConfigFile(string) (model.ProcessDefinition, error)
}

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	ctx context.Context,
	configChecker Service,
	cb func(path string) error,
) {
	procSpawnFn(func(ctx context.Context) error {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				err := configChecker.TryProcessConfigs(cb)
				if err != nil {
					return fmt.Errorf("failed to process config files: %w", err)
				}
			}
		}
	}, "Process Config Loader")
}
