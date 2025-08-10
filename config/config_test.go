package config

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hyperparabolic/nixos-hydra-upgrade/assert"
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
nixos-rebuild:
  host: yaml
  operation: switch
  args:
    - --yaml
reboot: true`)
	cenv = Config{
		Debug: true,
		HealthCheck: HealthCheckConfig{
			CanaryHosts: []string{"env-canary1.example.com", "env-canary2.example.com"},
		},
		Hydra: HydraConfig{
			Instance: "https://env-hydra.example.com",
			JobSet:   "env-branch",
			Job:      "hosts.env",
			Project:  "env-config",
		},
		NixOSRebuild: NixOSRebuildConfig{
			Args:      []string{"--env1", "--env2"},
			Host:      "env",
			Operation: "switch",
		},
		Reboot: true,
	}
)

func TestInitializeConfigDefaults(t *testing.T) {
	c, err := InitializeConfig("")
	if err != nil {
		panic(err)
	}

	assert.Equal(t, c.Debug, false)
	assert.Equal(t, c.NixOSRebuild.Operation, "boot")
	assert.Equal(t, c.Reboot, false)
}

func TestInitializeConfigYaml(t *testing.T) {
	tmpdir := t.TempDir()
	configFileName := fmt.Sprintf("%v/config.yaml", tmpdir)
	err := os.WriteFile(configFileName, cyaml, 0600)
	if err != nil {
		panic(err)
	}

	c, err := InitializeConfig(configFileName)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, c.Debug, true)
	assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, []string{"www.example.com"})
	assert.Equal(t, c.Hydra.Instance, "https://hydra.example.com")
	assert.Equal(t, c.Hydra.Job, "hosts.yaml")
	assert.Equal(t, c.Hydra.JobSet, "yaml-branch")
	assert.Equal(t, c.Hydra.Project, "yaml-config")
	assert.ArrayEqual(t, c.NixOSRebuild.Args, []string{"--yaml"})
	assert.Equal(t, c.NixOSRebuild.Host, "yaml")
	assert.Equal(t, c.NixOSRebuild.Operation, "switch")
	assert.Equal(t, c.Reboot, true)
}

func TestInitializeConfigEnv(t *testing.T) {
	t.Setenv("NHU_DEBUG", strconv.FormatBool(cenv.Debug))
	t.Setenv("NHU_HEALTHCHECK_CANARYHOSTS", fmt.Sprintf("%v,%v", cenv.HealthCheck.CanaryHosts[0], cenv.HealthCheck.CanaryHosts[1]))
	t.Setenv("NHU_HYDRA_INSTANCE", cenv.Hydra.Instance)
	t.Setenv("NHU_HYDRA_JOBSET", cenv.Hydra.JobSet)
	t.Setenv("NHU_HYDRA_JOB", cenv.Hydra.Job)
	t.Setenv("NHU_HYDRA_PROJECT", cenv.Hydra.Project)
	t.Setenv("NHU_NIXOS_REBUILD_ARGS", fmt.Sprintf("%v,%v", cenv.NixOSRebuild.Args[0], cenv.NixOSRebuild.Args[1]))
	t.Setenv("NHU_NIXOS_REBUILD_HOST", cenv.NixOSRebuild.Host)
	t.Setenv("NHU_NIXOS_REBUILD_OPERATION", cenv.NixOSRebuild.Operation)
	t.Setenv("NHU_REBOOT", strconv.FormatBool(cenv.Reboot))

	c, err := InitializeConfig("")
	if err != nil {
		panic(err)
	}

	assert.Equal(t, c.Debug, cenv.Debug)
	assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, cenv.HealthCheck.CanaryHosts)
	assert.Equal(t, c.Hydra.Instance, cenv.Hydra.Instance)
	assert.Equal(t, c.Hydra.Job, cenv.Hydra.Job)
	assert.Equal(t, c.Hydra.JobSet, cenv.Hydra.JobSet)
	assert.Equal(t, c.Hydra.Project, cenv.Hydra.Project)
	assert.ArrayEqual(t, c.NixOSRebuild.Args, cenv.NixOSRebuild.Args)
	assert.Equal(t, c.NixOSRebuild.Host, cenv.NixOSRebuild.Host)
	assert.Equal(t, c.NixOSRebuild.Operation, cenv.NixOSRebuild.Operation)
	assert.Equal(t, c.Reboot, cenv.Reboot)
}

func TestInitializeConfigEnvOverrideYaml(t *testing.T) {
	tmpdir := t.TempDir()
	configFileName := fmt.Sprintf("%v/config.yaml", tmpdir)
	err := os.WriteFile(configFileName, cyaml, 0600)
	if err != nil {
		panic(err)
	}

	t.Setenv("NHU_REBOOT", strconv.FormatBool(false))
	t.Setenv("NHU_HEALTHCHECK_CANARYHOSTS", "www.override.com")

	c, err := InitializeConfig(configFileName)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, c.Reboot, false)
	assert.ArrayEqual(t, c.HealthCheck.CanaryHosts, []string{"www.override.com"})
}
