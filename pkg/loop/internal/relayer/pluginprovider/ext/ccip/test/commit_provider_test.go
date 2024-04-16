package test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip"
	looptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
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

	grpcScaffold := looptest.NewGRPCScaffold(t, setupCommitProviderServer, ccip.NewCommitProviderClient)
	t.Cleanup(grpcScaffold.Close)
	roundTripCommitProviderTests(ctx, t, grpcScaffold.Client())
}

func roundTripCommitProviderTests(ctx context.Context, t *testing.T, client types.CCIPCommitProvider) {
	t.Run("CommitStore", func(t *testing.T) {
		commitClient, err := client.NewCommitStoreReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripCommitStoreTests(t, commitClient)
		require.NoError(t, commitClient.Close())
	})

	t.Run("OffRamp", func(t *testing.T) {
		offRampClient, err := client.NewOffRampReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripOffRampTests(t, offRampClient)
		require.NoError(t, offRampClient.Close())
	})

	t.Run("OnRamp", func(t *testing.T) {
		onRampClient, err := client.NewOnRampReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripOnRampTests(t, onRampClient)
		require.NoError(t, onRampClient.Close())
	})

	t.Run("PriceGetter", func(t *testing.T) {
		priceGetterClient, err := client.NewPriceGetter(ctx)
		require.NoError(t, err)
		roundTripPriceGetterTests(t, priceGetterClient)
		require.NoError(t, priceGetterClient.Close())
	})

	t.Run("PriceRegistry", func(t *testing.T) {
		priceRegistryClient, err := client.NewPriceRegistryReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripPriceRegistryTests(t, priceRegistryClient)
		require.NoError(t, priceRegistryClient.Close())
	})
}

func setupCommitProviderServer(t *testing.T, s *grpc.Server, b *loopnet.BrokerExt) *ccip.CommitProviderServer {
	commitProvider := ccip.NewCommitProviderServer(CommitProvider, b)
	ccippb.RegisterCommitCustomHandlersServer(s, commitProvider)
	return commitProvider
}

var _ looptest.SetupGRPCServer[*ccip.CommitProviderServer] = setupCommitProviderServer
var _ looptest.SetupGRPCClient[*ccip.CommitProviderClient] = ccip.NewCommitProviderClient
