package relayerset

import (
	"context"
	"fmt"

	stelpb "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// stellarClient wraps a StellarClient by attaching a RelayerID to every request via context metadata.
type stellarClient struct {
	relayID types.RelayID
	client  stelpb.StellarClient
}

var _ stelpb.StellarClient = (*stellarClient)(nil)

func (sc *stellarClient) GetLedgerEntries(ctx context.Context, in *stelpb.GetLedgerEntriesRequest, opts ...grpc.CallOption) (*stelpb.GetLedgerEntriesResponse, error) {
	return sc.client.GetLedgerEntries(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) GetLatestLedger(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*stelpb.GetLatestLedgerResponse, error) {
	return sc.client.GetLatestLedger(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) ReadContract(ctx context.Context, in *stelpb.ReadContractRequest, opts ...grpc.CallOption) (*stelpb.ReadContractResponse, error) {
	return sc.client.ReadContract(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) SubmitTransaction(ctx context.Context, in *stelpb.SubmitTransactionRequest, opts ...grpc.CallOption) (*stelpb.SubmitTransactionResponse, error) {
	return sc.client.SubmitTransaction(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) GetTransactionStatus(ctx context.Context, in *stelpb.GetTransactionStatusRequest, opts ...grpc.CallOption) (*stelpb.GetTransactionStatusResponse, error) {
	return sc.client.GetTransactionStatus(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) SimulateTransaction(ctx context.Context, in *stelpb.SimulateTransactionRequest, opts ...grpc.CallOption) (*stelpb.SimulateTransactionResponse, error) {
	return sc.client.SimulateTransaction(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) GetTransactionResult(ctx context.Context, in *stelpb.GetTransactionResultRequest, opts ...grpc.CallOption) (*stelpb.GetTransactionResultResponse, error) {
	return sc.client.GetTransactionResult(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *stellarClient) GetTransactionFee(ctx context.Context, in *stelpb.GetTransactionFeeRequest, opts ...grpc.CallOption) (*stelpb.GetTransactionFeeResponse, error) {
	return sc.client.GetTransactionFee(appendRelayID(ctx, sc.relayID), in, opts...)
}

// stellarServer implements stelpb.StellarServer by routing each RPC through the RelayerSet.
type stellarServer struct {
	stelpb.UnimplementedStellarServer
	parent *Server
}

var _ stelpb.StellarServer = (*stellarServer)(nil)

func (ss *stellarServer) GetLedgerEntries(ctx context.Context, req *stelpb.GetLedgerEntriesRequest) (*stelpb.GetLedgerEntriesResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	dReq, err := stelpb.ConvertGetLedgerEntriesRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid GetLedgerEntries request: %w", err)
	}
	dResp, err := svc.GetLedgerEntries(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertGetLedgerEntriesResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid GetLedgerEntries response: %w", err)
	}
	return pResp, nil
}

func (ss *stellarServer) GetLatestLedger(ctx context.Context, _ *emptypb.Empty) (*stelpb.GetLatestLedgerResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	dResp, err := svc.GetLatestLedger(ctx)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertGetLatestLedgerResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid GetLatestLedger response: %w", err)
	}
	return pResp, nil
}

func (ss *stellarServer) ReadContract(ctx context.Context, req *stelpb.ReadContractRequest) (*stelpb.ReadContractResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	dReq, err := stelpb.ConvertReadContractRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid ReadContract request: %w", err)
	}
	dResp, err := svc.ReadContract(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &stelpb.ReadContractResponse{
		Result:         dResp.Result,
		LedgerSequence: dResp.LedgerSequence,
		Error:          dResp.Error,
	}, nil
}

func (ss *stellarServer) SubmitTransaction(ctx context.Context, req *stelpb.SubmitTransactionRequest) (*stelpb.SubmitTransactionResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	dReq, err := stelpb.ConvertSubmitTransactionRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid SubmitTransaction request: %w", err)
	}
	dResp, err := svc.SubmitTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertSubmitTransactionResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid SubmitTransaction response: %w", err)
	}
	return pResp, nil
}

func (ss *stellarServer) GetTransactionStatus(ctx context.Context, req *stelpb.GetTransactionStatusRequest) (*stelpb.GetTransactionStatusResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	status, err := svc.GetTransactionStatus(ctx, req.GetTransactionId())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &stelpb.GetTransactionStatusResponse{Status: int32(status)}, nil
}

func (ss *stellarServer) SimulateTransaction(ctx context.Context, req *stelpb.SimulateTransactionRequest) (*stelpb.SimulateTransactionResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	dReq, err := stelpb.ConvertSimulateTransactionRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid SimulateTransaction request: %w", err)
	}
	dResp, err := svc.SimulateTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertSimulateTransactionResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid SimulateTransaction response: %w", err)
	}
	return pResp, nil
}

func (ss *stellarServer) GetTransactionResult(ctx context.Context, req *stelpb.GetTransactionResultRequest) (*stelpb.GetTransactionResultResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	dResp, err := svc.GetTransactionResult(ctx, req.GetTransactionId())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertTxResultToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid GetTransactionResult response: %w", err)
	}
	return pResp, nil
}

func (ss *stellarServer) GetTransactionFee(ctx context.Context, req *stelpb.GetTransactionFeeRequest) (*stelpb.GetTransactionFeeResponse, error) {
	svc, err := ss.parent.getStellarService(ctx)
	if err != nil {
		return nil, err
	}
	fee, err := svc.GetTransactionFee(ctx, req.GetTransactionId())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &stelpb.GetTransactionFeeResponse{Fee: fee.Int64()}, nil
}

// getStellarService extracts the RelayID from context metadata and returns the StellarService
// for the corresponding relayer.
func (s *Server) getStellarService(ctx context.Context) (types.StellarService, error) {
	id, err := readRelayID(ctx)
	if err != nil {
		return nil, err
	}
	idT := relayerset.RelayerId{Network: id.Network, ChainId: id.ChainID}
	r, err := s.getRelayer(ctx, &idT)
	if err != nil {
		return nil, err
	}
	return r.Stellar()
}
