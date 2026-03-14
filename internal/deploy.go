package shipinternal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Options struct {
	ServerIP string
	User     string
}

var loadDeployConfig = LoadDeployConfig
var runLocalCommand = runLocalShellCommand
var waitForSSHClient = WaitForSSH
var copyDeployFile = CopyFile
var runRemoteCommands = RunCommands

func Run(ctx context.Context, opts Options) error {
	deployConfig, err := loadDeployConfig()
	if err != nil {
		return err
	}

	for _, cleanupPath := range deployConfig.ResolvedCleanupPaths(".") {
		defer os.Remove(cleanupPath)
	}

	for _, command := range deployConfig.LocalCommands {
		if err := runLocalCommand(ctx, command); err != nil {
			return err
		}
	}

	uploads, err := deployConfig.ResolvedUploads(".")
	if err != nil {
		return err
	}

	if len(uploads) == 0 && len(deployConfig.RemoteCommands) == 0 {
		return nil
	}

	client, err := waitForSSHClient(ctx, opts.User, opts.ServerIP, 10*time.Second)
	if err != nil {
		return err
	}
	defer client.Close()

	for _, upload := range uploads {
		if err := copyDeployFile(ctx, client, upload.Source, upload.Destination, upload.Mode); err != nil {
			return err
		}
	}

	if err := runRemoteCommands(ctx, client, deployConfig.RemoteCommands); err != nil {
		return err
	}

	return nil
}

func runLocalShellCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-lc", command)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", strings.TrimSpace(command), err)
	}
	return nil
}
