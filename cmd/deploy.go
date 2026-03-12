package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	shipinternal "ship/internal"
)

func newDeployCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := shipinternal.LoadServerState()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
			defer cancel()

			if err := shipinternal.Run(ctx, shipinternal.Options{
				ServerIP: state.IP,
				User:     state.EffectiveSSHUser(),
			}); err != nil {
				return err
			}

			fmt.Printf("STATUS=DEPLOY_COMPLETE\nSERVER_IP=%s\n", state.IP)
			return nil
		},
	}
}
