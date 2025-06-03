package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"gopkg.in/yaml.v3"
)

type ConfigReader struct {
	dir string
}

func NewConfigReader(dir string) *ConfigReader {
	return &ConfigReader{
		dir: dir,
	}
}

func (cr *ConfigReader) ReadYAMLFiles() ([]string, error) {
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
