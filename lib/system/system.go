package system

import (
	"os"
	"os/exec"
)

func Reboot() {
	cmd := exec.Command("systemctl", "reboot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
