package test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip"
	looptest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	cciptypes "github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestStaticOnRamp(t *testing.T) {
	t.Parallel()

	// static test implementation is self consistent
	ctx := context.Background()
	assert.NoError(t, OnRampReader.Evaluate(ctx, OnRampReader))

	// error when the test implementation is evaluates something that differs from the static implementation
	botched := OnRampReader
	botched.addressResponse = "not the right address"
	err := OnRampReader.Evaluate(ctx, botched)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not the right address")
}

func TestOnRampGRPC(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)

	scaffold := looptest.NewGRPCScaffold(t, setupOnRampServer, setupOnRampClient)

	// test the client
	roundTripOnRampTests(ctx, t, scaffold.Client())
	// offramp implements dependency management, test that it closes properly
	t.Run("Dependency management", func(t *testing.T) {
		d := &looptest.MockDep{}
		scaffold.Server().AddDep(d)
		assert.False(t, d.IsClosed())
		scaffold.Client().Close()
		assert.True(t, d.IsClosed())
	})
}

func roundTripOnRampTests(ctx context.Context, t *testing.T, client cciptypes.OnRampReader) {
	t.Run("Address", func(t *testing.T) {
		got, err := client.Address(ctx)
		require.NoError(t, err)
		assert.Equal(t, OnRampReader.addressResponse, got)
	})

	t.Run("GetDynamicConfig", func(t *testing.T) {
		got, err := client.GetDynamicConfig(ctx)
		require.NoError(t, err)
		assert.Equal(t, OnRampReader.dynamicConfigResponse, got)
	})

	t.Run("GetSendRequestsBetweenSeqNums", func(t *testing.T) {
		got, err := client.GetSendRequestsBetweenSeqNums(ctx, OnRampReader.getSendRequestsBetweenSeqNums.SeqNumMin, OnRampReader.getSendRequestsBetweenSeqNums.SeqNumMax, OnRampReader.getSendRequestsBetweenSeqNums.Finalized)
		require.NoError(t, err)
		if !reflect.DeepEqual(OnRampReader.getSendRequestsBetweenSeqNumsResponse.EVM2EVMMessageWithTxMeta, got) {
			t.Errorf("expected %v, got %v", OnRampReader.getSendRequestsBetweenSeqNumsResponse.EVM2EVMMessageWithTxMeta, got)
		}
	})

	t.Run("IsSourceChainHealthy", func(t *testing.T) {
		got, err := client.IsSourceChainHealthy(ctx)
		require.NoError(t, err)
		assert.Equal(t, OnRampReader.isSourceChainHealthyResponse, got)
	})

	t.Run("IsSourceCursed", func(t *testing.T) {
		got, err := client.IsSourceCursed(ctx)
		require.NoError(t, err)
		assert.Equal(t, OnRampReader.isSourceCursedResponse, got)
	})

	t.Run("RouterAddress", func(t *testing.T) {
		got, err := client.RouterAddress(ctx)
		require.NoError(t, err)
		assert.Equal(t, OnRampReader.routerResponse, got)
	})

	t.Run("SourcePriceRegistryAddress", func(t *testing.T) {
		got, err := client.SourcePriceRegistryAddress(ctx)
		require.NoError(t, err)
		assert.Equal(t, OnRampReader.sourcePriceRegistryResponse, got)
	})
}

func setupOnRampServer(t *testing.T, server *grpc.Server, b *loopnet.BrokerExt) *ccip.OnRampReaderGRPCServer {
	onRamp := ccip.NewOnRampReaderGRPCServer(OnRampReader)
	ccippb.RegisterOnRampReaderServer(server, onRamp)
	return onRamp
}

func setupOnRampClient(b *loopnet.BrokerExt, conn grpc.ClientConnInterface) *ccip.OnRampReaderGRPCClient {
	return ccip.NewOnRampReaderGRPCClient(conn)
}
