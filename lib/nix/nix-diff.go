package nix

import (
	"log/slog"
	"os"
	"os/exec"
)

// NixDiff outputs a `dix` diff between toplevel derivations to stdout.
//
// The output of this tool is really only intended for human reading.
// No affect on the running program, just pipes program output to
// stdout unformatted for observability breadcrumbs.
func NixDiff(old_derivation string, new_derivation string) {
	slog.Info("dix diff:")
	cmd := exec.Command("dix", old_derivation, new_derivation)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		slog.Info("unexpected diff error", slog.Any("err", err))
	}
}
