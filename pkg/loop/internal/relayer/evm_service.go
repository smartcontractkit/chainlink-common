package relayer

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// evm service client
type evmRelayerClient struct {
	cl pb.EVMClient
}

// Direct calls
func (e *evmRelayerClient) CallContract(ctx context.Context, msg *evm.CallMsg, confidence primitives.ConfidenceLevel) ([]byte, error) {
	call, err := callMsgToProto(msg)
	if err != nil {
		return nil, err
	}
	conf, err := contractreader.ConfidenceToProto(confidence)
	if err != nil {
		return nil, err
	}
	reply, err := e.cl.CallContract(ctx, &pb.CallContractRequest{
		Call:            call,
		ConfidenceLevel: conf,
	})
	if err != nil {
		return nil, err
	}

	return reply.Data, nil
}

func (e *evmRelayerClient) GetLogs(ctx context.Context, filterQuery evm.EVMFilterQuery) ([]*evm.Log, error) {
	reply, err := e.cl.GetLogs(ctx, &pb.GetLogsRequest{
		FilterQuery: evmFilterToProto(filterQuery),
	})

	if err != nil {
		return nil, err
	}

	return protoToLogs(reply.Logs), nil
}

func (e *evmRelayerClient) BalanceAt(ctx context.Context, account string, blockNumber *big.Int) (*big.Int, error) {
	reply, err := e.cl.BalanceAt(ctx, &pb.BalanceAtRequest{
		Account:     account,
		BlockNumber: pb.NewBigIntFromInt(blockNumber),
	})

	if err != nil {
		return nil, err
	}

	return reply.Balance.Int(), nil
}

func (e *evmRelayerClient) EstimateGas(ctx context.Context, msg *evm.CallMsg) (uint64, error) {
	call, err := callMsgToProto(msg)
	if err != nil {
		return 0, err
	}

	reply, err := e.cl.EstimateGas(ctx, &pb.EstimateGasRequest{
		Msg: call,
	})
	if err != nil {
		return 0, err
	}

	return reply.Gas, nil
}

func (e *evmRelayerClient) TransactionByHash(ctx context.Context, hash string) (*evm.Transaction, error) {
	reply, err := e.cl.GetTransactionByHash(ctx, &pb.GetTransactionByHashRequest{
		Hash: hash,
	})
	if err != nil {
		return nil, err
	}

	return protoToTransaction(reply.Transaction)
}

func (e *evmRelayerClient) TransactionReceipt(ctx context.Context, txHash string) (*evm.Receipt, error) {
	reply, err := e.cl.GetTransactionReceipt(ctx, &pb.GetReceiptRequest{Hash: txHash})
	if err != nil {
		return nil, err
	}

	return protoToReceipt(reply.Receipt)
}

// ChainService
func (e *evmRelayerClient) LatestAndFinalizedHead(ctx context.Context) (latest types.Head, finalized types.Head, err error) {
	reply, err := e.cl.LatestAndFinalizedHead(ctx, &emptypb.Empty{})
	if err != nil {
		return types.Head{}, types.Head{}, err
	}

	return protoToHead(reply.Latest), protoToHead(reply.Finalized), err

}
func (e *evmRelayerClient) QueryLogsFromCache(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	//TODO BCFR-1328
	return nil, errors.New("unimplemented")
}

func (e *evmRelayerClient) SubscribeLogTrigger(ctx context.Context, filterQuery []query.Expression) (chan<- *evm.Log, error) {
	//TODO BCFR-1328
	return nil, errors.New("unimplmented")
}

func (e *evmRelayerClient) RegisterLogTracking(ctx context.Context, filter evm.FilterQuery) error {
	_, err := e.cl.RegisterLogTracking(ctx, &pb.RegisterLogTrackingRequest{Filter: lPfilterToProto(filter)})
	return err
}

func (e *evmRelayerClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.cl.UnregisterLogTracking(ctx, &pb.UnregisterLogTrackingRequest{FilterName: filterName})
	return err
}

func (e *evmRelayerClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := e.cl.GetTransactionStatus(ctx, &pb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, err
	}

	return types.TransactionStatus(reply.TransactionStatus), nil
}

func (e *evmRelayerClient) GetTransactionFee(ctx context.Context, transactionID string) (*types.TransactionFee, error) {
	reply, err := e.cl.GetTransactionFee(ctx, &pb.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, err
	}

	return &types.TransactionFee{
		TransactionFee: reply.TransationFee.Int(),
	}, nil
}

// evm service server
var _ pb.EVMServer = (*evmServiceServer)(nil)

type evmServiceServer struct {
	*net.BrokerExt
	pb.UnimplementedEVMServer

	impl types.EVMService
}

func (e *evmServiceServer) GetTransactionFee(ctx context.Context, request *pb.GetTransactionFeeRequest) (*pb.GetTransactionFeeReply, error) {
	reply, err := e.impl.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(reply.TransactionFee),
	}, nil
}
func (e *evmServiceServer) CallContract(ctx context.Context, req *pb.CallContractRequest) (*pb.CallContractReply, error) {
	confidence, err := contractreader.ConfidenceFromProto(req.ConfidenceLevel)
	if err != nil {
		return nil, err
	}

	call, err := protoToCallMsg(req.Call)
	if err != nil {
		return nil, err
	}

	data, err := e.impl.CallContract(ctx, call, confidence)
	if err != nil {
		return nil, err
	}

	return &pb.CallContractReply{
		Data: data,
	}, nil
}
func (e *evmServiceServer) GetLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsReply, error) {
	f, err := protoToEvmFilter(req.FilterQuery)
	if err != nil {
		return nil, err
	}
	logs, err := e.impl.GetLogs(ctx, f)
	if err != nil {
		return nil, err
	}

	return &pb.GetLogsReply{
		Logs: logsToProto(logs),
	}, nil
}
func (e *evmServiceServer) BalanceAt(ctx context.Context, req *pb.BalanceAtRequest) (*pb.BalanceAtReply, error) {
	balance, err := e.impl.BalanceAt(ctx, req.Account, req.BlockNumber.Int())
	if err != nil {
		return nil, err
	}

	return &pb.BalanceAtReply{
		Balance: pb.NewBigIntFromInt(balance),
	}, nil
}

func (e *evmServiceServer) EstimateGas(ctx context.Context, req *pb.EstimateGasRequest) (*pb.EstimateGasReply, error) {
	call, err := protoToCallMsg(req.Msg)
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, call)
	if err != nil {
		return nil, err
	}

	return &pb.EstimateGasReply{
		Gas: gas,
	}, nil
}

func (e *evmServiceServer) GetTransactionByHash(ctx context.Context, req *pb.GetTransactionByHashRequest) (*pb.GetTransactionByHashReply, error) {
	tx, err := e.impl.TransactionByHash(ctx, req.Hash)
	if err != nil {
		return nil, err
	}

	pbtx, err := transactionToProto(tx)
	if err != nil {
		return nil, err
	}
	return &pb.GetTransactionByHashReply{
		Transaction: pbtx,
	}, nil
}

func (e *evmServiceServer) GetTransactionReceipt(ctx context.Context, req *pb.GetReceiptRequest) (*pb.GetReceiptReply, error) {
	rec, err := e.impl.TransactionReceipt(ctx, req.Hash)
	if err != nil {
		return nil, err
	}

	pbrec, err := receiptToProto(rec)
	if err != nil {
		return nil, err
	}
	return &pb.GetReceiptReply{
		Receipt: pbrec,
	}, nil
}

func (e *evmServiceServer) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*pb.LatestAndFinalizedHeadReply, error) {
	latest, finalized, err := e.impl.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.LatestAndFinalizedHeadReply{
		Latest:    headToProto(latest),
		Finalized: headToProto(finalized),
	}, nil
}

func (e *evmServiceServer) QueryLogsFromCache(context.Context, *pb.QueryLogsFromCacheRequest) (*pb.QueryLogsFromCacheReply, error) {
	return nil, errors.New("method QueryLogsFromCache not implemented")
}

func (e *evmServiceServer) SubscribeLogTrigger(*pb.SubscribeLogTriggerRequest, grpc.ServerStreamingServer[pb.LogTriggerReply]) error {
	return errors.New("method SubscribeLogTrigger not implemented")
}

func (e *evmServiceServer) RegisterLogTracking(ctx context.Context, req *pb.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	f, err := protoToLpFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	err = e.impl.RegisterLogTracking(ctx, f)
	return nil, err
}

func (e *evmServiceServer) UnregisterLogTracking(ctx context.Context, req *pb.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	err := e.impl.UnregisterLogTracking(ctx, req.FilterName)

	return nil, err
}

func (e *evmServiceServer) GetTransactionStatus(ctx context.Context, req *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusReply, error) {
	status, err := e.impl.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionStatusReply{TransactionStatus: pb.TransactionStatus(status)}, nil
}

func newEVMServiceServer(impl types.EVMService, b *net.BrokerExt) *evmServiceServer {
	return &evmServiceServer{impl: impl, BrokerExt: b.WithName("EVMServiceServer")}
}

var errEmptyMsg = errors.New("call msg can't be empty")

func protoToHead(h *pb.Head) types.Head {
	return types.Head{
		Height:    h.Height,
		Hash:      h.Hash,
		Timestamp: h.Timestamp,
	}
}

func headToProto(h types.Head) *pb.Head {
	return &pb.Head{
		Height:    h.Height,
		Hash:      h.Hash,
		Timestamp: h.Timestamp,
	}
}

var errEmptyReceipt = errors.New("receipt is empty")

func receiptToProto(r *evm.Receipt) (*pb.EVMReceipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}

	return &pb.EVMReceipt{
		PostState:         r.PostState,
		Status:            r.Status,
		Logs:              logsToProto(r.Logs),
		TxHash:            r.TxHash,
		ContractAddress:   r.ContractAddress,
		GasUsed:           r.GasUsed,
		BlockHash:         r.BlockHash,
		BlockNumber:       pb.NewBigIntFromInt(r.BlockNumber),
		TxIndex:           r.TransactionIndex,
		EffectiveGasPrice: pb.NewBigIntFromInt(r.EffectiveGasPrice),
	}, nil
}

func protoToReceipt(r *pb.EVMReceipt) (*evm.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}
	return &evm.Receipt{
		PostState:         r.PostState,
		Status:            r.Status,
		Logs:              protoToLogs(r.Logs),
		TxHash:            r.TxHash,
		ContractAddress:   r.ContractAddress,
		GasUsed:           r.GasUsed,
		BlockHash:         r.BlockHash,
		BlockNumber:       r.BlockNumber.Int(),
		TransactionIndex:  r.TxIndex,
		EffectiveGasPrice: r.EffectiveGasPrice.Int(),
	}, nil
}

var errEmptyTx = errors.New("transaction is empty")

func transactionToProto(tx *evm.Transaction) (*pb.EVMTransaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &pb.EVMTransaction{
		To:       tx.To,
		Data:     tx.Data,
		Hash:     tx.Hash,
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: pb.NewBigIntFromInt(tx.GasPrice),
		Value:    pb.NewBigIntFromInt(tx.Value),
	}, nil
}

func protoToTransaction(tx *pb.EVMTransaction) (*evm.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evm.Transaction{
		To:       tx.To,
		Data:     tx.Data,
		Hash:     tx.Hash,
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: tx.GasPrice.Int(),
		Value:    tx.Value.Int(),
	}, nil
}

func callMsgToProto(m *evm.CallMsg) (*pb.CallMsg, error) {
	if m == nil {
		return nil, errEmptyMsg
	}

	return &pb.CallMsg{
		From: m.From,
		To:   m.To,
		Data: m.Data,
	}, nil
}

func protoToCallMsg(p *pb.CallMsg) (*evm.CallMsg, error) {
	if p == nil {
		return nil, errEmptyMsg
	}

	return &evm.CallMsg{
		From: p.From,
		Data: p.Data,
		To:   p.To,
	}, nil
}

func lPfilterToProto(f evm.FilterQuery) *pb.LPFilter {
	return &pb.LPFilter{
		Name:          f.Name,
		RetentionTime: int64(f.Retention),
		Addresses:     f.Addresses,
		EventSigs:     f.EventSigs,
		Topic2:        f.Topic2,
		Topic3:        f.Topic3,
		Topic4:        f.Topic4,
		MaxLogsKept:   f.MaxLogsKept,
		LogsPerBlock:  f.LogsPerBlock,
	}
}

func protoToLpFilter(f *pb.LPFilter) (evm.FilterQuery, error) {
	if f == nil {
		return evm.FilterQuery{}, errEmptyFilter
	}

	return evm.FilterQuery{
		Name:         f.Name,
		Retention:    time.Duration(f.RetentionTime),
		Addresses:    f.Addresses,
		EventSigs:    f.EventSigs,
		Topic2:       f.Topic2,
		Topic3:       f.Topic3,
		Topic4:       f.Topic4,
		MaxLogsKept:  f.MaxLogsKept,
		LogsPerBlock: f.LogsPerBlock,
	}, nil
}

var errEmptyFilter = errors.New("filter cant be empty")

func protoToEvmFilter(f *pb.EVMFilterQuery) (evm.EVMFilterQuery, error) {
	if f == nil {
		return evm.EVMFilterQuery{}, errEmptyFilter
	}
	return evm.EVMFilterQuery{
		BlockHash: f.BlockHash,
		FromBlock: f.FromBlock.Int(),
		ToBlock:   f.ToBlock.Int(),
		Addresses: f.Addresses,
		Topics:    f.Topics,
	}, nil
}

func evmFilterToProto(f evm.EVMFilterQuery) *pb.EVMFilterQuery {
	return &pb.EVMFilterQuery{
		BlockHash: f.BlockHash,
		FromBlock: pb.NewBigIntFromInt(f.FromBlock),
		ToBlock:   pb.NewBigIntFromInt(f.ToBlock),
		Addresses: f.Addresses,
		Topics:    f.Topics,
	}
}

func protoToLogs(logs []*pb.Log) []*evm.Log {
	ret := make([]*evm.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, protoToLog(l))
	}

	return ret
}

func logsToProto(logs []*evm.Log) []*pb.Log {
	ret := make([]*pb.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, logToProto(l))
	}

	return ret
}

func logToProto(l *evm.Log) *pb.Log {
	return &pb.Log{
		Index:       l.LogIndex,
		BlockHash:   l.BlockHash,
		BlockNumber: pb.NewBigIntFromInt(l.BlockNumber),
		Topics:      l.Topics,
		EventSig:    l.EventSig,
		Address:     l.Address,
		TxHash:      l.TxHash,
		Data:        l.Data,
		Removed:     l.Removed,
	}
}

func protoToLog(l *pb.Log) *evm.Log {
	return &evm.Log{
		LogIndex:    l.Index,
		BlockHash:   l.BlockHash,
		BlockNumber: l.BlockNumber.Int(),
		Topics:      l.Topics,
		EventSig:    l.EventSig,
		Address:     l.Address,
		TxHash:      l.TxHash,
		Data:        l.Data,
		Removed:     l.Removed,
	}
}
