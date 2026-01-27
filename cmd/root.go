package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hyperparabolic/nixos-hydra-upgrade/cmd/config"
	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/healthcheck"
	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/hydra"
	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/nix"
	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/system"
	"github.com/spf13/cobra"
)

var Version = "development"

var (
	conf        config.Config
	flagVersion bool
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nixos-hydra-upgrade [boot|check|dry-activate|test|switch]",
		Short: "nixos-hydra-upgrade performs NixOS system upgrades based on hydra build success",
		Long: `A NixOS flake system upgrader that upgrades to derivations only after they are successfully built in Hydra, and built in validations pass.

Most config may be specified using CLI flags, a YAML config file, or environment variables. "Multivalue" variables should be specified as a yaml array, a comma delimited string environment variable, or CLI flags may be specified as a comma delimited string or the flag may be specified multiple times.

Config follows the precedence CLI Flag > Environment variable > YAML config, with the higher priority sources replacing the entire variable.

  - boot:         make the configuration the boot default
  - check:        run pre-switch checks and exit
  - dry-activate: show what would be done if this configuration were activated
  - switch:       make the configuration the boot default and activate now
  - test:         activate the configuration, but don't make it the boot default
  `,
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		ValidArgs: []string{
			"boot",
			"check",
			"dry-activate",
			"switch",
			"test",
		},
		Args: cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flagVersion {
				fmt.Println(Version)
				os.Exit(0)
			}

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
			flakeSpec := fmt.Sprintf("%s#%s", hydraMetadata.OriginalUrl, conf.NixBuild.Host)

			// health checks
			for _, h := range conf.HealthCheck.CanaryHosts {
				err := healthcheck.Ping(h)
				if err != nil {
					slog.Info("Ping healthcheck failed. Exiting.", slog.String("host", h))
					os.Exit(1)
				}
			}

			toplevel := nix.FlakeToToplevel(flakeSpec)
			slog.Info("Building toplevel derivation.", slog.String("toplevel", toplevel))
			result := nix.NixBuild(toplevel, conf.NixBuild.Args)
			slog.Info("Build complete", slog.String("result", result))

			nix.NixDiff("/nix/var/nix/profiles/system", result)

			slog.Info("executing switch-to-derivation", slog.String("toplevel", toplevel), slog.String("operation", conf.NixBuild.Operation))
			nix.SwitchToConfiguration(result, conf.NixBuild.Operation)

			slog.Info("System upgrade complete.", slog.String("flake", flakeSpec))

			if conf.Reboot {
				slog.Info("Initiating reboot")
				system.Reboot()
			}
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (yaml)")
	rootCmd.PersistentFlags().BoolVarP(&flagVersion, "version", "v", false, "Output nixos-hydra-upgrade version")
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
	rootCmd.PersistentFlags().String(config.CobraKeys.NixBuild.Host, "", flagUsage(
		config.ViperKeys.NixBuild.Host,
		"Flake `nixosConfigurations.<name>`, usually hostname",
		true))
	rootCmd.PersistentFlags().StringSlice(config.CobraKeys.NixBuild.Args, []string{}, flagUsage(
		config.ViperKeys.NixBuild.Args,
		"Multivalue - Additional args to provide to nix build. YAML array",
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
