package shipinternal

import (
	"context"
	"fmt"
)

type CreateRequest struct {
	Name   string
	Region string
	Size   string
	Image  string
}

type Provider interface {
	CreateServer(ctx context.Context, req CreateRequest) (ServerState, error)
	DestroyServer(ctx context.Context, state ServerState) error
}

func New(name string) (Provider, error) {
	switch name {
	case "", "digitalocean":
		return NewDigitalOcean(), nil
	case "hetzner":
		return NewHetzner(), nil
	case "vultr":
		return NewVultr(), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", name)
	}
}
