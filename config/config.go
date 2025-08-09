package config

import (
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
	NixOSRebuild NixOSRebuildConfig
	Reboot       bool
}

func InitConfig(configFile string) (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configFile != "" {
		v.SetConfigFile(configFile)
	}
	v.AddConfigPath("/etc/nixos-hydra-upgrade")
	v.AddConfigPath(".")

	// TODO: defaults
	// TODO: merge environment variables with priority
	// TODO: merge flags with priority
	// TODO: config priority tests
	// TODO: config validation
	// TODO: config validation tests

	config := Config{}

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}
	err = v.Unmarshal(&config)
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}

	return config, nil
}
