// Package grpcpool opens several independent gRPC connections to the same
// backend address and round-robins RPCs across them, exposed as a single
// grpc.ClientConnInterface so every caller (each module's service-client
// constructor) uses it exactly like a plain grpc.NewClient(address, ...)
// connection.
//
// Why not grpc-go's built-in round_robin balancer: it operates on
// resolver.Endpoint identity, which is keyed purely by the set of addresses
// within an endpoint — handing it N endpoints that all resolve to the same
// address collapses them back into a single endpoint/connection. This
// package implements "N independent connections to the exact same address"
// directly instead.
package grpcpool

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
)

// PoolSize is how many independent connections NewPooledClient opens to the
// same backend address. A single grpc.ClientConn multiplexes all traffic
// over one HTTP/2 connection — spreading traffic across several independent
// connections dilutes the known golang.org/x/net HPACK encoder panic risk
// under heavy cumulative stream churn on one long-lived connection.
const PoolSize = 16

// PooledConn implements grpc.ClientConnInterface over a fixed set of
// independent *grpc.ClientConns, picking one per call via a simple atomic
// round-robin counter.
type PooledConn struct {
	conns   []*grpc.ClientConn
	counter uint64
}

// NewPooledClient dials PoolSize independent connections to address, each
// with the given dial options, and returns them as one pooled
// grpc.ClientConnInterface.
func NewPooledClient(address string, dialOpts ...grpc.DialOption) (*PooledConn, error) {
	conns := make([]*grpc.ClientConn, 0, PoolSize)
	for i := 0; i < PoolSize; i++ {
		conn, err := grpc.NewClient(address, dialOpts...)
		if err != nil {
			for _, c := range conns {
				_ = c.Close()
			}
			return nil, err
		}
		conns = append(conns, conn)
	}
	return &PooledConn{conns: conns}, nil
}

func (p *PooledConn) next() *grpc.ClientConn {
	i := atomic.AddUint64(&p.counter, 1)
	return p.conns[i%uint64(len(p.conns))]
}

// Invoke implements grpc.ClientConnInterface.
func (p *PooledConn) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	return p.next().Invoke(ctx, method, args, reply, opts...)
}

// NewStream implements grpc.ClientConnInterface.
func (p *PooledConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return p.next().NewStream(ctx, desc, method, opts...)
}

// Close closes every underlying connection in the pool.
func (p *PooledConn) Close() error {
	var firstErr error
	for _, c := range p.conns {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
