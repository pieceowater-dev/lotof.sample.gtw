package transport

import "context"

type Transport interface {
	Send(ctx context.Context, request interface{}) (interface{}, error)
}

type Type string

const (
	GRPC Type = "grpc"
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) CreateTransport(transportType Type, address string) Transport {
	switch transportType {
	case GRPC:
		return NewGRPCTransport(address)
	default:
		return nil
	}
}
