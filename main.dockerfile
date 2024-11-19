FROM golang:1.23 as dev
LABEL authors="pieceowater"

WORKDIR /app
COPY . .
RUN go mod download
RUN go run github.com/99designs/gqlgen generate
RUN go build -o /app ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app .
CMD ["./app"]