APP_NAME = lotof.sample.gtw
BUILD_DIR = bin
MAIN_FILE = cmd/server/main.go
PROTOC = protoc
PROTOC_GEN_GO = $(GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GRPC_GO = $(GOPATH)/bin/protoc-gen-go-grpc
PROTOC_PKG = github.com/pieceowater-dev/lotof.sample.proto
PROTOC_PKG_PATH = $(shell go list -m -f '{{.Dir}}' $(PROTOC_PKG))
PROTOC_DIR = protos
PROTOC_OUT_DIR = ./internal/core/grpc/generated
GQLGEN = go run github.com/99designs/gqlgen
DOCKER_COMPOSE = docker-compose

.PHONY: all clean build run update grpc-gen grpc-clean grpc-update gql-gen gql-clean compose-up compose-down

# Setup the environment
setup: grpc-update
	@echo "Setup completed!"; \
	go mod tidy

# Default build target
all: build grpc-gen gql-gen

# Update dependencies
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

# gRPC code generation
grpc-gen:
	mkdir -p $(PROTOC_OUT_DIR)
	$(PROTOC) \
		-I $(PROTOC_PKG_PATH)/$(PROTOC_DIR) \
		--go_out=$(PROTOC_OUT_DIR) \
		--go-grpc_out=$(PROTOC_OUT_DIR) \
		$(PROTOC_PKG_PATH)/$(PROTOC_DIR)/*/*/*.proto

# Clean gRPC generated files
grpc-clean:
	rm -rf $(PROTOC_OUT_DIR)

# Update gRPC dependencies
grpc-update:
	go get -u $(PROTOC_PKG)@latest

# GQLGEN generation
gql-gen:
	$(GQLGEN) generate
	git add -A

# Clean GQLGEN generated files
gql-clean:
	rm -rf internal/graph/generated.go internal/graph/model/models_gen.go

# Docker build target
build-docker:
	docker build -t $(APP_NAME) .

# Build Docker image and run the container
build-and-run-docker: build-docker
	docker stop $(APP_NAME)
	docker rm $(APP_NAME)
	docker run -d -p 8080:8080 \
		--network lotofsamplesvc_pieceonetwork \
		--name $(APP_NAME) \
		$(APP_NAME)

# Start Docker Compose services
compose-up:
	$(DOCKER_COMPOSE) up -d

# Stop Docker Compose services
compose-down:
	$(DOCKER_COMPOSE) down