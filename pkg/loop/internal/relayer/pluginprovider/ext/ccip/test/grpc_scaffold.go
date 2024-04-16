package test

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/hashicorp/consul/sdk/freeport"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	loopnettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net/test"
)

type minimalClient interface {
	io.Closer
	goplugin.GRPCClientConn
}

type grpcScaffold[T minimalClient, S any] struct {
	t *testing.T
	//server *grpc.Server
	server S
	client T

	grpcServer *grpc.Server
}

func (t *grpcScaffold[T, S]) Close() {
	// close the client and server
	// note: some implementations in our suite of LOOPPs release server resources on client.Close()
	// this happens when the interface allocates resources via ~ `NewXXX` (e.g. NewCommitStoreReader)
	// the order of these two lines is important.
	require.NoError(t.t, t.client.Close(), "failed to close client")
}

func (t *grpcScaffold[T, S]) Client() T {
	return t.client
}

func (t *grpcScaffold[T, S]) Server() S {
	return t.server
}

func newGRPCScaffold[T minimalClient, S any](t *testing.T, serverFn setupGRPCServer[S], clientFn setupGRPCClient[T]) *grpcScaffold[T, S] {
	lis := tcpListener(t)
	grpcServer := grpc.NewServer()
	t.Cleanup(grpcServer.Stop)

	lggr := logger.Test(t)
	broker := &loopnettest.Broker{T: t}
	brokerExt := &loopnet.BrokerExt{
		Broker:       broker,
		BrokerConfig: loopnet.BrokerConfig{Logger: lggr, StopCh: make(chan struct{})},
	}

	s := serverFn(t, grpcServer, brokerExt)
	go func() {
		// the cleanup call to grpcServer.Stop will unblock this
		require.NoError(t, grpcServer.Serve(lis))
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "failed to dial %s", lis.Addr().String())
	t.Cleanup(func() { require.NoError(t, conn.Close(), "failed to close connection") })

	client := clientFn(brokerExt, conn)

	return &grpcScaffold[T, S]{t: t, server: s, client: client, grpcServer: grpcServer}
}

func tcpListener(t *testing.T) net.Listener {
	port := freeport.GetOne(t)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	require.NoError(t, err, "failed to listen on port %d", port)
	t.Cleanup(func() { lis.Close() })
	return lis
}

// setupGRPCServer is a function that sets up a grpc server with a given broker
// typical and expected usage is to instantiate a grpc server implementation with a static test interface implementation
// and then register that grpc server
// e.g.
// ```
//
//	func setupCCIPCommitProviderGRPCServer(t *testing.T, s *grpc.Server, b *loopnet.BrokerExt) *grpc.Server {
//	  commitProvider := ccip.NewCommitProviderServer(CommitProvider, b)
//	  ccippb.RegisterCommitCustomHandlersServer(s, commitProvider)
//	  return s
//	}
type setupGRPCServer[S any] func(t *testing.T, s *grpc.Server, b *loopnet.BrokerExt) S
type setupGRPCClient[T minimalClient] func(b *loopnet.BrokerExt, conn grpc.ClientConnInterface) T

// mockDep is a mock dependency that can be used to test that a grpc client closes its dependencies
type mockDep struct {
	closeCalled bool
}

func (m *mockDep) Close() error {
	m.closeCalled = true
	return nil
}
