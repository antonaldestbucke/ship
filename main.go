package main

import (
	"os"

	"ship/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
