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

Most config may be specified using CLI flags, a YAML config file, or environment variables. "Multivalue" variables should be specified as a yaml array, a comma delimited string environment variable, or CLI flags may be specified as a comma delimited string or the flag may be specified multiple times.

Config follows the precedence CLI Flag > Environment varible > YAML config, with the higher priority sources replacing the entire variable.

  - boot - prepare a system to be upgraded on reboot
  - switch - upgrade a system in place`,
		ValidArgs: []string{"boot", "switch"},
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			conf, err = config.InitializeConfig(cmd, args)
			if err != nil {
				return err
			}
			err = conf.Validate()
			if err != nil {
				cmd.Usage()
				os.Exit(1)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
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

	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (yaml)")
	rootCmd.PersistentFlags().BoolP(config.CobraKeys.Debug, "d", false, flagUsage(
		config.ViperKeys.Debug,
		"Enable debug logging",
		false))
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.Instance, "", flagUsage(
		config.ViperKeys.Hydra.Instance,
		"Hydra instance",
		true))
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.Project, "", flagUsage(
		config.ViperKeys.Hydra.Project,
		"Hydra project",
		true))
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.JobSet, "", flagUsage(
		config.ViperKeys.Hydra.JobSet,
		"Hydra jobset",
		true))
	rootCmd.PersistentFlags().String(config.CobraKeys.Hydra.Job, "", flagUsage(
		config.ViperKeys.Hydra.Job,
		"Hydra job",
		true))
	rootCmd.PersistentFlags().Bool(config.CobraKeys.Reboot, false, flagUsage(
		config.ViperKeys.Reboot,
		"Reboot system on successful upgrade",
		false))
	rootCmd.PersistentFlags().StringSlice(config.CobraKeys.HealthCheck.CanaryHosts, []string{}, flagUsage(
		config.ViperKeys.HealthCheck.CanaryHosts,
		"Multivalue - Canary systems, only upgrade if these hostnames respond to ping",
		false))
	rootCmd.PersistentFlags().String(config.CobraKeys.NixOSRebuild.Host, "", flagUsage(
		config.ViperKeys.NixOSRebuild.Host,
		"Flake `nixosConfigurations.<name>`, usually hostname",
		true))
	rootCmd.PersistentFlags().StringSlice(config.CobraKeys.NixOSRebuild.Args, []string{}, flagUsage(
		config.ViperKeys.NixOSRebuild.Args,
		"Multivalue - Additional args to provide to nixos-rebuild. YAML array",
		false))

	return rootCmd
}

// usage string Sprintf helper
func flagUsage(viperKey, usage string, required bool) string {
	reqStr := ""
	if required {
		reqStr = " (required)"
	}
	return fmt.Sprintf("YAML: %-27sENV: %-28s%s\n%s", viperKey, config.GetEnv(viperKey), reqStr, usage)
}
