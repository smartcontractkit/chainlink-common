package relayer

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/protobuf/types/known/emptypb"

	stelpb "github.com/smartcontractkit/chainlink-common/pkg/chains/stellar"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

var _ types.StellarService = (*StellarClient)(nil)

// StellarClient wraps a stelpb.StellarClient gRPC stub and exposes it as types.StellarService.
type StellarClient struct {
	grpcClient stelpb.StellarClient
}

// NewStellarClient returns a StellarClient that delegates to the provided gRPC client.
func NewStellarClient(client stelpb.StellarClient) *StellarClient {
	return &StellarClient{grpcClient: client}
}

func (sc *StellarClient) GetLedgerEntries(ctx context.Context, req stellar.GetLedgerEntriesRequest) (stellar.GetLedgerEntriesResponse, error) {
	pReq, err := stelpb.ConvertGetLedgerEntriesRequestToProto(req)
	if err != nil {
		return stellar.GetLedgerEntriesResponse{}, fmt.Errorf("invalid GetLedgerEntries request: %w", err)
	}
	pResp, err := sc.grpcClient.GetLedgerEntries(ctx, pReq)
	if err != nil {
		return stellar.GetLedgerEntriesResponse{}, net.WrapRPCErr(err)
	}
	resp, err := stelpb.ConvertGetLedgerEntriesResponseFromProto(pResp)
	if err != nil {
		return stellar.GetLedgerEntriesResponse{}, fmt.Errorf("invalid GetLedgerEntries response: %w", err)
	}
	return resp, nil
}

func (sc *StellarClient) GetLatestLedger(ctx context.Context) (stellar.GetLatestLedgerResponse, error) {
	pResp, err := sc.grpcClient.GetLatestLedger(ctx, &emptypb.Empty{})
	if err != nil {
		return stellar.GetLatestLedgerResponse{}, net.WrapRPCErr(err)
	}
	resp, err := stelpb.ConvertGetLatestLedgerResponseFromProto(pResp)
	if err != nil {
		return stellar.GetLatestLedgerResponse{}, fmt.Errorf("invalid GetLatestLedger response: %w", err)
	}
	return resp, nil
}

func (sc *StellarClient) ReadContract(ctx context.Context, req stellar.ReadContractRequest) (stellar.ReadContractResponse, error) {
	pReq, err := stelpb.ConvertReadContractRequestToProto(req)
	if err != nil {
		return stellar.ReadContractResponse{}, fmt.Errorf("invalid ReadContract request: %w", err)
	}
	pResp, err := sc.grpcClient.ReadContract(ctx, pReq)
	if err != nil {
		return stellar.ReadContractResponse{}, net.WrapRPCErr(err)
	}
	return stellar.ReadContractResponse{
		Result:         pResp.Result,
		LedgerSequence: pResp.LedgerSequence,
		Error:          pResp.Error,
	}, nil
}

func (sc *StellarClient) SubmitTransaction(ctx context.Context, req stellar.SubmitTransactionRequest) (*stellar.SubmitTransactionResponse, error) {
	pReq, err := stelpb.ConvertSubmitTransactionRequestToProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid SubmitTransaction request: %w", err)
	}
	pResp, err := sc.grpcClient.SubmitTransaction(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	resp, err := stelpb.ConvertSubmitTransactionResponseFromProto(pResp)
	if err != nil {
		return nil, fmt.Errorf("invalid SubmitTransaction response: %w", err)
	}
	return resp, nil
}

func (sc *StellarClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	pReq := &stelpb.GetTransactionStatusRequest{TransactionId: transactionID}
	pResp, err := sc.grpcClient.GetTransactionStatus(ctx, pReq)
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}
	return types.TransactionStatus(pResp.GetStatus()), nil
}

func (sc *StellarClient) SimulateTransaction(ctx context.Context, req stellar.SimulateTransactionRequest) (stellar.SimulateTransactionResponse, error) {
	pReq, err := stelpb.ConvertSimulateTransactionRequestToProto(req)
	if err != nil {
		return stellar.SimulateTransactionResponse{}, fmt.Errorf("invalid SimulateTransaction request: %w", err)
	}
	pResp, err := sc.grpcClient.SimulateTransaction(ctx, pReq)
	if err != nil {
		return stellar.SimulateTransactionResponse{}, net.WrapRPCErr(err)
	}
	resp, err := stelpb.ConvertSimulateTransactionResponseFromProto(pResp)
	if err != nil {
		return stellar.SimulateTransactionResponse{}, fmt.Errorf("invalid SimulateTransaction response: %w", err)
	}
	return resp, nil
}

func (sc *StellarClient) GetTransactionResult(ctx context.Context, transactionID string) (stellar.TxResult, error) {
	pReq := &stelpb.GetTransactionResultRequest{TransactionId: transactionID}
	pResp, err := sc.grpcClient.GetTransactionResult(ctx, pReq)
	if err != nil {
		return stellar.TxResult{}, net.WrapRPCErr(err)
	}
	resp, err := stelpb.ConvertTxResultFromProto(pResp)
	if err != nil {
		return stellar.TxResult{}, fmt.Errorf("invalid GetTransactionResult response: %w", err)
	}
	return resp, nil
}

func (sc *StellarClient) GetTransactionFee(ctx context.Context, transactionID string) (*big.Int, error) {
	pReq := &stelpb.GetTransactionFeeRequest{TransactionId: transactionID}
	pResp, err := sc.grpcClient.GetTransactionFee(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return big.NewInt(pResp.GetFee()), nil
}

// stellarServer wraps types.StellarService and exposes it as a stelpb.StellarServer gRPC endpoint.
type stellarServer struct {
	stelpb.UnimplementedStellarServer

	*net.BrokerExt

	impl types.StellarService
}

var _ stelpb.StellarServer = (*stellarServer)(nil)

func newStellarServer(impl types.StellarService, b *net.BrokerExt) *stellarServer {
	return &stellarServer{impl: impl, BrokerExt: b.WithName("StellarServer")}
}

func (s *stellarServer) GetLedgerEntries(ctx context.Context, req *stelpb.GetLedgerEntriesRequest) (*stelpb.GetLedgerEntriesResponse, error) {
	dReq, err := stelpb.ConvertGetLedgerEntriesRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid GetLedgerEntries request: %w", err)
	}
	dResp, err := s.impl.GetLedgerEntries(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertGetLedgerEntriesResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid GetLedgerEntries response: %w", err)
	}
	return pResp, nil
}

func (s *stellarServer) GetLatestLedger(ctx context.Context, _ *emptypb.Empty) (*stelpb.GetLatestLedgerResponse, error) {
	dResp, err := s.impl.GetLatestLedger(ctx)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertGetLatestLedgerResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid GetLatestLedger response: %w", err)
	}
	return pResp, nil
}

func (s *stellarServer) ReadContract(ctx context.Context, req *stelpb.ReadContractRequest) (*stelpb.ReadContractResponse, error) {
	dReq, err := stelpb.ConvertReadContractRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid ReadContract request: %w", err)
	}
	dResp, err := s.impl.ReadContract(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &stelpb.ReadContractResponse{
		Result:         dResp.Result,
		LedgerSequence: dResp.LedgerSequence,
		Error:          dResp.Error,
	}, nil
}

func (s *stellarServer) SubmitTransaction(ctx context.Context, req *stelpb.SubmitTransactionRequest) (*stelpb.SubmitTransactionResponse, error) {
	dReq, err := stelpb.ConvertSubmitTransactionRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid SubmitTransaction request: %w", err)
	}
	dResp, err := s.impl.SubmitTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertSubmitTransactionResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid SubmitTransaction response: %w", err)
	}
	return pResp, nil
}

func (s *stellarServer) GetTransactionStatus(ctx context.Context, req *stelpb.GetTransactionStatusRequest) (*stelpb.GetTransactionStatusResponse, error) {
	status, err := s.impl.GetTransactionStatus(ctx, req.GetTransactionId())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &stelpb.GetTransactionStatusResponse{Status: int32(status)}, nil
}

func (s *stellarServer) SimulateTransaction(ctx context.Context, req *stelpb.SimulateTransactionRequest) (*stelpb.SimulateTransactionResponse, error) {
	dReq, err := stelpb.ConvertSimulateTransactionRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid SimulateTransaction request: %w", err)
	}
	dResp, err := s.impl.SimulateTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertSimulateTransactionResponseToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid SimulateTransaction response: %w", err)
	}
	return pResp, nil
}

func (s *stellarServer) GetTransactionResult(ctx context.Context, req *stelpb.GetTransactionResultRequest) (*stelpb.GetTransactionResultResponse, error) {
	dResp, err := s.impl.GetTransactionResult(ctx, req.GetTransactionId())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	pResp, err := stelpb.ConvertTxResultToProto(dResp)
	if err != nil {
		return nil, fmt.Errorf("invalid GetTransactionResult response: %w", err)
	}
	return pResp, nil
}

func (s *stellarServer) GetTransactionFee(ctx context.Context, req *stelpb.GetTransactionFeeRequest) (*stelpb.GetTransactionFeeResponse, error) {
	fee, err := s.impl.GetTransactionFee(ctx, req.GetTransactionId())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return &stelpb.GetTransactionFeeResponse{Fee: fee.Int64()}, nil
}
