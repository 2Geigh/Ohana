{
  description = "Development Dependencies for Ohana: The Homepage of the Living Web";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";
  };

  outputs = {nixpkgs, ...}: let
    systems = ["aarch64-darwin" "x86_64-linux"];
  in {
    # nixpkgs.lib.genAttrs -> Define devShells as an attribute set whos attributes are each `system` in `systems`
    devShells = nixpkgs.lib.genAttrs systems (system: {
      default = nixpkgs.legacyPackages.${system}.mkShell {
        packages = with nixpkgs.legacyPackages.${system}; [
          beamPackages.elixir
          beamPackages.erlang
          cargo
          docker
          docker-compose
          go_1_27
          jdk25
          maven
          nodejs
          php
          pnpm
          python3
          rustc
          uv
        ];

        shellHook = ''
          echo "Welcome to Ohana!"
        '';
      };
    });
  };
}
