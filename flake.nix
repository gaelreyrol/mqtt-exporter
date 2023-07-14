{
  description = "A Nix-flake-based Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    }:

    flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs { inherit system; };
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

        vendorHash = "sha256-IGnZdyaq50Ja3LzCzruk19bPUgeN0wuN+tc6jk9Ck5A=";

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
    });
}