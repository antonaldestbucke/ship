package shipinternal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/actionutil"
)

type hetznerProvider struct{}

func NewHetzner() Provider {
	return hetznerProvider{}
}

func (hetznerProvider) CreateServer(ctx context.Context, req CreateRequest) (ServerState, error) {
	client, err := newHetznerClientFromEnv()
	if err != nil {
		return ServerState{}, err
	}

	locationName := firstNonEmpty(req.Region, "nbg1")
	serverTypeName := firstNonEmpty(req.Size, "cx22")
	imageName := firstNonEmpty(req.Image, "ubuntu-22.04")

	location, _, err := client.Location.GetByName(ctx, locationName)
	if err != nil {
		return ServerState{}, fmt.Errorf("get Hetzner location %q: %w", locationName, err)
	}
	if location == nil {
		return ServerState{}, fmt.Errorf("Hetzner location %q not found", locationName)
	}

	serverType, _, err := client.ServerType.GetByName(ctx, serverTypeName)
	if err != nil {
		return ServerState{}, fmt.Errorf("get Hetzner server type %q: %w", serverTypeName, err)
	}
	if serverType == nil {
		return ServerState{}, fmt.Errorf("Hetzner server type %q not found", serverTypeName)
	}

	image, _, err := client.Image.GetByName(ctx, imageName)
	if err != nil {
		return ServerState{}, fmt.Errorf("get Hetzner image %q: %w", imageName, err)
	}
	if image == nil {
		return ServerState{}, fmt.Errorf("Hetzner image %q not found", imageName)
	}

	localKey, err := DiscoverLocalSSHKey()
	if err != nil {
		return ServerState{}, err
	}
	sshKey, err := ensureHetznerSSHKey(ctx, client, localKey)
	if err != nil {
		return ServerState{}, err
	}

	createResult, _, err := client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name:       req.Name,
		ServerType: serverType,
		Image:      image,
		Location:   location,
		SSHKeys:    []*hcloud.SSHKey{sshKey},
	})
	if err != nil {
		return ServerState{}, fmt.Errorf("create Hetzner server: %w", err)
	}

	if err := client.Action.WaitFor(ctx, actionutil.AppendNext(createResult.Action, createResult.NextActions)...); err != nil {
		return ServerState{}, fmt.Errorf("wait for Hetzner server create action: %w", err)
	}

	ip, err := waitForHetznerIPv4(ctx, client, createResult.Server.ID, 10*time.Second)
	if err != nil {
		return ServerState{}, err
	}

	return ServerState{
		Provider: "hetzner",
		ServerID: fmt.Sprintf("%d", createResult.Server.ID),
		IP:       ip,
		SSHUser:  "root",
	}, nil
}

func ensureHetznerSSHKey(ctx context.Context, client *hcloud.Client, localKey LocalSSHKey) (*hcloud.SSHKey, error) {
	keys, err := client.SSHKey.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list Hetzner SSH keys: %w", err)
	}
	for _, key := range keys {
		if key.Fingerprint == localKey.FingerprintMD5 || strings.TrimSpace(key.PublicKey) == localKey.PublicKey {
			return key, nil
		}
	}

	key, _, err := client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      "ship-" + localKey.Name,
		PublicKey: localKey.PublicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create Hetzner SSH key: %w", err)
	}
	return key, nil
}

func (hetznerProvider) DestroyServer(ctx context.Context, state ServerState) error {
	client, err := newHetznerClientFromEnv()
	if err != nil {
		return err
	}

	var serverID int64
	if _, err := fmt.Sscanf(state.ServerID, "%d", &serverID); err != nil {
		return fmt.Errorf("invalid Hetzner server_id %q: %w", state.ServerID, err)
	}

	server, _, err := client.Server.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("get Hetzner server: %w", err)
	}
	if server == nil {
		return nil
	}

	result, _, err := client.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return fmt.Errorf("delete Hetzner server: %w", err)
	}
	if result != nil && result.Action != nil {
		if err := client.Action.WaitFor(ctx, result.Action); err != nil {
			return fmt.Errorf("wait for Hetzner delete action: %w", err)
		}
	}
	return nil
}

func newHetznerClientFromEnv() (*hcloud.Client, error) {
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		return nil, errors.New("HCLOUD_TOKEN is not set")
	}
	return hcloud.NewClient(hcloud.WithToken(token), hcloud.WithApplication("ship", "0.1.0")), nil
}

func waitForHetznerIPv4(ctx context.Context, client *hcloud.Client, serverID int64, interval time.Duration) (string, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		server, _, err := client.Server.GetByID(ctx, serverID)
		if err != nil {
			return "", fmt.Errorf("get Hetzner server: %w", err)
		}
		if server != nil && !server.PublicNet.IPv4.IsUnspecified() {
			return server.PublicNet.IPv4.IP.String(), nil
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait for Hetzner server IPv4: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
