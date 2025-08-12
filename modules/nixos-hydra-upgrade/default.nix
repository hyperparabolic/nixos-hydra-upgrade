{
  inputs,
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.system.autoUpgradeHydra;
  nixosHydraUpgradePackages = inputs.nixos-hydra-upgrade.packages.${pkgs.system};
  settingsFormat = pkgs.formats.yaml {};
in {
  options = {
    system.autoUpgradeHydra = {
      enable = lib.mkEnableOption "Whether to perform system upgrades with nixos-hydra-upgrade.";

      environmentFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          Environment file (see {manpage}`systemd.exec(5)`
          "EnvironmentFile=" section for the syntax) to define service environment variables.
          This option may be used to safely include secrets without exposure in the nix store.

          See [usage](https://github.com/hyperparabolic/nixos-hydra-upgrade/blob/main/README.md#usage)
          for environment variable details.
        '';
      };

      dates = lib.mkOption {
        type = lib.types.str;
        default = "04:40";
        example = "daily";
        description = ''
          How often or when an upgrade occurs.

          The format is described in
          {manpage}`systemd.time(7)`.
        '';
      };

      settings = lib.mkOption {
        description = ''
          Configuration for nixos-hydra-upgrade, see [usage](https://github.com/hyperparabolic/nixos-hydra-upgrade/blob/main/README.md#usage)
          for details on all supported values.
        '';
        type = lib.types.submodule {
          freeformType = settingsFormat.type;
          options = {
            debug = lib.mkOption {
              type = lib.types.bool;
              description = "enable debug logging";
              default = false;
            };
            healthChecks = lib.mkOption {
              description = ''
                Options to specify health checks to perform before system upgrade.
              '';
              type = lib.types.submodule {
                freeformType = settingsFormat.type;
                options = {
                  canaryHosts = lib.mkOption {
                    type = lib.types.listOf lib.types.str;
                    default = [];
                    description = ''
                      Hosts specified here are checked with a simple ICMP ping
                      healthcheck. If all hosts do not reply the upgrade will
                      be aborted. Probably does not properly escape every valid
                      url.
                    '';
                  };
                };
              };
            };
            hydra = lib.mkOption {
              description = ''
                Options to specify the hydra build.
              '';
              type = lib.types.submodule {
                freeformType = settingsFormat.type;
                options = {
                  instance = lib.mkOption {
                    type = lib.types.str;
                    example = "https://hydra.oak.decent.id";
                    description = "hydra instance to query";
                  };
                  project = lib.mkOption {
                    type = lib.types.str;
                    description = "hydra project";
                  };
                  jobset = lib.mkOption {
                    type = lib.types.str;
                    description = "hydra jobset";
                  };
                  job = lib.mkOption {
                    type = lib.types.str;
                    description = "hydra job";
                  };
                };
              };
            };
            nixos-rebuild = lib.mkOption {
              description = ''
                Options to pass to `nixos-rebuild`.
              '';
              type = lib.types.submodule {
                freeformType = settingsFormat.type;
                options = {
                  operation = lib.mkOption {
                    type = lib.types.enum [
                      "switch"
                      "boot"
                    ];
                    default = "boot";
                    description = "{command}`nixos-rebuild` operation to execute";
                  };
                  host = lib.mkOption {
                    type = lib.types.str;
                    description = "system hostname";
                    default = config.networking.hostName;
                  };
                  args = lib.mkOption {
                    type = lib.types.listOf lib.types.str;
                    default = [];
                    example = [
                      "--accept-flake-config"
                    ];
                    description = "Additional flags to pass to {command}`nixos-rebuild`";
                  };
                };
              };
            };
            reboot = lib.mkOption {
              default = false;
              type = lib.types.bool;
              description = ''
                Wether to reboot the system following a successful {command}`nixos-rebuild`.
              '';
            };
          };
        };
        default = {};
      };
    };
  };

  config = lib.mkIf cfg.enable {
    environment.etc."nixos-hydra-upgrade" = {
      mode = "0440";
      source = settingsFormat.generate "nixos-hydra-upgrade.yaml" cfg.settings;
      target = "nixos-hydra-upgrade/config.yaml";
    };
    systemd.services.nixos-hydra-upgrade =
      {
        description = "NixOS Upgrade with hydra build validation and health check support.";

        restartIfChanged = false;
        unitConfig.X-StopOnRemoval = false;
        serviceConfig.Type = "oneshot";

        environment =
          config.nix.envVars
          // {
            inherit (config.environment.sessionVariables) NIX_PATH;
            HOME = "/root";
          }
          // config.networking.proxy.envVars;

        path = [
          config.nix.package
          config.system.build.nixos-rebuild
        ];

        script = "${lib.getExe nixosHydraUpgradePackages.default}";

        startAt = cfg.dates;

        after = ["network-online.target"];
        wants = ["network-online.target"];
      }
      // lib.optionalAttrs (cfg.environmentFile != null) {
        EnvironmentFile = cfg.environmentFile;
      };
  };
}
