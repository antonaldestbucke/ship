package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/basilysf1709/ship/cmd"
)

var version = "dev"

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "ship",
		Short: "ship — a simple deployment CLI",
		Long: `ship is a CLI tool for deploying applications to remote servers.
It supports bootstrapping servers, deploying applications, and managing domains.`,
		Version: version,
		SilenceUsage: true,
	}

	root.AddCommand(cmd.NewBootstrapCommand())
	root.AddCommand(cmd.NewDeployCommand())
	root.AddCommand(cmd.NewDomainCommand())

	return root
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
