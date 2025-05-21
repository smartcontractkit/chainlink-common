package relayer

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
)

var _ types.EVMService = (*evmClient)(nil)

type evmClient struct {
	cl evmpb.EVMClient
}

func (e *evmClient) GetTransactionFee(ctx context.Context, transactionID string) (*evm.TransactionFee, error) {
	reply, err := e.cl.GetTransactionFee(ctx, &evmpb.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, err
	}

	return &evm.TransactionFee{
		TransactionFee: reply.TransationFee.Int(),
	}, nil
}

func (e *evmClient) CallContract(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int) ([]byte, error) {
	call, err := callMsgToProto(msg)
	if err != nil {
		return nil, err
	}

	reply, err := e.cl.CallContract(ctx, &evmpb.CallContractRequest{
		Call:        call,
		BlockNumber: pb.NewBigIntFromInt(blockNumber),
	})

	if err != nil {
		return nil, err
	}

	return reply.Data.GetAbi(), nil
}

func (e *evmClient) FilterLogs(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error) {
	reply, err := e.cl.FilterLogs(ctx, &evmpb.FilterLogsRequest{
		FilterQuery: evmFilterToProto(filterQuery),
	})

	if err != nil {
		return nil, err
	}

	return protoToLogs(reply.Logs), nil
}

func (e *evmClient) BalanceAt(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error) {
	reply, err := e.cl.BalanceAt(ctx, &evmpb.BalanceAtRequest{
		Account:     &evmpb.Address{Address: account[:]},
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

	reply, err := e.cl.EstimateGas(ctx, &evmpb.EstimateGasRequest{
		Msg: call,
	})
	if err != nil {
		return 0, err
	}

	return reply.Gas, nil
}

func (e *evmClient) TransactionByHash(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
	reply, err := e.cl.GetTransactionByHash(ctx, &evmpb.GetTransactionByHashRequest{
		Hash: &evmpb.Hash{Hash: hash[:]},
	})
	if err != nil {
		return nil, err
	}

	return protoToTransaction(reply.Transaction)
}

func (e *evmClient) TransactionReceipt(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error) {
	reply, err := e.cl.GetTransactionReceipt(ctx, &evmpb.GetReceiptRequest{Hash: &evmpb.Hash{Hash: txHash[:]}})
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
func (e *evmClient) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	query, err := expressionsToProto(filterQuery)
	if err != nil {
		return nil, err
	}

	sort, err := chaincommonpb.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	conf, err := chaincommonpb.ConvertConfidenceToProto(confidenceLevel)
	if err != nil {
		return nil, err
	}

	reply, err := e.cl.QueryTrackedLogs(ctx, &evmpb.QueryTrackedLogsRequest{
		Expression:      query,
		LimitAndSort:    sort,
		ConfidenceLevel: conf,
	})

	if err != nil {
		return nil, err
	}

	return protoToLogs(reply.Logs), nil
}

func (e *evmClient) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	_, err := e.cl.RegisterLogTracking(ctx, &evmpb.RegisterLogTrackingRequest{Filter: lPfilterToProto(filter)})
	return err
}

func (e *evmClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.cl.UnregisterLogTracking(ctx, &evmpb.UnregisterLogTrackingRequest{FilterName: filterName})
	return err
}

func (e *evmClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := e.cl.GetTransactionStatus(ctx, &pb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, err
	}

	return types.TransactionStatus(reply.TransactionStatus), nil
}

var _ evmpb.EVMServer = (*evmServer)(nil)

type evmServer struct {
	evmpb.UnimplementedEVMServer

	*net.BrokerExt

	impl types.EVMService
}

func newEVMServer(impl types.EVMService, b *net.BrokerExt) *evmServer {
	return &evmServer{impl: impl, BrokerExt: b.WithName("EVMServer")}
}

func (e *evmServer) GetTransactionFee(ctx context.Context, request *evmpb.GetTransactionFeeRequest) (*evmpb.GetTransactionFeeReply, error) {
	reply, err := e.impl.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(reply.TransactionFee),
	}, nil
}

func (e *evmServer) CallContract(ctx context.Context, req *evmpb.CallContractRequest) (*evmpb.CallContractReply, error) {
	call, err := protoToCallMsg(req.Call)
	if err != nil {
		return nil, err
	}

	data, err := e.impl.CallContract(ctx, call, req.BlockNumber.Int())
	if err != nil {
		return nil, err
	}

	return &evmpb.CallContractReply{
		Data: &evmpb.ABIPayload{Abi: data},
	}, nil
}
func (e *evmServer) FilterLogs(ctx context.Context, req *evmpb.FilterLogsRequest) (*evmpb.FilterLogsReply, error) {
	f, err := protoToEvmFilter(req.FilterQuery)
	if err != nil {
		return nil, err
	}
	logs, err := e.impl.FilterLogs(ctx, f)
	if err != nil {
		return nil, err
	}

	return &evmpb.FilterLogsReply{
		Logs: logsToProto(logs),
	}, nil
}
func (e *evmServer) BalanceAt(ctx context.Context, req *evmpb.BalanceAtRequest) (*evmpb.BalanceAtReply, error) {
	balance, err := e.impl.BalanceAt(ctx, protoToAddress(req.Account), req.BlockNumber.Int())
	if err != nil {
		return nil, err
	}

	return &evmpb.BalanceAtReply{
		Balance: pb.NewBigIntFromInt(balance),
	}, nil
}

func (e *evmServer) EstimateGas(ctx context.Context, req *evmpb.EstimateGasRequest) (*evmpb.EstimateGasReply, error) {
	call, err := protoToCallMsg(req.Msg)
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, call)
	if err != nil {
		return nil, err
	}

	return &evmpb.EstimateGasReply{
		Gas: gas,
	}, nil
}

func (e *evmServer) GetTransactionByHash(ctx context.Context, req *evmpb.GetTransactionByHashRequest) (*evmpb.GetTransactionByHashReply, error) {
	tx, err := e.impl.TransactionByHash(ctx, protoToHash(req.GetHash()))
	if err != nil {
		return nil, err
	}

	pbtx, err := transactionToProto(tx)
	if err != nil {
		return nil, err
	}
	return &evmpb.GetTransactionByHashReply{
		Transaction: pbtx,
	}, nil
}

func (e *evmServer) GetTransactionReceipt(ctx context.Context, req *evmpb.GetReceiptRequest) (*evmpb.GetReceiptReply, error) {
	rec, err := e.impl.TransactionReceipt(ctx, protoToHash(req.GetHash()))
	if err != nil {
		return nil, err
	}

	pbrec, err := receiptToProto(rec)
	if err != nil {
		return nil, err
	}
	return &evmpb.GetReceiptReply{
		Receipt: pbrec,
	}, nil
}

func (e *evmServer) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*evmpb.LatestAndFinalizedHeadReply, error) {
	latest, finalized, err := e.impl.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmpb.LatestAndFinalizedHeadReply{
		Latest:    headToProto(latest),
		Finalized: headToProto(finalized),
	}, nil
}

func (e *evmServer) QueryTrackedLogs(ctx context.Context, req *evmpb.QueryTrackedLogsRequest) (*evmpb.QueryTrackedLogsReply, error) {
	exprs, err := protoToExpressions(req.Expression)
	if err != nil {
		return nil, err
	}

	conf, err := chaincommonpb.ConfidenceFromProto(req.ConfidenceLevel)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := chaincommonpb.ConvertLimitAndSortFromProto(req.LimitAndSort)

	logs, err := e.impl.QueryTrackedLogs(ctx, exprs, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmpb.QueryTrackedLogsReply{
		Logs: logsToProto(logs),
	}, nil
}

func (e *evmServer) RegisterLogTracking(ctx context.Context, req *evmpb.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	f, err := protoToLpFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	err = e.impl.RegisterLogTracking(ctx, f)
	return nil, err
}

func (e *evmServer) UnregisterLogTracking(ctx context.Context, req *evmpb.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	err := e.impl.UnregisterLogTracking(ctx, req.FilterName)

	return nil, err
}

func (e *evmServer) GetTransactionStatus(ctx context.Context, req *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusReply, error) {
	status, err := e.impl.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionStatusReply{TransactionStatus: pb.TransactionStatus(status)}, nil
}

var errEmptyMsg = errors.New("call msg can't be empty")

func protoToHead(h *evmpb.Head) evm.Head {
	return evm.Head{
		Timestamp:  h.Timestamp,
		Hash:       protoToHash(h.GetHash()),
		ParentHash: protoToHash(h.GetParentHash()),
		Number:     h.BlockNumber.Int(),
	}
}

func headToProto(h evm.Head) *evmpb.Head {
	return &evmpb.Head{
		Timestamp:   h.Timestamp,
		BlockNumber: pb.NewBigIntFromInt(h.Number),
		Hash:        toProtoHash(h.Hash),
		ParentHash:  toProtoHash(h.ParentHash),
	}
}

var errEmptyReceipt = errors.New("receipt is empty")

func receiptToProto(r *evm.Receipt) (*evmpb.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}

	return &evmpb.Receipt{
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

func protoToReceipt(r *evmpb.Receipt) (*evm.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}
	return &evm.Receipt{
		Status:            r.Status,
		Logs:              protoToLogs(r.Logs),
		TxHash:            protoToHash(r.GetTxHash()),
		ContractAddress:   protoToAddress(r.GetContractAddress()),
		GasUsed:           r.GasUsed,
		BlockHash:         protoToHash(r.GetBlockHash()),
		BlockNumber:       r.BlockNumber.Int(),
		TransactionIndex:  r.TxIndex,
		EffectiveGasPrice: r.EffectiveGasPrice.Int(),
	}, nil
}

var errEmptyTx = errors.New("transaction is empty")

func transactionToProto(tx *evm.Transaction) (*evmpb.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evmpb.Transaction{
		To:       toProtoAddress(tx.To),
		Data:     toProtoABI(tx.Data),
		Hash:     toProtoHash(tx.Hash),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: pb.NewBigIntFromInt(tx.GasPrice),
		Value:    pb.NewBigIntFromInt(tx.Value),
	}, nil
}

func protoToTransaction(tx *evmpb.Transaction) (*evm.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evm.Transaction{
		To:       protoToAddress(tx.GetTo()),
		Data:     tx.Data.GetAbi(),
		Hash:     protoToHash(tx.GetHash()),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: tx.GasPrice.Int(),
		Value:    tx.Value.Int(),
	}, nil
}

func callMsgToProto(m *evm.CallMsg) (*evmpb.CallMsg, error) {
	if m == nil {
		return nil, errEmptyMsg
	}

	return &evmpb.CallMsg{
		From: toProtoAddress(m.From),
		To:   toProtoAddress(m.To),
		Data: toProtoABI(m.Data),
	}, nil
}

func protoToCallMsg(p *evmpb.CallMsg) (*evm.CallMsg, error) {
	if p == nil {
		return nil, errEmptyMsg
	}

	return &evm.CallMsg{
		From: protoToAddress(p.GetFrom()),
		Data: p.GetData().GetAbi(),
		To:   protoToAddress(p.GetTo()),
	}, nil
}

func lPfilterToProto(f evm.LPFilterQuery) *evmpb.LPFilter {
	return &evmpb.LPFilter{
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

func protoToLpFilter(f *evmpb.LPFilter) (evm.LPFilterQuery, error) {
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

func protoToEvmFilter(f *evmpb.FilterQuery) (evm.FilterQuery, error) {
	if f == nil {
		return evm.FilterQuery{}, errEmptyFilter
	}
	return evm.FilterQuery{
		BlockHash: protoToHash(f.GetBlockHash()),
		FromBlock: f.GetFromBlock().Int(),
		ToBlock:   f.GetToBlock().Int(),
		Addresses: protoToAddreses(f.Addresses),
		Topics:    protoToTopics(f.Topics),
	}, nil
}

func evmFilterToProto(f evm.FilterQuery) *evmpb.FilterQuery {
	return &evmpb.FilterQuery{
		BlockHash: toProtoHash(f.BlockHash),
		FromBlock: pb.NewBigIntFromInt(f.FromBlock),
		ToBlock:   pb.NewBigIntFromInt(f.ToBlock),
		Addresses: toProtoAddresses(f.Addresses),
		Topics:    toProtoTopics(f.Topics),
	}
}

func protoToLogs(logs []*evmpb.Log) []*evm.Log {
	ret := make([]*evm.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, protoToLog(l))
	}

	return ret
}

func logsToProto(logs []*evm.Log) []*evmpb.Log {
	ret := make([]*evmpb.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, logToProto(l))
	}

	return ret
}

func logToProto(l *evm.Log) *evmpb.Log {
	return &evmpb.Log{
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

func protoToLog(l *evmpb.Log) *evm.Log {
	return &evm.Log{
		LogIndex:    l.Index,
		BlockHash:   protoToHash(l.GetBlockHash()),
		BlockNumber: l.BlockNumber.Int(),
		Topics:      protoToHashes(l.Topics),
		EventSig:    protoToHash(l.GetEventSig()),
		Address:     protoToAddress(l.GetAddress()),
		TxHash:      protoToHash(l.GetTxHash()),
		Data:        l.Data.GetAbi(),
		Removed:     l.Removed,
	}
}

func toProtoHash(h evm.Hash) *evmpb.Hash {
	return &evmpb.Hash{Hash: h[:]}
}

func toProtoTopics(ss [][]evm.Hash) []*evmpb.Topics {
	ret := make([]*evmpb.Topics, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, &evmpb.Topics{Topic: toProtoHashes(s)})
	}

	return ret
}

func toProtoHashes(ss []evm.Hash) []*evmpb.Hash {
	ret := make([]*evmpb.Hash, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, toProtoHash(s))
	}
	return ret
}

func protoToTopics(topics []*evmpb.Topics) [][]evm.Hash {
	ret := make([][]evm.Hash, 0, len(topics))
	for _, topic := range topics {
		ret = append(ret, protoToHashes(topic.Topic))
	}

	return ret
}

func protoToHashes(hs []*evmpb.Hash) []evm.Hash {
	ret := make([]evm.Hash, 0, len(hs))
	for _, h := range hs {
		ret = append(ret, protoToHash(h))
	}

	return ret
}

func toProtoAddress(a evm.Address) *evmpb.Address {
	return &evmpb.Address{Address: a[:]}
}

func toProtoAddresses(ss []evm.Address) []*evmpb.Address {
	ret := make([]*evmpb.Address, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, toProtoAddress(s))
	}
	return ret
}

func protoToAddreses(s []*evmpb.Address) []evm.Address {
	ret := make([]evm.Address, 0, len(s))
	for _, a := range s {
		ret = append(ret, protoToAddress(a))
	}

	return ret
}

func toProtoABI(data []byte) *evmpb.ABIPayload {
	return &evmpb.ABIPayload{Abi: data}
}

func protoToHash(hp *evmpb.Hash) evm.Hash {
	var h evm.Hash
	copy(h[:], hp.GetHash())
	return h
}

func protoToAddress(ap *evmpb.Address) evm.Address {
	var a evm.Address
	copy(a[:], ap.GetAddress())
	return a
}
func hashedValueComparersToProto(cs []evmprimitives.HashedValueComparator) []*evmpb.HashValueComparator {
	ret := make([]*evmpb.HashValueComparator, 0, len(cs))
	for _, c := range cs {
		ret = append(ret, &evmpb.HashValueComparator{
			Operator: chaincommonpb.ComparisonOperator(c.Operator),
			Values:   toProtoHashes(c.Values),
		})
	}

	return ret
}

func protoToHashedValueComparers(hvc []*evmpb.HashValueComparator) []evmprimitives.HashedValueComparator {
	ret := make([]evmprimitives.HashedValueComparator, 0, len(hvc))
	for _, c := range hvc {
		ret = append(ret, evmprimitives.HashedValueComparator{
			Values:   protoToHashes(c.Values),
			Operator: primitives.ComparisonOperator(c.Operator),
		})
	}

	return ret
}

func expressionsToProto(expressions []query.Expression) ([]*evmpb.Expression, error) {
	query := make([]*evmpb.Expression, 0, len(expressions))
	for idx, expr := range expressions {
		exprpb, err := expressionToProto(expr)
		if err != nil {
			return nil, fmt.Errorf("err to convert expr idx %d err: %v", idx, err)
		}
		query = append(query, exprpb)
	}

	return query, nil
}

func expressionToProto(expression query.Expression) (*evmpb.Expression, error) {
	pbExpression := &evmpb.Expression{}
	if expression.IsPrimitive() {
		p := &chaincommonpb.Primitive{}
		ep := &evmpb.Primitive{}
		switch primitive := expression.Primitive.(type) {
		case *primitives.Comparator:
			return nil, errors.New("comparator primitive is not supported for EVMService")
		case *primitives.Block:
			p.Primitive = &chaincommonpb.Primitive_Block{
				Block: &chaincommonpb.Block{
					BlockNumber: primitive.Block,
					Operator:    chaincommonpb.ComparisonOperator(primitive.Operator),
				}}

			putGeneralPrimitive(pbExpression, p)
		case *primitives.Confidence:
			pbConfidence, err := chaincommonpb.ConvertConfidenceToProto(primitive.ConfidenceLevel)
			if err != nil {
				return nil, err
			}

			p.Primitive = &chaincommonpb.Primitive_Confidence{
				Confidence: pbConfidence,
			}

			putGeneralPrimitive(pbExpression, p)
		case *primitives.Timestamp:
			p.Primitive = &chaincommonpb.Primitive_Timestamp{
				Timestamp: &chaincommonpb.Timestamp{
					Timestamp: primitive.Timestamp,
					Operator:  chaincommonpb.ComparisonOperator(primitive.Operator),
				}}

			putGeneralPrimitive(pbExpression, p)
		case *primitives.TxHash:
			p.Primitive = &chaincommonpb.Primitive_TxHash{
				TxHash: &chaincommonpb.TxHash{
					TxHash: primitive.TxHash,
				}}

			putGeneralPrimitive(pbExpression, p)
		case *evmprimitives.Address:
			ep.Primitive = &evmpb.Primitive_ContractAddress{ContractAddress: &evmpb.ContractAddress{
				Address: &evmpb.Address{Address: primitive.Address[:]},
			}}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByTopic:
			ep.Primitive = &evmpb.Primitive_EventByTopic{
				EventByTopic: &evmpb.EventByTopic{
					Topic:                primitive.Topic,
					HashedValueComparers: hashedValueComparersToProto(primitive.HashedValueComprarers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByWord:
			ep.Primitive = &evmpb.Primitive_EventByWord{
				EventByWord: &evmpb.EventByWord{
					WordIndex:            uint32(primitive.WordIndex),
					HashedValueComparers: hashedValueComparersToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventSig:
			ep.Primitive = &evmpb.Primitive_EventSig{
				EventSig: &evmpb.EventSig{
					EventSig: &evmpb.Hash{Hash: primitive.EventSig[:]},
				},
			}

			putEVMPrimitive(pbExpression, ep)
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &evmpb.Expression_BooleanExpression{BooleanExpression: &evmpb.BooleanExpression{}}
	var expressions []*evmpb.Expression
	for _, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := expressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pbExpr)
	}
	pbExpression.Evaluator = &evmpb.Expression_BooleanExpression{
		BooleanExpression: &evmpb.BooleanExpression{
			BooleanOperator: chaincommonpb.BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func protoToExpressions(expressions []*evmpb.Expression) ([]query.Expression, error) {
	exprs := make([]query.Expression, 0, len(expressions))
	for idx, exprpb := range expressions {
		expr, err := protoToExpression(exprpb)
		if err != nil {
			return nil, fmt.Errorf("err to convert expr idx %d err: %v", idx, err)
		}

		exprs = append(exprs, expr)
	}

	return exprs, nil
}

func protoToExpression(pbExpression *evmpb.Expression) (query.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *evmpb.Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			convertedExpression, err := protoToExpression(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}
		if pbEvaluatedExpr.BooleanExpression.BooleanOperator == chaincommonpb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *evmpb.Expression_Primitive:
		switch primitive := pbEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *evmpb.Primitive_GeneralPrimitive:
			return protoToGeneralExpr(primitive.GeneralPrimitive)
		default:
			return protoToEVMExpr(pbEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", pbEvaluatedExpr)
	}
}

func protoToGeneralExpr(pbEvaluatedExpr *chaincommonpb.Primitive) (query.Expression, error) {
	switch primitive := pbEvaluatedExpr.GetPrimitive().(type) {
	case *chaincommonpb.Primitive_Comparator:
		return query.Expression{}, errors.New("comparator primitive is not supported for EVMService")
	case *chaincommonpb.Primitive_Confidence:
		confidence, err := chaincommonpb.ConfidenceFromProto(primitive.Confidence)
		return query.Confidence(confidence), err
	case *chaincommonpb.Primitive_Block:
		return query.Block(primitive.Block.BlockNumber, primitives.ComparisonOperator(primitive.Block.Operator)), nil
	case *chaincommonpb.Primitive_TxHash:
		return query.TxHash(primitive.TxHash.TxHash), nil
	case *chaincommonpb.Primitive_Timestamp:
		return query.Timestamp(primitive.Timestamp.Timestamp, primitives.ComparisonOperator(primitive.Timestamp.Operator)), nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
	}
}

func protoToEVMExpr(pbEvaluatedExpr *evmpb.Primitive) (query.Expression, error) {
	switch primitive := pbEvaluatedExpr.GetPrimitive().(type) {
	case *evmpb.Primitive_ContractAddress:
		address := protoToAddress(primitive.ContractAddress.GetAddress())
		return evmprimitives.NewAddressFilter(address), nil
	case *evmpb.Primitive_EventSig:
		hash := protoToHash(primitive.EventSig.GetEventSig())
		return evmprimitives.NewEventSigFilter(hash), nil
	case *evmpb.Primitive_EventByTopic:
		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(),
				protoToHashedValueComparers(primitive.EventByTopic.GetHashedValueComparers())),
			nil
	case *evmpb.Primitive_EventByWord:
		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()),
				protoToHashedValueComparers(primitive.EventByWord.GetHashedValueComparers())),
			nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *evmpb.Expression, p *chaincommonpb.Primitive) {
	exp.Evaluator = &evmpb.Expression_Primitive{Primitive: &evmpb.Primitive{Primitive: &evmpb.Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *evmpb.Expression, p *evmpb.Primitive) {
	exp.Evaluator = &evmpb.Expression_Primitive{Primitive: &evmpb.Primitive{Primitive: p.Primitive}}
}
