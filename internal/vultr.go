package shipinternal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vultr/govultr/v3"
	"golang.org/x/oauth2"
)

type vultrProvider struct{}

func NewVultr() Provider {
	return vultrProvider{}
}

func (vultrProvider) CreateServer(ctx context.Context, req CreateRequest) (ServerState, error) {
	client, err := newVultrClientFromEnv(ctx)
	if err != nil {
		return ServerState{}, err
	}

	localKey, err := DiscoverLocalSSHKey()
	if err != nil {
		return ServerState{}, err
	}
	keyID, err := ensureVultrSSHKey(ctx, client, localKey)
	if err != nil {
		return ServerState{}, err
	}

	osID, err := findVultrOSID(ctx, client, firstNonEmpty(req.Image, "Ubuntu 22.04 x64"))
	if err != nil {
		return ServerState{}, err
	}

	instance, _, err := client.Instance.Create(ctx, &govultr.InstanceCreateReq{
		Region:   firstNonEmpty(req.Region, "ewr"),
		Plan:     firstNonEmpty(req.Size, "vc2-2c-4gb"),
		Label:    req.Name,
		Hostname: req.Name,
		OsID:     osID,
		SSHKeys:  []string{keyID},
	})
	if err != nil {
		return ServerState{}, fmt.Errorf("create Vultr instance: %w", err)
	}

	ip, err := waitForVultrMainIP(ctx, client, instance.ID, 15*time.Second)
	if err != nil {
		return ServerState{}, err
	}

	return ServerState{
		Provider: "vultr",
		ServerID: instance.ID,
		IP:       ip,
		SSHUser:  "root",
	}, nil
}

func ensureVultrSSHKey(ctx context.Context, client *govultr.Client, localKey LocalSSHKey) (string, error) {
	keys, err := listAllVultrSSHKeys(ctx, client)
	if err != nil {
		return "", fmt.Errorf("list Vultr SSH keys: %w", err)
	}
	for _, key := range keys {
		if strings.TrimSpace(key.SSHKey) == localKey.PublicKey {
			return key.ID, nil
		}
	}

	key, _, err := client.SSHKey.Create(ctx, &govultr.SSHKeyReq{
		Name:   "ship-" + localKey.Name,
		SSHKey: localKey.PublicKey,
	})
	if err != nil {
		return "", fmt.Errorf("create Vultr SSH key: %w", err)
	}
	return key.ID, nil
}

func (vultrProvider) DestroyServer(ctx context.Context, state ServerState) error {
	client, err := newVultrClientFromEnv(ctx)
	if err != nil {
		return err
	}
	if err := client.Instance.Delete(ctx, state.ServerID); err != nil {
		return fmt.Errorf("delete Vultr instance: %w", err)
	}
	return nil
}

func newVultrClientFromEnv(ctx context.Context) (*govultr.Client, error) {
	token := os.Getenv("VULTR_API_KEY")
	if token == "" {
		return nil, errors.New("VULTR_API_KEY is not set")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, tokenSource)
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return govultr.NewClient(httpClient), nil
}

func listAllVultrSSHKeys(ctx context.Context, client *govultr.Client) ([]govultr.SSHKey, error) {
	opts := &govultr.ListOptions{PerPage: 500}
	var keys []govultr.SSHKey
	for {
		page, meta, _, err := client.SSHKey.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		keys = append(keys, page...)
		if meta == nil || meta.Links == nil || meta.Links.Next == "" {
			return keys, nil
		}
		opts.Cursor = meta.Links.Next
	}
}

func findVultrOSID(ctx context.Context, client *govultr.Client, imageName string) (int, error) {
	opts := &govultr.ListOptions{PerPage: 500}
	for {
		oses, meta, _, err := client.OS.List(ctx, opts)
		if err != nil {
			return 0, fmt.Errorf("list Vultr operating systems: %w", err)
		}
		for _, osEntry := range oses {
			if strings.EqualFold(osEntry.Name, imageName) {
				return osEntry.ID, nil
			}
		}
		for _, osEntry := range oses {
			if strings.Contains(strings.ToLower(osEntry.Name), strings.ToLower(imageName)) {
				return osEntry.ID, nil
			}
		}
		if meta == nil || meta.Links == nil || meta.Links.Next == "" {
			break
		}
		opts.Cursor = meta.Links.Next
	}
	return 0, fmt.Errorf("Vultr image %q not found", imageName)
}

func waitForVultrMainIP(ctx context.Context, client *govultr.Client, instanceID string, interval time.Duration) (string, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		instance, _, err := client.Instance.Get(ctx, instanceID)
		if err != nil {
			return "", fmt.Errorf("get Vultr instance: %w", err)
		}
		if instance != nil && instance.MainIP != "" && strings.EqualFold(instance.Status, "active") {
			return instance.MainIP, nil
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("wait for Vultr instance to become active: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
