package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
)

type Store interface {
	AddProcessedFile(context.Context, string) error
	MarkCompleted(context.Context, string) error
	FileExists(context.Context, string) (bool, error)
	RunInAtomically(context.Context, func(ctx context.Context) error) error
}

type Reader interface {
	ParseConfigFile(string) (model.ProcessDefinition, error)
	ReadYAMLFiles() ([]string, error)
}

type Validator interface {
	Validate(model.ProcessDefinition) error
}

type Sender interface {
	Send(model.ProcessDefinition) error
}

type Service struct {
	store     Store
	reader    Reader
	validator Validator
	sender    Sender
}

func NewService(store Store, reader Reader, validator Validator, sender Sender) *Service {
	return &Service{
		store:     store,
		reader:    reader,
		validator: validator,
		sender:    sender,
	}
}

// TryProcessConfigs scans the directory for YAML files, checks whether they have already been processed,
// and if not, validates and registers them in the store. Successfully handled files are marked as completed.
func (s *Service) TryProcessConfigs() error {
	files, err := s.reader.ReadYAMLFiles()
	if err != nil {
		return fmt.Errorf("failed to read process config files: %w", err)
	}

	for _, path := range files {
		filename := filepath.Base(path)

		err := s.store.RunInAtomically(context.Background(), func(ctx context.Context) error {
			exists, err := s.store.FileExists(ctx, filename)
			if err != nil {
				return fmt.Errorf("failed to check if file exists: %w", err)
			}
			if exists {
				logger.GetLogger().Warnf("Skipping file '%s' — already processed", filename)
				return nil
			}

			if err := s.store.AddProcessedFile(ctx, filename); err != nil {
				return fmt.Errorf("adding file failed: %w", err)
			}

			process, err := s.reader.ParseConfigFile(path)
			if err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}

			if err := s.validator.Validate(process); err != nil {
				return fmt.Errorf("validation error: %w", err)
			}

			if err := s.sender.Send(process); err != nil {
				return fmt.Errorf("webapi send error: %w", err)
			}

			if err := s.store.MarkCompleted(ctx, filename); err != nil {
				return fmt.Errorf("marking file as complete failed: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("db transaction failed: %w", err)
		}
	}

	return nil
}
