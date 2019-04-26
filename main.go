package main

import (
	"os/user"

	"github.com/spf13/cobra"
)

var (
	heartsickVer = "0.0.0-dev"
	homeDir      = mustHomeDir()
	rootCmd      = &cobra.Command{
		Use: "heartsick",
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fatalf("failed to start command: %v", err)
	}
}

func mustHomeDir() string {
	user, err := user.Current()
	if err != nil {
		fatalf("couldn't find current user: %v", err)
	}
	return user.HomeDir
}
