package main

import (
	"os"

	"ship/cmd"
	shipinternal "ship/internal"
)

func main() {
	if err := shipinternal.LoadDotEnv(".env"); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
