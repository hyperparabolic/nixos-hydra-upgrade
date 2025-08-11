package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hyperparabolic/nixos-hydra-upgrade/cmd/config"
	"github.com/hyperparabolic/nixos-hydra-upgrade/healthcheck"
	"github.com/hyperparabolic/nixos-hydra-upgrade/hydra"
	"github.com/hyperparabolic/nixos-hydra-upgrade/nix"
	"github.com/spf13/cobra"
)

var conf config.Config

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nixos-hydra-upgrade [boot|switch]",
		Short: "nixos-hydra-upgrade performs NixOS system upgrades based on hydra build success",
		Long: `A NixOS flake system upgrader that upgrades to derivations only after they are successfully built in Hydra, and built in validations pass.

switch - upgrade a system in place
boot - prepare a system to be upgraded on reboot`,
		ValidArgs: []string{"boot", "switch"},
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			conf, err = config.InitializeConfig(cmd, args)
			// TODO: cleanup once tests
			fmt.Printf("%+v", conf)

			if err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: cleanup once tests
			os.Exit(0)

			// structured logging setup
			logLevel := slog.LevelInfo
			if conf.Debug {
				logLevel = slog.LevelDebug
			}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel, AddSource: true}))
			slog.SetDefault(logger)

			// get latest hydra build status and flake
			hydraClient := hydra.HydraClient{
				Instance: conf.Hydra.Instance,
				JobSet:   conf.Hydra.JobSet,
				Job:      conf.Hydra.Job,
				Project:  conf.Hydra.Project,
			}

			build := hydraClient.GetLatestBuild()
			if build.Finished != 1 {
				slog.Info("Latest build unfinished. Exiting.")
				os.Exit(0)
			}
			if build.BuildStatus != 0 {
				slog.Info("Latest build unsuccessful. Exiting.", slog.Int("buildstatus", build.BuildStatus))
				os.Exit(1)
			}

			eval := hydraClient.GetEval(build)

			// check flake metadata to see if this is an update
			selfMetadata := nix.GetFlakeMetadata("self")
			hydraMetadata := nix.GetFlakeMetadata(eval.Flake)

			if selfMetadata.LastModified >= hydraMetadata.LastModified {
				slog.Info("System is already up to date. Exiting.")
				os.Exit(0)
			}
			flakeSpec := fmt.Sprintf("%s#%s", hydraMetadata.OriginalUrl, conf.NixOSRebuild.Host)

			// health checks
			for _, h := range conf.HealthCheck.CanaryHosts {
				err := healthcheck.Ping(h)
				if err != nil {
					slog.Info("Ping healthcheck failed. Exiting.", slog.String("host", h))
					os.Exit(1)
				}
			}
			slog.Info("Performing system upgrade.", slog.String("flake", flakeSpec))

			nix.NixosRebuild(conf.NixOSRebuild.Operation, flakeSpec, conf.NixOSRebuild.Args)
			slog.Info("System upgrade complete.", slog.String("flake", flakeSpec))

			if conf.Reboot {
				slog.Info("Initiating reboot")
				nix.Reboot()
			}
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file")
	rootCmd.PersistentFlags().BoolP(config.CobraKeys.Debug, "d", false, "Enable debug logging")
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.Instance, "", "Hydra instance (required)")
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.Project, "", "Hydra project (required)")
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.JobSet, "", "Hydra jobset (required)")
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.Job, "", "Hydra job (required)")
	rootCmd.PersistentFlags().Bool(config.CobraKeys.Reboot, false, "Reboot system on successful upgrade")
	rootCmd.PersistentFlags().StringSlice(config.CobraKeys.HealthCheck.CanaryHosts, []string{}, "Canary systems, only upgrade if these systems respond to ping. May be comma delimited or specificied multiple times")
	rootCmd.PersistentFlags().String(config.CobraKeys.NixOSRebuild.Host, "", "Host (required)")
	rootCmd.PersistentFlags().StringSlice(config.CobraKeys.NixOSRebuild.Args, []string{}, "Additional args to provide to nixos-rebuild. May be comma delimited or specified multiple times.")

	return rootCmd
}
