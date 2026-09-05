# syntax=docker/dockerfile:1.7
FROM golang:1.27-alpine AS builder

# Install build dependencies (git, make, protoc and plugins for code generation)
RUN apk add --no-cache git make protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Copy go.mod and go.sum for dependency caching
WORKDIR /app
COPY go.mod go.sum ./

# Download dependencies (including private ones)
RUN --mount=type=secret,id=gh_token,required=true \
    git config --global url."https://$(cat /run/secrets/gh_token):x-oauth-basic@github.com/".insteadOf "https://github.com/" && \
    go mod download

# Copy the rest of the source code
COPY . .

# Clean up Go modules and vendor cache to reduce image size
RUN go mod tidy

# Generate gRPC and gqlgen code (own + external hub proto package)
RUN make generate

# Build the binary with optimizations for minimal size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/service ./cmd/server/main.go

# Final minimal image
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/service /app/service
EXPOSE 8080
ENTRYPOINT ["/app/service"]

# For local build:
#   echo "YOUR_GITHUB_PAT" > ~/.gh_token
# then:
#   docker buildx build --secret id=gh_token,src=$HOME/.gh_token -t my.app .
# For GitHub Actions:
#   build-push-action passes the secret via build-arg or secret

