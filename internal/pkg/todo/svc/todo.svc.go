package svc

import (
	"app/internal/core/graph/model"
	pb "app/internal/core/grpc/generated"
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type TodoService struct {
	grpcAddress string
}

func NewTodoService() *TodoService {
	grpcAddress := "localhost:50051" // os.Getenv("LOTOF_SAMPLE_SVC_GRPC_ADDRESS")
	return &TodoService{grpcAddress: grpcAddress}
}

func (s *TodoService) Todos() ([]*model.Todo, error) {
	// Establish gRPC connection
	conn, err := grpc.Dial(s.grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("failed to connect to gRPC server: " + err.Error())
	}
	defer conn.Close()

	// Create gRPC client
	client := pb.NewTodoServiceClient(conn)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Send request
	req := &pb.GetTodosRequest{}
	res, err := client.GetTodos(ctx, req)
	if err != nil {
		log.Printf("Error calling GetTodos: %v", err)
		return nil, err
	}

	// Convert gRPC response to local model
	var todos []*model.Todo
	for _, t := range res.Todos {
		todos = append(todos, &model.Todo{
			ID:       t.Id,
			Text:     t.Text,
			Category: model.TodoCategoryEnum(t.Category.String()),
			Done:     t.Done,
		})
	}

	return todos, nil
}
