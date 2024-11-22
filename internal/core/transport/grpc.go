package transport

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"reflect"
)

// GRPCTransport handles both client and server-side transport
type GRPCTransport struct {
	address string
	server  *grpc.Server
}

func NewGRPCTransport(address string) *GRPCTransport {
	return &GRPCTransport{address: address}
}

// CreateClient dynamically creates a gRPC client using the passed constructor.
func (g *GRPCTransport) CreateClient(clientConstructor any) (any, error) {
	// Dial the gRPC connection.
	conn, err := grpc.Dial(g.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("failed to connect to gRPC server: " + err.Error())
	}

	// Use reflection to call the constructor function dynamically
	constructorValue := reflect.ValueOf(clientConstructor)
	if constructorValue.Kind() != reflect.Func {
		return nil, errors.New("clientConstructor must be a function")
	}

	// Call the constructor to create the client (pass the connection as argument)
	clientValues := constructorValue.Call([]reflect.Value{reflect.ValueOf(conn)})

	// Ensure that the client creation was successful and return the client
	if len(clientValues) > 0 {
		return clientValues[0].Interface(), nil
	}
	return nil, errors.New("failed to create client")
}

// Send sends a dynamic gRPC request based on method name and request type
func (g *GRPCTransport) Send(ctx context.Context, client any, serviceMethod string, request any) (any, error) {
	// Use reflection to get the method from the client dynamically
	clientValue := reflect.ValueOf(client)
	method := clientValue.MethodByName(serviceMethod)
	if !method.IsValid() {
		return nil, errors.New("invalid service method: " + serviceMethod)
	}

	// Ensure the request is passed as a reflect.Value
	reqValue := reflect.ValueOf(request)
	if reqValue.IsValid() {
		// Call the method dynamically, passing the context and the request
		returnValues := method.Call([]reflect.Value{reflect.ValueOf(ctx), reqValue})
		if len(returnValues) > 1 && returnValues[1].Interface() != nil {
			return nil, returnValues[1].Interface().(error)
		}
		// Return the response from the method call
		return returnValues[0].Interface(), nil
	}
	return nil, errors.New("invalid request type for method")
}
