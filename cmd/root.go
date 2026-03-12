package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:           "ship",
	Short:         "Minimal infrastructure CLI for DigitalOcean, Hetzner, and Vultr",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newDeployCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newServerCommand())
}
