package sshutil

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConnectionConfig struct {
	Host         string
	Port         string
	ClientConfig *ssh.ClientConfig
}

func BuildConnectionConfig(params map[string]string) (*SSHConnectionConfig, error) {
	host := params["host"]
	port := params["port"]
	user := params["user"]
	password := params["password"]
	keyPath := params["keyPath"]

	if host == "" || user == "" {
		return nil, fmt.Errorf("host and user are required for SSH connection")
	}
	if port == "" {
		port = "22"
	}

	auth, err := getSSHAuthMethod(password, keyPath)
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // for local testing it is okay
		Timeout:         10 * time.Second,
	}

	return &SSHConnectionConfig{
		Host:         host,
		Port:         port,
		ClientConfig: clientConfig,
	}, nil
}

func getSSHAuthMethod(password, keyPath string) (ssh.AuthMethod, error) {
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		return ssh.PublicKeys(signer), nil
	}
	if password != "" {
		return ssh.Password(password), nil
	}
	return nil, fmt.Errorf("either password or keyPath must be provided")
}
