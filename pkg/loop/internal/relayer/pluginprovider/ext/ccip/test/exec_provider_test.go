package test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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
		ep := ExecutionProvider(logger.Test(t))
		// static test implementation is self consistent
		assert.NoError(t, ep.Evaluate(ctx, ep))

		// error when the test implementation evaluates something that differs from form itself
		botched := ExecutionProvider(logger.Test(t))
		botched.priceRegistryReader = staticPriceRegistryReader{}
		err := ep.Evaluate(ctx, botched)
		require.Error(t, err)
		var evalErr evaluationError
		require.True(t, errors.As(err, &evalErr), "expected error to be an evaluationError")
		assert.Equal(t, priceRegistryComponent, evalErr.component)
	})
	t.Run("Self consistent AssertEqual", func(t *testing.T) {
		ep := ExecutionProvider(logger.Test(t))
		// no parallel because the AssertEqual is parallel
		ep.AssertEqual(ctx, t, ep)
	})
}

func TestExecProviderGRPC(t *testing.T) {
	t.Parallel()

	grpcScaffold := looptest.NewGRPCScaffold(t, setupExecProviderServer, ccip.NewExecProviderClient)
	t.Cleanup(grpcScaffold.Close)
	roundTripExecProviderTests(t, grpcScaffold.Client())
}

func roundTripExecProviderTests(t *testing.T, client types.CCIPExecProvider) {
	t.Run("CommitStore", func(t *testing.T) {
		commitClient, err := client.NewCommitStoreReader(tests.Context(t), "ignored")
		require.NoError(t, err)
		roundTripCommitStoreTests(t, commitClient)
		require.NoError(t, commitClient.Close())
	})

	t.Run("OffRamp", func(t *testing.T) {
		offRampClient, err := client.NewOffRampReader(tests.Context(t), "ignored")
		require.NoError(t, err)
		roundTripOffRampTests(t, offRampClient)
		require.NoError(t, offRampClient.Close())
	})

	t.Run("OnRamp", func(t *testing.T) {
		onRampClient, err := client.NewOnRampReader(tests.Context(t), "ignored", 0, 0)
		require.NoError(t, err)
		roundTripOnRampTests(t, onRampClient)
		require.NoError(t, onRampClient.Close())
	})

	t.Run("PriceRegistry", func(t *testing.T) {
		priceRegistryClient, err := client.NewPriceRegistryReader(tests.Context(t), "ignored")
		require.NoError(t, err)
		roundTripPriceRegistryTests(t, priceRegistryClient)
		require.NoError(t, priceRegistryClient.Close())
	})

	t.Run("TokenData", func(t *testing.T) {
		tokenDataClient, err := client.NewTokenDataReader(tests.Context(t), "ignored")
		require.NoError(t, err)
		roundTripTokenDataTests(t, tokenDataClient)
		require.NoError(t, tokenDataClient.Close())
	})

	t.Run("TokenPool", func(t *testing.T) {
		tokenReaderClient, err := client.NewTokenPoolBatchedReader(tests.Context(t), "ignored", 0)
		require.NoError(t, err)
		roundTripTokenPoolTests(t, tokenReaderClient)
		require.NoError(t, tokenReaderClient.Close())
	})

	t.Run("SourceNativeToken", func(t *testing.T) {
		token, err := client.SourceNativeToken(tests.Context(t), "ignored")
		require.NoError(t, err)
		assert.Equal(t, ExecutionProvider(logger.Test(t)).sourceNativeTokenResponse, token)
	})

	t.Run("GetTransactionStatus", func(t *testing.T) {
		status, err := client.GetTransactionStatus(tests.Context(t), "ignored")
		require.NoError(t, err)
		assert.Equal(t, ExecutionProvider(logger.Test(t)).transactionStatusResponse, status)
	})
}

func setupExecProviderServer(t *testing.T, server *grpc.Server, b *loopnet.BrokerExt) *ccip.ExecProviderServer {
	execProvider := ccip.NewExecProviderServer(ExecutionProvider(logger.Test(t)), b)
	ccippb.RegisterExecutionCustomHandlersServer(server, execProvider)
	return execProvider
}
