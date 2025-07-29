{
  inputs,
  config,
  lib,
  pkgs,
}: let
  cfg = config.system.autoUpgradeHydra;
  nixosHydraUpgradePackages = inputs.nixos-hydra-upgrade.packages.${pkgs.system};
in {
  /*
  Automatic system upgrades with nixos-hydra-upgrade.

  A lot of this is a drop-in replacement for `system.autoUpgrade`, but a few options are
  dropped and a few new are introduced:
  - Channels are not supported.
  - The flake option is replaced with a hydra job that discovers the flake.
  - Delay options aren't implemented (not against supporting them, but I don't use them).
  - rebootWindow isn't implemented (ditto)
  - persistent isn't implemented (ditto)
  */
  options = {
    system.autoUpgradeHydra = {
      enable = lib.mkEnableOption "Whether to perform system upgrades with nixos-hydra-upgrade.";

      operation = lib.mkOption {
        type = lib.types.enum [
          "switch"
          "boot"
        ];
        default = "switch";
        description = lib.mdDoc "{command}`nixos-rebuild` operation to execute";
      };

      host = lib.mkOption {
        type = lib.types.str;
        description = lib.mdDoc "system hostname";
        default = config.networking.hostName;
      };

      hydra = lib.mkOption {
        description = lib.mdDoc ''
          Options to specify the hydra build.
        '';
        type = lib.types.submodule {
          options = {
            instance = lib.mkOption {
              type = lib.types.str;
              example = "https://hydra.oak.decent.id";
              description = lib.mdDoc "hydra instance to query";
            };
            project = lib.mkOption {
              type = lib.types.str;
              description = lib.mdDoc "hydra project";
            };
            jobset = lib.mkOption {
              type = lib.types.str;
              description = lib.mdDoc "hydra jobset";
            };
            job = lib.mkOption {
              type = lib.types.str;
              description = lib.mdDoc "hydra job";
            };
          };
        };
      };

      healthChecks = lib.mkOption {
        description = lib.mdDoc ''
          Options to specify health checks to perform before system upgrade.
        '';

        type = lib.types.submodule {
          options = {
            canaryHosts = {
              type = lib.types.listOf lib.types.str;
              default = [];
              description = lib.mdDoc ''
                Hosts specified here are checked with a simple ICMP ping
                healthcheck. If all hosts do not reply the upgrade will
                be aborted. Probably does not properly escape every valid
                url.
              '';
            };
          };
        };
      };

      flags = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [];
        example = [
          "--accept-flake-config"
        ];
        description = lib.mdDoc "Additional flags to pass to {command}`nixos-rebuild`";
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

      allowReboot = lib.mkOption {
        default = false;
        type = lib.types.bool;
        description = lib.mdDoc ''
          Wether to reboot the system following a successful {command}`nixos-rebuild`.
        '';
      };
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.nixos-hydra-upgrade = {
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

      path = with pkgs; [
        config.nix.package
        # config options for this are undocumented and presumably unstable?
        # doesn't support development versions of nixos-rebuild I guess.
        nixos-rebuild
      ];

      script = let
        nixos-hydra-upgrade = "${lib.getExe nixosHydraUpgradePackages.default}";
        reboot-arg = lib.optionalString cfg.allowReboot "--reboot";
        canary-args =
          lib.optionalString
          (builtins.length cfg.healthChecks.canaryHosts > 0)
          "${(lib.strings.concatMapStringsSep " " (x: "--canary=\"" + x + "\"") cfg.healthChecks.canaryHosts)}";
        nixos-rebuild-args =
          lib.optionalString
          (builtins.length cfg.flags > 0)
          "${(lib.strings.concatMapStringsSep " " (x: "--passthru-args=\"" + x + "\"") cfg.flags)}";
      in ''
        ${nixos-hydra-upgrade} ${cfg.operation} \
          --host="${cfg.host}" \
          --instance="${cfg.hydra.instance}" \
          --project="${cfg.hydra.project}" \
          --jobset="${cfg.hydra.jobset}" \
          --job="${cfg.hydra.job}" ${canary-args} ${nixos-rebuild-args} ${reboot-arg}
      '';

      startAt = cfg.dates;

      after = ["network-online.target"];
      wants = ["network-online.target"];
    };
  };
}
