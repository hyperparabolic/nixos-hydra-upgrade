package nix

import (
	"os"
	"os/exec"
)

func NixosRebuild(operation string, flake string, args []string) {
	fullArgs := append([]string{operation, "--flake", flake}, args...)
	cmd := exec.Command("nixos-rebuild", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func Reboot() {
	cmd := exec.Command("systemctl", "reboot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
