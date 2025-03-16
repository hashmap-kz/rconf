package runner

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashmap-kz/go-texttable/pkg/table"
	rconf "github.com/hashmap-kz/rconf/internal/client"
	"github.com/hashmap-kz/rconf/internal/cmd"
	"github.com/hashmap-kz/rconf/internal/connstr"
	"github.com/hashmap-kz/rconf/internal/remote"
	"github.com/hashmap-kz/rconf/internal/resolver"
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

// ProcessHost handles script execution on a single host.
func ProcessHost(task *HostTask) {
	defer task.wg.Done()
	task.semaphore <- struct{}{}
	defer func() { <-task.semaphore }()

	fmt.Printf("[HOST: %s] 🔄 Connecting...\n", task.Host)
	client, err := rconf.NewSSHClient(connstr.ConnInfo{
		User:     task.User,
		Password: task.Password,
		Host:     task.Host,
		Port:     task.Port,
	}, task.PrivateKeyPath, task.PrivateKeyPassphrase)
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

	failedScripts := []string{}

	for script, content := range task.ScriptContents {
		remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(script))
		fmt.Printf("[HOST: %s] ⏳ Uploading %s...\n", task.Host, script)

		err := client.UploadScript(content, remotePath)
		if err != nil {
			slogger.Error("Failed to upload script",
				slog.String("host", task.Host),
				slog.String("script", script),
				slog.Any("error", err),
			)
			fmt.Printf("[HOST: %s] ❌ Upload failed for %s\n", task.Host, script)
			failedScripts = append(failedScripts, script)
			continue
		}

		fmt.Printf("[HOST: %s] 🚀 Executing %s...\n", task.Host, script)
		output, err := client.ExecuteScript(remotePath)
		if err != nil {
			slogger.Error("Execution failed",
				slog.String("host", task.Host),
				slog.String("script", script),
				slog.Any("error", err),
				slog.String("output", output),
			)
			fmt.Printf("[HOST: %s] ❌ Execution failed for %s\n", task.Host, script)
			failedScripts = append(failedScripts, script)
			continue
		}

		fmt.Printf("[HOST: %s] ✅ Successfully executed %s\n", task.Host, script)
	}

	if len(failedScripts) > 0 {
		task.Results.Store(task.Host, fmt.Sprintf("Failed: %s", strings.Join(failedScripts, ", ")))
	} else {
		task.Results.Store(task.Host, "Success")
	}
}

// Run executes scripts on multiple hosts with concurrency control.
func Run(cfg *cmd.Config) {
	InitLogger(cfg.LogFile)
	scriptContents, err := ReadScriptsIntoMemory(cfg.Filenames, cfg.Recursive)
	if err != nil {
		slogger.Error("Failed to read scripts", slog.Any("error", err))
		os.Exit(1)
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
			os.Exit(1)
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
		go ProcessHost(task)
	}

	wg.Wait()

	PrintSummary(results)
}

// PrintSummary prints the execution results in a well-formatted table using tabwriter.
func PrintSummary(results *sync.Map) {
	fmt.Println("\n=== Execution Summary ===")

	tbl := table.NewTextTable()
	tbl.DefineColumn("HOST", table.LEFT, table.LEFT)
	tbl.DefineColumn("RESULT", table.RIGHT, table.RIGHT)

	// Iterate over results and print each row
	results.Range(func(key, value interface{}) bool {
		tbl.InsertAllAndFinishRow(
			fmt.Sprintf("%v", key),
			fmt.Sprintf("%v", value),
		)
		return true
	})

	fmt.Println(tbl.Print())
}

// ReadScriptsIntoMemory reads all scripts (including from directories) before execution and stores their contents.
func ReadScriptsIntoMemory(scriptPaths []string, recursive bool) (map[string][]byte, error) {
	scriptContents := make(map[string][]byte)

	files, err := resolver.ResolveAllFiles(scriptPaths, recursive)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if resolver.IsURL(f) {
			data, err := remote.ReadRemoteFileContent(f)
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
