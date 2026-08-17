# Application name and build directory
APP_NAME = lotof.sample.gtw
BUILD_DIR = bin
MAIN_FILE = cmd/server/main.go

# Protobuf compiler and plugins
PROTOC = protoc
PROTOC_GEN_GO = $(GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GRPC_GO = $(GOPATH)/bin/protoc-gen-go-grpc
PROTOC_PKG = github.com/pieceowater-dev/lotof.sample.proto
EXTERNAL_PROTOC_PKG = github.com/pieceowater-dev/lotof.hub.proto
PROTOC_PKG_PATH = $(shell go list -m -f '{{.Dir}}' $(PROTOC_PKG))
EXTERNAL_PROTOC_PKG_PATH = $(shell go list -m -f '{{.Dir}}' $(EXTERNAL_PROTOC_PKG))
PROTOC_DIR = protos
PROTOC_OUT_DIR = ./internal/core/grpc/generated
# Auto-detect system WKT include path (present on Alpine after protobuf-dev, absent on macOS where homebrew handles it)
PROTOC_WKT_INCLUDE = $(shell [ -f /usr/include/google/protobuf/timestamp.proto ] && echo "-I /usr/include" || echo "")

# GQLGEN tool for GraphQL code generation
GQLGEN = go run github.com/99designs/gqlgen

# Docker Compose tool
DOCKER_COMPOSE = docker-compose

.PHONY: all clean build run update grpc-gen grpc-gen-external grpc-clean grpc-update gql-gen gql-clean generate test compose-up compose-down

# Setup the environment by updating gRPC dependencies
setup: grpc-update
	@echo "Setup completed!"; \
	go mod tidy

# Default build target
all: build grpc-gen gql-gen

# Update Go module dependencies
update:
	go mod tidy

# Build the project
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

# Run the application
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) gql-clean grpc-clean

# Alias used by CI to run all code generation
generate: grpc-gen grpc-gen-external gql-gen

# Build check (no tests yet)
test:
	go build ./...

# gRPC code generation from proto files
grpc-gen:
	@echo "Generating gRPC code from proto files..."
	mkdir -p $(PROTOC_OUT_DIR)
	find $(PROTOC_PKG_PATH)/$(PROTOC_DIR) -name "*.proto" | xargs $(PROTOC) \
		-I $(PROTOC_PKG_PATH)/$(PROTOC_DIR) \
		$(PROTOC_WKT_INCLUDE) \
		--go_out=paths=source_relative:$(PROTOC_OUT_DIR) \
		--go-grpc_out=paths=source_relative:$(PROTOC_OUT_DIR)
	@echo "gRPC code generation completed!"

# gRPC code generation from external (hub) proto files — needed for auth token exchange
grpc-gen-external:
	@echo "Generating gRPC code from external proto files..."
	mkdir -p $(PROTOC_OUT_DIR)
	find $(EXTERNAL_PROTOC_PKG_PATH)/$(PROTOC_DIR) -name "*.proto" | xargs $(PROTOC) \
		-I $(EXTERNAL_PROTOC_PKG_PATH)/$(PROTOC_DIR) \
		$(PROTOC_WKT_INCLUDE) \
		--go_out=paths=source_relative:$(PROTOC_OUT_DIR) \
		--go-grpc_out=paths=source_relative:$(PROTOC_OUT_DIR)
	@echo "External gRPC code generation completed!"

# Clean gRPC generated files
grpc-clean:
	rm -rf $(PROTOC_OUT_DIR)

# Update gRPC dependencies
grpc-update:
	go clean -modcache
	go get -u $(PROTOC_PKG)@latest

# GQLGEN code generation
gql-gen:
	$(GQLGEN) generate
	git add -A

# Clean GQLGEN generated files
gql-clean:
	rm -rf internal/core/graph

# Build Docker image
build-docker:
	docker build -t $(APP_NAME) .

# Build Docker image and run the container
build-and-run-docker: build-docker
	docker stop $(APP_NAME)
	docker rm $(APP_NAME)
	docker run -d -p 8080:8080 \
		--network lotofsamplegtw_pieceonetwork \
		--name $(APP_NAME) \
		$(APP_NAME)

# Start Docker Compose services
compose-up:
	$(DOCKER_COMPOSE) up -d

# Stop Docker Compose services
compose-down:
	$(DOCKER_COMPOSE) down
