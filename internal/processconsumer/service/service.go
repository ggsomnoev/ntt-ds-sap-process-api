package service

import (
	"context"
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/google/uuid"
)

//counterfeiter:generate . Store
type Store interface {
	AddMessage(context.Context, model.Message) error
	MarkCompleted(context.Context, uuid.UUID) error
	MessageExists(context.Context, uuid.UUID) (bool, error)
	RunInAtomically(context.Context, func(context.Context) error) error
}

//counterfeiter:generate . Executor
type Executor interface {
	Run(ctx context.Context, task model.Task) error
}

type Service struct {
	store     Store
	executors map[model.ClassType]Executor
}

func NewService(
	store Store,
	executors map[model.ClassType]Executor,
) *Service {
	return &Service{
		store:     store,
		executors: executors,
	}
}

func (s *Service) Run(ctx context.Context, message model.Message) error {
	return s.store.RunInAtomically(ctx, func(ctx context.Context) error {
		exists, err := s.store.MessageExists(ctx, message.UUID)
		if err != nil {
			return fmt.Errorf("failed to check message: %w", err)
		}
		if exists {
			logger.GetLogger().Infof("Skipping already processed UUID: %s", message.UUID)
			return nil
		}

		if err := s.store.AddMessage(ctx, message); err != nil {
			return fmt.Errorf("failed to persist process event: %w", err)
		}

		logger.GetLogger().Infof("Running process: %s...", message.ProcessDefinition.Name)
		if err := s.runProcessDefinition(ctx, message.ProcessDefinition); err != nil {
			return fmt.Errorf("task execution failed: %w", err)
		}
		logger.GetLogger().Info("Process executed successfully!")

		if err := s.store.MarkCompleted(ctx, message.UUID); err != nil {
			return fmt.Errorf("failed to mark process as completed: %w", err)
		}

		return nil
	})
}

func (s *Service) runProcessDefinition(ctx context.Context, def model.ProcessDefinition) error {
	taskStatus := make(map[string]bool)

	for _, task := range def.Tasks {
		for _, dep := range task.WaitFor {
			if !taskStatus[dep] {
				logger.GetLogger().Warnf("Task %s cannot run before dependency %s", task.Name, dep)
				return nil
			}
		}

		executor, ok := s.executors[task.Class]
		if !ok {
			return fmt.Errorf("no executor registered for class type: %s", task.Class)
		}

		if err := executor.Run(ctx, task); err != nil {
			return fmt.Errorf("failed to run task %s: %w", task.Name, err)
		}

		taskStatus[task.Name] = true
	}

	return nil
}
