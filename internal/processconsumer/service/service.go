package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
			logger.GetLogger().Infof("Skipping already processed process with UUID: %s", message.UUID)
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

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errMsg error

		taskStatus = make(map[string]bool)
		statusMu   sync.Mutex
	)

	for _, task := range def.Tasks {
		wg.Add(1)
		go func(t model.Task) {
			defer wg.Done()
			if err := s.runTask(ctx, processID, task, &taskStatus, &statusMu); err != nil {
				mu.Lock()
				if errMsg == nil {
					errMsg = err
				}
				mu.Unlock()
			}
		}(task)
	}

	wg.Wait()

	if errMsg != nil {
		_ = s.processStore.UpdateProcessStatus(ctx, processID, model.StatusFailed)
		return fmt.Errorf("failed task: %w", errMsg)
	}

	if err := s.processStore.UpdateProcessStatus(ctx, processID, model.StatusCompleted); err != nil {
		return fmt.Errorf("failed to update process status: %w", err)
	}

	return nil
}

func (s *Service) runTask(
	ctx context.Context,
	processID uuid.UUID,
	task model.Task,
	taskStatus *map[string]bool,
	statusMu *sync.Mutex,
) error {
	// Wait for all dependencies
	for {
		statusMu.Lock()
		allDepsDone := true
		for _, dep := range task.WaitFor {
			if !(*taskStatus)[dep] {
				allDepsDone = false
				break
			}
		}
		statusMu.Unlock()

		if allDepsDone {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	executor, ok := s.executors[task.Class]
	if !ok {
		msg := fmt.Sprintf("No executor registered for class type: %s", task.Class)
		s.processStore.AppendProcessLog(ctx, processID, msg)
		return errors.New(msg)
	}

	if err := executor.Run(ctx, task); err != nil {
		msg := fmt.Sprintf("Failed to run task %s: %v", task.Name, err)
		s.processStore.AppendProcessLog(ctx, processID, msg)
		return fmt.Errorf("executor error: %w", err)
	}

	s.processStore.AppendProcessLog(ctx, processID, fmt.Sprintf("Task %s completed", task.Name))

	statusMu.Lock()
	(*taskStatus)[task.Name] = true
	statusMu.Unlock()

	return nil
}
