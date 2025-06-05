package reader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

func (cr *ConfigReader) GetProcessNameFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var process model.ProcessDefinition
	if err := yaml.Unmarshal(data, &process); err != nil {
		return "", fmt.Errorf("failed to unmarshal file %s: %w", path, err)
	}

	if process.Name == "" {
		return "", fmt.Errorf("process name missing in file %s", path)
	}

	return process.Name, nil
}

func (cr *ConfigReader) ApplyTemplatingToTasks(tasks []model.Task, inputs map[string]string) ([]model.Task, error) {
	var rendered []model.Task
	for _, t := range tasks {
		params := make(map[string]string)
		for key, tmplStr := range t.Parameters {
			tmpl, err := template.New(key).Parse(tmplStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse template for task %s param %s: %w", t.Name, key, err)
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, inputs); err != nil {
				return nil, fmt.Errorf("failed to render template for task %s param %s: %w", t.Name, key, err)
			}
			params[key] = buf.String()
		}
		t.Parameters = params
		rendered = append(rendered, t)
	}
	return rendered, nil
}
