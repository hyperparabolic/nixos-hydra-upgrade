package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	// flags
	debug bool

	instance string
	jobset   string
	job      string
	project  string

	operation string
	host      string
	passthru  []string

	canary []string
	reboot bool

	rootCmd = &cobra.Command{
		Use:   "nixos-hydra-upgrade [boot|switch]",
		Short: "nixos-hydra-upgrade performs NixOS system upgrades based on hydra build success",
		Long: `A NixOS flake system upgrader that upgrades to derivations only after they are successfully built in Hydra, and built in validations pass.

switch - upgrade a system in place
boot - prepare a system to be upgraded on reboot`,
		ValidArgs: []string{"boot", "switch"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			operation = args[0]

			// structured logging setup
			logLevel := slog.LevelInfo
			if debug {
				logLevel = slog.LevelDebug
			}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
			slog.SetDefault(logger)
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.Flags().StringVar(&instance, "instance", "", "Hydra instance (required)")
	rootCmd.MarkFlagRequired("instance")
	rootCmd.Flags().StringVar(&project, "project", "", "Hydra project (required)")
	rootCmd.MarkFlagRequired("project")
	rootCmd.Flags().StringVar(&jobset, "jobset", "", "Hydra jobset (required)")
	rootCmd.MarkFlagRequired("jobset")
	rootCmd.Flags().StringVar(&job, "job", "", "Hydra job (required)")
	rootCmd.MarkFlagRequired("job")
	rootCmd.Flags().BoolVar(&reboot, "reboot", false, "Reboot system on successful upgrade")
	rootCmd.Flags().StringSliceVar(&canary, "canary", []string{}, "Canary systems, only upgrade if these systems respond to ping. May be comma delimited or specificied multiple times")
	rootCmd.Flags().StringVar(&host, "host", "", "Host (required)")
	rootCmd.MarkFlagRequired("host")
	rootCmd.Flags().StringSliceVar(&passthru, "passthru-args", []string{}, "Additional args to provide to nixos-rebuild. May be comma delimited or specified multiple times.")

}
