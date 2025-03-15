package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// Config holds SSH execution details.
type Config struct {
	User        string
	PrivateKey  string
	Scripts     []string
	Hosts       []string
	WorkerLimit int
	LogFile     string
}

// Structured logger
var slogger *slog.Logger

// HostTask encapsulates all information needed to process a host.
type HostTask struct {
	User           string
	Host           string
	PrivateKey     string
	ScriptContents map[string][]byte
	Results        *sync.Map
	WG             *sync.WaitGroup
	Semaphore      chan struct{}
}

// InitLogger initializes structured logging with slog.
func InitLogger(logFile string) {
	file, err := os.OpenFile(logFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		os.Exit(1)
	}
	writer := io.MultiWriter(file)
	slogger = slog.New(slog.NewTextHandler(writer, nil))
}

// SSHClient wraps an SSH client and SFTP session.
type SSHClient struct {
	client *ssh.Client
	sftp   *sftp.Client
}

// NewSSHClient establishes an SSH and SFTP connection.
func NewSSHClient(user, host, privateKeyPath string) (*SSHClient, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host+":22", config)
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
func (s *SSHClient) ExecuteScript(remotePath string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	out, err := session.CombinedOutput(fmt.Sprintf("sudo chmod +x %s && sudo %s", remotePath, remotePath))
	if err != nil {
		return string(out), fmt.Errorf("failed to execute script: %w", err)
	}

	return string(out), nil
}

// ProcessHost handles script execution on a single host.
func ProcessHost(task *HostTask) {
	defer task.WG.Done()
	task.Semaphore <- struct{}{}
	defer func() { <-task.Semaphore }()

	fmt.Printf("[HOST: %s] 🔄 Connecting...\n", task.Host)
	client, err := NewSSHClient(task.User, task.Host, task.PrivateKey)
	if err != nil {
		slogger.Error("SSH connection failed", slog.String("host", task.Host), slog.Any("error", err))
		fmt.Printf("[HOST: %s] ❌ SSH connection failed\n", task.Host)
		task.Results.Store(task.Host, "SSH Failed")
		return
	}
	defer func() {
		fmt.Printf("[HOST: %s] 🔄 Disconnecting...\n", task.Host)
		client.Close()
	}()

	successScripts := []string{}
	failedScripts := []string{}

	for script, content := range task.ScriptContents {
		remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(script))
		fmt.Printf("[HOST: %s] ⏳ Uploading %s...\n", task.Host, script)

		err := client.UploadScript(content, remotePath)
		if err != nil {
			slogger.Error("Failed to upload script", slog.String("host", task.Host), slog.String("script", script), slog.Any("error", err))
			fmt.Printf("[HOST: %s] ❌ Upload failed for %s\n", task.Host, script)
			failedScripts = append(failedScripts, script)
			continue
		}

		fmt.Printf("[HOST: %s] 🚀 Executing %s...\n", task.Host, script)
		output, err := client.ExecuteScript(remotePath)
		if err != nil {
			slogger.Error("Execution failed", slog.String("host", task.Host), slog.String("script", script), slog.Any("error", err), slog.String("output", output))
			fmt.Printf("[HOST: %s] ❌ Execution failed for %s\n", task.Host, script)
			failedScripts = append(failedScripts, script)
			continue
		}

		fmt.Printf("[HOST: %s] ✅ Successfully executed %s\n", task.Host, script)
		successScripts = append(successScripts, script)
	}

	if len(failedScripts) > 0 {
		task.Results.Store(task.Host, fmt.Sprintf("Failed: %s", strings.Join(failedScripts, ", ")))
	} else {
		task.Results.Store(task.Host, "Success")
	}
}

// Run executes scripts on multiple hosts with concurrency control.
func Run(cfg Config) {
	InitLogger(cfg.LogFile)
	scriptContents, err := ReadScriptsIntoMemory(cfg.Scripts)
	if err != nil {
		slogger.Error("Failed to read scripts", slog.Any("error", err))
		os.Exit(1)
	}

	fmt.Println("\n🚀 Starting script execution...\n")

	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.WorkerLimit)
	results := &sync.Map{}

	for _, host := range cfg.Hosts {
		wg.Add(1)
		task := &HostTask{
			User:           cfg.User,
			Host:           host,
			PrivateKey:     cfg.PrivateKey,
			ScriptContents: scriptContents,
			Results:        results,
			WG:             &wg,
			Semaphore:      sem,
		}
		go ProcessHost(task)
	}

	wg.Wait()

	PrintSummary(results)
}

// PrintSummary prints the execution results in a simple table format.
func PrintSummary(results *sync.Map) {
	fmt.Println("\n=== Execution Summary ===")
	fmt.Println("+----------------+----------------------+")
	fmt.Println("| HOST           | EXECUTION RESULT     |")
	fmt.Println("+----------------+----------------------+")
	results.Range(func(key, value interface{}) bool {
		fmt.Printf("| %-14s | %-20s |\n", key, value)
		return true
	})
	fmt.Println("+----------------+----------------------+")
}

// ReadScriptsIntoMemory reads all scripts (including from directories) before execution and stores their contents.
func ReadScriptsIntoMemory(scriptPaths []string) (map[string][]byte, error) {
	scriptContents := make(map[string][]byte)

	for _, path := range scriptPaths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			err := filepath.WalkDir(path, func(subPath string, d os.DirEntry, err error) error {
				if !d.IsDir() && strings.HasSuffix(d.Name(), ".sh") {
					content, err := os.ReadFile(subPath)
					if err == nil {
						scriptContents[subPath] = content
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			content, err := os.ReadFile(path)
			if err == nil {
				scriptContents[path] = content
			}
		}
	}
	return scriptContents, nil
}

// main function using Cobra CLI
func main() {
	var cfg Config

	rootCmd := &cobra.Command{
		Use:   "ssh-executor",
		Short: "Execute local scripts on remote hosts via SSH",
		Run: func(cmd *cobra.Command, args []string) {
			Run(cfg)
		},
	}

	rootCmd.Flags().StringVarP(&cfg.User, "user", "u", "", "SSH user (required)")
	rootCmd.Flags().StringVarP(&cfg.PrivateKey, "key", "k", "", "Path to SSH private key (required)")
	rootCmd.Flags().StringSliceVarP(&cfg.Scripts, "scripts", "s", nil, "List of script paths or directories (required)")
	rootCmd.Flags().StringSliceVarP(&cfg.Hosts, "hosts", "H", nil, "List of remote hosts (required)")
	rootCmd.Flags().IntVarP(&cfg.WorkerLimit, "workers", "w", 2, "Max concurrent SSH connections")
	rootCmd.Flags().StringVarP(&cfg.LogFile, "log", "l", "ssh_execution.log", "Log file path")

	_ = rootCmd.MarkFlagRequired("user")
	_ = rootCmd.MarkFlagRequired("key")
	_ = rootCmd.MarkFlagRequired("scripts")
	_ = rootCmd.MarkFlagRequired("hosts")

	_ = rootCmd.Execute()
}
