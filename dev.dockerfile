FROM golang:1.23 as dev
LABEL authors="pieceowater"
WORKDIR /app
COPY . .
RUN go mod download
RUN go run github.com/99designs/gqlgen generate
CMD ["go", "run", "./cmd/server/main.go"]