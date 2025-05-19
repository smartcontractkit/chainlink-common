package relayer

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"

	"google.golang.org/protobuf/types/known/emptypb"

	evmcap "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type EVMClient struct {
	grpcClient evmcap.EVMClient
}

func NewEVMCClient(grpcClient evmcap.EVMClient) *EVMClient {
	return &EVMClient{
		grpcClient: grpcClient,
	}
}

var _ types.EVMService = (*EVMClient)(nil)

func (e *EVMClient) GetTransactionFee(ctx context.Context, transactionID string) (*evm.TransactionFee, error) {
	reply, err := e.grpcClient.GetTransactionFee(ctx, &evmcap.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, err
	}

	return &evm.TransactionFee{
		TransactionFee: valuespb.NewIntFromBigInt(reply.TransationFee),
	}, nil
}

func (e *EVMClient) CallContract(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int) ([]byte, error) {
	call, err := CallMsgToProto(msg)
	if err != nil {
		return nil, err
	}

	reply, err := e.grpcClient.CallContract(ctx, &evmcap.CallContractRequest{
		Call:        call,
		BlockNumber: valuespb.NewBigIntFromInt(blockNumber),
	})

	if err != nil {
		return nil, err
	}

	return reply.Data.GetAbi(), nil
}

func (e *EVMClient) FilterLogs(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error) {
	reply, err := e.grpcClient.FilterLogs(ctx, &evmcap.FilterLogsRequest{
		FilterQuery: evmFilterToProto(filterQuery),
	})

	if err != nil {
		return nil, err
	}

	return protoToLogs(reply.Logs), nil
}

func (e *EVMClient) BalanceAt(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error) {
	reply, err := e.grpcClient.BalanceAt(ctx, &evmcap.BalanceAtRequest{
		Account:     &evmcap.Address{Address: account[:]},
		BlockNumber: valuespb.NewBigIntFromInt(blockNumber),
	})

	if err != nil {
		return nil, err
	}

	return valuespb.NewIntFromBigInt(reply.Balance), nil
}

func (e *EVMClient) EstimateGas(ctx context.Context, msg *evm.CallMsg) (uint64, error) {
	call, err := CallMsgToProto(msg)
	if err != nil {
		return 0, err
	}

	reply, err := e.grpcClient.EstimateGas(ctx, &evmcap.EstimateGasRequest{
		Msg: call,
	})
	if err != nil {
		return 0, err
	}

	return reply.Gas, nil
}

func (e *EVMClient) TransactionByHash(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
	reply, err := e.grpcClient.TransactionByHash(ctx, &evmcap.TransactionByHashRequest{
		Hash: &evmcap.Hash{Hash: hash[:]},
	})
	if err != nil {
		return nil, err
	}

	return ProtoToTransaction(reply.Transaction)
}

func (e *EVMClient) TransactionReceipt(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error) {
	reply, err := e.grpcClient.TransactionReceipt(ctx, &evmcap.TransactionReceiptRequest{Hash: &evmcap.Hash{Hash: txHash[:]}})
	if err != nil {
		return nil, err
	}

	return protoToReceipt(reply.Receipt)
}

func (e *EVMClient) LatestAndFinalizedHead(ctx context.Context) (latest evm.Head, finalized evm.Head, err error) {
	reply, err := e.grpcClient.LatestAndFinalizedHead(ctx, &emptypb.Empty{})
	if err != nil {
		return evm.Head{}, evm.Head{}, err
	}

	return protoToHead(reply.Latest), protoToHead(reply.Finalized), err

}
func (e *EVMClient) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	q, err := expressionsToProto(filterQuery)
	if err != nil {
		return nil, err
	}

	sort, err := contractreader.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, err
	}

	conf, err := contractreader.ConfidenceToProto(confidenceLevel)
	if err != nil {
		return nil, err
	}

	reply, err := e.grpcClient.QueryTrackedLogs(ctx, &evmcap.QueryTrackedLogsRequest{
		Expression:      q,
		LimitAndSort:    sort,
		ConfidenceLevel: conf,
	})

	if err != nil {
		return nil, err
	}

	return protoToLogs(reply.Logs), nil
}

func (e *EVMClient) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	_, err := e.grpcClient.RegisterLogTracking(ctx, &evmcap.RegisterLogTrackingRequest{Filter: lPfilterToProto(filter)})
	return err
}

func (e *EVMClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.grpcClient.UnregisterLogTracking(ctx, &evmcap.UnregisterLogTrackingRequest{FilterName: filterName})
	return err
}

func (e *EVMClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := e.grpcClient.GetTransactionStatus(ctx, &pb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, err
	}

	return types.TransactionStatus(reply.TransactionStatus), nil
}

type EvmServer struct {
	evmcap.UnimplementedEVMServer

	*net.BrokerExt

	impl types.EVMService
}

var _ evmcap.EVMServer = (*EvmServer)(nil)

func NewEVMServer(impl types.EVMService, b *net.BrokerExt) *EvmServer {
	return &EvmServer{impl: impl, BrokerExt: b.WithName("EVMServer")}
}

func (e *EvmServer) GetTransactionFee(ctx context.Context, request *evmcap.GetTransactionFeeRequest) (*evmcap.GetTransactionFeeReply, error) {
	reply, err := e.impl.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmcap.GetTransactionFeeReply{
		TransationFee: valuespb.NewBigIntFromInt(reply.TransactionFee),
	}, nil
}

func (e *EvmServer) CallContract(ctx context.Context, req *evmcap.CallContractRequest) (*evmcap.CallContractReply, error) {
	call, err := ProtoToCallMsg(req.Call)
	if err != nil {
		return nil, err
	}

	data, err := e.impl.CallContract(ctx, call, valuespb.NewIntFromBigInt(req.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmcap.CallContractReply{
		Data: &evmcap.ABIPayload{Abi: data},
	}, nil
}
func (e *EvmServer) FilterLogs(ctx context.Context, req *evmcap.FilterLogsRequest) (*evmcap.FilterLogsReply, error) {
	f, err := ProtoToEvmFilter(req.FilterQuery)
	if err != nil {
		return nil, err
	}
	logs, err := e.impl.FilterLogs(ctx, f)
	if err != nil {
		return nil, err
	}

	return &evmcap.FilterLogsReply{
		Logs: LogsToProto(logs),
	}, nil
}
func (e *EvmServer) BalanceAt(ctx context.Context, req *evmcap.BalanceAtRequest) (*evmcap.BalanceAtReply, error) {
	balance, err := e.impl.BalanceAt(ctx, ProtoToAddress(req.Account), valuespb.NewIntFromBigInt(req.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmcap.BalanceAtReply{
		Balance: valuespb.NewBigIntFromInt(balance),
	}, nil
}

func (e *EvmServer) EstimateGas(ctx context.Context, req *evmcap.EstimateGasRequest) (*evmcap.EstimateGasReply, error) {
	call, err := ProtoToCallMsg(req.Msg)
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, call)
	if err != nil {
		return nil, err
	}

	return &evmcap.EstimateGasReply{
		Gas: gas,
	}, nil
}

func (e *EvmServer) GetTransactionByHash(ctx context.Context, req *evmcap.TransactionByHashRequest) (*evmcap.TransactionByHashReply, error) {
	tx, err := e.impl.TransactionByHash(ctx, ProtoToHash(req.GetHash()))
	if err != nil {
		return nil, err
	}

	pbtx, err := TransactionToProto(tx)
	if err != nil {
		return nil, err
	}
	return &evmcap.TransactionByHashReply{
		Transaction: pbtx,
	}, nil
}

func (e *EvmServer) GetTransactionReceipt(ctx context.Context, req *evmcap.TransactionReceiptRequest) (*evmcap.TransactionReceiptReply, error) {
	rec, err := e.impl.TransactionReceipt(ctx, ProtoToHash(req.GetHash()))
	if err != nil {
		return nil, err
	}

	pbrec, err := ReceiptToProto(rec)
	if err != nil {
		return nil, err
	}
	return &evmcap.TransactionReceiptReply{
		Receipt: pbrec,
	}, nil
}

func (e *EvmServer) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*evmcap.LatestAndFinalizedHeadReply, error) {
	latest, finalized, err := e.impl.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmcap.LatestAndFinalizedHeadReply{
		Latest:    HeadToProto(latest),
		Finalized: HeadToProto(finalized),
	}, nil
}

func (e *EvmServer) QueryTrackedLogs(ctx context.Context, req *evmcap.QueryTrackedLogsRequest) (*evmcap.QueryTrackedLogsReply, error) {
	exprs, err := ProtoToExpressions(req.Expression)
	if err != nil {
		return nil, err
	}

	limitAndSort, err := evmcap.ConvertLimitAndSortFromProto(req.LimitAndSort)
	if err != nil {
		return nil, err
	}

	conf, err := evmcap.ConfidenceFromProto(req.ConfidenceLevel)
	if err != nil {
		return nil, err
	}

	logs, err := e.impl.QueryTrackedLogs(ctx, exprs, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmcap.QueryTrackedLogsReply{
		Logs: LogsToProto(logs),
	}, nil
}

func (e *EvmServer) RegisterLogTracking(ctx context.Context, req *evmcap.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	f, err := ProtoToLpFilter(req.Filter)
	if err != nil {
		return nil, err
	}
	return nil, e.impl.RegisterLogTracking(ctx, f)
}

func (e *EvmServer) UnregisterLogTracking(ctx context.Context, req *evmcap.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	return nil, e.impl.UnregisterLogTracking(ctx, req.FilterName)
}

func (e *EvmServer) GetTransactionStatus(ctx context.Context, req *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusReply, error) {
	s, err := e.impl.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionStatusReply{TransactionStatus: pb.TransactionStatus(s)}, nil
}

var errEmptyMsg = errors.New("call msg can't be empty")

func protoToHead(h *evmcap.Head) evm.Head {
	return evm.Head{
		Timestamp:  h.Timestamp,
		Hash:       ProtoToHash(h.GetHash()),
		ParentHash: ProtoToHash(h.GetParentHash()),
		Number:     valuespb.NewIntFromBigInt(h.BlockNumber),
	}
}

func HeadToProto(h evm.Head) *evmcap.Head {
	return &evmcap.Head{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        toProtoHash(h.Hash),
		ParentHash:  toProtoHash(h.ParentHash),
	}
}

var errEmptyReceipt = errors.New("receipt is empty")

func ReceiptToProto(r *evm.Receipt) (*evmcap.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}

	return &evmcap.Receipt{
		Status:            r.Status,
		Logs:              LogsToProto(r.Logs),
		TxHash:            toProtoHash(r.TxHash),
		ContractAddress:   toProtoAddress(r.ContractAddress),
		GasUsed:           r.GasUsed,
		BlockHash:         toProtoHash(r.BlockHash),
		BlockNumber:       valuespb.NewBigIntFromInt(r.BlockNumber),
		TxIndex:           r.TransactionIndex,
		EffectiveGasPrice: valuespb.NewBigIntFromInt(r.EffectiveGasPrice),
	}, nil
}

func protoToReceipt(r *evmcap.Receipt) (*evm.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}
	return &evm.Receipt{
		Status:            r.Status,
		Logs:              protoToLogs(r.Logs),
		TxHash:            ProtoToHash(r.GetTxHash()),
		ContractAddress:   ProtoToAddress(r.GetContractAddress()),
		GasUsed:           r.GasUsed,
		BlockHash:         ProtoToHash(r.GetBlockHash()),
		BlockNumber:       valuespb.NewIntFromBigInt(r.BlockNumber),
		TransactionIndex:  r.TxIndex,
		EffectiveGasPrice: valuespb.NewIntFromBigInt(r.EffectiveGasPrice),
	}, nil
}

var errEmptyTx = errors.New("transaction is empty")

func TransactionToProto(tx *evm.Transaction) (*evmcap.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evmcap.Transaction{
		To:       toProtoAddress(tx.To),
		Data:     toProtoABI(tx.Data),
		Hash:     toProtoHash(tx.Hash),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: valuespb.NewBigIntFromInt(tx.GasPrice),
		Value:    valuespb.NewBigIntFromInt(tx.Value),
	}, nil
}

func ProtoToTransaction(tx *evmcap.Transaction) (*evm.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evm.Transaction{
		To:       ProtoToAddress(tx.GetTo()),
		Data:     tx.GetData().GetAbi(),
		Hash:     ProtoToHash(tx.GetHash()),
		Nonce:    tx.GetNonce(),
		Gas:      tx.GetGas(),
		GasPrice: valuespb.NewIntFromBigInt(tx.GetGasPrice()),
		Value:    valuespb.NewIntFromBigInt(tx.GetValue()),
	}, nil
}

func CallMsgToProto(m *evm.CallMsg) (*evmcap.CallMsg, error) {
	if m == nil {
		return nil, errEmptyMsg
	}

	return &evmcap.CallMsg{
		From: toProtoAddress(m.From),
		To:   toProtoAddress(m.To),
		Data: toProtoABI(m.Data),
	}, nil
}

func ProtoToCallMsg(p *evmcap.CallMsg) (*evm.CallMsg, error) {
	if p == nil {
		return nil, errEmptyMsg
	}

	return &evm.CallMsg{
		From: ProtoToAddress(p.GetFrom()),
		Data: p.GetData().GetAbi(),
		To:   ProtoToAddress(p.GetTo()),
	}, nil
}

func lPfilterToProto(f evm.LPFilterQuery) *evmcap.LPFilter {
	return &evmcap.LPFilter{
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

func ProtoToLpFilter(f *evmcap.LPFilter) (evm.LPFilterQuery, error) {
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

func ProtoToEvmFilter(f *evmcap.FilterQuery) (evm.FilterQuery, error) {
	if f == nil {
		return evm.FilterQuery{}, errEmptyFilter
	}
	return evm.FilterQuery{
		BlockHash: ProtoToHash(f.GetBlockHash()),
		FromBlock: valuespb.NewIntFromBigInt(f.GetFromBlock()),
		ToBlock:   valuespb.NewIntFromBigInt(f.GetToBlock()),
		Addresses: protoToAddreses(f.Addresses),
		Topics:    protoToTopics(f.Topics),
	}, nil
}

func evmFilterToProto(f evm.FilterQuery) *evmcap.FilterQuery {
	return &evmcap.FilterQuery{
		BlockHash: toProtoHash(f.BlockHash),
		FromBlock: valuespb.NewBigIntFromInt(f.FromBlock),
		ToBlock:   valuespb.NewBigIntFromInt(f.ToBlock),
		Addresses: toProtoAddresses(f.Addresses),
		Topics:    toProtoTopics(f.Topics),
	}
}

func protoToLogs(logs []*evmcap.Log) []*evm.Log {
	ret := make([]*evm.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, protoToLog(l))
	}

	return ret
}

func LogsToProto(logs []*evm.Log) []*evmcap.Log {
	ret := make([]*evmcap.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, logToProto(l))
	}

	return ret
}

func logToProto(l *evm.Log) *evmcap.Log {
	return &evmcap.Log{
		Index:       l.LogIndex,
		BlockHash:   toProtoHash(l.BlockHash),
		BlockNumber: valuespb.NewBigIntFromInt(l.BlockNumber),
		Topics:      toProtoHashes(l.Topics),
		EventSig:    toProtoHash(l.EventSig),
		Address:     toProtoAddress(l.Address),
		TxHash:      toProtoHash(l.TxHash),
		Data:        toProtoABI(l.Data),
		Removed:     l.Removed,
	}
}

func protoToLog(l *evmcap.Log) *evm.Log {
	return &evm.Log{
		LogIndex:    l.GetIndex(),
		BlockHash:   ProtoToHash(l.GetBlockHash()),
		BlockNumber: valuespb.NewIntFromBigInt(l.GetBlockNumber()),
		Topics:      protoToHashes(l.Topics),
		EventSig:    ProtoToHash(l.GetEventSig()),
		Address:     ProtoToAddress(l.GetAddress()),
		TxHash:      ProtoToHash(l.GetTxHash()),
		Data:        l.Data.GetAbi(),
		Removed:     l.GetRemoved(),
	}
}

func toProtoHash(h evm.Hash) *evmcap.Hash {
	return &evmcap.Hash{Hash: h[:]}
}

func toProtoTopics(ss [][]evm.Hash) []*evmcap.Topics {
	ret := make([]*evmcap.Topics, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, &evmcap.Topics{Topic: toProtoHashes(s)})
	}

	return ret
}

func toProtoHashes(ss []evm.Hash) []*evmcap.Hash {
	ret := make([]*evmcap.Hash, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, toProtoHash(s))
	}
	return ret
}

func protoToTopics(topics []*evmcap.Topics) [][]evm.Hash {
	ret := make([][]evm.Hash, 0, len(topics))
	for _, topic := range topics {
		ret = append(ret, protoToHashes(topic.Topic))
	}

	return ret
}

func protoToHashes(hs []*evmcap.Hash) []evm.Hash {
	ret := make([]evm.Hash, 0, len(hs))
	for _, h := range hs {
		ret = append(ret, ProtoToHash(h))
	}

	return ret
}

func toProtoAddress(a evm.Address) *evmcap.Address {
	return &evmcap.Address{Address: a[:]}
}

func toProtoAddresses(ss []evm.Address) []*evmcap.Address {
	ret := make([]*evmcap.Address, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, toProtoAddress(s))
	}
	return ret
}

func protoToAddreses(s []*evmcap.Address) []evm.Address {
	ret := make([]evm.Address, 0, len(s))
	for _, a := range s {
		ret = append(ret, ProtoToAddress(a))
	}

	return ret
}

func toProtoABI(data []byte) *evmcap.ABIPayload {
	return &evmcap.ABIPayload{Abi: data}
}

func ProtoToHash(hp *evmcap.Hash) evm.Hash {
	var h evm.Hash
	copy(h[:], hp.GetHash())
	return h
}

func ProtoToAddress(ap *evmcap.Address) evm.Address {
	var a evm.Address
	copy(a[:], ap.GetAddress())
	return a
}
func hashedValueComparersToProto(cs []evmprimitives.HashedValueComparator) []*evmcap.HashValueComparator {
	ret := make([]*evmcap.HashValueComparator, 0, len(cs))
	for _, c := range cs {
		ret = append(ret, &evmcap.HashValueComparator{
			Operator: pb.ComparisonOperator(c.Operator),
			Values:   toProtoHashes(c.Values),
		})
	}

	return ret
}

func protoToHashedValueComparers(hvc []*evmcap.HashValueComparator) []evmprimitives.HashedValueComparator {
	ret := make([]evmprimitives.HashedValueComparator, 0, len(hvc))
	for _, c := range hvc {
		ret = append(ret, evmprimitives.HashedValueComparator{
			Values:   protoToHashes(c.Values),
			Operator: primitives.ComparisonOperator(c.Operator),
		})
	}

	return ret
}

func expressionsToProto(expressions []query.Expression) ([]*evmcap.Expression, error) {
	q := make([]*evmcap.Expression, 0, len(expressions))
	for idx, expr := range expressions {
		exprpb, err := expressionToProto(expr)
		if err != nil {
			return nil, fmt.Errorf("err to convert expr idx %d err: %v", idx, err)
		}
		q = append(q, exprpb)
	}

	return q, nil
}

func expressionToProto(expression query.Expression) (*evmcap.Expression, error) {
	pbExpression := &evmcap.Expression{}
	if expression.IsPrimitive() {
		p := &pb.Primitive{}
		ep := &evmcap.Primitive{}
		switch primitive := expression.Primitive.(type) {
		case *primitives.Comparator:
			return nil, errors.New("comparator primitive is not supported for EVMService")
		case *primitives.Block:
			p.Primitive = &pb.Primitive_Block{
				Block: &pb.Block{
					BlockNumber: primitive.Block,
					Operator:    pb.ComparisonOperator(primitive.Operator),
				}}

			putGeneralPrimitive(pbExpression, p)
		case *primitives.Confidence:
			pbConfidence, err := contractreader.ConfidenceToProto(primitive.ConfidenceLevel)
			if err != nil {
				return nil, err
			}

			p.Primitive = &pb.Primitive_Confidence{
				Confidence: pbConfidence,
			}

			putGeneralPrimitive(pbExpression, p)
		case *primitives.Timestamp:
			p.Primitive = &pb.Primitive_Timestamp{
				Timestamp: &pb.Timestamp{
					Timestamp: primitive.Timestamp,
					Operator:  pb.ComparisonOperator(primitive.Operator),
				}}

			putGeneralPrimitive(pbExpression, p)
		case *primitives.TxHash:
			p.Primitive = &pb.Primitive_TxHash{
				TxHash: &pb.TxHash{
					TxHash: primitive.TxHash,
				}}

			putGeneralPrimitive(pbExpression, p)
		case *evmprimitives.Address:
			ep.Primitive = &evmcap.Primitive_ContractAddress{ContractAddress: &evmcap.ContractAddress{
				Address: &evmcap.Address{Address: primitive.Address[:]},
			}}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByTopic:
			ep.Primitive = &evmcap.Primitive_EventByTopic{
				EventByTopic: &evmcap.EventByTopic{
					Topic:                primitive.Topic,
					HashedValueComparers: hashedValueComparersToProto(primitive.HashedValueComprarers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByWord:
			ep.Primitive = &evmcap.Primitive_EventByWord{
				EventByWord: &evmcap.EventByWord{
					WordIndex:            uint32(primitive.WordIndex),
					HashedValueComparers: hashedValueComparersToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventSig:
			ep.Primitive = &evmcap.Primitive_EventSig{
				EventSig: &evmcap.EventSig{
					EventSig: &evmcap.Hash{Hash: primitive.EventSig[:]},
				},
			}

			putEVMPrimitive(pbExpression, ep)
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &evmcap.Expression_BooleanExpression{BooleanExpression: &evmcap.BooleanExpression{}}
	var expressions []*evmcap.Expression
	for _, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := expressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pbExpr)
	}
	pbExpression.Evaluator = &evmcap.Expression_BooleanExpression{
		BooleanExpression: &evmcap.BooleanExpression{
			BooleanOperator: pb.BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func ProtoToExpressions(expressions []*evmcap.Expression) ([]query.Expression, error) {
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

func protoToExpression(pbExpression *evmcap.Expression) (query.Expression, error) {
	switch pbEvaluatedExpr := pbExpression.Evaluator.(type) {
	case *evmcap.Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range pbEvaluatedExpr.BooleanExpression.Expression {
			convertedExpression, err := protoToExpression(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}
		if pbEvaluatedExpr.BooleanExpression.BooleanOperator == pb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *evmcap.Expression_Primitive:
		switch primitive := pbEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *evmcap.Primitive_GeneralPrimitive:
			return protoToGeneralExpr(primitive.GeneralPrimitive)
		default:
			return protoToEVMExpr(pbEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", pbEvaluatedExpr)
	}
}

func protoToGeneralExpr(pbEvaluatedExpr *pb.Primitive) (query.Expression, error) {
	switch primitive := pbEvaluatedExpr.GetPrimitive().(type) {
	case *pb.Primitive_Comparator:
		return query.Expression{}, errors.New("comparator primitive is not supported for EVMService")
	case *pb.Primitive_Confidence:
		confidence, err := evmcap.ConfidenceFromProto(primitive.Confidence)
		return query.Confidence(confidence), err
	case *pb.Primitive_Block:
		return query.Block(primitive.Block.BlockNumber, primitives.ComparisonOperator(primitive.Block.Operator)), nil
	case *pb.Primitive_TxHash:
		return query.TxHash(primitive.TxHash.TxHash), nil
	case *pb.Primitive_Timestamp:
		return query.Timestamp(primitive.Timestamp.Timestamp, primitives.ComparisonOperator(primitive.Timestamp.Operator)), nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
	}
}

func protoToEVMExpr(pbEvaluatedExpr *evmcap.Primitive) (query.Expression, error) {
	switch primitive := pbEvaluatedExpr.GetPrimitive().(type) {
	case *evmcap.Primitive_ContractAddress:
		address := ProtoToAddress(primitive.ContractAddress.GetAddress())
		return evmprimitives.NewAddressFilter(address), nil
	case *evmcap.Primitive_EventSig:
		hash := ProtoToHash(primitive.EventSig.GetEventSig())
		return evmprimitives.NewEventSigFilter(hash), nil
	case *evmcap.Primitive_EventByTopic:
		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(),
				protoToHashedValueComparers(primitive.EventByTopic.GetHashedValueComparers())),
			nil
	case *evmcap.Primitive_EventByWord:
		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()),
				protoToHashedValueComparers(primitive.EventByWord.GetHashedValueComparers())),
			nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *evmcap.Expression, p *pb.Primitive) {
	exp.Evaluator = &evmcap.Expression_Primitive{Primitive: &evmcap.Primitive{Primitive: &evmcap.Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *evmcap.Expression, p *evmcap.Primitive) {
	exp.Evaluator = &evmcap.Expression_Primitive{Primitive: &evmcap.Primitive{Primitive: p.Primitive}}
}
