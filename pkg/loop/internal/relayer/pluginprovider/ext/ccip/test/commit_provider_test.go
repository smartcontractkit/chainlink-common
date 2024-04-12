package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/hashicorp/consul/sdk/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	loopnettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net/test"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestStaticCommitProvider(t *testing.T) {
	ctx := tests.Context(t)
	t.Run("Self consistent Evaluate", func(t *testing.T) {
		t.Parallel()
		// static test implementation is self consistent
		assert.NoError(t, CommitProvider.Evaluate(ctx, CommitProvider))

		// error when the test implementation evaluates something that differs from form itself
		botched := CommitProvider
		botched.priceRegistryReader = staticPriceRegistryReader{}
		err := CommitProvider.Evaluate(ctx, botched)
		require.Error(t, err)
		var evalErr evaluationError
		require.True(t, errors.As(err, &evalErr), "expected error to be an evaluationError")
		assert.Equal(t, priceRegistryComponent, evalErr.component)
	})
	t.Run("Self consistent AssertEqual", func(t *testing.T) {
		// no parallel because the AssertEqual is parallel
		CommitProvider.AssertEqual(ctx, t, CommitProvider)
	})
}

func TestCommitProviderGRPC(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)

	grpcScaffold := newGRPCScaffold(t, setupCommitProviderServer, ccip.NewCommitProviderClient)
	t.Cleanup(grpcScaffold.Close)
	// this is the meat of the test
	roundTripCommitProviderTests(ctx, t, grpcScaffold.Client())
}

func roundTripCommitProviderTests(ctx context.Context, t *testing.T, client types.CCIPCommitProvider) {
	t.Run("CommitStore", func(t *testing.T) {
		commitClient, err := client.NewCommitStoreReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripCommitStoreTests(ctx, t, commitClient)
		require.NoError(t, commitClient.Close())
	})

	t.Run("OffRamp", func(t *testing.T) {
		offRampClient, err := client.NewOffRampReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripOffRampTests(ctx, t, offRampClient)
		require.NoError(t, offRampClient.Close())
	})

	t.Run("OnRamp", func(t *testing.T) {
		onRampClient, err := client.NewOnRampReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripOnRampTests(ctx, t, onRampClient)
		require.NoError(t, onRampClient.Close())
	})

	t.Run("PriceGetter", func(t *testing.T) {
		priceGetterClient, err := client.NewPriceGetter(ctx)
		require.NoError(t, err)
		roundTripPriceGetterTests(ctx, t, priceGetterClient)
		require.NoError(t, priceGetterClient.Close())
	})

	t.Run("PriceRegistry", func(t *testing.T) {
		priceRegistryClient, err := client.NewPriceRegistryReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripPriceRegistryTests(ctx, t, priceRegistryClient)
		require.NoError(t, priceRegistryClient.Close())
	})
}

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
	wg         *sync.WaitGroup
}

func (t *grpcScaffold[T, S]) Close() {
	// close the client and server
	// note: some implementations in our suite of LOOPPs release server resources on client.Close()
	// this happens when the interface allocates resources via ~ `NewXXX` (e.g. NewCommitStoreReader)
	// the order of these two lines is important.
	require.NoError(t.t, t.client.Close(), "failed to close client")
	t.grpcServer.Stop()

	t.wg.Wait()
}

func (t *grpcScaffold[T, S]) Client() T {
	return t.client
}

func (t *grpcScaffold[T, S]) Server() S {
	return t.server
}

func newGRPCScaffold[T minimalClient, S any](t *testing.T, setup setupGRPCServer[S], clientFn setupGRPCClient[T]) *grpcScaffold[T, S] {

	lis := tcpListener(t)
	grpcServer := grpc.NewServer()

	lggr := logger.Test(t)
	broker := &loopnettest.Broker{T: t}
	brokerExt := &loopnet.BrokerExt{
		Broker:       broker,
		BrokerConfig: loopnet.BrokerConfig{Logger: lggr, StopCh: make(chan struct{})},
	}

	s := setup(t, grpcServer, brokerExt)
	// start the server and shutdown handler
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, grpcServer.Serve(lis))
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "failed to dial %s", lis.Addr().String())
	t.Cleanup(func() { conn.Close() })
	client := clientFn(brokerExt, conn)

	return &grpcScaffold[T, S]{t: t, server: s, client: client, grpcServer: grpcServer, wg: &wg}
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

func setupCommitProviderServer(t *testing.T, s *grpc.Server, b *loopnet.BrokerExt) *ccip.CommitProviderServer {
	commitProvider := ccip.NewCommitProviderServer(CommitProvider, b)
	ccippb.RegisterCommitCustomHandlersServer(s, commitProvider)
	return commitProvider
}
