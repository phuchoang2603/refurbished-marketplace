{
  description = "Development shell for refurbished marketplace";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems =
        f:
        nixpkgs.lib.genAttrs systems (
          system:
          f (
            import nixpkgs {
              inherit system;
              config.allowUnfree = true;
            }
          )
        );
    in
    {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gofumpt
            golangci-lint
            delve

            goose
            sqlc

            protobuf
            protoc-gen-go
            protoc-gen-go-grpc

            kubectl
            kubernetes-helm
            tilt

            # confluent-kafka-go (CGO): vendored librdkafka still links against libsasl2, openssl, etc.
            pkg-config
            rdkafka
            cyrus_sasl
            openssl
          ];

          shellHook = ''
            export GOPATH="$HOME/go"
            export PATH="$GOPATH/bin:$PATH"

            # Help CGO find SASL / TLS libs (fixes: ld: library not found for -lsasl2).
            export PKG_CONFIG_PATH=${
              pkgs.lib.makeSearchPath "lib/pkgconfig" [
                "${pkgs.rdkafka.dev}"
                "${pkgs.cyrus_sasl.dev}"
                "${pkgs.openssl.dev}"
              ]
            }
            export CGO_CFLAGS="-I${pkgs.rdkafka.dev}/include"
            export CGO_LDFLAGS="-L${pkgs.cyrus_sasl.out}/lib -lsasl2 -L${pkgs.openssl.out}/lib -lssl -lcrypto"
          '';
        };
      });
    };
}
