package model

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
	Class      string            `yaml:"class" json:"class"`
	Parameters map[string]string `yaml:"parameters" json:"parameters"`
	WaitFor    []string          `yaml:"waitfor,omitempty" json:"waitfor,omitempty"`
}
