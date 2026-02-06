package relayer

import (
	"context"
	"fmt"

	solpb "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ types.SolanaService = (*SolClient)(nil)

type SolClient struct {
	grpcClient solpb.SolanaClient
}

func NewSolanaClient(client solpb.SolanaClient) *SolClient {
	return &SolClient{
		grpcClient: client,
	}
}

func (sc *SolClient) GetLatestLPBlock(ctx context.Context) (*solana.LPBlock, error) {
	resp, err := sc.grpcClient.GetLatestLPBlock(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solana.LPBlock{
		Slot: resp.GetSlot(),
	}, nil
}

func (sc *SolClient) GetFiltersNames(ctx context.Context) ([]string, error) {
	resp, err := sc.grpcClient.GetFiltersNames(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return resp.GetItems(), nil
}

func (sc *SolClient) SubmitTransaction(ctx context.Context, req solana.SubmitTransactionRequest) (*solana.SubmitTransactionReply, error) {
	pReq := solpb.ConvertSubmitTransactionRequestToProto(req)

	pResp, err := sc.grpcClient.SubmitTransaction(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := solpb.ConvertSubmitTransactionReplyFromProto(pResp)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return dResp, nil
}

func (sc *SolClient) RegisterLogTracking(ctx context.Context, req solana.LPFilterQuery) error {
	filter := solpb.ConvertLPFilterQueryToProto(&req)
	_, err := sc.grpcClient.RegisterLogTracking(ctx, &solpb.RegisterLogTrackingRequest{
		Filter: filter,
	})
	return net.WrapRPCErr(err)
}

func (sc *SolClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := sc.grpcClient.UnregisterLogTracking(ctx, &solpb.UnregisterLogTrackingRequest{FilterName: filterName})
	return net.WrapRPCErr(err)
}

func (sc *SolClient) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort) ([]*solana.Log, error) {
	pExprs, err := solpb.ConvertExpressionsToProto(filterQuery)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoLimitAndSort, err := chaincommonpb.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	pReq := &solpb.QueryTrackedLogsRequest{
		FilterQuery:  pExprs,
		LimitAndSort: protoLimitAndSort,
	}

	pResp, err := sc.grpcClient.QueryTrackedLogs(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	out := make([]*solana.Log, 0, len(pResp.Logs))
	for _, pl := range pResp.Logs {
		dl, err := solpb.ConvertLogFromProto(pl)
		if err != nil {
			return nil, net.WrapRPCErr(err)
		}
		out = append(out, dl)
	}

	return out, nil
}

func (sc *SolClient) GetBalance(ctx context.Context, req solana.GetBalanceRequest) (*solana.GetBalanceReply, error) {
	pReq := solpb.ConvertGetBalanceRequestToProto(req)
	pResp, err := sc.grpcClient.GetBalance(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetBalanceReplyFromProto(pResp), nil
}

func (sc *SolClient) GetAccountInfoWithOpts(ctx context.Context, req solana.GetAccountInfoRequest) (*solana.GetAccountInfoReply, error) {
	pReq := &solpb.GetAccountInfoWithOptsRequest{
		Account: req.Account[:],
		Opts:    solpb.ConvertGetAccountInfoOptsToProto(req.Opts),
	}
	pResp, err := sc.grpcClient.GetAccountInfoWithOpts(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	acc, err := solpb.ConvertAccountFromProto(pResp.Value)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply := &solana.GetAccountInfoReply{
		RPCContext: solpb.ConvertRPCContextFromProto(pResp.RpcContext),
		Value:      acc,
	}

	return reply, nil
}

func (sc *SolClient) GetMultipleAccountsWithOpts(ctx context.Context, req solana.GetMultipleAccountsRequest) (*solana.GetMultipleAccountsReply, error) {
	pReq := solpb.ConvertGetMultipleAccountsRequestToProto(&req)
	pResp, err := sc.grpcClient.GetMultipleAccountsWithOpts(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	reply, err := solpb.ConvertGetMultipleAccountsReplyFromProto(pResp)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply, nil
}

func (sc *SolClient) GetBlock(ctx context.Context, req solana.GetBlockRequest) (*solana.GetBlockReply, error) {
	pReq := solpb.ConvertGetBlockRequestToProto(&req)
	pResp, err := sc.grpcClient.GetBlock(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := solpb.ConvertGetBlockOptsReplyFromProto(pResp)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply, nil
}

func (sc *SolClient) GetSlotHeight(ctx context.Context, req solana.GetSlotHeightRequest) (*solana.GetSlotHeightReply, error) {
	pReq := solpb.ConvertGetSlotHeightRequestToProto(req)
	pResp, err := sc.grpcClient.GetSlotHeight(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetSlotHeightReplyFromProto(pResp), nil
}

func (sc *SolClient) GetTransaction(ctx context.Context, req solana.GetTransactionRequest) (*solana.GetTransactionReply, error) {
	pReq := solpb.ConvertGetTransactionRequestToProto(req)
	pResp, err := sc.grpcClient.GetTransaction(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	reply, err := solpb.ConvertGetTransactionReplyFromProto(pResp)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply, nil
}

func (sc *SolClient) GetFeeForMessage(ctx context.Context, req solana.GetFeeForMessageRequest) (*solana.GetFeeForMessageReply, error) {
	pReq := solpb.ConvertGetFeeForMessageRequestToProto(&req)
	pResp, err := sc.grpcClient.GetFeeForMessage(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetFeeForMessageReplyFromProto(pResp), nil
}

func (sc *SolClient) GetSignatureStatuses(ctx context.Context, req solana.GetSignatureStatusesRequest) (*solana.GetSignatureStatusesReply, error) {
	pReq := solpb.ConvertGetSignatureStatusesRequestToProto(&req)
	pResp, err := sc.grpcClient.GetSignatureStatuses(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetSignatureStatusesReplyFromProto(pResp), nil
}

func (sc *SolClient) SimulateTX(ctx context.Context, req solana.SimulateTXRequest) (*solana.SimulateTXReply, error) {
	pReq := solpb.ConvertSimulateTXRequestToProto(req)
	pResp, err := sc.grpcClient.SimulateTX(ctx, pReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := solpb.ConvertSimulateTXReplyFromProto(pResp)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply, nil
}

type solServer struct {
	solpb.UnimplementedSolanaServer

	*net.BrokerExt

	impl types.SolanaService
}

var _ solpb.SolanaServer = (*solServer)(nil)

func newSolServer(impl types.SolanaService, b *net.BrokerExt) *solServer {
	return &solServer{impl: impl, BrokerExt: b.WithName("SolanaServer")}
}

func (s *solServer) GetLatestLPBlock(ctx context.Context, _ *emptypb.Empty) (*solpb.GetLatestLPBlockReply, error) {
	dResp, err := s.impl.GetLatestLPBlock(ctx)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.GetLatestLPBlockReply{
		Slot: dResp.Slot,
	}, nil
}

func (s *solServer) GetFiltersNames(ctx context.Context, _ *emptypb.Empty) (*solpb.GetFiltersNamesReply, error) {
	names, err := s.impl.GetFiltersNames(ctx)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.GetFiltersNamesReply{
		Items: names,
	}, nil
}

func (s *solServer) SubmitTransaction(ctx context.Context, req *solpb.SubmitTransactionRequest) (*solpb.SubmitTransactionReply, error) {
	dReq, err := solpb.ConvertSubmitTransactionRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := s.impl.SubmitTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	pResp := solpb.ConvertSubmitTransactionReplyToProto(dResp)
	return pResp, nil
}

func (s *solServer) RegisterLogTracking(ctx context.Context, req *solpb.RegisterLogTrackingRequest) (*solpb.RegisterLogTrackingReply, error) {
	if req.Filter == nil {
		return nil, net.WrapRPCErr(fmt.Errorf("missing filter"))
	}

	filter, err := solpb.ConvertLPFilterQueryFromProto(req.Filter)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	if err := s.impl.RegisterLogTracking(ctx, *filter); err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.RegisterLogTrackingReply{}, nil
}

func (s *solServer) UnregisterLogTracking(ctx context.Context, req *solpb.UnregisterLogTrackingRequest) (*solpb.UnregisterLogTrackingReply, error) {
	if err := s.impl.UnregisterLogTracking(ctx, req.GetFilterName()); err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.UnregisterLogTrackingReply{}, nil
}

func (s *solServer) QueryTrackedLogs(ctx context.Context, req *solpb.QueryTrackedLogsRequest) (*solpb.QueryTrackedLogsReply, error) {
	dExprs, err := solpb.ConvertExpressionsFromProto(req.GetFilterQuery())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	ls, err := chaincommonpb.ConvertLimitAndSortFromProto(req.GetLimitAndSort())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	logs, err := s.impl.QueryTrackedLogs(ctx, dExprs, ls)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	out := make([]*solpb.Log, 0, len(logs))
	for _, l := range logs {
		out = append(out, solpb.ConvertLogToProto(l))
	}

	return &solpb.QueryTrackedLogsReply{Logs: out}, nil
}

func (s *solServer) GetBalance(ctx context.Context, req *solpb.GetBalanceRequest) (*solpb.GetBalanceReply, error) {
	dReq, err := solpb.ConvertGetBalanceRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := s.impl.GetBalance(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetBalanceReplyToProto(dResp), nil
}

func (s *solServer) GetAccountInfoWithOpts(ctx context.Context, req *solpb.GetAccountInfoWithOptsRequest) (*solpb.GetAccountInfoWithOptsReply, error) {
	addr, err := solpb.ConvertPublicKeyFromProto(req.GetAccount())
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	opts := solpb.ConvertGetAccountInfoOptsFromProto(req.GetOpts())

	dReq := solana.GetAccountInfoRequest{Account: addr, Opts: opts}
	dResp, err := s.impl.GetAccountInfoWithOpts(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &solpb.GetAccountInfoWithOptsReply{
		RpcContext: solpb.ConvertRPCContextToProto(dResp.RPCContext),
		Value:      solpb.ConvertAccountToProto(dResp.Value),
	}, nil
}

func (s *solServer) GetMultipleAccountsWithOpts(ctx context.Context, req *solpb.GetMultipleAccountsWithOptsRequest) (*solpb.GetMultipleAccountsWithOptsReply, error) {
	dReq := solpb.ConvertGetMultipleAccountsRequestFromProto(req)
	dResp, err := s.impl.GetMultipleAccountsWithOpts(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetMultipleAccountsReplyToProto(dResp), nil
}

func (s *solServer) GetBlock(ctx context.Context, req *solpb.GetBlockRequest) (*solpb.GetBlockReply, error) {
	dReq := solpb.ConvertGetBlockRequestFromProto(req)
	dResp, err := s.impl.GetBlock(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetBlockReplyToProto(dResp), nil
}

func (s *solServer) GetSlotHeight(ctx context.Context, req *solpb.GetSlotHeightRequest) (*solpb.GetSlotHeightReply, error) {
	dReq := solpb.ConvertGetSlotHeightRequestFromProto(req)
	dResp, err := s.impl.GetSlotHeight(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetSlotHeightReplyToProto(dResp), nil
}

func (s *solServer) GetTransaction(ctx context.Context, req *solpb.GetTransactionRequest) (*solpb.GetTransactionReply, error) {
	dReq, err := solpb.ConvertGetTransactionRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := s.impl.GetTransaction(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetTransactionReplyToProto(dResp), nil
}

func (s *solServer) GetFeeForMessage(ctx context.Context, req *solpb.GetFeeForMessageRequest) (*solpb.GetFeeForMessageReply, error) {
	dReq := solpb.ConvertGetFeeForMessageRequestFromProto(req)
	dResp, err := s.impl.GetFeeForMessage(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}
	return solpb.ConvertGetFeeForMessageReplyToProto(dResp), nil
}

func (s *solServer) GetSignatureStatuses(ctx context.Context, req *solpb.GetSignatureStatusesRequest) (*solpb.GetSignatureStatusesReply, error) {
	dReq, err := solpb.ConvertGetSignatureStatusesRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := s.impl.GetSignatureStatuses(ctx, *dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertGetSignatureStatusesReplyToProto(dResp), nil
}

func (s *solServer) SimulateTX(ctx context.Context, req *solpb.SimulateTXRequest) (*solpb.SimulateTXReply, error) {
	dReq, err := solpb.ConvertSimulateTXRequestFromProto(req)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	dResp, err := s.impl.SimulateTX(ctx, dReq)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return solpb.ConvertSimulateTXReplyToProto(dResp), nil
}
