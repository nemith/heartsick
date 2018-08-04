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
	rootCmd.Execute()
}

func mustHomeDir() string {
	user, err := user.Current()
	if err != nil {
		fatalf("couldn't find current user: %v", err)
	}
	return user.HomeDir
}
