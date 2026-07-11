{
  pkgs,
  config,
  lib,
  ...
}:

let
  homeDir = builtins.getEnv "HOME";
  colimaSocket = "${homeDir}/.config/colima/k8s/docker.sock";
in
{
  dotenv.enable = true;

  env = {
    DOCKER_HOST = "unix://${colimaSocket}";
    TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE = "/var/run/docker.sock";
    DOPPLER_PROJECT = "refurbished-marketplace";
    DOPPLER_CONFIG = "dev";
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
    doppler

    # ai stuff
    nodejs
    openspec

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
        echo "Syncing go.work..."
        go work sync
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

  tasks = {
    "codegen:proto" = {
      exec = "generate-proto";
      before = [ "devenv:enterShell" ];

      execIfModified = [
        "services/**/proto/**/*.proto"
        "shared/**/proto/**/*.proto"
      ];
    };

    "codegen:sqlc" = {
      exec = "sqlc-gen";
      before = [ "devenv:enterShell" ];

      execIfModified = [
        "services/**/sqlc.yaml"
        "services/**/*.sql"
      ];
    };

    "go:tidy" = {
      exec = "tidy";
      before = [ "devenv:enterShell" ];

      execIfModified = [
        "services/**/go.mod"
        "shared/**/go.mod"
        "tools/**/go.mod"
        "go.work"
      ];
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
