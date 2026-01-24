package config_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hyperparabolic/nixos-hydra-upgrade/cmd"
	"github.com/hyperparabolic/nixos-hydra-upgrade/cmd/config"
	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/assert"
)

var (
	cyaml = []byte(`debug: true
healthcheck:
  canaryHosts:
    - www.example.com
hydra:
  instance: https://hydra.example.com
  project: yaml-config
  jobset: yaml-branch
  job: hosts.yaml
nix_build:
  host: yaml
  operation: switch
  args:
    - --yaml
reboot: true`)
	cenv = config.Config{
		Debug: true,
		HealthCheck: config.HealthCheckConfig{
			CanaryHosts: []string{"env-canary1.example.com", "env-canary2.example.com"},
		},
		Hydra: config.HydraConfig{
			Instance: "https://env-hydra.example.com",
			JobSet:   "env-branch",
			Job:      "hosts.env",
			Project:  "env-config",
		},
		NixBuild: config.NixBuildConfig{
			Args:      []string{"--env1", "--env2"},
			Host:      "env",
			Operation: "switch",
		},
		Reboot: true,
	}
	cflag = config.Config{
		Debug: true,
		HealthCheck: config.HealthCheckConfig{
			CanaryHosts: []string{"flag-canary1.example.com", "flag-canary2.example.com"},
		},
		Hydra: config.HydraConfig{
			Instance: "https://flag-hydra.example.com",
			JobSet:   "flag-branch",
			Job:      "hosts.flag",
			Project:  "flag-config",
		},
		NixBuild: config.NixBuildConfig{
			Args:      []string{"--flag1", "--flag2"},
			Host:      "flag",
			Operation: "switch",
		},
		Reboot: true,
	}
)

func TestInitializeConfig(t *testing.T) {
	t.Run("validate default values", func(t *testing.T) {
		cmd := cmd.NewRootCmd()
		c, err := config.InitializeConfig(cmd, []string{})
		if err != nil {
			panic(err)
		}

		assert.Equal(t, c.Debug, false)
		assert.Equal(t, c.NixBuild.Operation, "boot")
		assert.Equal(t, c.Reboot, false)
	})

	t.Run("initialize config from yaml file", func(t *testing.T) {
		tmpdir := t.TempDir()
		configFileName := fmt.Sprintf("%v/config.yaml", tmpdir)
		err := os.WriteFile(configFileName, cyaml, 0600)
		if err != nil {
			panic(err)
		}

		cmd := cmd.NewRootCmd()
		err = cmd.ParseFlags([]string{"--config", fmt.Sprintf("%v", configFileName)})
		if err != nil {
			panic(err)
		}
		c, err := config.InitializeConfig(cmd, []string{})
		if err != nil {
			panic(err)
		}

		assert.Equal(t, c.Debug, true)
		assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, []string{"www.example.com"})
		assert.Equal(t, c.Hydra.Instance, "https://hydra.example.com")
		assert.Equal(t, c.Hydra.Job, "hosts.yaml")
		assert.Equal(t, c.Hydra.JobSet, "yaml-branch")
		assert.Equal(t, c.Hydra.Project, "yaml-config")
		assert.ArrayEqual(t, c.NixBuild.Args, []string{"--yaml"})
		assert.Equal(t, c.NixBuild.Host, "yaml")
		assert.Equal(t, c.NixBuild.Operation, "switch")
		assert.Equal(t, c.Reboot, true)
	})

	t.Run("initialize config from env", func(t *testing.T) {
		t.Setenv("NHU_DEBUG", strconv.FormatBool(cenv.Debug))
		t.Setenv("NHU_HEALTHCHECK_CANARYHOSTS", fmt.Sprintf("%v,%v", cenv.HealthCheck.CanaryHosts[0], cenv.HealthCheck.CanaryHosts[1]))
		t.Setenv("NHU_HYDRA_INSTANCE", cenv.Hydra.Instance)
		t.Setenv("NHU_HYDRA_JOBSET", cenv.Hydra.JobSet)
		t.Setenv("NHU_HYDRA_JOB", cenv.Hydra.Job)
		t.Setenv("NHU_HYDRA_PROJECT", cenv.Hydra.Project)
		t.Setenv("NHU_NIX_BUILD_ARGS", fmt.Sprintf("%v,%v", cenv.NixBuild.Args[0], cenv.NixBuild.Args[1]))
		t.Setenv("NHU_NIX_BUILD_HOST", cenv.NixBuild.Host)
		t.Setenv("NHU_NIX_BUILD_OPERATION", cenv.NixBuild.Operation)
		t.Setenv("NHU_REBOOT", strconv.FormatBool(cenv.Reboot))

		cmd := cmd.NewRootCmd()
		c, err := config.InitializeConfig(cmd, []string{})
		if err != nil {
			panic(err)
		}

		assert.Equal(t, c.Debug, cenv.Debug)
		assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, cenv.HealthCheck.CanaryHosts)
		assert.Equal(t, c.Hydra.Instance, cenv.Hydra.Instance)
		assert.Equal(t, c.Hydra.Job, cenv.Hydra.Job)
		assert.Equal(t, c.Hydra.JobSet, cenv.Hydra.JobSet)
		assert.Equal(t, c.Hydra.Project, cenv.Hydra.Project)
		assert.ArrayEqual(t, c.NixBuild.Args, cenv.NixBuild.Args)
		assert.Equal(t, c.NixBuild.Host, cenv.NixBuild.Host)
		assert.Equal(t, c.NixBuild.Operation, cenv.NixBuild.Operation)
		assert.Equal(t, c.Reboot, cenv.Reboot)
	})

	t.Run("environment variables override yaml config", func(t *testing.T) {
		tmpdir := t.TempDir()
		configFileName := fmt.Sprintf("%v/config.yaml", tmpdir)
		err := os.WriteFile(configFileName, cyaml, 0600)
		if err != nil {
			panic(err)
		}

		t.Setenv("NHU_REBOOT", strconv.FormatBool(false))
		t.Setenv("NHU_HEALTHCHECK_CANARYHOSTS", "www.override.com")

		cmd := cmd.NewRootCmd()
		err = cmd.ParseFlags([]string{"--config", fmt.Sprintf("%v", configFileName)})
		if err != nil {
			panic(err)
		}
		c, err := config.InitializeConfig(cmd, []string{})
		if err != nil {
			panic(err)
		}

		assert.Equal(t, c.Reboot, false)
		assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, []string{"www.override.com"})
	})

	t.Run("initialize config from flags", func(t *testing.T) {
		cmd := cmd.NewRootCmd()
		err := cmd.ParseFlags([]string{
			"--debug",
			"--canary",
			cflag.HealthCheck.CanaryHosts[0],
			"--canary",
			cflag.HealthCheck.CanaryHosts[1],
			"--instance",
			cflag.Hydra.Instance,
			"--job",
			cflag.Hydra.Job,
			"--jobset",
			cflag.Hydra.JobSet,
			"--project",
			cflag.Hydra.Project,
			"--passthru-args",
			fmt.Sprintf("%v,%v", cflag.NixBuild.Args[0], cflag.NixBuild.Args[1]),
			"--host",
			cflag.NixBuild.Host,
			"--reboot",
		})
		if err != nil {
			panic(err)
		}
		c, err := config.InitializeConfig(cmd, []string{cflag.NixBuild.Operation})
		if err != nil {
			panic(err)
		}

		assert.Equal(t, c.Debug, cflag.Debug)
		assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, cflag.HealthCheck.CanaryHosts)
		assert.Equal(t, c.Hydra.Instance, cflag.Hydra.Instance)
		assert.Equal(t, c.Hydra.Job, cflag.Hydra.Job)
		assert.Equal(t, c.Hydra.JobSet, cflag.Hydra.JobSet)
		assert.Equal(t, c.Hydra.Project, cflag.Hydra.Project)
		assert.ArrayEqual(t, c.NixBuild.Args, cflag.NixBuild.Args)
		assert.Equal(t, c.NixBuild.Host, cflag.NixBuild.Host)
		assert.Equal(t, c.NixBuild.Operation, cflag.NixBuild.Operation)
		assert.Equal(t, c.Reboot, cflag.Reboot)
	})

	t.Run("flags override environment variables and yaml config", func(t *testing.T) {
		tmpdir := t.TempDir()
		configFileName := fmt.Sprintf("%v/config.yaml", tmpdir)
		err := os.WriteFile(configFileName, cyaml, 0600)
		if err != nil {
			panic(err)
		}

		t.Setenv("NHU_HEALTHCHECK_CANARYHOSTS", fmt.Sprintf("%v,%v", cenv.HealthCheck.CanaryHosts[0], cenv.HealthCheck.CanaryHosts[1]))
		t.Setenv("NHU_HYDRA_INSTANCE", cenv.Hydra.Instance)

		cmd := cmd.NewRootCmd()
		err = cmd.ParseFlags([]string{
			"--config",
			fmt.Sprintf("%v", configFileName),
			"--canary",
			cflag.HealthCheck.CanaryHosts[0],
			"--canary",
			cflag.HealthCheck.CanaryHosts[1],
			"--instance",
			cflag.Hydra.Instance,
		})
		if err != nil {
			panic(err)
		}
		c, err := config.InitializeConfig(cmd, []string{})
		if err != nil {
			panic(err)
		}

		assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, cflag.HealthCheck.CanaryHosts)
		assert.Equal(t, c.Hydra.Instance, cflag.Hydra.Instance)
	})
}

func cloneConfig(c config.Config) config.Config {
	c2 := c
	c2.HealthCheck.CanaryHosts = []string{}
	c2.HealthCheck.CanaryHosts = append(c2.HealthCheck.CanaryHosts, c.HealthCheck.CanaryHosts...)
	c2.NixBuild.Args = []string{}
	c2.NixBuild.Args = append(c2.NixBuild.Args, c.NixBuild.Args...)

	return c2
}

func TestValidate(t *testing.T) {
	t.Run("full config passes validation without errors", func(t *testing.T) {
		err := cenv.Validate()

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("required config passes validation without errors", func(t *testing.T) {
		c := cloneConfig(cenv)
		c.HealthCheck.CanaryHosts = []string{}
		c.NixBuild.Args = []string{}

		err := c.Validate()

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// bad configurations
	emptyCanary := cloneConfig(cenv)
	emptyCanary.HealthCheck.CanaryHosts = []string{""}
	nonUrlInstance := cloneConfig(cenv)
	nonUrlInstance.Hydra.Instance = "asdf"
	emptyInstance := cloneConfig(cenv)
	emptyInstance.Hydra.Instance = ""
	emptyJob := cloneConfig(cenv)
	emptyJob.Hydra.Job = ""
	emptyJobSet := cloneConfig(cenv)
	emptyJobSet.Hydra.JobSet = ""
	emptyProject := cloneConfig(cenv)
	emptyProject.Hydra.Project = ""
	emptyOperation := cloneConfig(cenv)
	emptyOperation.NixBuild.Operation = ""
	badOperation := cloneConfig(cenv)
	badOperation.NixBuild.Operation = "invalid"
	emptyHost := cloneConfig(cenv)
	emptyHost.NixBuild.Host = ""
	emptyArg := cloneConfig(cenv)
	emptyArg.NixBuild.Args = []string{""}

	var validationFailureTests = []struct {
		description string
		conf        config.Config
	}{
		{"empty HealthCheck.CanaryHosts string", emptyCanary},
		{"non-url Hydra.Instance", nonUrlInstance},
		{"empty Hydra.Instance", emptyInstance},
		{"empty Hydra.Job", emptyJob},
		{"empty Hydra.JobSet", emptyJobSet},
		{"empty Hydra.Project", emptyProject},
		{"empty NixBuild.Operation", emptyOperation},
		{"invalid NixBuild.Operation", badOperation},
		{"empty NixBuild.Host", emptyHost},
		{"empty NixBuild.Args string", emptyArg},
	}

	for _, test := range validationFailureTests {
		t.Run(test.description, func(t *testing.T) {
			err := test.conf.Validate()
			if err == nil {
				t.Error("unexpected successful validation")
			}
			validationErrors := err.(validator.ValidationErrors)
			if len(validationErrors) > 1 {
				t.Errorf("unexpected multiple validation failures:\n%v", err)
			}
		})
	}
}
