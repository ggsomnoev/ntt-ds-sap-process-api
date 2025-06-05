package service

import (
	"context"
	"fmt"
)

//counterfeiter:generate . Store
type Store interface {
	SaveProcessDefinitionMeta(context.Context, string, string) error
}

//counterfeiter:generate . Reader
type Reader interface {
	ReadYAMLFiles() ([]string, error)
	GetProcessNameFromFile(string) (string, error)
}

type Service struct {
	store  Store
	reader Reader
}

func NewService(store Store, reader Reader) *Service {
	return &Service{
		store:  store,
		reader: reader,
	}
}

func (s *Service) IndexProcessDefinitions(ctx context.Context) error {
	files, err := s.reader.ReadYAMLFiles()
	if err != nil {
		return fmt.Errorf("failed to read process config files: %w", err)
	}

	for _, path := range files {
		processName, err := s.reader.GetProcessNameFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to get process name: %w", err)
		}

		err = s.store.SaveProcessDefinitionMeta(ctx, processName, path)
		if err != nil {
			return fmt.Errorf("failed to store process definition metadata: %w", err)
		}
	}

	return nil
}
