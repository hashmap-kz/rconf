package runner

import (
	"os"
	"testing"

	"github.com/hashmap-kz/rconf/internal/cmd"
	"github.com/stretchr/testify/assert"
)

const (
	sshPrivateKey       = "../../integration/id_ed25519"
	integrationTestEnv  = "RCONF_INTEGRATION_TESTS_AVAILABLE"
	integrationTestFlag = "0xcafebabe"
)

func TestRunner(t *testing.T) {
	if os.Getenv(integrationTestEnv) != integrationTestFlag {
		t.Log("integration test was skipped due to configuration")
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
