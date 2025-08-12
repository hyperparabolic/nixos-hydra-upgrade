{
  description = "NixOS system upgrades on hydra build success";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = {
    self,
    nixpkgs,
    systems,
    ...
  }: let
    forEachSystem = nixpkgs.lib.genAttrs (import systems);
    pkgsFor = forEachSystem (system: import nixpkgs {inherit system;});

    pname = "nixos-hydra-upgrade";
  in {
    devShells = forEachSystem (system: let
      pkgs = pkgsFor.${system};
    in {
      default = pkgsFor.${system}.mkShell {
        packages = with pkgs; [
          go
          gopls
          go-tools
          gotools
        ];
      };
    });

    nixosModules = {
      default = self.nixosModules.nixos-hydra-upgrade;
      nixos-hydra-upgrade = import ./modules/nixos-hydra-upgrade;
    };

    packages = forEachSystem (system: let
      pkgs = pkgsFor.${system};
      version = "v0.2.0";
    in {
      default = pkgsFor.${system}.buildGoModule {
        inherit pname version;
        src = ./.;
        vendorHash = "sha256-fe0ZKSyRkenh2dph7jFIpwrNT8auqu69iFv2JjSgMlo=";

        ldflags = [
          "-X 'github.com/hyperparabolic/nixos-hydra-upgrade/cmd.Version=${version}'"
        ];

        meta = {
          homepage = "https://github.com/hyperparabolic/nixos-hydra-upgrade";
          description = "nixos upgrader that queries hydra and performs health checks";
          license = pkgs.lib.licenses.mit;
          mainProgram = "nixos-hydra-upgrade";
          platforms = pkgs.lib.platforms.linux;
        };

        nativeBuildInputs = with pkgs; [
          installShellFiles
          go
        ];

        postInstall = pkgs.lib.optionalString (pkgs.stdenv.buildPlatform.canExecute pkgs.stdenv.hostPlatform) ''
          installShellCompletion --cmd nixos-hydra-upgrade \
            --bash <($out/bin/nixos-hydra-upgrade completion bash) \
            --fish <($out/bin/nixos-hydra-upgrade completion fish) \
            --zsh <($out/bin/nixos-hydra-upgrade completion zsh)

          $out/bin/nixos-hydra-upgrade docs man > nixos-hydra-upgrade.1
          installManPage nixos-hydra-upgrade.1
        '';
      };
    });
  };
}
