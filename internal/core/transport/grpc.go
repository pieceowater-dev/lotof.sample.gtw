package transport

import (
	"context"
	"errors"
	"time"

	pb "app/internal/core/grpc/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCTransport struct {
	address string
}

func NewGRPCTransport(address string) *GRPCTransport {
	return &GRPCTransport{address: address}
}

func (g *GRPCTransport) Send(ctx context.Context, request interface{}) (interface{}, error) {
	conn, err := grpc.Dial(g.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("failed to connect to gRPC server: " + err.Error())
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	client := pb.NewTodoServiceClient(conn)

	req, ok := request.(*pb.GetTodosRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := client.GetTodos(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
