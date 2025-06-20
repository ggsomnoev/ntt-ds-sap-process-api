package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	UUID              uuid.UUID         `json:"uuid"`
	ProcessDefinition ProcessDefinition `json:"process_definition"`
}

type ProcessDefinition struct {
	Name   string  `yaml:"name" json:"name"`
	Params []Param `yaml:"params" json:"params"`
	Tasks  []Task  `yaml:"tasks" json:"tasks"`
}

type Param struct {
	Name        string `yaml:"name" json:"name"`
	Mandatory   bool   `yaml:"mandatory" json:"mandatory"`
	Description string `yaml:"description" json:"description"`
	DefValue    string `yaml:"defvalue" json:"defvalue"`
}

type Task struct {
	Name       string            `yaml:"name" json:"name"`
	Class      ClassType         `yaml:"class" json:"class"`
	Parameters map[string]string `yaml:"parameters" json:"parameters"`
	WaitFor    []string          `yaml:"waitfor,omitempty" json:"waitfor,omitempty"`
}

type ClassType string

const (
	LocalCmd ClassType = "localCmd"
	SshCmd   ClassType = "sshCmd"
	ScpCmd   ClassType = "scpCmd"
)

var validClassTypes = map[string]ClassType{
	"localcmd": LocalCmd,
	"sshcmd":   SshCmd,
	"scpcmd":   ScpCmd,
}

func (c *ClassType) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid class type: %w", err)
	}

	normalized := strings.ToLower(raw)
	val, ok := validClassTypes[normalized]
	if !ok {
		return fmt.Errorf("unsupported class type: %s", raw)
	}

	*c = val
	return nil
}

type ProcessStatus string

const (
	StatusRunning   ProcessStatus = "running"
	StatusCompleted ProcessStatus = "completed"
	StatusFailed    ProcessStatus = "failed"
	StatusStopped   ProcessStatus = "stopped"
)

type ProcessRun struct {
	ID         uuid.UUID         `json:"id"`
	Definition ProcessDefinition `json:"definition"`
	Status     ProcessStatus     `json:"status"`
	StartedAt  time.Time         `json:"started_at"`
	EndedAt    *time.Time        `json:"ended_at,omitempty"`
}

type ProcessLog struct {
	ID        int       `json:"id"`
	ProcessID uuid.UUID `json:"process_id"`
	Log       string    `json:"log"`
	CreatedAt time.Time `json:"created_at"`
}
