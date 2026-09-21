{
  description = "myOps - unified web interface for self-hosted Gitea and Drone CI/CD";

  # The NixOS channel snapshot. flake.lock pins the exact revision and content
  # hash, so this stays reproducible; `nix flake update` advances it.
  inputs.nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";

  outputs =
    { nixpkgs, ... }:
    let
      # Platforms the development shell is offered on. macOS is the primary
      # target; Linux is kept so CI can reuse the same shell.
      systems = [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # Go toolchain and editor/CI tooling.
            go
            gopls
            golangci-lint
            gotools

            # Live reload used by `make dev`.
            air

            # Tailwind CSS v4 standalone CLI. The project deliberately avoids
            # npm, so this replaces `npx @tailwindcss/cli`.
            tailwindcss_4

            # Build prerequisites: Makefile, git metadata, TLS roots for the
            # Go module proxy.
            gnumake
            git
            cacert
          ];

          # Pin Go to the toolchain that Nix installed instead of letting it
          # download a different one at build time.
          GOTOOLCHAIN = "local";

          shellHook = ''
            echo "myOps development shell"
            echo "  go           $(go version | cut -d' ' -f3)"
            echo "  tailwindcss  $(tailwindcss --help 2>/dev/null | head -n1 | tr -d '≈ ')"
            echo "  golangci-lint $(golangci-lint --version 2>/dev/null | head -n1 | awk '{print $4}')"
            echo ""
            echo "  make dev      run the server with hot reload"
            echo "  make build    build Tailwind CSS and the Go binary"
            echo "  make lint     run golangci-lint"
          '';
        };
      });

      # `nix fmt` runs the formatter from the flake root with no arguments, and
      # bare `nixfmt` reads stdin instead of discovering files. Find the files
      # ourselves so a plain `nix fmt` does the expected thing.
      formatter = forAllSystems (
        pkgs:
        pkgs.writeShellScriptBin "nixfmt-tree" ''
          set -euo pipefail
          if [ "$#" -gt 0 ]; then
            exec ${pkgs.nixfmt}/bin/nixfmt "$@"
          fi
          mapfile -t files < <(find . -name '*.nix' -not -path './.git/*' -not -path './.direnv/*')
          if [ "''${#files[@]}" -eq 0 ]; then
            echo "nixfmt-tree: no .nix files found" >&2
            exit 0
          fi
          exec ${pkgs.nixfmt}/bin/nixfmt "''${files[@]}"
        ''
      );
    };
}
