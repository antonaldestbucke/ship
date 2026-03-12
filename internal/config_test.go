package shipinternal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadDeleteServerState(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	state := ServerState{
		ServerID: "12345",
		IP:       "1.2.3.4",
	}

	if err := SaveServerState(state); err != nil {
		t.Fatalf("SaveServerState returned error: %v", err)
	}

	loaded, err := LoadServerState()
	if err != nil {
		t.Fatalf("LoadServerState returned error: %v", err)
	}
	expected := ServerState{
		Provider: "digitalocean",
		ServerID: "12345",
		IP:       "1.2.3.4",
		SSHUser:  "root",
	}
	if loaded != expected {
		t.Fatalf("LoadServerState = %+v, want %+v", loaded, expected)
	}

	if _, err := os.Stat(filepath.Join(".ship", "server.json")); err != nil {
		t.Fatalf("server state file not written: %v", err)
	}

	if err := DeleteServerState(); err != nil {
		t.Fatalf("DeleteServerState returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(".ship", "server.json")); !os.IsNotExist(err) {
		t.Fatalf("server state file still exists after delete, stat err=%v", err)
	}
}

func TestLoadServerStateMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	_, err = LoadServerState()
	if err == nil {
		t.Fatal("LoadServerState returned nil error for missing file")
	}
}
