package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"gopkg.in/yaml.v3"
)

type Store interface {
	AddProcessedFile(context.Context, string) error
	MarkCompleted(context.Context, string) error
	FileExists(context.Context, string) (bool, error)
	RunInAtomically(context.Context, func(ctx context.Context) error) error
}

type ConfigReader struct {
	dir   string
	store Store
}

func NewConfigReader(dir string, store Store) *ConfigReader {
	return &ConfigReader{
		dir:   dir,
		store: store,
	}
}

// TryProcessConfigs scans the directory for YAML files, checks whether they have already been processed,
// and if not, validates and registers them in the store. Successfully handled files are marked as completed.
func (cr *ConfigReader) TryProcessConfigs(cb func(path string) error) error {
	files, err := cr.readYAMLFiles()
	if err != nil {
		return fmt.Errorf("failed to read process config files: %w", err)
	}

	for _, path := range files {
		filename := filepath.Base(path)

		err := cr.store.RunInAtomically(context.Background(), func(ctx context.Context) error {
			exists, err := cr.store.FileExists(ctx, filename)
			if err != nil {
				return fmt.Errorf("failed to check if file exists: %w", err)
			}
			if exists {
				logger.GetLogger().Warnf("Skipping file '%s' — already processed", filename)
				return nil
			}

			if err := cr.store.AddProcessedFile(ctx, filename); err != nil {
				return fmt.Errorf("adding file failed: %w", err)
			}

			if err := cb(path); err != nil {
				return fmt.Errorf("failed during cb execution: %w", err)
			}

			if err := cr.store.MarkCompleted(ctx, filename); err != nil {
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

func (cr *ConfigReader) readYAMLFiles() ([]string, error) {
	entries, err := os.ReadDir(cr.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml") {
			files = append(files, filepath.Join(cr.dir, entry.Name()))
		}
	}

	return files, nil
}

func (cr *ConfigReader) ParseConfigFile(path string) (model.ProcessDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.ProcessDefinition{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var process model.ProcessDefinition
	if err := yaml.Unmarshal(data, &process); err != nil {
		return model.ProcessDefinition{}, fmt.Errorf("failed to unmarshal file %s: %w", path, err)
	}

	return process, nil
}
