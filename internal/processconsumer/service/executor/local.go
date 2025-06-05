package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
)

type LocalCmdExecutor struct {
}

func NewLocalCmdService() *LocalCmdExecutor {
	return &LocalCmdExecutor{}
}

func (le *LocalCmdExecutor) Run(ctx context.Context, task model.Task) error {
	cmdStr, ok := task.Parameters["command"]
	if !ok || strings.TrimSpace(cmdStr) == "" {
		return fmt.Errorf("missing 'command' parameter in task %q", task.Name)
	}

	logger.GetLogger().Infof("Executing local command: %q", cmdStr)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	output := stdoutBuf.String()
	errOutput := stderrBuf.String()

	if err != nil {
		return fmt.Errorf("local command execution failed for task %q: %w. Error output: %s", task.Name, err, errOutput)
	}

	logger.GetLogger().Infof("Local command for task %q executed successfully. Output: %s.", task.Name, output)
	return nil
}
