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
          go
          gotools
          golangci-lint
        ];
      };

      packages.default = pkgs.buildGoModule {
        pname = "mqtt-exporter";
        version = "dev";

        src = ./.;

        # vendorHash = pkgs.lib.fakeHash;
        vendorHash = "sha256-SA2sjZfisHLpDm1820GToerHLbE1oQ2obl9pmsiyRqE=";

        ldflags = [
          "-s"
          "-w"
        ];

        meta = with pkgs.lib; {
          description = "Export MQTT messages to Promotheus";
          homepage = "https://github.com/gaelreyrol/mqtt-exporter";
          license = licenses.mit;
          maintainers = with maintainers; [ gaelreyrol ];
          mainProgram = "mqtt_exporter";
        };
      };

      formatter = treefmtEval.config.build.wrapper;

      checks = {
        formatting = treefmtEval.config.build.check self;
      };
    });
}
