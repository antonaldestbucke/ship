package shipinternal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type digitalOceanProvider struct{}

func NewDigitalOcean() Provider {
	return digitalOceanProvider{}
}

func (digitalOceanProvider) CreateServer(ctx context.Context, req CreateRequest) (ServerState, error) {
	client, err := newDigitalOceanClientFromEnv()
	if err != nil {
		return ServerState{}, err
	}

	keys, _, err := client.Keys.List(ctx, &godo.ListOptions{PerPage: 200})
	if err != nil {
		return ServerState{}, fmt.Errorf("list DigitalOcean SSH keys: %w", err)
	}
	if len(keys) == 0 {
		return ServerState{}, errors.New("no SSH keys found in DigitalOcean account")
	}

	sshKeys := make([]godo.DropletCreateSSHKey, 0, len(keys))
	for _, key := range keys {
		sshKeys = append(sshKeys, godo.DropletCreateSSHKey{ID: key.ID})
	}

	droplet, _, err := client.Droplets.Create(ctx, &godo.DropletCreateRequest{
		Name:   req.Name,
		Region: firstNonEmpty(req.Region, "nyc3"),
		Size:   firstNonEmpty(req.Size, "s-2vcpu-4gb"),
		Image: godo.DropletCreateImage{
			Slug: firstNonEmpty(req.Image, "ubuntu-22-04-x64"),
		},
		SSHKeys: sshKeys,
	})
	if err != nil {
		return ServerState{}, fmt.Errorf("create DigitalOcean droplet: %w", err)
	}

	ip, err := waitForDigitalOceanIP(ctx, client, droplet.ID, 15*time.Second)
	if err != nil {
		return ServerState{}, err
	}

	return ServerState{
		Provider: "digitalocean",
		ServerID: fmt.Sprintf("%d", droplet.ID),
		IP:       ip,
		SSHUser:  "root",
	}, nil
}

func (digitalOceanProvider) DestroyServer(ctx context.Context, state ServerState) error {
	client, err := newDigitalOceanClientFromEnv()
	if err != nil {
		return err
	}

	var dropletID int
	if _, err := fmt.Sscanf(state.ServerID, "%d", &dropletID); err != nil {
		return fmt.Errorf("invalid DigitalOcean server_id %q: %w", state.ServerID, err)
	}

	if _, err := client.Droplets.Delete(ctx, dropletID); err != nil {
		return fmt.Errorf("delete DigitalOcean droplet: %w", err)
	}
	return nil
}

func newDigitalOceanClientFromEnv() (*godo.Client, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, errors.New("DIGITALOCEAN_TOKEN is not set")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return godo.NewClient(oauth2.NewClient(context.Background(), tokenSource)), nil
}

func waitForDigitalOceanIP(ctx context.Context, client *godo.Client, dropletID int, interval time.Duration) (string, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		droplet, _, err := client.Droplets.Get(ctx, dropletID)
		if err != nil {
			return "", fmt.Errorf("get DigitalOcean droplet: %w", err)
		}

		if droplet.Status == "active" {
			for _, network := range droplet.Networks.V4 {
				if network.Type == "public" && net.ParseIP(network.IPAddress) != nil {
					return network.IPAddress, nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait for DigitalOcean droplet to become active: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
