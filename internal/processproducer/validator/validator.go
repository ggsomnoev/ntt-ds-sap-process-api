package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
)

type ProcessValidator struct{}

func NewProcessValidator() *ProcessValidator {
	return &ProcessValidator{}
}

func (pv *ProcessValidator) Validate(proc model.ProcessDefinition) error {
	if strings.TrimSpace(proc.Name) == "" {
		return errors.New("process name must not be empty")
	}

	taskNames := make(map[string]struct{})
	for _, task := range proc.Tasks {
		if strings.TrimSpace(task.Name) == "" {
			return errors.New("task name must not be empty")
		}
		if strings.TrimSpace(string(task.Class)) == "" {
			return fmt.Errorf("task '%s' class must not be empty", task.Name)
		}
		if _, exists := taskNames[task.Name]; exists {
			return fmt.Errorf("duplicate task name found: %s", task.Name)
		}
		taskNames[task.Name] = struct{}{}

		for _, dep := range task.WaitFor {
			if dep == task.Name {
				return fmt.Errorf("task '%s' cannot wait for itself", task.Name)
			}
		}
	}

	paramNames := make(map[string]struct{})
	for _, param := range proc.Params {
		if strings.TrimSpace(param.Name) == "" {
			return errors.New("param name must not be empty")
		}
		if _, exists := paramNames[param.Name]; exists {
			return fmt.Errorf("duplicate param name found: %s", param.Name)
		}
		paramNames[param.Name] = struct{}{}
	}

	for _, task := range proc.Tasks {
		for _, dep := range task.WaitFor {
			if _, ok := taskNames[dep]; !ok {
				return fmt.Errorf("task '%s' waits for unknown task '%s'", task.Name, dep)
			}
		}
	}

	return nil
}

func (v *ProcessValidator) ValidateMandatoryParams(def model.ProcessDefinition, userParams map[string]string) error {
	var missing []string

	for _, param := range def.Params {
		if param.Mandatory {
			if val, ok := userParams[param.Name]; !ok || strings.TrimSpace(val) == "" {
				missing = append(missing, param.Name)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing mandatory parameters: %s", strings.Join(missing, ", "))
	}
	return nil
}
