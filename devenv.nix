{
  pkgs,
  config,
  ...
}:

let
  homeDir = builtins.getEnv "HOME";
  colimaSocket = "${homeDir}/.config/colima/default/docker.sock";
in
{
  env = {
    DOCKER_HOST = "unix://${colimaSocket}";
    TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE = "/var/run/docker.sock";
  };

  packages = with pkgs; [
    # sql
    sqlc
    sqls
    goose

    # grpc
    protols
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc

    # templ
    templ
    tailwindcss

    # kubernetes
    kubectl
    kubernetes-helm
    tilt

    # ai stuff
    nodejs
    openspec
    playwright-test
    playwright-driver

    # formatter
    gofumpt
    sqruff
  ];

  languages.go = {
    enable = true;
    delve = {
      enable = true;
    };
    lsp = {
      enable = true;
    };
  };

  files = {
    ".sqruff".ini = {
      sqruff = {
        dialect = "postgres";
      };
    };
  };

  scripts = {
    generate-proto = {
      exec = ''
        cd "${config.git.root}"

        set -e
        PROTO_FILES=$(find services shared -type f -path '*/proto/*/v1/*.proto')
        if [ -z "$PROTO_FILES" ]; then
          echo "No proto files found"
          exit 0
        fi

        for file in $PROTO_FILES; do
          echo "Generating $file"
          protoc \
            --proto_path=. \
            --go_out=. --go_opt=paths=source_relative \
            --go-grpc_out=. --go-grpc_opt=paths=source_relative \
            "$file"
        done
      '';
    };

    tidy = {
      exec = ''
        echo "Tidying Go modules..."
        for dir in $(find shared services -name go.mod -exec dirname {} \; | sort); do
          echo "Tidying $dir..."
          (cd "$dir" && go mod tidy)
        done
      '';
    };

    sqlc-gen = {
      exec = ''
        echo "Bootstrap sql queries..."
        for dir in $(find services -maxdepth 2 -name sqlc.yaml -exec dirname {} \;); do
          (cd "$dir" && sqlc generate)
        done
      '';
    };
  };

  git-hooks.hooks = {
    treefmt.enable = true;
  };

  treefmt = {
    enable = true;
    config.programs = {
      nixfmt.enable = true;
      gofumpt = {
        enable = true;
        extra = true;
        excludes = [
          "vendor/*"
          "**/proto/*.go"
          "**/database/*.go"
          "*_templ.go"
        ];
      };
      sqruff.enable = true;
      oxfmt = {
        enable = true;
        excludes = [
          "*.css"
        ];
      };

    };
  };
}
