{
  description = "Go Enterprise APIs for Circular";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
      };
      go-pkg = pkgs.go_1_22;
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = [
          go-pkg
          pkgs.gopls
        ];
      };
    };
}