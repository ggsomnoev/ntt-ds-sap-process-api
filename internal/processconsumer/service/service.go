package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/google/uuid"
)

//go:generate counterfeiter . ProcessStore
type ProcessStore interface {
	InsertProcess(context.Context, model.ProcessRun) error
	AppendProcessLog(context.Context, uuid.UUID, string) error
	UpdateProcessStatus(context.Context, uuid.UUID, model.ProcessStatus) error
}

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
	store        Store
	processStore ProcessStore
	executors    map[model.ClassType]Executor
}

func NewService(
	store Store,
	processStore ProcessStore,
	executors map[model.ClassType]Executor,
) *Service {
	return &Service{
		store:        store,
		processStore: processStore,
		executors:    executors,
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
	processID := uuid.New()
	startedAt := time.Now()

	process := model.ProcessRun{
		ID:         processID,
		Definition: def,
		Status:     model.StatusRunning,
		StartedAt:  startedAt,
	}

	if err := s.processStore.InsertProcess(ctx, process); err != nil {
		return fmt.Errorf("failed to insert process: %w", err)
	}

	taskStatus := make(map[string]bool)

	for _, task := range def.Tasks {
		for _, dep := range task.WaitFor {
			if !taskStatus[dep] {
				msg := fmt.Sprintf("Task %s cannot run before dependency %s", task.Name, dep)
				logger.GetLogger().Warnf(msg)
				s.processStore.AppendProcessLog(ctx, processID, msg)
				return nil
			}
		}

		executor, ok := s.executors[task.Class]
		if !ok {
			msg := fmt.Sprintf("No executor registered for class type: %s", task.Class)
			s.processStore.AppendProcessLog(ctx, processID, msg)
			return fmt.Errorf(msg)
		}

		if err := executor.Run(ctx, task); err != nil {
			msg := fmt.Sprintf("Failed to run task %s: %v", task.Name, err)
			s.processStore.AppendProcessLog(ctx, processID, msg)
			return fmt.Errorf("executor error: %w", err)
		}

		s.processStore.AppendProcessLog(ctx, processID, fmt.Sprintf("Task %s completed", task.Name))
		taskStatus[task.Name] = true
	}

	process.Status = model.StatusCompleted

	if err := s.processStore.UpdateProcessStatus(ctx, process.ID, process.Status); err != nil {
		return fmt.Errorf("failed to update process status: %w", err)
	}

	return nil
}
