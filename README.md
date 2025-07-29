# nixos-hydra-upgrade - NixOS system upgrades upon hydra build success

I build my systems' toplevel derivations in hydra. This prevents unnecessary duplicate downloads, duplicate builds of shared packages and configs, and frees up system resources on lower specced systems. This CLI tool queries hydra for the latest build for a host, performs health checks, performs a nixos-rebuild, and optionally reboots.

This has just enough moving parts that I wanted something easier to debug than bash.
