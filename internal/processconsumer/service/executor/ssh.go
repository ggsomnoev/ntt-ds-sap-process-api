package executor

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/executor/sshutil"

	"golang.org/x/crypto/ssh"
)

type SSHCmdExecutor struct{}

func NewSSHCmdExecutor() *SSHCmdExecutor {
	return &SSHCmdExecutor{}
}

func (se *SSHCmdExecutor) Run(ctx context.Context, task model.Task) error {
	command := task.Parameters["command"]
	if command == "" {
		return fmt.Errorf("missing 'command' parameter in task %q", task.Name)
	}

	connCfg, err := sshutil.BuildConnectionConfig(task.Parameters)
	if err != nil {
		return fmt.Errorf("failed to build SSH config: %w", err)
	}

	addr := fmt.Sprintf("%s:%s", connCfg.Host, connCfg.Port)
	client, err := ssh.Dial("tcp", addr, connCfg.ClientConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.GetLogger().Warnf("failed to close SSH client: %v", err)
		}
	}()

	logger.GetLogger().Infof("Executing SSH command on %s: %q", addr, command)

	if err := runCommand(ctx, client, task.Name, command); err != nil {
		return err
	}

	logger.GetLogger().Infof("SSH command for task %q succeeded", task.Name)
	return nil
}

func runCommand(ctx context.Context, client *ssh.Client, taskName, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			logger.GetLogger().Warnf("failed to close session: %v", err)
		}
	}()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return ctx.Err()
	case err = <-done:
	}

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if err != nil {
		return fmt.Errorf("SSH command failed for task %q: %w. Stderr: %s", taskName, err, stderr)
	}

	if stdout != "" {
		logger.GetLogger().Infof("Output for task %q:\n%s", taskName, stdout)
	}
	return nil
}
