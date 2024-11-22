# lotof.sample.gateway

Welcome to the **Lotof Sample Gateway** project! This project provides a structure to generate, build, and run a [GraphQL](https://graphql.org/) server using [gqlgen](https://github.com/99designs/gqlgen) and [Docker](https://www.docker.com/) for containerization. Below you'll find a comprehensive guide to get started with this project.

## Prerequisites

Before you begin, ensure you have the following installed on your machine:

- [Go](https://golang.org/doc/install) (version 1.23 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Makefile Targets

The project uses a Makefile to automate common tasks. Below are the available make targets and their functionalities:

### `make all`

- **Description**: Generates the necessary Go files for gqlgen, gRPC, and runs the server.
- **Commands**:
    - Executes the `gqlgen` and `grpcgen` targets.
    - Starts the server using `make run`.

### `make gql-gen`

- **Description**: Generates the Go code required for your GraphQL server.
- **Commands**:
    - Executes `$(GQLGEN) generate` to generate GraphQL server code.
    - Cleans up dependencies with `go mod tidy`.

### `make grpc-gen`

- **Description**: Generates Go stubs for gRPC services.
- **Commands**:
    - Uses `$(PROTOC)` with plugins `protoc-gen-go` and `protoc-gen-go-grpc`.
    - Places the generated files in `./internal/core/grpc/generated`.

### `make run`

- **Description**: Runs the GraphQL server.
- **Commands**:
    - Executes `go run ./cmd/server/main.go` to start the server.

### `make build-dev`

- **Description**: Builds a Docker image for the development environment.
- **Commands**:
    - Builds Docker image using `dev.dockerfile` and names it `gateway-dev`.

### `make build-main`

- **Description**: Builds a Docker image for the production environment.
- **Commands**:
    - Builds Docker image using `main.dockerfile` and names it `gateway-prod`.

### `make compose-up`

- **Description**: Starts the services defined in `docker-compose.yml`.
- **Commands**:
    - Runs `$(DOCKER_COMPOSE) up -d` to start services in detached mode.

### `make compose-down`

- **Description**: Stops and removes the services defined in `docker-compose.yml`.
- **Commands**:
    - Runs `$(DOCKER_COMPOSE) down` to stop the services.

### `make clean`

- **Description**: Cleans up generated files.
- **Commands**:
    - Removes the files in `internal/core/grpc/generated` and GraphQL-related generated files.

## Getting Started

1. Clone the repository:

   ```bash
   git clone https://github.com/pieceowater-dev/lotof.sample.gtw.git
   cd lotof.sample.gtw
   ```

2.	Install dependencies:

   ```bash
   go mod tidy
   ```

3. **Generate GraphQL files** and start the server:

   ```bash
   make all
   ```

4. (Optional) To build Docker images for development or production:

   ```bash
   make build
   ```

5.	(Optional) To manage Docker services:

      ```bash
      make compose-up   # Start services
      make compose-down # Stop services
      ```

## Notes

- Customize `Dockerfile` as needed for your project's specific requirements.
- The server entry point is at `./cmd/server/main.go`.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author
![PCWT Dev Logo](https://avatars.githubusercontent.com/u/168465239?s=50)
### [PCWT Dev](https://github.com/pieceowater-dev)