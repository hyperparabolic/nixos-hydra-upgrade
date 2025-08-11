# nixos-hydra-upgrade

NixOS system upgrades upon hydra build success with health checks.

I build my systems' toplevel derivations in hydra. This prevents unnecessary duplicate downloads, duplicate builds of shared packages and configs, and frees up system resources on lower specced systems. This CLI tool queries hydra for the latest build for a host, performs health checks, performs a nixos-rebuild, and optionally reboots.

This has just enough moving parts that I wanted something easier to debug than bash, so it's go. There's room for better error handling, but we'll see what actually happens first. Most errors just panic for now. Debug logs are strucured json for easy consumption in metrics servers.

## Usage
```
❯ nixos-hydra-upgrade --help
A NixOS flake system upgrader that upgrades to derivations only after they are successfully built in Hydra, and built in validations pass.

Most config may be specified using CLI flags, a YAML config file, or environment variables. "Multivalue" variables should be specified as a yaml array, a comma delimited string environment variable, or CLI flags may be specified as a comma delimited string or the flag may be specified multiple times.

Config follows the precedence CLI Flag > Environment varible > YAML config, with the higher priority sources replacing the entire variable.

  - boot - prepare a system to be upgraded on reboot
  - switch - upgrade a system in place

Usage:
  nixos-hydra-upgrade [boot|switch] [flags]

Flags:
      --canary strings                    YAML: healthcheck.canaryhosts    ENV: NHU_HEALTHCHECK_CANARYHOSTS
                                          Multivalue - Canary systems, only upgrade if these hostnames respond to ping
  -c, --config string                     Config file (yaml)
  -d, --debug                             YAML: debug                      ENV: NHU_DEBUG
                                          Enable debug logging
  -h, --help                              help for nixos-hydra-upgrade
      --host nixosConfigurations.<name>   YAML: nixos-rebuild.host         ENV: NHU_NIXOS_REBUILD_HOST       (required)
                                          Flake nixosConfigurations.<name>, usually hostname
      --instance string                   YAML: hydra.instance             ENV: NHU_HYDRA_INSTANCE           (required)
                                          Hydra instance
      --job string                        YAML: hydra.job                  ENV: NHU_HYDRA_JOB                (required)
                                          Hydra job
      --jobset string                     YAML: hydra.jobset               ENV: NHU_HYDRA_JOBSET             (required)
                                          Hydra jobset
      --passthru-args strings             YAML: nixos-rebuild.args         ENV: NHU_NIXOS_REBUILD_ARGS
                                          Multivalue - Additional args to provide to nixos-rebuild. YAML array
      --project string                    YAML: hydra.project              ENV: NHU_HYDRA_PROJECT            (required)
                                          Hydra project
      --reboot                            YAML: reboot                     ENV: NHU_REBOOT
                                          Reboot system on successful upgrade
```

## hydra build / eval

This cli makes requests against a hydra instances to check on individual jobs / builds to check for latest success, and discovers the associated flake from the builds evals. This currently only supports flakes, and does not support channels.

## health checks

Probably going to extend this to more options. These need to be converted to a fan-out / fan-in pattern and run concurrently when I implement more. Keeping it simple and concurrent for the first go with just ping.

### ICMP ping

Hosts specified with the `--canary` cli flag or `system.autoUpgradeHydra.healthChecks.canaryHosts` are pinged as a precondition for upgrade.

## config

A complete config looks something like:

```nix
{
  imports = [
    inputs.nixos-hydra-upgrade.nixosModules.nixos-hydra-upgrade
  ];

  system.autoUpgradeHydra = {
    enable = true;
    operation = "boot";
    host = "hostname";
    hydra = {
      instance = "https://your.hydra.example.com";
      project = "nix-config";
      jobset = "main";
      job = "hostname";
    };
    healthChecks = {
      canaryHosts = [
        # dependent service hostnames, urls, or ip addresses
        "yourcanaryhostname"
      ];
    };
    flags = [
      # any extra options you want to pass to `nixos-rebuild`
    ];
    # systemd.time#CALENDAR EVENTS
    dates = "*-*-* 04:40:00";
    allowReboot = false;
  };
}

```

Open to more full compatibility with `system.autoUpgrade`, but just starting with the stuff I actually use.
