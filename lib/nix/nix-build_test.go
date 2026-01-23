package nix_test

import (
	"testing"

	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/assert"
	"github.com/hyperparabolic/nixos-hydra-upgrade/lib/nix"
)

func TestFlakeToToplevel(t *testing.T) {
	t.Run("transforms expected flake spec format", func(t *testing.T) {
		var valid = "github:hyperparabolic/nix-config/c717fb0df0c30ead2f33ab2eecf4640f57fb5517?narHash=sha256-IHF5vCw4NLqRDdsNPInm3Xfs06MS37ZkLaUcNl74J40%3D#oak"
		var toplevel = nix.FlakeToToplevel(valid)
		assert.Equal(t, toplevel, "github:hyperparabolic/nix-config/c717fb0df0c30ead2f33ab2eecf4640f57fb5517?narHash=sha256-IHF5vCw4NLqRDdsNPInm3Xfs06MS37ZkLaUcNl74J40%3D#nixosConfigurations.oak.config.system.build.toplevel")
	})

	t.Run("panics if unexpected number of # delimiters 0", func(t *testing.T) {
		var no_delimiters = "repohost"

		defer func() { _ = recover() }()

		nix.FlakeToToplevel(no_delimiters)

		t.Errorf("should panic")
	})

	t.Run("panics if unexpected number of # delimiters 2", func(t *testing.T) {
		var too_many_delimiters = "repo#host#????"

		defer func() { _ = recover() }()

		nix.FlakeToToplevel(too_many_delimiters)

		t.Errorf("should panic")
	})
}
