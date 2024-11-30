package svc

import (
	"app/internal/core/cfg"
	"app/internal/core/graph/model"
	pb "app/internal/core/grpc/generated"
	"context"
	"errors"
	"github.com/pieceowater-dev/lotof.lib.gossiper/v2"
	"log"
)

type TodoService struct {
	transport gossiper.Transport
	client    pb.TodoServiceClient // Add client as a property
}

func NewTodoService() *TodoService {
	factory := gossiper.NewTransportFactory()
	grpcTransport := factory.CreateTransport(
		gossiper.GRPC,
		cfg.Inst().LotofSampleSvcGrpcAddress,
	)

	// Create the client only once and store it as a property
	clientConstructor := pb.NewTodoServiceClient
	client, err := grpcTransport.CreateClient(clientConstructor)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	return &TodoService{
		transport: grpcTransport,
		client:    client.(pb.TodoServiceClient), // Cast to the correct type
	}
}

func (s *TodoService) Todos() ([]*model.Todo, error) {
	ctx := context.Background()
	request := &pb.GetTodosRequest{} // Dynamic request for GetTodos

	// Send the request using the client stored in the TodoService instance
	response, err := s.transport.Send(ctx, s.client, "GetTodos", request)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	// Assert the response to the correct type
	res, ok := response.(*pb.GetTodosResponse)
	if !ok {
		return nil, errors.New("invalid response type from gRPC transport")
	}

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
