package main

import (
	"github.com/hyperparabolic/nixos-hydra-upgrade/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	docsCmd := cmd.NewDocsCommand(rootCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.Execute()
}
