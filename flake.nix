{
  description = "A Nix-flake-based Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix.url = "github:numtide/treefmt-nix";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , treefmt-nix
    }:

    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; };
      treefmtEval = treefmt-nix.lib.evalModule pkgs {
        projectRootFile = "flake.nix";
        programs.nixpkgs-fmt.enable = true;
        programs.gofmt.enable = true;
      };
    in
    {
      devShells.default = pkgs.mkShellNoCC {
        packages = with pkgs; [
          mosquitto
          go
          gotools
          golangci-lint
          inetutils
        ];
      };

      packages.default = pkgs.callPackage ./default.nix { };

      formatter = treefmtEval.config.build.wrapper;

      checks = {
        formatting = treefmtEval.config.build.check self;

        mqtt-exporter = self.packages.${system}.default;
      };
    });
}
