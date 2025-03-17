package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashmap-kz/rconf/internal/version"

	"github.com/hashmap-kz/rconf/internal/cmd"
	"github.com/hashmap-kz/rconf/internal/runner"
	"github.com/spf13/cobra"
)

func Execute() error {
	var cfg cmd.Config

	rootCmd := &cobra.Command{
		Use:     "rconf",
		Short:   "Execute local scripts on remote hosts via SSH",
		Version: version.Version,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runner.Run(&cfg)
		},
	}

	rootCmd.Flags().StringVarP(&cfg.PrivateKeyPath, "pkey", "i", "", "Path to SSH private key (required when pkey-auth is used)")
	rootCmd.Flags().StringVarP(&cfg.PrivateKeyPassphrase, "pkey-pass", "", "", "Passphrase to SSH private key (required when pkey is password-protected)")
	rootCmd.Flags().StringSliceVarP(&cfg.Filenames, "filename", "f", nil, "List of script paths or directories (required)")
	rootCmd.Flags().StringSliceVarP(&cfg.ConnStrings, "conn", "H", nil, strings.TrimSpace(`
List of remote hosts (required)
Format: username:password@host:port?key1=value1&key2=value2
- password is optional
- port is optional (default 22)
- query-opts are optional (available: sudo)
`))
	rootCmd.Flags().IntVarP(&cfg.WorkerLimit, "workers", "w", 2, "Max concurrent SSH connections")
	rootCmd.Flags().StringVarP(&cfg.LogFile, "log", "l", "rconf.log", "Log file path")
	rootCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "R", true, "Process the directory used in -f, --filename recursively")

	requiredFlags := []string{"filename", "conn"}
	for _, flag := range requiredFlags {
		if err := rootCmd.MarkFlagRequired(flag); err != nil {
			fmt.Printf("Failed to mark '%s' flag as required: %v", flag, err)
			os.Exit(1)
		}
	}

	return rootCmd.Execute()
}
