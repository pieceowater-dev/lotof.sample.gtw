package svc

import (
	"app/internal/core/cfg"
	"app/internal/core/graph/model"
	pb "app/internal/core/grpc/generated"
	"app/internal/core/transport"
	"context"
	"errors"
	"log"
)

type TodoService struct {
	transport transport.Transport
}

func NewTodoService() *TodoService {
	factory := transport.NewFactory()
	grpcTransport := factory.CreateTransport(
		transport.GRPC,
		cfg.Inst().LotofSampleSvcGrpcAddress,
	)
	return &TodoService{
		transport: grpcTransport,
	}
}

func (s *TodoService) Todos() ([]*model.Todo, error) {
	ctx := context.Background()
	request := &pb.GetTodosRequest{}
	response, err := s.transport.Send(ctx, request)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

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
