package executor

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/processconsumer/service/executor/sshutil"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SCPCmdExecutor struct{}

func NewSCPCmdExecutor() *SCPCmdExecutor {
	return &SCPCmdExecutor{}
}

func (e *SCPCmdExecutor) Run(ctx context.Context, task model.Task) error {
	local := task.Parameters["localPath"]
	remote := task.Parameters["remotePath"]
	if local == "" || remote == "" {
		return fmt.Errorf("missing 'localPath' or 'remotePath' in task %q", task.Name)
	}

	connCfg, err := sshutil.BuildConnectionConfig(task.Parameters)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", connCfg.Host, connCfg.Port)
	client, err := ssh.Dial("tcp", addr, connCfg.ClientConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	logger.GetLogger().Infof("About to transfer file to %s...", addr)

	if err := transferFile(client, local, remote, connCfg.Host, connCfg.Port); err != nil {
		return err
	}

	logger.GetLogger().Infof("Remote file copy for task %q succeeded", task.Name)
	return nil
}

func transferFile(client *ssh.Client, localPath, remotePath, host, port string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to start SFTP session: %w", err)
	}
	defer sftpClient.Close()

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	content, err := io.ReadAll(localFile)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	if _, err := remoteFile.Write(content); err != nil {
		return fmt.Errorf("failed to write to remote file: %w", err)
	}

	logger.GetLogger().Infof("File successfully copied to %s:%s as %s", host, port, remotePath)
	return nil
}
