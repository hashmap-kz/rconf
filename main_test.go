package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"

	"github.com/ory/dockertest/v3/docker"
)

const testScript = `#!/bin/sh
echo "Hello, SSH!"`

// TODO: this should be configured in a Makefile
const integrationTestsAvailable = false

func TestSSHOperations(t *testing.T) {
	if !integrationTestsAvailable {
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "rastasheep/ubuntu-sshd",
		Tag:          "18.04",
		Env:          []string{"ROOT_PASSWORD=root"},
		ExposedPorts: []string{"22"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"22/tcp": {{HostIP: "0.0.0.0", HostPort: "2222"}}, // Change "2222" to any available port
		},
		Mounts: []string{
			fmt.Sprintf("%s:/root/.ssh/id_ed25519", "/root/.ssh/id_ed25519"),
			fmt.Sprintf("%s:/root/.ssh/authorized_keys", "/root/.ssh/id_ed25519.pub"),
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	defer pool.Purge(resource)

	var sshClient *SSHClient
	time.Sleep(3 * time.Second) // Give the container some time to start

	hostPort := resource.GetPort("22/tcp")

	t.Run("Test Script Upload", func(t *testing.T) {
		sshClient, err = NewSSHClient("root", "localhost", hostPort, "/root/.ssh/id_ed25519")
		if err != nil {
			t.Fatalf("Failed to establish SSH connection: %s", err)
		}
		defer sshClient.Close()

		remotePath := "/tmp/test_script.sh"
		err := sshClient.UploadScript([]byte(testScript), remotePath)
		if err != nil {
			t.Fatalf("Failed to upload script: %s", err)
		}

		output, err := sshClient.ExecuteScript(remotePath)
		if err != nil {
			t.Fatalf("Failed to execute script: %s", err)
		}
		if output != "Hello, SSH!\n" {
			t.Fatalf("Unexpected script output: %s", output)
		}
	})
}
