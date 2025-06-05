package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SCPCmdExecutor struct{}

func NewSCPCmdExecutor() *SCPCmdExecutor {
	return &SCPCmdExecutor{}
}

func (e *SCPCmdExecutor) Run(ctx context.Context, task model.Task) error {
	host := task.Parameters["host"]
	port := task.Parameters["port"]
	user := task.Parameters["user"]
	password := task.Parameters["password"]
	keyPath := task.Parameters["keyPath"]
	localPath := task.Parameters["localPath"]
	remotePath := task.Parameters["remotePath"]

	if host == "" || port == "" || user == "" || localPath == "" || remotePath == "" {
		return fmt.Errorf("scpCmd missing one or more required parameters")
	}

	var auth ssh.AuthMethod
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	} else if password != "" {
		auth = ssh.Password(password)
	} else {
		return fmt.Errorf("either password or keyPath must be provided")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For local testing it is okay!
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("failed to start SFTP session: %w", err)
	}
	defer sftpClient.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	bytesWritten, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	_, err = remoteFile.Write(bytesWritten)
	if err != nil {
		return fmt.Errorf("failed to write to remote file: %w", err)
	}

	logger.GetLogger().Infof("File successfully copied to %s:%s as %s", host, port, remotePath)
	return nil
}
