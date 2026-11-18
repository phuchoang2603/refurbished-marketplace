{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:

let
  services = [
    "cart"
    "checkout"
    "inventory"
    "orders"
    "products"
    "search"
    "users"
    "web"
  ];
  sharedDirs = [
    "auth"
    "err"
    "messaging"
    "observe"
    "proto"
    "runtime"
    "testutil"
  ];
  modulePath = "github.com/phuchoang2603/refurbished-marketplace";
  isolateService = service: {
    files = [ "**/services/${service}/**" ];
    list-mode = "lax";
    deny = map (other: {
      pkg = "${modulePath}/services/${other}";
      desc = "services must not import other services";
    }) (lib.remove service services);
  };
in

{
  imports = map (m: "${inputs.devenv-modules}/${m}.nix") [
    "base"
    "go"
    "sqlc"
    "proto"
    "kubernetes"
  ];

  env = {
    DOPPLER_PROJECT = "refurbished-marketplace";
    DOPPLER_CONFIG = "prd";
  };

  packages = with pkgs; [

    templ
    tailwindcss

    doppler
    (govulncheck.override {
      buildGoLatestModule = buildGoModule.override { go = config.languages.go.package; };
    })
    openspec

    sqruff
  ];

  enterShell = ''
    colima_socket="$HOME/.config/colima/default/docker.sock"
    if [ -S "$colima_socket" ]; then
      export DOCKER_HOST="unix://$colima_socket"
      export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
    fi
  '';

  repo.golangci.linters = {
    enable = [ "depguard" ];
    settings.depguard.rules =
      lib.genAttrs' services (service: lib.nameValuePair "service-${service}" (isolateService service))
      // {
        shared-isolated = {
          files = map (dir: "**/shared/${dir}/**") sharedDirs;
          list-mode = "lax";
          deny = [
            {
              pkg = "${modulePath}/services";
              desc = "shared must not import services";
            }
          ];
        };
      };
  };

  files = {
    ".sqruff".ini = {
      sqruff = {
        dialect = "postgres";
      };
    };
  };

  tasks = {
    "devenv:treefmt:run".before = lib.mkForce [ ];

    "go:tidy" = {
      exec = ''
        cd "${config.git.root}"
        echo "Syncing go.work..."
        go work sync
      '';

      execIfModified = [
        "services/**/go.mod"
        "shared/**/go.mod"
        "tools/**/go.mod"
        "go.work"
      ];
    };

    "codegen:templ" = {
      exec = ''
        cd "${config.git.root}/services/web"
        templ generate
      '';

      execIfModified = [
        "services/web/**/*.templ"
      ];
    };

    "codegen:tailwind" = {
      exec = ''
        cd "${config.git.root}/services/web"
        tailwindcss -c tailwind.config.js -i tailwind.css -o static/app.css
      '';

      execIfModified = [
        "services/web/tailwind.css"
        "services/web/tailwind.config.js"
        "services/web/**/*.templ"
      ];
    };
  };

  treefmt = {
    enable = true;
    config.programs = {
      gofumpt = {
        excludes = [
          "vendor/*"
          "**/database/*.go"
          "*_templ.go"
        ];
      };
      sqruff.enable = true;
      oxfmt = {
        enable = true;
        excludes = [
          "*.css"
          "**/templates/*.yaml"
          "**/templates/*.tpl"
        ];
      };

    };
  };
}
