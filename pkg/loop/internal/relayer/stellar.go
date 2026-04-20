package relayer

import (
	"context"
	"fmt"

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
