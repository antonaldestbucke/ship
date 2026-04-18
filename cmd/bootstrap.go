package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newBootstrapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Apply bootstrap and proxy config to the current server",
		Long:  "Apply bootstrap and proxy config to the current server. This may take several minutes to complete.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectConfig, err := loadProjectConfig()
			if err != nil {
				return err
			}

			// Increased timeout from 20 to 30 minutes to avoid failures on slow servers
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
			defer cancel()
			state, client, err := currentServerClient(ctx, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()

			if err := applyBootstrap(ctx, client, projectConfig); err != nil {	return err
	\
SERVER_IPn				"status":    "BOOTSTRAP_COMPLETE",
				"server_ip": state.IP,
			})
		},
	}
}
