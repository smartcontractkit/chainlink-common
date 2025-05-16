package relayerset

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	evm2 "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type evm struct {
	relayID types.RelayID
	client  *Client
}

var _ evm2.EVMClient = (*evm)(nil)

func (e evm) GetTransactionFee(ctx context.Context, in *evm2.GetTransactionFeeRequest, opts ...grpc.CallOption) (*evm2.GetTransactionFeeReply, error) {
	return e.client.relayerSetClient.EVMGetTransactionFee(ctx, &relayerset.EVMGetTransactionFeeRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in},
		opts...)
}

func (e evm) CallContract(ctx context.Context, in *evm2.CallContractRequest, opts ...grpc.CallOption) (*evm2.CallContractReply, error) {
	return e.client.relayerSetClient.EVMCallContract(ctx, &relayerset.EVMCallContractRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) FilterLogs(ctx context.Context, in *evm2.FilterLogsRequest, opts ...grpc.CallOption) (*evm2.FilterLogsReply, error) {
	return e.client.relayerSetClient.EVMFilterLogs(ctx, &relayerset.EVMFilterLogsRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) BalanceAt(ctx context.Context, in *evm2.BalanceAtRequest, opts ...grpc.CallOption) (*evm2.BalanceAtReply, error) {
	return e.client.relayerSetClient.EVMBalanceAt(ctx, &relayerset.EVMBalanceAtRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) EstimateGas(ctx context.Context, in *evm2.EstimateGasRequest, opts ...grpc.CallOption) (*evm2.EstimateGasReply, error) {
	return e.client.relayerSetClient.EVMEstimateGas(ctx, &relayerset.EVMEstimateGasRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) TransactionByHash(ctx context.Context, in *evm2.TransactionByHashRequest, opts ...grpc.CallOption) (*evm2.TransactionByHashReply, error) {
	return e.client.relayerSetClient.EVMTransactionByHash(ctx, &relayerset.EVMTransactionByHashRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) TransactionReceipt(ctx context.Context, in *evm2.TransactionReceiptRequest, opts ...grpc.CallOption) (*evm2.TransactionReceiptReply, error) {
	return e.client.relayerSetClient.EVMTransactionReceipt(ctx, &relayerset.EVMReceiptRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) LatestAndFinalizedHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*evm2.LatestAndFinalizedHeadReply, error) {
	return e.client.relayerSetClient.EVMLatestAndFinalizedHead(ctx, &relayerset.LatestHeadRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
	}, opts...)
}

func (e evm) QueryTrackedLogs(ctx context.Context, in *evm2.QueryTrackedLogsRequest, opts ...grpc.CallOption) (*evm2.QueryTrackedLogsReply, error) {
	return e.client.relayerSetClient.EVMQueryTrackedLogs(ctx, &relayerset.EVMQueryTrackedLogsRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) RegisterLogTracking(ctx context.Context, in *evm2.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.relayerSetClient.EVMRegisterLogTracking(ctx, &relayerset.EVMRegisterLogTrackingRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) UnregisterLogTracking(ctx context.Context, in *evm2.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return e.client.relayerSetClient.EVMUnregisterLogTracking(ctx, &relayerset.EVMUnregisterLogTrackingRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}

func (e evm) GetTransactionStatus(ctx context.Context, in *pb.GetTransactionStatusRequest, opts ...grpc.CallOption) (*pb.GetTransactionStatusReply, error) {
	return e.client.relayerSetClient.EVMGetTransactionStatus(ctx, &relayerset.EVMGetTransactionStatusRequest{
		RelayerId: &relayerset.RelayerId{
			Network: e.relayID.Network,
			ChainId: e.relayID.ChainID,
		},
		Request: in,
	}, opts...)
}
