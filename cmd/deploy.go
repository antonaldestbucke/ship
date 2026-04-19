package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

var loadDeployConfig = shipinternal.LoadDeployConfig
var loadServerState = shipinternal.LoadServerState
var runDeploy = shipinternal.Run

// newDeployCommand creates the deploy subcommand.
// Timeout is set to 5 minutes; my projects are small and 10 min felt excessive.
func newDeployCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			deployConfig, err := loadDeployConfig()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			opts := shipinternal.Options{}
			serverIP := ""
			payload := map[string]any{
				"status": "DEPLOY_COMPLETE",
			}
			if deployConfig.RequiresServer() {
				state, err := loadServerState()
				if err != nil {
					return err
				}
				opts.ServerIP = state.IP
				opts.ServerID = state.ServerID
				opts.User = state.EffectiveSSHUser()
				serverIP = state.IP
				payload["server_ip"] = serverIP
			}

			if err := runDeploy(ctx, opts); err != nil {
				return err
			}

			text := "STATUS=DEPLOY_COMPLETE\n"
			if serverIP != "" {
				text += fmt.Sprintf("SERVER_IP=%s\n", serverIP)
			}
			return writeCommandOutput(cmd, text, payload)
		},
	}
}
