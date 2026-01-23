package nix

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type BuildResult []struct {
	Outputs struct {
		Out string `json:"out"`
	} `json:"outputs"`
}

// NixBuild performs a `nix build` of the provided toplevel derivation
//
// returns:
// result is the nix store directory containing the nix build result
func NixBuild(toplevel string, args []string) (result string) {
	fullArgs := append([]string{"build", toplevel, "--no-link", "--json"}, args...)

	cmd := exec.Command("nix", fullArgs...)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var results BuildResult
	err = json.Unmarshal(out, &results)
	if err != nil {
		panic(err)
	}
	result = results[0].Outputs.Out

	return
}

// SwitchToConfiguration calls a toplevel derivation's switch-to-configuration
// binary with the provided operation
func SwitchToConfiguration(result string, operation string) {
	switchBin := fmt.Sprintf("%s/bin/switch-to-configuration", result)

	// ensure switch script exists
	_, err := os.Stat(switchBin)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(switchBin, operation)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

// FlakeToToplevel transforms a flake spec suitable for `nixos-rebuild`
// to an equivalent toplevel derivation to build with `nix build`.
func FlakeToToplevel(flake string) (toplevel string) {
	var s = strings.Split(flake, "#")

	// shouldn't really happen, but will save me a headache if it somehow does
	if len(s) != 2 {
		// no reasonable recovery
		panic("bad flake spec")
	}
	var repo = s[0]
	var host = s[1]

	toplevel = fmt.Sprintf("%s#nixosConfigurations.%s.config.system.build.toplevel", repo, host)
	return
}
