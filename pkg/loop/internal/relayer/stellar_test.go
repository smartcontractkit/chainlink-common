package relayer

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	stelpb "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

func TestStellarDomainRoundTripThroughGRPC(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()

	svc := &staticStellarService{}
	stelpb.RegisterStellarServer(s, newStellarServer(svc, &loopnet.BrokerExt{
		BrokerConfig: loopnet.BrokerConfig{
			Logger: logger.Test(t),
		},
	}))

	go func() { _ = s.Serve(lis) }()
	defer s.Stop()

	ctx := t.Context()
	conn, err := grpc.DialContext(ctx, "bufnet", //nolint:staticcheck
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := &StellarClient{grpcClient: stelpb.NewStellarClient(conn)}

	t.Run("GetLedgerEntries_WithLiveUntil", func(t *testing.T) {
		liveUntil := uint32(500)
		svc.getLedgerEntries = func(_ context.Context, req stellartypes.GetLedgerEntriesRequest) (stellartypes.GetLedgerEntriesResponse, error) {
			require.Equal(t, []stellartypes.XDR{"a2V5MQ=="}, req.Keys) // base64("key1")
			return stellartypes.GetLedgerEntriesResponse{
				LatestLedger: 50,
				Entries: []stellartypes.LedgerEntryResult{
					{
						KeyXDR:             "a2V5MQ==", // base64("key1")
						DataXDR:            "ZGF0YTE=", // base64("data1")
						LastModifiedLedger: 30,
						LiveUntilLedgerSeq: &liveUntil,
					},
				},
			}, nil
		}

		resp, err := client.GetLedgerEntries(ctx, stellartypes.GetLedgerEntriesRequest{Keys: []stellartypes.XDR{"a2V5MQ=="}})
		require.NoError(t, err)
		require.Equal(t, uint32(50), resp.LatestLedger)
		require.Len(t, resp.Entries, 1)
		require.Equal(t, stellartypes.XDR("a2V5MQ=="), resp.Entries[0].KeyXDR)
		require.Equal(t, stellartypes.XDR("ZGF0YTE="), resp.Entries[0].DataXDR)
		require.Equal(t, uint32(30), resp.Entries[0].LastModifiedLedger)
		require.NotNil(t, resp.Entries[0].LiveUntilLedgerSeq)
		require.Equal(t, liveUntil, *resp.Entries[0].LiveUntilLedgerSeq)
	})

	t.Run("GetLedgerEntries_NoLiveUntil", func(t *testing.T) {
		svc.getLedgerEntries = func(_ context.Context, _ stellartypes.GetLedgerEntriesRequest) (stellartypes.GetLedgerEntriesResponse, error) {
			return stellartypes.GetLedgerEntriesResponse{
				LatestLedger: 60,
				Entries: []stellartypes.LedgerEntryResult{
					{
						KeyXDR:             "a2V5Mg==", // base64("key2")
						DataXDR:            "data2XDR", // valid 8-char base64
						LastModifiedLedger: 40,
						LiveUntilLedgerSeq: nil,
					},
				},
			}, nil
		}

		resp, err := client.GetLedgerEntries(ctx, stellartypes.GetLedgerEntriesRequest{Keys: []stellartypes.XDR{"a2V5Mg=="}})
		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		require.Nil(t, resp.Entries[0].LiveUntilLedgerSeq)
		require.Equal(t, uint32(60), resp.LatestLedger)
	})

	t.Run("GetLedgerEntries_MixedLiveUntil", func(t *testing.T) {
		// Two entries in one response: one with LiveUntilLedgerSeq set, one without.
		// Guards against the loop in ConvertGetLedgerEntriesResponseFromProto carrying
		// the HasLiveUntilLedgerSeq bool from one entry into the next.
		// "azE=", "azI=", "ZDE=", "ZDI=" are valid 4-char base64 values.
		liveUntil := uint32(777)
		svc.getLedgerEntries = func(_ context.Context, _ stellartypes.GetLedgerEntriesRequest) (stellartypes.GetLedgerEntriesResponse, error) {
			return stellartypes.GetLedgerEntriesResponse{
				LatestLedger: 70,
				Entries: []stellartypes.LedgerEntryResult{
					{KeyXDR: "azE=", DataXDR: "ZDE=", LastModifiedLedger: 10, LiveUntilLedgerSeq: &liveUntil},
					{KeyXDR: "azI=", DataXDR: "ZDI=", LastModifiedLedger: 20, LiveUntilLedgerSeq: nil},
				},
			}, nil
		}

		resp, err := client.GetLedgerEntries(ctx, stellartypes.GetLedgerEntriesRequest{Keys: []stellartypes.XDR{"azE=", "azI="}})
		require.NoError(t, err)
		require.Len(t, resp.Entries, 2)
		require.NotNil(t, resp.Entries[0].LiveUntilLedgerSeq)
		require.Equal(t, liveUntil, *resp.Entries[0].LiveUntilLedgerSeq)
		require.Nil(t, resp.Entries[1].LiveUntilLedgerSeq)
	})

	t.Run("GetLatestLedger", func(t *testing.T) {
		svc.getLatestLedger = func(_ context.Context) (stellartypes.GetLatestLedgerResponse, error) {
			return stellartypes.GetLatestLedgerResponse{
				Hash:            "deadbeef", // valid 4-byte hex
				ProtocolVersion: 21,
				Sequence:        1234,
				LedgerCloseTime: 9876543210,
			}, nil
		}

		resp, err := client.GetLatestLedger(ctx)
		require.NoError(t, err)
		require.Equal(t, stellartypes.LedgerHash("deadbeef"), resp.Hash)
		require.Equal(t, uint32(21), resp.ProtocolVersion)
		require.Equal(t, uint32(1234), resp.Sequence)
		require.Equal(t, int64(9876543210), resp.LedgerCloseTime)
	})
}

type staticStellarService struct {
	types.UnimplementedStellarService
	getLedgerEntries func(ctx context.Context, req stellartypes.GetLedgerEntriesRequest) (stellartypes.GetLedgerEntriesResponse, error)
	getLatestLedger  func(ctx context.Context) (stellartypes.GetLatestLedgerResponse, error)
}

func (s *staticStellarService) GetLedgerEntries(ctx context.Context, req stellartypes.GetLedgerEntriesRequest) (stellartypes.GetLedgerEntriesResponse, error) {
	if s.getLedgerEntries == nil {
		return s.UnimplementedStellarService.GetLedgerEntries(ctx, req)
	}
	return s.getLedgerEntries(ctx, req)
}

func (s *staticStellarService) GetLatestLedger(ctx context.Context) (stellartypes.GetLatestLedgerResponse, error) {
	if s.getLatestLedger == nil {
		return s.UnimplementedStellarService.GetLatestLedger(ctx)
	}
	return s.getLatestLedger(ctx)
}
