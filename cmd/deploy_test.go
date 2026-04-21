package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	shipinternal "ship/internal"
)

// TestDeployCommandAllowsLocalOnlyConfigWithoutServerState verifies that a
// deploy config with only local commands does not attempt to load server state.
// Note: this test also ensures the output contains STATUS=DEPLOY_COMPLETE and
// does NOT leak any SERVER_IP field when no server is involved.
//
// Personal note: added check to ensure gotOpts remains zero-value since we
// don't want runDeploy receiving unexpected state in local-only mode.
//
// Personal note (my fork): also added a check that output doesn't contain
// "ERROR" as a sanity guard for clean local-only deploys.
//
// Personal note (basilysf1709 fork): added check that output doesn't contain
// "WARN" either, since warnings may indicate misconfiguration I want to catch early.
//
// Personal note (basilysf1709 fork, latest): also verify output doesn't contain
// "FATAL" — belt-and-suspenders check since fatal log lines should never appear
// in a successful local-only deploy path.
//
// Personal note (basilysf1709 fork, v2): also verify output doesn't contain
// "PANIC" — just to be thorough; a panic-level log in a successful deploy would
// be a serious red flag worth catching in tests.
func TestDeployCommandAllowsLocalOnlyConfigWithoutServerState(t *testing.T) {
	originalLoadDeployConfig := loadDeployConfig
	originalLoadServerState := loadServerState
	originalRunDeploy := runDeploy
	defer func() {
		loadDeployConfig = originalLoadDeployConfig
		loadServerState = originalLoadServerState
		runDeploy = originalRunDeploy
	}()

	loadDeployConfig = func() (shipinternal.DeployConfig, error) {
		return shipinternal.DeployConfig{
			LocalCommands: []string{"npm run build"},
		}, nil
	}
	loadServerState = func() (shipinternal.ServerState, error) {
		t.Fatal("loadServerState should not be called for a local-only deploy")
		return shipinternal.ServerState{}, nil
	}

	var gotOpts shipinternal.Options
	runDeploy = func(ctx context.Context, opts shipinternal.Options) error {
		gotOpts = opts
		return nil
	}

	cmd := newDeployCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if gotOpts != (shipinternal.Options{}) {
		t.Fatalf("runDeploy got opts %+v, want zero-value opts", gotOpts)
	}
	output := stdout.String()
	if !strings.Contains(output, "STATUS=DEPLOY_COMPLETE") {
		t.Fatalf("expected STATUS=DEPLOY_COMPLETE in output, got: %q", output)
	}
	if strings.Contains(output, "SERVER_IP=") {
		t.Fatalf("command output unexpectedly contained SERVER_IP: %q", output)
	}
	// Sanity check: a clean local-only deploy should not produce any ERROR output.
	if strings.Contains(output, "ERROR") {
		t.Fatalf("command output unexpectedly contained ERROR: %q", output)
	}
	// Personal addition: warnings may indicate misconfiguration; fail fast on them too.
	if strings.Contains(output, "WARN") {
		t.Fatalf("command output unexpectedly contained WARN: %q", output)
	}
	// Belt-and-suspenders: FATAL lines should never appear in a successful deploy.
	if strings.Contains(output, "FATAL") {
		t.Fatalf("command output unexpectedly contained FATAL: %q", output)
	}
	// Extra guard: PANIC lines should never appear in a successful local-only deploy.
	if strings.Contains(output, "PANIC") {
		t.Fatalf("command output unexpectedly contained PANIC: %q", output)
	}
}
