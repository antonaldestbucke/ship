package shipinternal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const shipDir = ".ship"
const serverFile = "server.json"

type ServerState struct {
	Provider string `json:"provider,omitempty"`
	ServerID string `json:"server_id"`
	IP       string `json:"ip"`
	SSHUser  string `json:"ssh_user,omitempty"`
}

func serverStatePath() string {
	return filepath.Join(shipDir, serverFile)
}

func SaveServerState(state ServerState) error {
	if err := os.MkdirAll(shipDir, 0o755); err != nil {
		return fmt.Errorf("create .ship directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal server state: %w", err)
	}

	if err := os.WriteFile(serverStatePath(), data, 0o600); err != nil {
		return fmt.Errorf("write .ship/server.json: %w", err)
	}
	return nil
}

func LoadServerState() (ServerState, error) {
	data, err := os.ReadFile(serverStatePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ServerState{}, errors.New("missing .ship/server.json; create a server first")
		}
		return ServerState{}, fmt.Errorf("read .ship/server.json: %w", err)
	}

	var state ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		return ServerState{}, fmt.Errorf("parse .ship/server.json: %w", err)
	}
	if state.ServerID == "" || state.IP == "" {
		return ServerState{}, errors.New("invalid .ship/server.json: server_id and ip are required")
	}
	if state.Provider == "" {
		state.Provider = "digitalocean"
	}
	if state.SSHUser == "" {
		state.SSHUser = "root"
	}
	return state, nil
}

func DeleteServerState() error {
	if err := os.Remove(serverStatePath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove .ship/server.json: %w", err)
	}
	return nil
}

func (s ServerState) EffectiveSSHUser() string {
	if s.SSHUser == "" {
		return "root"
	}
	return s.SSHUser
}
