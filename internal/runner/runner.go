package runner

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashmap-kz/rconf/internal/cmd"
	"github.com/hashmap-kz/rconf/internal/connstr"
	"github.com/hashmap-kz/rconf/internal/resolver"
	rconf "github.com/hashmap-kz/rconf/internal/sshclient"
)

// Structured logger
var slogger *slog.Logger

// HostTask encapsulates all information needed to process a host.
type HostTask struct {
	User                 string
	Password             string
	Host                 string
	Port                 string
	PrivateKeyPath       string
	PrivateKeyPassphrase string
	ScriptContents       map[string][]byte
	Results              *sync.Map

	wg        *sync.WaitGroup
	semaphore chan struct{}
}

// Run executes scripts on multiple hosts with concurrency control.
func Run(cfg *cmd.Config) error {
	checkConfigDefaults(cfg)
	initLogger(cfg.LogFile)

	scriptContents, err := readScriptsIntoMemory(cfg.Filenames, cfg.Recursive)
	if err != nil {
		slogger.Error("Failed to read scripts", slog.Any("error", err))
		return err
	}

	fmt.Println("\n🚀 Starting script execution...")

	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.WorkerLimit)
	results := &sync.Map{}

	// prepare tasks

	tasks := make([]*HostTask, 0, len(cfg.ConnStrings))
	for _, connStr := range cfg.ConnStrings {
		connInfo, err := connstr.ParseConnectionString(connStr)
		if err != nil {
			slogger.Error("Failed to read conninfo", slog.Any("error", err))
			return err
		}
		task := &HostTask{
			User:                 connInfo.User,
			Password:             connInfo.Password,
			Host:                 connInfo.Host,
			Port:                 connInfo.Port,
			PrivateKeyPath:       cfg.PrivateKeyPath,
			PrivateKeyPassphrase: cfg.PrivateKeyPassphrase,
			ScriptContents:       scriptContents,
			Results:              results,
			wg:                   &wg,
			semaphore:            sem,
		}
		tasks = append(tasks, task)
	}

	// run tasks

	for _, task := range tasks {
		wg.Add(1)
		go processHost(task)
	}

	wg.Wait()

	printSummary(results)
	return nil
}

// checkConfigDefaults checks and sets default values when they're empty
func checkConfigDefaults(cfg *cmd.Config) {
	if cfg.LogFile == "" {
		cfg.LogFile = "rconf.log"
	}
	if cfg.WorkerLimit <= 0 {
		cfg.WorkerLimit = 2
	}
}

// initLogger initializes structured logging with slog.
func initLogger(logFile string) {
	file, err := os.OpenFile(logFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		os.Exit(1)
	}
	writer := io.MultiWriter(file)
	slogger = slog.New(slog.NewTextHandler(writer, nil))
}

// processHost handles script execution on a single host.
func processHost(task *HostTask) {
	defer task.wg.Done()
	task.semaphore <- struct{}{}
	defer func() { <-task.semaphore }()

	hostInfoLog := fmt.Sprintf("%s:%s", task.Host, task.Port)

	fmt.Printf("[HOST: %s] 🔄 Connecting...\n", hostInfoLog)
	client, err := rconf.NewSSHClient(connstr.ConnInfo{
		User:     task.User,
		Password: task.Password,
		Host:     task.Host,
		Port:     task.Port,
	}, task.PrivateKeyPath, task.PrivateKeyPassphrase)
	if err != nil {
		slogger.Error("SSH connection failed", slog.String("host", hostInfoLog), slog.Any("error", err))
		fmt.Printf("[HOST: %s] ❌ SSH connection failed\n", hostInfoLog)
		task.Results.Store(hostInfoLog, "SSH Failed")
		return
	}
	defer func() {
		fmt.Printf("[HOST: %s] 🔄 Disconnecting...\n", hostInfoLog)
		client.Close()
	}()

	failedScripts := []string{}

	for script, content := range task.ScriptContents {
		remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(script))
		fmt.Printf("[HOST: %s] ⏳ Uploading %s...\n", hostInfoLog, script)

		err := client.UploadScript(content, remotePath)
		if err != nil {
			slogger.Error("Failed to upload script",
				slog.String("host", hostInfoLog),
				slog.String("script", script),
				slog.Any("error", err),
			)
			fmt.Printf("[HOST: %s] ❌ Upload failed for %s\n", hostInfoLog, script)
			failedScripts = append(failedScripts, script)
			continue
		}

		fmt.Printf("[HOST: %s] 🚀 Executing %s...\n", hostInfoLog, script)
		output, err := client.ExecuteScript(remotePath)
		if err != nil {
			slogger.Error("Execution failed",
				slog.String("host", hostInfoLog),
				slog.String("script", script),
				slog.Any("error", err),
				slog.String("output", output),
			)
			fmt.Printf("[HOST: %s] ❌ Execution failed for %s\n", hostInfoLog, script)
			failedScripts = append(failedScripts, script)
			continue
		}

		fmt.Printf("[HOST: %s] ✅ Successfully executed %s\n", hostInfoLog, script)
	}

	if len(failedScripts) > 0 {
		task.Results.Store(hostInfoLog, fmt.Sprintf("❌ Failed: %s", strings.Join(failedScripts, ", ")))
	} else {
		task.Results.Store(hostInfoLog, "✅ Success")
	}
}

// printSummary prints the execution results in a well-formatted table using tabwriter.
func printSummary(results *sync.Map) {
	fmt.Println("\n=== Execution Summary ===")

	// Iterate over results and print each row
	results.Range(func(key, value interface{}) bool {
		fmt.Printf("%v %v\n", key, value)
		return true
	})
}

// readScriptsIntoMemory reads all scripts (including from directories) before execution and stores their contents.
func readScriptsIntoMemory(scriptPaths []string, recursive bool) (map[string][]byte, error) {
	scriptContents := make(map[string][]byte)

	files, err := resolver.ResolveAllFiles(scriptPaths, recursive)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if resolver.IsURL(f) {
			data, err := resolver.ReadRemoteFileContent(f)
			if err != nil {
				return nil, err
			}
			scriptContents[f] = data
		} else {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			scriptContents[f] = data
		}
	}

	return scriptContents, nil
}
