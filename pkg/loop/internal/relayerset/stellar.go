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
	return stelpb.ConvertGetLedgerEntriesResponseToProto(dResp), nil
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
	return stelpb.ConvertGetLatestLedgerResponseToProto(dResp), nil
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
