package runner

import (
	"os"
	"testing"

	"github.com/hashmap-kz/rconf/internal/cmd"
	"github.com/stretchr/testify/assert"
)

// TODO: this should be configured in a Makefile
const (
	integrationTestsAvailable = true
	sshPrivateKey             = "../../integration/id_ed25519"
)

func TestRunner(t *testing.T) {
	if !integrationTestsAvailable {
		return
	}

	// prepare pkey mod
	err := os.Chmod(sshPrivateKey, 0o600)
	assert.NoError(t, err)

	// init config
	config := cmd.Config{
		Filenames: []string{
			"../../integration/scripts/01-basic.sh",
		},
		ConnStrings: []string{
			"root@localhost:12222",
			"root@localhost:12223",
		},
		PrivateKeyPath: sshPrivateKey,
		WorkerLimit:    2,
	}

	Run(&config)
}
