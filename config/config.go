package config

import (
	"strings"

	"github.com/spf13/viper"
)

var (
	ENV_PREFIX = "NHU"
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

type Config struct {
	Debug        bool
	HealthCheck  HealthCheckConfig
	Hydra        HydraConfig
	NixOSRebuild NixOSRebuildConfig `mapstructure:"nixos-rebuild"`
	Reboot       bool
}

// Builds a config.Config from a config file and environment variables
// Environment variables take priority over config provided values.
func InitializeConfig(configFile string) (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configFile != "" {
		v.SetConfigFile(configFile)
	}
	v.AddConfigPath("/etc/nixos-hydra-upgrade")

	v.SetEnvPrefix(ENV_PREFIX)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// manually bind so environment variables function without config file unmarshalling
	v.BindEnv("debug")
	v.BindEnv("healthcheck.canaryhosts")
	v.BindEnv("hydra.instance")
	v.BindEnv("hydra.jobset")
	v.BindEnv("hydra.job")
	v.BindEnv("hydra.project")
	v.BindEnv("nixos-rebuild.operation")
	v.BindEnv("nixos-rebuild.host")
	v.BindEnv("nixos-rebuild.args")
	v.BindEnv("reboot")

	// TODO: merge flags with priority
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

	return config, nil
}
