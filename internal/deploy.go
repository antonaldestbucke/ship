package shipinternal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	ServerIP string
	User     string
}

func Run(ctx context.Context, opts Options) error {
	if err := runLocalCommand(ctx, "docker", "build", "-t", "app", "."); err != nil {
		return err
	}

	archivePath := filepath.Join(".", "app.tar")
	if err := runLocalCommand(ctx, "docker", "save", "app", "-o", archivePath); err != nil {
		return err
	}
	defer os.Remove(archivePath)

	client, err := WaitForSSH(ctx, opts.User, opts.ServerIP, 10*time.Second)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := CopyFile(ctx, client, archivePath, "/root/app.tar", 0o644); err != nil {
		return err
	}

	if err := RunCommands(ctx, client, []string{
		"docker load -i /root/app.tar",
		"docker stop app || true",
		"docker rm app || true",
		"docker run -d --name app -p 80:80 app",
	}); err != nil {
		return err
	}

	return nil
}

func runLocalCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", strings.Join(append([]string{name}, args...), " "), err)
	}
	return nil
}
