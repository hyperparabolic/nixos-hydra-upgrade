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
    in {
      default = pkgsFor.${system}.buildGoModule {
        inherit pname;
        version = "0.0.1";
        src = ./.;
        vendorHash = "sha256-kc/b/DrghxC2k3T+pana7k5f3Hh1Uyxk/TVxTNuHZbY=";

        meta = {
          homepage = "https://github.com/hyperparabolic/nixos-hydra-upgrade";
          description = "nixos upgrader that queries hydra and performs health checks";
          license = pkgs.lib.licenses.mit;
          mainProgram = "nixos-hydra-upgrade";
          platforms = pkgs.lib.platforms.linux;
        };
      };
    });
  };
}
