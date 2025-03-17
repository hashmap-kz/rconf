package cmd

// Config holds SSH execution details.
type Config struct {
	Filenames            []string
	ConnStrings          []string // ssh://user:pass@host:port
	PrivateKeyPath       string
	PrivateKeyPassphrase string
	WorkerLimit          int
	LogFile              string
	Recursive            bool
	Version              string
}
