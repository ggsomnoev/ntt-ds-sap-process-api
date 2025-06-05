package executor

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"

	"golang.org/x/crypto/ssh"
)

type SSHCmdExecutor struct{}

func NewSSHCmdExecutor() *SSHCmdExecutor {
	return &SSHCmdExecutor{}
}

func (se *SSHCmdExecutor) Run(ctx context.Context, task model.Task) error {
	cmdStr, ok := task.Parameters["command"]
	if !ok || strings.TrimSpace(cmdStr) == "" {
		return fmt.Errorf("missing 'command' parameter in task %q", task.Name)
	}

	host := task.Parameters["ssh_host"]
	user := task.Parameters["ssh_user"]
	password := task.Parameters["ssh_password"]
	port := task.Parameters["ssh_port"]
	if port == "" {
		port = "22"
	}

	if host == "" || user == "" || password == "" {
		return fmt.Errorf("missing SSH credentials in task %q", task.Name)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For local testing it is okay!
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	logger.GetLogger().Infof("Executing SSH command on %s: %q", addr, cmdStr)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmdStr)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return ctx.Err()
	case err = <-done:
	}

	output := stdoutBuf.String()
	errOutput := stderrBuf.String()

	if err != nil {
		return fmt.Errorf("SSH command failed for task %q: %w. Stderr: %s", task.Name, err, errOutput)
	}

	logger.GetLogger().Infof("SSH command for task %q succeeded. Output: %s", task.Name, output)
	return nil
}
