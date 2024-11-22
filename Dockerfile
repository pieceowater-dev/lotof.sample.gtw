# Stage 1: Build the application and generate code
FROM golang:1.23 AS builder

WORKDIR /app

# Install dependencies
RUN apt-get update && \
    apt-get install -y unzip wget curl libc6 protobuf-compiler && \
    apt-get clean

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Install tools for code generation
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/99designs/gqlgen@latest

# Generate gRPC and GraphQL code
RUN make grpc-update
RUN make grpc-gen
RUN make gql-gen

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/server/main.go

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/app .

LABEL authors="pieceowater"

# Command to run the application
CMD ["./app"]

EXPOSE 8080

# Clean up unnecessary files to reduce image size
RUN rm -rf /app && \
    rm -rf /root/.cache