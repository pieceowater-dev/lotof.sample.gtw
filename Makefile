GQLGEN = go run github.com/99designs/gqlgen
DOCKER_COMPOSE = docker-compose

.PHONY: all generate run clean

all: gqlgen run

gqlgen:
	$(GQLGEN) generate && go mod tidy

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

clean:
	rm -rf internal/graph/generated.go internal/graph/model/models_gen.go