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

func TestStaticExecProvider(t *testing.T) {
	ctx := tests.Context(t)
	t.Run("Self consistent Evaluate", func(t *testing.T) {
		t.Parallel()
		// static test implementation is self consistent
		assert.NoError(t, ExecutionProvider.Evaluate(ctx, ExecutionProvider))

		// error when the test implementation evaluates something that differs from form itself
		botched := ExecutionProvider
		botched.priceRegistryReader = staticPriceRegistryReader{}
		err := ExecutionProvider.Evaluate(ctx, botched)
		require.Error(t, err)
		var evalErr evaluationError
		require.True(t, errors.As(err, &evalErr), "expected error to be an evaluationError")
		assert.Equal(t, priceRegistryComponent, evalErr.component)
	})
	t.Run("Self consistent AssertEqual", func(t *testing.T) {
		// no parallel because the AssertEqual is parallel
		ExecutionProvider.AssertEqual(ctx, t, ExecutionProvider)
	})
}

func TestExecProviderGRPC(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)

	grpcScaffold := looptest.NewGRPCScaffold(t, setupExecProviderServer, ccip.NewExecProviderClient)
	t.Cleanup(grpcScaffold.Close)
	roundTripExecProviderTests(ctx, t, grpcScaffold.Client())
}

func roundTripExecProviderTests(ctx context.Context, t *testing.T, client types.CCIPExecProvider) {
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

	t.Run("PriceRegistry", func(t *testing.T) {
		priceRegistryClient, err := client.NewPriceRegistryReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripPriceRegistryTests(ctx, t, priceRegistryClient)
		require.NoError(t, priceRegistryClient.Close())
	})

	t.Run("TokenData", func(t *testing.T) {
		tokenDataClient, err := client.NewTokenDataReader(ctx, "ignored")
		require.NoError(t, err)
		roundTripTokenDataTests(ctx, t, tokenDataClient)
		require.NoError(t, tokenDataClient.Close())
	})

	t.Run("TokenPool", func(t *testing.T) {
		tokenReaderClient, err := client.NewTokenPoolBatchedReader(ctx)
		require.NoError(t, err)
		roundTripTokenPoolTests(ctx, t, tokenReaderClient)
		require.NoError(t, tokenReaderClient.Close())
	})

	t.Run("SourceNativeToken", func(t *testing.T) {
		token, err := client.SourceNativeToken(ctx)
		require.NoError(t, err)
		assert.Equal(t, ExecutionProvider.sourceNativeTokenResponse, token)
	})
}

func setupExecProviderServer(t *testing.T, server *grpc.Server, b *loopnet.BrokerExt) *ccip.ExecProviderServer {
	execProvider := ccip.NewExecProviderServer(ExecutionProvider, b)
	ccippb.RegisterExecutionCustomHandlersServer(server, execProvider)
	return execProvider
}
