package relayer

import (
	"context"
	"encoding/base64"
	"errors"
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
			require.Equal(t, []string{"a2V5MQ=="}, req.Keys) // base64("key1")
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

		resp, err := client.GetLedgerEntries(ctx, stellartypes.GetLedgerEntriesRequest{Keys: []string{"a2V5MQ=="}})
		require.NoError(t, err)
		require.Equal(t, uint32(50), resp.LatestLedger)
		require.Len(t, resp.Entries, 1)
		require.Equal(t, "a2V5MQ==", resp.Entries[0].KeyXDR)
		require.Equal(t, "ZGF0YTE=", resp.Entries[0].DataXDR)
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

		resp, err := client.GetLedgerEntries(ctx, stellartypes.GetLedgerEntriesRequest{Keys: []string{"a2V5Mg=="}})
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

		resp, err := client.GetLedgerEntries(ctx, stellartypes.GetLedgerEntriesRequest{Keys: []string{"azE=", "azI="}})
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
		require.Equal(t, "deadbeef", resp.Hash)
		require.Equal(t, uint32(21), resp.ProtocolVersion)
		require.Equal(t, uint32(1234), resp.Sequence)
		require.Equal(t, int64(9876543210), resp.LedgerCloseTime)
	})

	t.Run("ReadContract_roundtrip", func(t *testing.T) {
		sym := "transfer"
		argVal := stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &sym}

		// Result is an opaque base64-XDR blob to the relayer; the test only verifies
		// the string is delivered intact across gRPC.
		const resultB64 = "AAAABAAAACo="

		svc.readContract = func(_ context.Context, req stellartypes.ReadContractRequest) (stellartypes.ReadContractResponse, error) {
			require.Equal(t, "CABC123", req.ContractID)
			require.Equal(t, "my_fn", req.Function)
			require.Equal(t, uint32(100), req.LedgerSequence)
			require.Len(t, req.Args, 1)
			require.Equal(t, stellartypes.ScValTypeSymbol, req.Args[0].Type)
			require.NotNil(t, req.Args[0].Symbol)
			require.Equal(t, "transfer", *req.Args[0].Symbol)
			return stellartypes.ReadContractResponse{
				Result:         resultB64,
				LedgerSequence: 101,
			}, nil
		}

		resp, err := client.ReadContract(ctx, stellartypes.ReadContractRequest{
			ContractID:     "CABC123",
			Function:       "my_fn",
			Args:           []stellartypes.ScVal{argVal},
			LedgerSequence: 100,
		})
		require.NoError(t, err)
		require.Equal(t, uint32(101), resp.LedgerSequence)
		require.Empty(t, resp.Error)
		require.Equal(t, resultB64, resp.Result)
	})

	t.Run("SubmitTransaction_roundtrip", func(t *testing.T) {
		sym := "transfer"
		argVal := stellartypes.ScVal{Type: stellartypes.ScValTypeSymbol, Symbol: &sym}

		svc.submitTransaction = func(_ context.Context, req stellartypes.SubmitTransactionRequest) (*stellartypes.SubmitTransactionResponse, error) {
			require.Equal(t, "idem-key", req.IdempotencyKey)
			require.Equal(t, "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF", req.FromAddress)
			require.Equal(t, "CABC123", req.ContractID)
			require.Equal(t, "my_fn", req.Function)
			require.Len(t, req.Args, 1)
			require.Equal(t, stellartypes.ScValTypeSymbol, req.Args[0].Type)
			require.Equal(t, uint32(5), req.LedgerBoundsOffset)
			return &stellartypes.SubmitTransactionResponse{
				TxStatus:         stellartypes.TxSuccess,
				TxHash:           "hash123",
				TxIdempotencyKey: "idem-key",
			}, nil
		}

		reply, err := client.SubmitTransaction(ctx, stellartypes.SubmitTransactionRequest{
			IdempotencyKey:     "idem-key",
			FromAddress:        "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF",
			ContractID:         "CABC123",
			Function:           "my_fn",
			Args:               []stellartypes.ScVal{argVal},
			LedgerBoundsOffset: 5,
		})
		require.NoError(t, err)
		require.Equal(t, stellartypes.TxSuccess, reply.TxStatus)
		require.Equal(t, "hash123", reply.TxHash)
		require.Equal(t, "idem-key", reply.TxIdempotencyKey)
	})

	t.Run("SubmitTransaction_withResultXDR", func(t *testing.T) {
		svc.submitTransaction = func(_ context.Context, _ stellartypes.SubmitTransactionRequest) (*stellartypes.SubmitTransactionResponse, error) {
			return &stellartypes.SubmitTransactionResponse{
				TxStatus:         stellartypes.TxSuccess,
				TxHash:           "hash-with-xdr",
				TxIdempotencyKey: "idem-xdr",
				ResultXDR:        base64.StdEncoding.EncodeToString([]byte("result")),
				ResultMetaXDR:    base64.StdEncoding.EncodeToString([]byte("meta")),
			}, nil
		}

		reply, err := client.SubmitTransaction(ctx, stellartypes.SubmitTransactionRequest{
			ContractID: "CABC123",
			Function:   "my_fn",
		})
		require.NoError(t, err)
		require.Equal(t, "hash-with-xdr", reply.TxHash)
		require.Equal(t, base64.StdEncoding.EncodeToString([]byte("result")), reply.ResultXDR)
		require.Equal(t, base64.StdEncoding.EncodeToString([]byte("meta")), reply.ResultMetaXDR)
	})

	t.Run("SubmitTransaction_invalidRequest", func(t *testing.T) {
		_, err := client.SubmitTransaction(ctx, stellartypes.SubmitTransactionRequest{Function: "fn"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid SubmitTransaction request")
	})

	t.Run("SubmitTransaction_implError", func(t *testing.T) {
		svc.submitTransaction = func(_ context.Context, _ stellartypes.SubmitTransactionRequest) (*stellartypes.SubmitTransactionResponse, error) {
			return nil, errors.New("submit failed")
		}

		_, err := client.SubmitTransaction(ctx, stellartypes.SubmitTransactionRequest{
			ContractID: "CABC123",
			Function:   "my_fn",
		})
		require.Error(t, err)
	})

	t.Run("ReadContract_noArgs_noResult", func(t *testing.T) {
		svc.readContract = func(_ context.Context, req stellartypes.ReadContractRequest) (stellartypes.ReadContractResponse, error) {
			require.Empty(t, req.Args)
			return stellartypes.ReadContractResponse{
				Error:          "contract error: not found",
				LedgerSequence: 200,
			}, nil
		}

		resp, err := client.ReadContract(ctx, stellartypes.ReadContractRequest{
			ContractID: "CXYZ",
			Function:   "noop",
		})
		require.NoError(t, err)
		require.Equal(t, "contract error: not found", resp.Error)
		require.Empty(t, resp.Result)
		require.Equal(t, uint32(200), resp.LedgerSequence)
	})
}

type staticStellarService struct {
	types.UnimplementedStellarService
	getLedgerEntries  func(ctx context.Context, req stellartypes.GetLedgerEntriesRequest) (stellartypes.GetLedgerEntriesResponse, error)
	getLatestLedger   func(ctx context.Context) (stellartypes.GetLatestLedgerResponse, error)
	readContract      func(ctx context.Context, req stellartypes.ReadContractRequest) (stellartypes.ReadContractResponse, error)
	submitTransaction func(ctx context.Context, req stellartypes.SubmitTransactionRequest) (*stellartypes.SubmitTransactionResponse, error)
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

func (s *staticStellarService) ReadContract(ctx context.Context, req stellartypes.ReadContractRequest) (stellartypes.ReadContractResponse, error) {
	if s.readContract == nil {
		return s.UnimplementedStellarService.ReadContract(ctx, req)
	}
	return s.readContract(ctx, req)
}

func (s *staticStellarService) SubmitTransaction(ctx context.Context, req stellartypes.SubmitTransactionRequest) (*stellartypes.SubmitTransactionResponse, error) {
	if s.submitTransaction == nil {
		return s.UnimplementedStellarService.SubmitTransaction(ctx, req)
	}
	return s.submitTransaction(ctx, req)
}
