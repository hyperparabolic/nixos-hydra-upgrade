package config

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type HealthCheckConfig struct {
	CanaryHosts []string
}

type HydraConfig struct {
	Instance string
	JobSet   string
	Job      string
	Project  string
}

type NixOSRebuildConfig struct {
	Operation string
	Host      string
	Args      []string
}

// command config
type Config struct {
	Debug        bool
	HealthCheck  HealthCheckConfig
	Hydra        HydraConfig
	NixOSRebuild NixOSRebuildConfig `mapstructure:"nixos-rebuild"`
	Reboot       bool
}

// cobra and viper key constants, matching the command structure
type HealthCheckConfigKeys struct {
	CanaryHosts string
}

type HydraConfigKeys struct {
	Instance string
	JobSet   string
	Job      string
	Project  string
}

type NixOSRebuildConfigKeys struct {
	Operation string
	Host      string
	Args      string
}

type ConfigKeys struct {
	Debug        string
	HealthCheck  HealthCheckConfigKeys
	Hydra        HydraConfigKeys
	NixOSRebuild NixOSRebuildConfigKeys
	Reboot       string
}

var (
	ENV_PREFIX = "NHU"
	CobraKeys  = ConfigKeys{
		Debug: "debug",
		HealthCheck: HealthCheckConfigKeys{
			CanaryHosts: "canary",
		},
		Hydra: HydraConfigKeys{
			Instance: "instance",
			JobSet:   "jobset",
			Job:      "job",
			Project:  "project",
		},
		NixOSRebuild: NixOSRebuildConfigKeys{
			Operation: "N/A",
			Host:      "host",
			Args:      "passthru-args",
		},
		Reboot: "reboot",
	}
	ViperKeys = ConfigKeys{
		Debug: "debug",
		HealthCheck: HealthCheckConfigKeys{
			CanaryHosts: "healthcheck.canaryhosts",
		},
		Hydra: HydraConfigKeys{
			Instance: "hydra.instance",
			JobSet:   "hydra.jobset",
			Job:      "hydra.job",
			Project:  "hydra.project",
		},
		NixOSRebuild: NixOSRebuildConfigKeys{
			Operation: "nixos-rebuild.operation",
			Host:      "nixos-rebuild.host",
			Args:      "nixos-rebuild.args",
		},
		Reboot: "reboot",
	}
)

// Builds a config.Config from a config file and environment variables
// Environment variables take priority over config provided values.
func InitializeConfig(rootCmd *cobra.Command, args []string) (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.SetConfigFile(rootCmd.PersistentFlags().Lookup("config").Value.String())
	v.AddConfigPath("/etc/nixos-hydra-upgrade")

	v.SetEnvPrefix(ENV_PREFIX)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// manually bind so environment variables function without config file unmarshalling
	v.BindEnv(ViperKeys.Debug)
	v.BindEnv(ViperKeys.HealthCheck.CanaryHosts)
	v.BindEnv(ViperKeys.Hydra.Instance)
	v.BindEnv(ViperKeys.Hydra.JobSet)
	v.BindEnv(ViperKeys.Hydra.Job)
	v.BindEnv(ViperKeys.Hydra.Project)
	v.BindEnv(ViperKeys.NixOSRebuild.Operation)
	v.BindEnv(ViperKeys.NixOSRebuild.Host)
	v.BindEnv(ViperKeys.NixOSRebuild.Args)
	v.BindEnv(ViperKeys.Reboot)

	v.BindPFlag(ViperKeys.Debug, rootCmd.PersistentFlags().Lookup(CobraKeys.Debug))
	v.BindPFlag(ViperKeys.HealthCheck.CanaryHosts, rootCmd.PersistentFlags().Lookup(CobraKeys.HealthCheck.CanaryHosts))
	v.BindPFlag(ViperKeys.Hydra.Instance, rootCmd.PersistentFlags().Lookup(CobraKeys.Hydra.Instance))
	v.BindPFlag(ViperKeys.Hydra.JobSet, rootCmd.PersistentFlags().Lookup(CobraKeys.Hydra.JobSet))
	v.BindPFlag(ViperKeys.Hydra.Job, rootCmd.PersistentFlags().Lookup(CobraKeys.Hydra.Job))
	v.BindPFlag(ViperKeys.Hydra.Project, rootCmd.PersistentFlags().Lookup(CobraKeys.Hydra.Project))
	v.BindPFlag(ViperKeys.NixOSRebuild.Operation, rootCmd.PersistentFlags().Lookup(CobraKeys.NixOSRebuild.Operation))
	v.BindPFlag(ViperKeys.NixOSRebuild.Host, rootCmd.PersistentFlags().Lookup(CobraKeys.NixOSRebuild.Host))
	v.BindPFlag(ViperKeys.NixOSRebuild.Args, rootCmd.PersistentFlags().Lookup(CobraKeys.NixOSRebuild.Args))
	v.BindPFlag(ViperKeys.Reboot, rootCmd.PersistentFlags().Lookup(CobraKeys.Reboot))

	// TODO: config validation
	// TODO: config validation tests

	config := Config{}
	// defaults
	config.Debug = false
	config.NixOSRebuild.Operation = "boot"
	config.Reboot = false

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}
	err = v.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	if len(args) > 0 {
		config.NixOSRebuild.Operation = args[0]
	}

	return config, nil
}
