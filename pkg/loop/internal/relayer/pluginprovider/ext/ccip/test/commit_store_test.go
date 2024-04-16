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

func TestStaticCommitStore(t *testing.T) {
	t.Parallel()

	// static test implementation is self consistent
	ctx := context.Background()
	assert.NoError(t, CommitStoreReader.Evaluate(ctx, CommitStoreReader))

	// error when the test implementation is evaluates something that differs from the static implementation
	botched := CommitStoreReader
	botched.changeConfigResponse = "not the right conifg"
	err := CommitStoreReader.Evaluate(ctx, botched)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not the right conifg")
}

func TestCommitStoreGRPC(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	scaffold := looptest.NewGRPCScaffold(t, setupCommitStoreServer, ccip.NewCommitStoreReaderGRPCClient)
	roundTripCommitStoreTests(ctx, t, scaffold.Client())
	// commit store implements dependency management, test that it closes properly
	t.Run("Dependency management", func(t *testing.T) {
		d := &looptest.MockDep{}
		scaffold.Server().AddDep(d)
		assert.False(t, d.IsClosed())
		scaffold.Client().Close()
		assert.True(t, d.IsClosed())
	})
}

// roundTripCommitStoreTests tests the round trip of the client<->server.
// it should exercise all the methods of the client.
// do not add client.Close to this test, test that from the driver test
func roundTripCommitStoreTests(ctx context.Context, t *testing.T, client cciptypes.CommitStoreReader) {
	t.Run("ChangeConfig", func(t *testing.T) {
		gotAddr, err := client.ChangeConfig(ctx, CommitStoreReader.changeConfigRequest.onchainConfig, CommitStoreReader.changeConfigRequest.offchainConfig)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.changeConfigResponse, gotAddr)
	})

	t.Run("DecodeCommitReport", func(t *testing.T) {
		report, err := client.DecodeCommitReport(ctx, CommitStoreReader.decodeCommitReportRequest)
		require.NoError(t, err)
		if !reflect.DeepEqual(CommitStoreReader.decodeCommitReportResponse, report) {
			t.Errorf("expected %v, got %v", CommitStoreReader.decodeCommitReportResponse, report)
		}
	})

	// reuse the test data for the encode method
	t.Run("EncodeCommtReport", func(t *testing.T) {
		report, err := client.EncodeCommitReport(ctx, CommitStoreReader.decodeCommitReportResponse)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.decodeCommitReportRequest, report)
	})

	// exercise all the gas price estimator methods
	t.Run("GasPriceEstimator", func(t *testing.T) {
		estimator, err := client.GasPriceEstimator(ctx)
		require.NoError(t, err)

		t.Run("GetGasPrice", func(t *testing.T) {
			price, err := estimator.GetGasPrice(ctx)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.getGasPriceResponse, price)
		})

		t.Run("DenoteInUSD", func(t *testing.T) {
			usd, err := estimator.DenoteInUSD(
				GasPriceEstimatorCommit.denoteInUSDRequest.p,
				GasPriceEstimatorCommit.denoteInUSDRequest.wrappedNativePrice,
			)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.denoteInUSDResponse.result, usd)
		})

		t.Run("Deviates", func(t *testing.T) {
			deviates, err := estimator.Deviates(
				GasPriceEstimatorCommit.deviatesRequest.p1,
				GasPriceEstimatorCommit.deviatesRequest.p2,
			)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.deviatesResponse, deviates)
		})

		t.Run("Median", func(t *testing.T) {
			median, err := estimator.Median(GasPriceEstimatorCommit.medianRequest.gasPrices)
			require.NoError(t, err)
			assert.Equal(t, GasPriceEstimatorCommit.medianResponse, median)
		})
	})

	t.Run("GetAcceptedCommitReportGteTimestamp", func(t *testing.T) {
		report, err := client.GetAcceptedCommitReportsGteTimestamp(ctx,
			CommitStoreReader.getAcceptedCommitReportsGteTimestampRequest.timestamp,
			CommitStoreReader.getAcceptedCommitReportsGteTimestampRequest.confirmations)
		require.NoError(t, err)
		if !reflect.DeepEqual(CommitStoreReader.getAcceptedCommitReportsGteTimestampResponse, report) {
			t.Errorf("expected %v, got %v", CommitStoreReader.getAcceptedCommitReportsGteTimestampResponse, report)
		}
	})

	t.Run("GetCommitReportMatchingSeqNum", func(t *testing.T) {
		report, err := client.GetCommitReportMatchingSeqNum(ctx,
			CommitStoreReader.getCommitReportMatchingSeqNumRequest.seqNum,
			CommitStoreReader.getCommitReportMatchingSeqNumRequest.confirmations)
		require.NoError(t, err)
		// use the same response as the reportsGteTimestamp for simplicity
		if !reflect.DeepEqual(CommitStoreReader.getAcceptedCommitReportsGteTimestampResponse, report) {
			t.Errorf("expected %v, got %v", CommitStoreReader.getAcceptedCommitReportsGteTimestampRequest, report)
		}
	})

	t.Run("GetCommitStoreStaticConfig", func(t *testing.T) {
		config, err := client.GetCommitStoreStaticConfig(ctx)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.getCommitStoreStaticConfigResponse, config)
	})

	t.Run("GetExpectedNextSequenceNumber", func(t *testing.T) {
		seq, err := client.GetExpectedNextSequenceNumber(ctx)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.getExpectedNextSequenceNumberResponse, seq)
	})

	t.Run("GetLatestPriceEpochAndRound", func(t *testing.T) {
		got, err := client.GetLatestPriceEpochAndRound(ctx)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.getLatestPriceEpochAndRoundResponse, got)
	})

	t.Run("IsBlessed", func(t *testing.T) {
		got, err := client.IsBlessed(ctx, CommitStoreReader.isBlessedRequest)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.isBlessedResponse, got)
	})

	t.Run("IsDestChainHealthy", func(t *testing.T) {
		got, err := client.IsDestChainHealthy(ctx)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.isDestChainHealthyResponse, got)
	})

	t.Run("IsDown", func(t *testing.T) {
		got, err := client.IsDown(ctx)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.isDownResponse, got)
	})

	t.Run("OffchainConfig", func(t *testing.T) {
		config, err := client.OffchainConfig(ctx)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.offchainConfigResponse, config)
	})

	t.Run("VerifyExecutionReport", func(t *testing.T) {
		got, err := client.VerifyExecutionReport(ctx, CommitStoreReader.verifyExecutionReportRequest)
		require.NoError(t, err)
		assert.Equal(t, CommitStoreReader.verifyExecutionReportResponse, got)
	})
}

func setupCommitStoreServer(t *testing.T, s *grpc.Server, b *loopnet.BrokerExt) *ccip.CommitStoreGRPCServer {
	commitProvider, err := ccip.NewCommitStoreReaderGRPCServer(CommitStoreReader, b)
	require.NoError(t, err)
	ccippb.RegisterCommitStoreReaderServer(s, commitProvider)
	return commitProvider
}

var _ looptest.SetupGRPCServer[*ccip.CommitStoreGRPCServer] = setupCommitStoreServer
var _ looptest.SetupGRPCClient[*ccip.CommitStoreGRPCClient] = ccip.NewCommitStoreReaderGRPCClient
