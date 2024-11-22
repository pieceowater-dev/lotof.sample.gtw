GQLGEN = go run github.com/99designs/gqlgen
PROTOC = protoc
PROTOC_GEN_GO = $(GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GRPC_GO = $(GOPATH)/bin/protoc-gen-go-grpc
PROTOC_DIR = ./internal/core/grpc/protos/*/*.proto
PROTOC_OUT_DIR = ./internal/core/grpc/generated
DOCKER_COMPOSE = docker-compose

.PHONY: all generate run clean

all: gqlgen grpcgen run



gqlgen:
	$(GQLGEN) generate && go mod tidy && git add *

gqlclean:
	rm -rf internal/graph/generated.go internal/graph/model/models_gen.go



grpcgen: grpcclean
	mkdir -p $(PROTOC_OUT_DIR)
	$(PROTOC) --go_out=$(PROTOC_OUT_DIR) --go-grpc_out=$(PROTOC_OUT_DIR) $(PROTOC_DIR)

grpcclean:
	rm -rf internal/graph/generated.go internal/graph/model/models_gen.go $(PROTOC_OUT_DIR)



run:
	go run ./cmd/server/main.go



build-dev:
	docker build -f dev.dockerfile -t gateway-dev .

build-main:
	docker build -f main.dockerfile -t gateway-prod .



compose-up:
	$(DOCKER_COMPOSE) up -d

compose-down:
	$(DOCKER_COMPOSE) down
