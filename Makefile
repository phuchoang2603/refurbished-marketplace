.PHONY: generate-proto tidy

generate-proto:
	@set -e; \
	PROTO_FILES=$$(find services shared -type f -path '*/proto/*/v1/*.proto'); \
	if [ -z "$$PROTO_FILES" ]; then \
		echo "No proto files found"; \
		exit 0; \
	fi; \
	for file in $$PROTO_FILES; do \
		echo "Generating $$file"; \
		protoc \
			--proto_path=. \
			--go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			"$$file"; \
	done

tidy:
	@echo "Tidying shared module..."
	@cd shared && go mod tidy
	@echo "Tidying service modules..."
	@for dir in $$(find services -maxdepth 2 -name go.mod -exec dirname {} \;); do \
		echo "Tidying $$dir..."; \
		(cd $$dir && go mod tidy); \
	done
