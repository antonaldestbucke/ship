package shipinternal

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestRunLocalOnlyDeploySkipsSSH(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalRunLocalCommand := runLocalCommand
	originalWaitForSSHClient := waitForSSHClient
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		runLocalCommand = originalRunLocalCommand
		waitForSSHClient = originalWaitForSSHClient
	}()

	var ran []string
	loadDeployConfig = func() (DeployConfig, error) {
		return DeployConfig{
			LocalCommands: []string{
				"npm ci",
				"npm run build",
			},
		}, nil
	}
	runLocalCommand = func(ctx context.Context, command string) error {
		ran = append(ran, command)
		return nil
	}
	waitForSSHClient = func(ctx context.Context, user, host string, interval time.Duration) (*ssh.Client, error) {
		t.Fatal("waitForSSHClient should not be called for a local-only deploy")
		return nil, nil
	}

	err := Run(context.Background(), Options{
		ServerIP: "1.2.3.4",
		User:     "root",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(ran) != 2 {
		t.Fatalf("Run executed %d local commands, want 2", len(ran))
	}
	if ran[0] != "npm ci" || ran[1] != "npm run build" {
		t.Fatalf("Run executed local commands %+v", ran)
	}
}
