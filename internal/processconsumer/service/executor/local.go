package executor

import (
	"context"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
)

type LocalCmdExecutor struct{}

func NewLocalCmdService() *LocalCmdExecutor {
	return &LocalCmdExecutor{}
}

func (le *LocalCmdExecutor) Run(ctx context.Context, task model.Task) error {
	logger.GetLogger().Infof("Got local task: %v", task)
	return nil
}
