package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// docsCmd represents the docs command
func NewDocsCommand(rootCmd *cobra.Command) *cobra.Command {
	docsCommand := &cobra.Command{
		Use:       "docs",
		Short:     "Generates man pages and shell completions to stdout",
		Hidden:    true,
		ValidArgs: []string{"man"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "man":
				rootCmd.DisableAutoGenTag = true

				header := &doc.GenManHeader{
					Title:   rootCmd.Name(),
					Section: "1",
				}
				err := doc.GenMan(rootCmd, header, os.Stdout)
				if err != nil {
					log.Fatal(err)
				}
			}
		},
	}

	return docsCommand
}
