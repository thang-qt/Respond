{
  description = "Respond.im dev environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            nodejs
            pnpm
            postgresql
            go-migrate
            air
            ripgrep
            git
            hurl
          ];
          shellHook = ''
            export PATH="$PWD/scripts:$PATH"
            export PGHOST="/tmp"
            echo "Postgres helpers: db-start, db-stop"
          '';
        };
      }
    );
}
