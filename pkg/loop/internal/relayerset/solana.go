package relayerset

import (
	"context"
	"fmt"

	solpb "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// solClient wraps the SolanaRelayerSetClient by attaching a RelayerID to SolClient requests.
// The attached RelayerID is stored in the context metadata.
type solClient struct {
	relayID types.RelayID
	client  solpb.SolanaClient
}

var _ solpb.SolanaClient = (*solClient)(nil)

func (sc *solClient) GetAccountInfoWithOpts(ctx context.Context, in *solpb.GetAccountInfoWithOptsRequest, opts ...grpc.CallOption) (*solpb.GetAccountInfoWithOptsReply, error) {
	return sc.client.GetAccountInfoWithOpts(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetBalance(ctx context.Context, in *solpb.GetBalanceRequest, opts ...grpc.CallOption) (*solpb.GetBalanceReply, error) {
	return sc.client.GetBalance(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetBlock(ctx context.Context, in *solpb.GetBlockRequest, opts ...grpc.CallOption) (*solpb.GetBlockReply, error) {
	return sc.client.GetBlock(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetLatestLPBlock(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*solpb.GetLatestLPBlockReply, error) {
	return sc.client.GetLatestLPBlock(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetFeeForMessage(ctx context.Context, in *solpb.GetFeeForMessageRequest, opts ...grpc.CallOption) (*solpb.GetFeeForMessageReply, error) {
	return sc.client.GetFeeForMessage(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetMultipleAccountsWithOpts(ctx context.Context, in *solpb.GetMultipleAccountsWithOptsRequest, opts ...grpc.CallOption) (*solpb.GetMultipleAccountsWithOptsReply, error) {
	return sc.client.GetMultipleAccountsWithOpts(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetSignatureStatuses(ctx context.Context, in *solpb.GetSignatureStatusesRequest, opts ...grpc.CallOption) (*solpb.GetSignatureStatusesReply, error) {
	return sc.client.GetSignatureStatuses(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetSlotHeight(ctx context.Context, in *solpb.GetSlotHeightRequest, opts ...grpc.CallOption) (*solpb.GetSlotHeightReply, error) {
	return sc.client.GetSlotHeight(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetTransaction(ctx context.Context, in *solpb.GetTransactionRequest, opts ...grpc.CallOption) (*solpb.GetTransactionReply, error) {
	return sc.client.GetTransaction(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) QueryTrackedLogs(ctx context.Context, in *solpb.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*solpb.QueryTrackedLogsReply, error) {
	return sc.client.QueryTrackedLogs(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) RegisterLogTracking(ctx context.Context, in *solpb.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*solpb.RegisterLogTrackingReply, error) {
	return sc.client.RegisterLogTracking(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) SimulateTX(ctx context.Context, in *solpb.SimulateTXRequest, opts ...grpc.CallOption) (*solpb.SimulateTXReply, error) {
	return sc.client.SimulateTX(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) SubmitTransaction(ctx context.Context, in *solpb.SubmitTransactionRequest, opts ...grpc.CallOption) (*solpb.SubmitTransactionReply, error) {
	return sc.client.SubmitTransaction(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) UnregisterLogTracking(ctx context.Context, in *solpb.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*solpb.UnregisterLogTrackingReply, error) {
	return sc.client.UnregisterLogTracking(appendRelayID(ctx, sc.relayID), in, opts...)
}

func (sc *solClient) GetFiltersNames(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*solpb.GetFiltersNamesReply, error) {
	return sc.client.GetFiltersNames(appendRelayID(ctx, sc.relayID), in, opts...)
}

type solServer struct {
	solpb.UnimplementedSolanaServer
	parent *Server
}

var _ solpb.SolanaServer = (*solServer)(nil)

// Server handlers
func (ss *solServer) GetLatestLPBlock(ctx context.Context, _ *emptypb.Empty) (*solpb.GetLatestLPBlockReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dResp, err := solService.GetLatestLPBlock(ctx)
	if err != nil {
		return nil, err
	}

	return &solpb.GetLatestLPBlockReply{
		Slot: dResp.Slot,
	}, nil
}

func (ss *solServer) SubmitTransaction(ctx context.Context, req *solpb.SubmitTransactionRequest) (*solpb.SubmitTransactionReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq, err := solpb.ConvertSubmitTransactionRequestFromProto(req)
	if err != nil {
		return nil, err
	}

	dResp, err := solService.SubmitTransaction(ctx, dReq)
	if err != nil {
		return nil, err
	}

	pResp := solpb.ConvertSubmitTransactionReplyToProto(dResp)
	return pResp, nil
}

func (ss *solServer) RegisterLogTracking(ctx context.Context, req *solpb.RegisterLogTrackingRequest) (*solpb.RegisterLogTrackingReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}
	if req.Filter == nil {
		return nil, fmt.Errorf("missing filter")
	}

	filter, err := solpb.ConvertLPFilterQueryFromProto(req.Filter)
	if err != nil {
		return nil, err
	}

	if err := solService.RegisterLogTracking(ctx, *filter); err != nil {
		return nil, err
	}

	return &solpb.RegisterLogTrackingReply{}, nil
}

func (ss *solServer) UnregisterLogTracking(ctx context.Context, req *solpb.UnregisterLogTrackingRequest) (*solpb.UnregisterLogTrackingReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}
	if err := solService.UnregisterLogTracking(ctx, req.GetFilterName()); err != nil {
		return nil, err
	}

	return &solpb.UnregisterLogTrackingReply{}, nil
}

func (ss *solServer) QueryTrackedLogs(ctx context.Context, req *solpb.QueryTrackedLogsRequest) (*solpb.QueryTrackedLogsReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dExprs, err := solpb.ConvertExpressionsFromProto(req.GetFilterQuery())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	ls, err := chaincommonpb.ConvertLimitAndSortFromProto(req.GetLimitAndSort())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	logs, err := solService.QueryTrackedLogs(ctx, dExprs, ls)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	out := make([]*solpb.Log, 0, len(logs))
	for _, l := range logs {
		out = append(out, solpb.ConvertLogToProto(l))
	}

	return &solpb.QueryTrackedLogsReply{Logs: out}, nil
}

func (ss *solServer) GetBalance(ctx context.Context, req *solpb.GetBalanceRequest) (*solpb.GetBalanceReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq, err := solpb.ConvertGetBalanceRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := solService.GetBalance(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetBalanceReplyToProto(dResp), nil
}

func (ss *solServer) GetAccountInfoWithOpts(ctx context.Context, req *solpb.GetAccountInfoWithOptsRequest) (*solpb.GetAccountInfoWithOptsReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	addr, err := solpb.ConvertPublicKeyFromProto(req.GetAccount())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	opts := solpb.ConvertGetAccountInfoOptsFromProto(req.GetOpts())

	dReq := solana.GetAccountInfoRequest{Account: addr, Opts: opts}
	dResp, err := solService.GetAccountInfoWithOpts(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.GetAccountInfoWithOptsReply{
		RpcContext: solpb.ConvertRPCContextToProto(dResp.RPCContext),
		Value:      solpb.ConvertAccountToProto(dResp.Value),
	}, nil
}

func (ss *solServer) GetMultipleAccountsWithOpts(ctx context.Context, req *solpb.GetMultipleAccountsWithOptsRequest) (*solpb.GetMultipleAccountsWithOptsReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq := solpb.ConvertGetMultipleAccountsRequestFromProto(req)
	dResp, err := solService.GetMultipleAccountsWithOpts(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetMultipleAccountsReplyToProto(dResp), nil
}

func (ss *solServer) GetBlock(ctx context.Context, req *solpb.GetBlockRequest) (*solpb.GetBlockReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq := solpb.ConvertGetBlockRequestFromProto(req)
	dResp, err := solService.GetBlock(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetBlockReplyToProto(dResp), nil
}

func (ss *solServer) GetSlotHeight(ctx context.Context, req *solpb.GetSlotHeightRequest) (*solpb.GetSlotHeightReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq := solpb.ConvertGetSlotHeightRequestFromProto(req)
	dResp, err := solService.GetSlotHeight(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetSlotHeightReplyToProto(dResp), nil
}

func (ss *solServer) GetTransaction(ctx context.Context, req *solpb.GetTransactionRequest) (*solpb.GetTransactionReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq, err := solpb.ConvertGetTransactionRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := solService.GetTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetTransactionReplyToProto(dResp), nil
}

func (ss *solServer) GetFeeForMessage(ctx context.Context, req *solpb.GetFeeForMessageRequest) (*solpb.GetFeeForMessageReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq := solpb.ConvertGetFeeForMessageRequestFromProto(req)
	dResp, err := solService.GetFeeForMessage(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetFeeForMessageReplyToProto(dResp), nil
}

func (ss *solServer) GetSignatureStatuses(ctx context.Context, req *solpb.GetSignatureStatusesRequest) (*solpb.GetSignatureStatusesReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq, err := solpb.ConvertGetSignatureStatusesRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := solService.GetSignatureStatuses(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetSignatureStatusesReplyToProto(dResp), nil
}

func (ss *solServer) SimulateTX(ctx context.Context, req *solpb.SimulateTXRequest) (*solpb.SimulateTXReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	dReq, err := solpb.ConvertSimulateTXRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := solService.SimulateTX(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertSimulateTXReplyToProto(dResp), nil
}

func (ss *solServer) GetFiltersNames(ctx context.Context, _ *emptypb.Empty) (*solpb.GetFiltersNamesReply, error) {
	solService, err := ss.parent.getSolService(ctx)
	if err != nil {
		return nil, err
	}

	names, err := solService.GetFiltersNames(ctx)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.GetFiltersNamesReply{
		Items: names,
	}, nil
}

func (s *Server) getSolService(ctx context.Context) (types.SolanaService, error) {
	id, err := readRelayID(ctx)
	if err != nil {
		return nil, err
	}
	idT := relayerset.RelayerId{Network: id.Network, ChainId: id.ChainID}
	r, err := s.getRelayer(ctx, &idT)
	if err != nil {
		return nil, err
	}

	return r.Solana()
}
