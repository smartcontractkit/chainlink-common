package relayerset

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	stelpb "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	coremocks "github.com/smartcontractkit/chainlink-common/pkg/types/core/mocks"
	stellarmocks "github.com/smartcontractkit/chainlink-common/pkg/types/mocks"
)

type stubStellarGRPC struct {
	submit func(ctx context.Context, in *stelpb.SubmitTransactionRequest, opts ...grpc.CallOption) (*stelpb.SubmitTransactionResponse, error)
}

func (s *stubStellarGRPC) GetLedgerEntries(context.Context, *stelpb.GetLedgerEntriesRequest, ...grpc.CallOption) (*stelpb.GetLedgerEntriesResponse, error) {
	panic("not implemented")
}

func (s *stubStellarGRPC) GetLatestLedger(context.Context, *emptypb.Empty, ...grpc.CallOption) (*stelpb.GetLatestLedgerResponse, error) {
	panic("not implemented")
}

func (s *stubStellarGRPC) ReadContract(context.Context, *stelpb.ReadContractRequest, ...grpc.CallOption) (*stelpb.ReadContractResponse, error) {
	panic("not implemented")
}

func (s *stubStellarGRPC) SubmitTransaction(ctx context.Context, in *stelpb.SubmitTransactionRequest, opts ...grpc.CallOption) (*stelpb.SubmitTransactionResponse, error) {
	return s.submit(ctx, in, opts...)
}

func TestStellarClient_SubmitTransaction_ForwardsRelayID(t *testing.T) {
	t.Parallel()

	relayID := types.RelayID{Network: "stellar-net", ChainID: "stellar-chain"}
	var gotNetwork, gotChain string

	inner := &stubStellarGRPC{
		submit: func(ctx context.Context, _ *stelpb.SubmitTransactionRequest, _ ...grpc.CallOption) (*stelpb.SubmitTransactionResponse, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			gotNetwork = md.Get(metadataNetwork)[0]
			gotChain = md.Get(metadataChain)[0]
			return &stelpb.SubmitTransactionResponse{TxStatus: stelpb.TxStatus_TX_STATUS_SUCCESS}, nil
		},
	}

	sc := &stellarClient{relayID: relayID, client: inner}
	_, err := sc.SubmitTransaction(context.Background(), &stelpb.SubmitTransactionRequest{
		ContractId: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "ping",
	})
	require.NoError(t, err)
	require.Equal(t, relayID.Network, gotNetwork)
	require.Equal(t, relayID.ChainID, gotChain)
}

func TestStellarServer_SubmitTransaction(t *testing.T) {
	t.Parallel()

	relayID := types.RelayID{Network: "N1", ChainID: "C1"}
	mockRelayer := coremocks.NewRelayer(t)
	mockSvc := stellarmocks.NewStellarService(t)

	mockRelayer.On("Stellar").Return(mockSvc, nil).Once()

	rs := &TestRelayerSet{relayers: map[types.RelayID]core.Relayer{relayID: mockRelayer}}
	srv, _ := NewRelayerSetServer(logger.Nop(), rs, &net.BrokerExt{
		BrokerConfig: net.BrokerConfig{Logger: logger.Nop()},
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		metadataNetwork, relayID.Network,
		metadataChain, relayID.ChainID,
	))

	domainReq := stellartypes.SubmitTransactionRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "transfer",
	}
	expected := &stellartypes.SubmitTransactionResponse{
		TxStatus:         stellartypes.TxSuccess,
		TxHash:           "hash-abc",
		TxIdempotencyKey: "idem-1",
	}
	mockSvc.EXPECT().SubmitTransaction(mock.Anything, domainReq).Return(expected, nil).Once()

	pReq, err := stelpb.ConvertSubmitTransactionRequestToProto(domainReq)
	require.NoError(t, err)

	resp, err := srv.stellar.SubmitTransaction(ctx, pReq)
	require.NoError(t, err)
	require.Equal(t, stelpb.TxStatus_TX_STATUS_SUCCESS, resp.GetTxStatus())
	require.Equal(t, expected.TxHash, resp.GetTxHash())
}

func TestStellarServer_SubmitTransaction_InvalidRequest(t *testing.T) {
	t.Parallel()

	relayID := types.RelayID{Network: "N1", ChainID: "C1"}
	mockRelayer := coremocks.NewRelayer(t)
	mockRelayer.On("Stellar").Return(stellarmocks.NewStellarService(t), nil).Once()

	rs := &TestRelayerSet{relayers: map[types.RelayID]core.Relayer{relayID: mockRelayer}}
	srv, _ := NewRelayerSetServer(logger.Nop(), rs, &net.BrokerExt{
		BrokerConfig: net.BrokerConfig{Logger: logger.Nop()},
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		metadataNetwork, relayID.Network,
		metadataChain, relayID.ChainID,
	))

	_, err := srv.stellar.SubmitTransaction(ctx, &stelpb.SubmitTransactionRequest{Function: "only-fn"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid SubmitTransaction request")
}

func TestStellarServer_SubmitTransaction_ServiceError(t *testing.T) {
	t.Parallel()

	relayID := types.RelayID{Network: "N1", ChainID: "C1"}
	mockRelayer := coremocks.NewRelayer(t)
	mockSvc := stellarmocks.NewStellarService(t)
	mockRelayer.On("Stellar").Return(mockSvc, nil).Once()

	rs := &TestRelayerSet{relayers: map[types.RelayID]core.Relayer{relayID: mockRelayer}}
	srv, _ := NewRelayerSetServer(logger.Nop(), rs, &net.BrokerExt{
		BrokerConfig: net.BrokerConfig{Logger: logger.Nop()},
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		metadataNetwork, relayID.Network,
		metadataChain, relayID.ChainID,
	))

	domainReq := stellartypes.SubmitTransactionRequest{
		ContractID: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABSC4",
		Function:   "fail",
	}
	mockSvc.EXPECT().SubmitTransaction(mock.Anything, domainReq).
		Return(nil, errors.New("txm unavailable")).Once()

	pReq, err := stelpb.ConvertSubmitTransactionRequestToProto(domainReq)
	require.NoError(t, err)

	_, err = srv.stellar.SubmitTransaction(ctx, pReq)
	require.Error(t, err)
}
