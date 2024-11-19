# lotof.sample.gtw

Welcome to the **Lotof Sample Gateway** project! This project provides a structure to generate, build, and run a [GraphQL](https://graphql.org/) server using [gqlgen](https://github.com/99designs/gqlgen) and [Docker](https://www.docker.com/) for containerization. Below you'll find a comprehensive guide to get started with this project.

## Prerequisites

Before you begin, ensure you have the following installed on your machine:

- [Go](https://golang.org/doc/install) (version 1.23 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Makefile Targets

The project uses a Makefile to automate common tasks. Below are the available make targets and their functionalities:

### `make all`

- **Description**: Generates the necessary Go files for gqlgen and runs the server.
- **Commands**:
    - Runs the `gqlgen` target.
    - Executes the server via `make run`.

### `make gqlgen`

- **Description**: Generates the Go code required for your GraphQL server.
- **Commands**:
    - Executes `$(GQLGEN) generate` to generate GraphQL server code.
    - Cleans up dependencies with `go mod tidy`.

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
    - Removes `internal/graph/generated.go` and `internal/graph/model/models_gen.go`.

## Getting Started

1. Clone the repository:

   ```bash
   git clone https://github.com/pieceowater-dev/lotof.sample.gtw.git
   cd lotof.sample.gtw
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. **Generate GraphQL files** and start the server:

   ```bash
   make all
   ```

4. (Optional) To build Docker images for development or production:

   ```bash
   make build-dev    # For development
   make build-main   # For production
   ```

    To manage Docker services:

      ```bash
      make compose-up   # Start services
      make compose-down # Stop services
      ```

## Notes

- Customize `dev.dockerfile` and `main.dockerfile` as needed for your project's specific requirements.
- The server entry point is at `./cmd/server/main.go`.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author
![PCWT Dev Logo](https://avatars.githubusercontent.com/u/168465239?s=50)
### [PCWT Dev](https://github.com/pieceowater-dev)