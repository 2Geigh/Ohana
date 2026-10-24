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
          nodejs
          python3
        ];

        shellHook = ''
          echo "Welcome to Ohana!"
        '';
      };
    });
  };
}
