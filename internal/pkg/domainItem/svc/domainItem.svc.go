package svc

import (
	"app/internal/core/cfg"
	"app/internal/core/graph/model"
	pb "app/internal/core/grpc/generated/lotof.sample.proto/lotof.sample.svc/domainItem"
	"context"
	"errors"
	"github.com/pieceowater-dev/lotof.lib.gossiper/v2"
	"log"
)

// DomainItemService handles the operations related to domain items.
type DomainItemService struct {
	transport gossiper.Transport
	client    pb.DomainItemServiceClient // gRPC client for domain item service.
}

// NewDomainItemService creates a new DomainItemService with the necessary transport and client.
func NewDomainItemService() *DomainItemService {
	factory := gossiper.NewTransportFactory()
	grpcTransport := factory.CreateTransport(
		gossiper.GRPC,
		cfg.Inst().LotofSampleSvcGrpcAddress,
	)

	// Create the client only once and store it as a property.
	clientConstructor := pb.NewDomainItemServiceClient
	client, err := grpcTransport.CreateClient(clientConstructor)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	return &DomainItemService{
		transport: grpcTransport,
		client:    client.(pb.DomainItemServiceClient), // Cast to the correct type.
	}
}

// Somethings fetches a list of somethings using the gRPC client.
func (s *DomainItemService) Somethings(ctx context.Context) ([]*model.Something, error) {
	request := &pb.GetSomethingRequest{} // Dynamic request.

	// Send the request using the client stored in the Service instance.
	response, err := s.transport.Send(ctx, s.client, "GetSomething", request)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil, err
	}

	// Assert the response to the correct type.
	res, ok := response.(*pb.GetSomethingResponse)
	if !ok {
		return nil, errors.New("invalid response type from gRPC transport")
	}

	var somethings []*model.Something
	for _, t := range res.Somethings {
		somethings = append(somethings, &model.Something{
			ID:       t.Id,
			SomeEnum: model.SomeEnum(t.SomeEnum),
		})
	}

	return somethings, nil
}
