package rconf

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashmap-kz/rconf/internal/connstr"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClient wraps an SSH client and SFTP session.
type SSHClient struct {
	client *ssh.Client
	sftp   *sftp.Client
}

// NewSSHClient establishes an SSH and SFTP connection.
func NewSSHClient(connInfoPass connstr.ConnInfo, pkeyPath, pkeyPass string) (*SSHClient, error) {
	authMethods, err := getAuthsMethods(connInfoPass.Password, pkeyPath, pkeyPass)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: connInfoPass.User,
		Auth: authMethods,
		//nolint:gosec
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", connInfoPass.Host, connInfoPass.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &SSHClient{client: client, sftp: sftpClient}, nil
}

// Close closes SSH and SFTP connections.
func (s *SSHClient) Close() {
	s.sftp.Close()
	s.client.Close()
}

func isPasswordProtectedPrivateKey(key []byte) bool {
	_, err := ssh.ParsePrivateKey(key)
	if err != nil {
		if err.Error() == (&ssh.PassphraseMissingError{}).Error() {
			return true
		}
	}
	return false
}

func getSigner(key []byte, passphrase string) (ssh.Signer, error) {
	if isPasswordProtectedPrivateKey(key) {
		if passphrase == "" {
			return nil, &ssh.PassphraseMissingError{}
		}
		signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
		return signer, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	return signer, err
}

// getAuthsMethods collects authentication with password or private_key+optional(passphrase)
func getAuthsMethods(password, pkeyPath, pkeyPass string) ([]ssh.AuthMethod, error) {
	var auths []ssh.AuthMethod

	// should be password or private-key
	if strings.TrimSpace(password) == "" && strings.TrimSpace(pkeyPath) == "" {
		return nil, fmt.Errorf("both password and private-key-path are empty")
	}

	// password-based-auth

	if password != "" {
		auths = append(auths, ssh.Password(password))
		return auths, nil
	}

	// pkey-based-auth

	key, err := os.ReadFile(pkeyPath)
	if err != nil {
		return nil, err
	}
	signer, err := getSigner(key, pkeyPass)
	if err != nil {
		return nil, err
	}
	auths = append(auths, ssh.PublicKeys(signer))
	return auths, nil
}

// UploadScript uploads a script to the remote host from memory.
func (s *SSHClient) UploadScript(scriptContent []byte, remotePath string) error {
	dstFile, err := s.sftp.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote script: %w", err)
	}
	defer dstFile.Close()

	_, err = dstFile.Write(scriptContent)
	if err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	return nil
}

// ExecuteScript executes a script on the remote host.
func (s *SSHClient) ExecuteScript(remotePath string, opts map[string][]string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("sudo chmod +x %s && sudo %s", remotePath, remotePath)
	if hasOpt(opts, "sudo", "false") {
		cmd = fmt.Sprintf("chmod +x %s && %s", remotePath, remotePath)
	}

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("failed to execute script: %w", err)
	}

	return string(out), nil
}

// internal

func hasOpt(opts map[string][]string, k, v string) bool {
	if val, exists := opts[k]; exists {
		for _, elem := range val {
			if elem == v {
				return true
			}
		}
	}
	return false
}
