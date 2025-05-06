package relayer

import (
	"context"
	"errors"
	"math/big"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	evmpb2 "github.com/smartcontractkit/chainlink-common/pkg/loop/pb/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

var _ types.EVMService = (*evmClient)(nil)

type evmClient struct {
	cl evmpb2.EVMClient
}

func (e *evmClient) GetTransactionFee(ctx context.Context, transactionID string) (*evm.TransactionFee, error) {
	reply, err := e.cl.GetTransactionFee(ctx, &evmpb2.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, err
	}

	return &evm.TransactionFee{
		TransactionFee: reply.TransationFee.Int(),
	}, nil
}

func (e *evmClient) CallContract(ctx context.Context, msg *evm.CallMsg, confidence primitives.ConfidenceLevel) ([]byte, error) {
	call, err := callMsgToProto(msg)
	if err != nil {
		return nil, err
	}
	conf, err := contractreader.ConfidenceToProto(confidence)
	if err != nil {
		return nil, err
	}
	reply, err := e.cl.CallContract(ctx, &evmpb2.CallContractRequest{
		Call:            call,
		ConfidenceLevel: conf,
	})
	if err != nil {
		return nil, err
	}

	return reply.Data.GetAbi(), nil
}

func (e *evmClient) GetLogs(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error) {
	reply, err := e.cl.GetLogs(ctx, &evmpb2.GetLogsRequest{
		FilterQuery: evmFilterToProto(filterQuery),
	})

	if err != nil {
		return nil, err
	}

	return protoToLogs(reply.Logs), nil
}

func (e *evmClient) BalanceAt(ctx context.Context, account string, blockNumber *big.Int) (*big.Int, error) {
	reply, err := e.cl.BalanceAt(ctx, &evmpb2.BalanceAtRequest{
		Account:     &evmpb2.Address{Address: account},
		BlockNumber: pb.NewBigIntFromInt(blockNumber),
	})

	if err != nil {
		return nil, err
	}

	return reply.Balance.Int(), nil
}

func (e *evmClient) EstimateGas(ctx context.Context, msg *evm.CallMsg) (uint64, error) {
	call, err := callMsgToProto(msg)
	if err != nil {
		return 0, err
	}

	reply, err := e.cl.EstimateGas(ctx, &evmpb2.EstimateGasRequest{
		Msg: call,
	})
	if err != nil {
		return 0, err
	}

	return reply.Gas, nil
}

func (e *evmClient) TransactionByHash(ctx context.Context, hash string) (*evm.Transaction, error) {
	reply, err := e.cl.GetTransactionByHash(ctx, &evmpb2.GetTransactionByHashRequest{
		Hash: &evmpb2.Hash{Hash: hash},
	})
	if err != nil {
		return nil, err
	}

	return protoToTransaction(reply.Transaction)
}

func (e *evmClient) TransactionReceipt(ctx context.Context, txHash string) (*evm.Receipt, error) {
	reply, err := e.cl.GetTransactionReceipt(ctx, &evmpb2.GetReceiptRequest{Hash: &evmpb2.Hash{Hash: txHash}})
	if err != nil {
		return nil, err
	}

	return protoToReceipt(reply.Receipt)
}

func (e *evmClient) LatestAndFinalizedHead(ctx context.Context) (latest evm.Head, finalized evm.Head, err error) {
	reply, err := e.cl.LatestAndFinalizedHead(ctx, &emptypb.Empty{})
	if err != nil {
		return evm.Head{}, evm.Head{}, err
	}

	return protoToHead(reply.Latest), protoToHead(reply.Finalized), err

}
func (e *evmClient) QueryLogsFromCache(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	//TODO BCFR-1328
	return nil, errors.New("unimplemented")
}

func (e *evmClient) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	_, err := e.cl.RegisterLogTracking(ctx, &evmpb2.RegisterLogTrackingRequest{Filter: lPfilterToProto(filter)})
	return err
}

func (e *evmClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.cl.UnregisterLogTracking(ctx, &evmpb2.UnregisterLogTrackingRequest{FilterName: filterName})
	return err
}

func (e *evmClient) GetTransactionStatus(ctx context.Context, transactionID string) (evm.TransactionStatus, error) {
	reply, err := e.cl.GetTransactionStatus(ctx, &evmpb2.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return evm.Unknown, err
	}

	return evm.TransactionStatus(reply.TransactionStatus), nil
}

var _ evmpb2.EVMServer = (*evmServer)(nil)

type evmServer struct {
	evmpb2.UnimplementedEVMServer

	*net.BrokerExt

	impl types.EVMService
}

func newEVMServer(impl types.EVMService, b *net.BrokerExt) *evmServer {
	return &evmServer{impl: impl, BrokerExt: b.WithName("EVMServer")}
}

func (e *evmServer) GetTransactionFee(ctx context.Context, request *evmpb2.GetTransactionFeeRequest) (*evmpb2.GetTransactionFeeReply, error) {
	reply, err := e.impl.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmpb2.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(reply.TransactionFee),
	}, nil
}

func (e *evmServer) CallContract(ctx context.Context, req *evmpb2.CallContractRequest) (*evmpb2.CallContractReply, error) {
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

	return &evmpb2.CallContractReply{
		Data: &evmpb2.ABIPayload{Abi: data},
	}, nil
}
func (e *evmServer) GetLogs(ctx context.Context, req *evmpb2.GetLogsRequest) (*evmpb2.GetLogsReply, error) {
	f, err := protoToEvmFilter(req.FilterQuery)
	if err != nil {
		return nil, err
	}
	logs, err := e.impl.GetLogs(ctx, f)
	if err != nil {
		return nil, err
	}

	return &evmpb2.GetLogsReply{
		Logs: logsToProto(logs),
	}, nil
}
func (e *evmServer) BalanceAt(ctx context.Context, req *evmpb2.BalanceAtRequest) (*evmpb2.BalanceAtReply, error) {
	balance, err := e.impl.BalanceAt(ctx, req.Account.Address, req.BlockNumber.Int())
	if err != nil {
		return nil, err
	}

	return &evmpb2.BalanceAtReply{
		Balance: pb.NewBigIntFromInt(balance),
	}, nil
}

func (e *evmServer) EstimateGas(ctx context.Context, req *evmpb2.EstimateGasRequest) (*evmpb2.EstimateGasReply, error) {
	call, err := protoToCallMsg(req.Msg)
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, call)
	if err != nil {
		return nil, err
	}

	return &evmpb2.EstimateGasReply{
		Gas: gas,
	}, nil
}

func (e *evmServer) GetTransactionByHash(ctx context.Context, req *evmpb2.GetTransactionByHashRequest) (*evmpb2.GetTransactionByHashReply, error) {
	tx, err := e.impl.TransactionByHash(ctx, req.Hash.Hash)
	if err != nil {
		return nil, err
	}

	pbtx, err := transactionToProto(tx)
	if err != nil {
		return nil, err
	}
	return &evmpb2.GetTransactionByHashReply{
		Transaction: pbtx,
	}, nil
}

func (e *evmServer) GetTransactionReceipt(ctx context.Context, req *evmpb2.GetReceiptRequest) (*evmpb2.GetReceiptReply, error) {
	rec, err := e.impl.TransactionReceipt(ctx, req.Hash.Hash)
	if err != nil {
		return nil, err
	}

	pbrec, err := receiptToProto(rec)
	if err != nil {
		return nil, err
	}
	return &evmpb2.GetReceiptReply{
		Receipt: pbrec,
	}, nil
}

func (e *evmServer) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*evmpb2.LatestAndFinalizedHeadReply, error) {
	latest, finalized, err := e.impl.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmpb2.LatestAndFinalizedHeadReply{
		Latest:    headToProto(latest),
		Finalized: headToProto(finalized),
	}, nil
}

func (e *evmServer) QueryLogsFromCache(context.Context, *evmpb2.QueryLogsFromCacheRequest) (*evmpb2.QueryLogsFromCacheReply, error) {
	return nil, errors.New("method QueryLogsFromCache not implemented")
}

func (e *evmServer) RegisterLogTracking(ctx context.Context, req *evmpb2.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	f, err := protoToLpFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	err = e.impl.RegisterLogTracking(ctx, f)
	return nil, err
}

func (e *evmServer) UnregisterLogTracking(ctx context.Context, req *evmpb2.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	err := e.impl.UnregisterLogTracking(ctx, req.FilterName)

	return nil, err
}

func (e *evmServer) GetTransactionStatus(ctx context.Context, req *evmpb2.GetTransactionStatusRequest) (*evmpb2.GetTransactionStatusReply, error) {
	status, err := e.impl.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmpb2.GetTransactionStatusReply{TransactionStatus: evmpb2.TransactionStatus(status)}, nil
}

var errEmptyMsg = errors.New("call msg can't be empty")

func protoToHead(h *evmpb2.Head) evm.Head {
	return evm.Head{
		Timestamp:  h.Timestamp,
		Hash:       h.Hash.Hash,
		ParentHash: h.ParentHash.Hash,
		Number:     h.BlockNumber.Int(),
	}
}

func headToProto(h evm.Head) *evmpb2.Head {
	return &evmpb2.Head{
		Timestamp:   h.Timestamp,
		BlockNumber: pb.NewBigIntFromInt(h.Number),
		Hash:        toProtoHash(h.Hash),
		ParentHash:  toProtoHash(h.ParentHash),
	}
}

var errEmptyReceipt = errors.New("receipt is empty")

func receiptToProto(r *evm.Receipt) (*evmpb2.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}

	return &evmpb2.Receipt{
		Status:            r.Status,
		Logs:              logsToProto(r.Logs),
		TxHash:            toProtoHash(r.TxHash),
		ContractAddress:   toProtoAddress(r.ContractAddress),
		GasUsed:           r.GasUsed,
		BlockHash:         toProtoHash(r.BlockHash),
		BlockNumber:       pb.NewBigIntFromInt(r.BlockNumber),
		TxIndex:           r.TransactionIndex,
		EffectiveGasPrice: pb.NewBigIntFromInt(r.EffectiveGasPrice),
	}, nil
}

func protoToReceipt(r *evmpb2.Receipt) (*evm.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}
	return &evm.Receipt{
		Status:            r.Status,
		Logs:              protoToLogs(r.Logs),
		TxHash:            r.TxHash.GetHash(),
		ContractAddress:   r.ContractAddress.GetAddress(),
		GasUsed:           r.GasUsed,
		BlockHash:         r.BlockHash.GetHash(),
		BlockNumber:       r.BlockNumber.Int(),
		TransactionIndex:  r.TxIndex,
		EffectiveGasPrice: r.EffectiveGasPrice.Int(),
	}, nil
}

var errEmptyTx = errors.New("transaction is empty")

func transactionToProto(tx *evm.Transaction) (*evmpb2.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evmpb2.Transaction{
		To:       toProtoAddress(tx.To),
		Data:     toProtoABI(tx.Data),
		Hash:     toProtoHash(tx.Hash),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: pb.NewBigIntFromInt(tx.GasPrice),
		Value:    pb.NewBigIntFromInt(tx.Value),
	}, nil
}

func protoToTransaction(tx *evmpb2.Transaction) (*evm.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evm.Transaction{
		To:       tx.To.GetAddress(),
		Data:     tx.Data.GetAbi(),
		Hash:     tx.Hash.GetHash(),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: tx.GasPrice.Int(),
		Value:    tx.Value.Int(),
	}, nil
}

func callMsgToProto(m *evm.CallMsg) (*evmpb2.CallMsg, error) {
	if m == nil {
		return nil, errEmptyMsg
	}

	return &evmpb2.CallMsg{
		From: toProtoAddress(m.From),
		To:   toProtoAddress(m.To),
		Data: toProtoABI(m.Data),
	}, nil
}

func protoToCallMsg(p *evmpb2.CallMsg) (*evm.CallMsg, error) {
	if p == nil {
		return nil, errEmptyMsg
	}

	return &evm.CallMsg{
		From: p.From.Address,
		Data: p.Data.Abi,
		To:   p.To.Address,
	}, nil
}

func lPfilterToProto(f evm.LPFilterQuery) *evmpb2.LPFilter {
	return &evmpb2.LPFilter{
		Name:          f.Name,
		RetentionTime: int64(f.Retention),
		Addresses:     toProtoAddresses(f.Addresses),
		EventSigs:     toProtoHashes(f.EventSigs),
		Topic2:        toProtoHashes(f.Topic2),
		Topic3:        toProtoHashes(f.Topic3),
		Topic4:        toProtoHashes(f.Topic4),
		MaxLogsKept:   f.MaxLogsKept,
		LogsPerBlock:  f.LogsPerBlock,
	}
}

func protoToLpFilter(f *evmpb2.LPFilter) (evm.LPFilterQuery, error) {
	if f == nil {
		return evm.LPFilterQuery{}, errEmptyFilter
	}

	return evm.LPFilterQuery{
		Name:         f.Name,
		Retention:    time.Duration(f.RetentionTime),
		Addresses:    protoToAddreses(f.Addresses),
		EventSigs:    protoToHashes(f.EventSigs),
		Topic2:       protoToHashes(f.Topic2),
		Topic3:       protoToHashes(f.Topic3),
		Topic4:       protoToHashes(f.Topic4),
		MaxLogsKept:  f.MaxLogsKept,
		LogsPerBlock: f.LogsPerBlock,
	}, nil
}

var errEmptyFilter = errors.New("filter cant be empty")

func protoToEvmFilter(f *evmpb2.FilterQuery) (evm.FilterQuery, error) {
	if f == nil {
		return evm.FilterQuery{}, errEmptyFilter
	}
	return evm.FilterQuery{
		BlockHash: f.BlockHash.GetHash(),
		FromBlock: f.FromBlock.Int(),
		ToBlock:   f.ToBlock.Int(),
		Addresses: protoToAddreses(f.Addresses),
		Topics:    protoToTopics(f.Topics),
	}, nil
}

func evmFilterToProto(f evm.FilterQuery) *evmpb2.FilterQuery {
	return &evmpb2.FilterQuery{
		BlockHash: toProtoHash(f.BlockHash),
		FromBlock: pb.NewBigIntFromInt(f.FromBlock),
		ToBlock:   pb.NewBigIntFromInt(f.ToBlock),
		Addresses: toProtoAddresses(f.Addresses),
		Topics:    toProtoTopics(f.Topics),
	}
}

func protoToLogs(logs []*evmpb2.Log) []*evm.Log {
	ret := make([]*evm.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, protoToLog(l))
	}

	return ret
}

func logsToProto(logs []*evm.Log) []*evmpb2.Log {
	ret := make([]*evmpb2.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, logToProto(l))
	}

	return ret
}

func logToProto(l *evm.Log) *evmpb2.Log {
	return &evmpb2.Log{
		Index:       l.LogIndex,
		BlockHash:   toProtoHash(l.BlockHash),
		BlockNumber: pb.NewBigIntFromInt(l.BlockNumber),
		Topics:      toProtoHashes(l.Topics),
		EventSig:    toProtoHash(l.EventSig),
		Address:     toProtoAddress(l.Address),
		TxHash:      toProtoHash(l.TxHash),
		Data:        toProtoABI(l.Data),
		Removed:     l.Removed,
	}
}

func protoToLog(l *evmpb2.Log) *evm.Log {
	return &evm.Log{
		LogIndex:    l.Index,
		BlockHash:   l.BlockHash.GetHash(),
		BlockNumber: l.BlockNumber.Int(),
		Topics:      protoToHashes(l.Topics),
		EventSig:    l.EventSig.GetHash(),
		Address:     l.Address.GetAddress(),
		TxHash:      l.TxHash.GetHash(),
		Data:        l.Data.GetAbi(),
		Removed:     l.Removed,
	}
}

func toProtoHash(s string) *evmpb2.Hash {
	if s == "" {
		return nil
	}
	return &evmpb2.Hash{Hash: s}
}

func toProtoTopics(ss [][]string) []*evmpb2.Topics {
	ret := make([]*evmpb2.Topics, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, &evmpb2.Topics{Topic: toProtoHashes(s)})
	}

	return ret
}

func toProtoHashes(ss []string) []*evmpb2.Hash {
	ret := make([]*evmpb2.Hash, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, toProtoHash(s))
	}
	return ret
}

func protoToTopics(topics []*evmpb2.Topics) [][]string {
	ret := make([][]string, 0, len(topics))
	for _, topic := range topics {
		ret = append(ret, protoToHashes(topic.Topic))
	}

	return ret
}

func protoToHashes(hs []*evmpb2.Hash) []string {
	ret := make([]string, 0, len(hs))
	for _, h := range hs {
		ret = append(ret, h.Hash)
	}

	return ret
}

func toProtoAddress(s string) *evmpb2.Address {
	if s == "" {
		return nil
	}
	return &evmpb2.Address{Address: s}
}

func toProtoAddresses(ss []string) []*evmpb2.Address {
	ret := make([]*evmpb2.Address, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, toProtoAddress(s))
	}
	return ret
}

func protoToAddreses(s []*evmpb2.Address) []string {
	ret := make([]string, 0, len(s))
	for _, a := range s {
		ret = append(ret, a.Address)
	}

	return ret
}

func toProtoABI(data []byte) *evmpb2.ABIPayload {
	return &evmpb2.ABIPayload{Abi: data}
}
