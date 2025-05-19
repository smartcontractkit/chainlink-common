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
		return nil, net.WrapRPCErr(err)
	}

	return &evm.TransactionFee{TransactionFee: valuespb.NewIntFromBigInt(reply.TransationFee)}, nil
}

func (e *EVMClient) CallContract(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int) ([]byte, error) {
	protoCallMsg, err := ConvertCallMsgToProto(msg)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.CallContract(ctx, &evmcap.CallContractRequest{
		Call:        protoCallMsg,
		BlockNumber: valuespb.NewBigIntFromInt(blockNumber),
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply.Data.GetAbi(), nil
}

func (e *EVMClient) FilterLogs(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error) {
	reply, err := e.grpcClient.FilterLogs(ctx, &evmcap.FilterLogsRequest{FilterQuery: ConvertFilterToProto(filterQuery)})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertLogsFromProto(reply.Logs), nil
}

func (e *EVMClient) BalanceAt(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error) {
	reply, err := e.grpcClient.BalanceAt(ctx, &evmcap.BalanceAtRequest{
		Account:     &evmcap.Address{Address: account[:]},
		BlockNumber: valuespb.NewBigIntFromInt(blockNumber),
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return valuespb.NewIntFromBigInt(reply.Balance), nil
}

func (e *EVMClient) EstimateGas(ctx context.Context, msg *evm.CallMsg) (uint64, error) {
	protoCallMsg, err := ConvertCallMsgToProto(msg)
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.EstimateGas(ctx, &evmcap.EstimateGasRequest{Msg: protoCallMsg})
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	return reply.Gas, nil
}

func (e *EVMClient) TransactionByHash(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
	reply, err := e.grpcClient.TransactionByHash(ctx, &evmcap.TransactionByHashRequest{Hash: &evmcap.Hash{Hash: hash[:]}})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertTransactionFromProto(reply.Transaction)
}

func (e *EVMClient) TransactionReceipt(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error) {
	reply, err := e.grpcClient.TransactionReceipt(ctx, &evmcap.TransactionReceiptRequest{Hash: &evmcap.Hash{Hash: txHash[:]}})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertReceiptFromProto(reply.Receipt)
}

func (e *EVMClient) LatestAndFinalizedHead(ctx context.Context) (latest evm.Head, finalized evm.Head, err error) {
	reply, err := e.grpcClient.LatestAndFinalizedHead(ctx, &emptypb.Empty{})
	if err != nil {
		return evm.Head{}, evm.Head{}, net.WrapRPCErr(err)
	}

	latest, err = convertHeadFromProto(reply.Latest)
	if err != nil {
		return evm.Head{}, evm.Head{}, net.WrapRPCErr(err)
	}

	finalized, err = convertHeadFromProto(reply.Finalized)
	if err != nil {
		return evm.Head{}, evm.Head{}, net.WrapRPCErr(err)
	}

	return latest, finalized, nil

}
func (e *EVMClient) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	protoExpressions, err := convertExpressionsToProto(filterQuery)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoLimitAndSort, err := contractreader.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoConfidenceLevel, err := contractreader.ConfidenceToProto(confidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.QueryTrackedLogs(ctx, &evmcap.QueryTrackedLogsRequest{
		Expression:      protoExpressions,
		LimitAndSort:    protoLimitAndSort,
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertLogsFromProto(reply.Logs), nil
}

func (e *EVMClient) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	_, err := e.grpcClient.RegisterLogTracking(ctx, &evmcap.RegisterLogTrackingRequest{Filter: convertLPFilterToProto(filter)})
	return net.WrapRPCErr(err)
}

func (e *EVMClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.grpcClient.UnregisterLogTracking(ctx, &evmcap.UnregisterLogTrackingRequest{FilterName: filterName})
	return net.WrapRPCErr(err)
}

func (e *EVMClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := e.grpcClient.GetTransactionStatus(ctx, &pb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, net.WrapRPCErr(err)
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
	txFee, err := e.impl.GetTransactionFee(ctx, request.TransactionId)
	if err != nil {
		return nil, err
	}

	return &evmcap.GetTransactionFeeReply{TransationFee: valuespb.NewBigIntFromInt(txFee.TransactionFee)}, nil
}

func (e *EvmServer) CallContract(ctx context.Context, req *evmcap.CallContractRequest) (*evmcap.CallContractReply, error) {
	callMsg, err := ConvertCallMsgFromProto(req.Call)
	if err != nil {
		return nil, err
	}

	data, err := e.impl.CallContract(ctx, callMsg, valuespb.NewIntFromBigInt(req.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmcap.CallContractReply{Data: &evmcap.ABIPayload{Abi: data}}, nil
}
func (e *EvmServer) FilterLogs(ctx context.Context, req *evmcap.FilterLogsRequest) (*evmcap.FilterLogsReply, error) {
	filter, err := ConvertFilterFromProto(req.FilterQuery)
	if err != nil {
		return nil, err
	}

	logs, err := e.impl.FilterLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &evmcap.FilterLogsReply{Logs: ConvertLogsToProto(logs)}, nil
}
func (e *EvmServer) BalanceAt(ctx context.Context, req *evmcap.BalanceAtRequest) (*evmcap.BalanceAtReply, error) {
	balance, err := e.impl.BalanceAt(ctx, ConvertAddressFromProto(req.Account), valuespb.NewIntFromBigInt(req.BlockNumber))
	if err != nil {
		return nil, err
	}

	return &evmcap.BalanceAtReply{Balance: valuespb.NewBigIntFromInt(balance)}, nil
}

func (e *EvmServer) EstimateGas(ctx context.Context, req *evmcap.EstimateGasRequest) (*evmcap.EstimateGasReply, error) {
	callMsg, err := ConvertCallMsgFromProto(req.Msg)
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, err
	}

	return &evmcap.EstimateGasReply{Gas: gas}, nil
}

func (e *EvmServer) GetTransactionByHash(ctx context.Context, req *evmcap.TransactionByHashRequest) (*evmcap.TransactionByHashReply, error) {
	tx, err := e.impl.TransactionByHash(ctx, ConvertHashFromProto(req.GetHash()))
	if err != nil {
		return nil, err
	}

	protoTx, err := ConvertTransactionToProto(tx)
	if err != nil {
		return nil, err
	}

	return &evmcap.TransactionByHashReply{Transaction: protoTx}, nil
}

func (e *EvmServer) GetTransactionReceipt(ctx context.Context, req *evmcap.TransactionReceiptRequest) (*evmcap.TransactionReceiptReply, error) {
	receipt, err := e.impl.TransactionReceipt(ctx, ConvertHashFromProto(req.GetHash()))
	if err != nil {
		return nil, err
	}

	protoReceipt, err := ConvertReceiptToProto(receipt)
	if err != nil {
		return nil, err
	}

	return &evmcap.TransactionReceiptReply{Receipt: protoReceipt}, nil
}

func (e *EvmServer) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*evmcap.LatestAndFinalizedHeadReply, error) {
	latest, finalized, err := e.impl.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmcap.LatestAndFinalizedHeadReply{
		Latest:    ConvertHeadToProto(latest),
		Finalized: ConvertHeadToProto(finalized),
	}, nil
}

func (e *EvmServer) QueryTrackedLogs(ctx context.Context, req *evmcap.QueryTrackedLogsRequest) (*evmcap.QueryTrackedLogsReply, error) {
	expressions, err := ConvertExpressionsFromProto(req.Expression)
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

	logs, err := e.impl.QueryTrackedLogs(ctx, expressions, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmcap.QueryTrackedLogsReply{Logs: ConvertLogsToProto(logs)}, nil
}

func (e *EvmServer) RegisterLogTracking(ctx context.Context, req *evmcap.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	lpFilter, err := ConvertLPFilterFromProto(req.Filter)
	if err != nil {
		return nil, err
	}
	return nil, e.impl.RegisterLogTracking(ctx, lpFilter)
}

func (e *EvmServer) UnregisterLogTracking(ctx context.Context, req *evmcap.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	return nil, e.impl.UnregisterLogTracking(ctx, req.FilterName)
}

func (e *EvmServer) GetTransactionStatus(ctx context.Context, req *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusReply, error) {
	txStatus, err := e.impl.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionStatusReply{TransactionStatus: pb.TransactionStatus(txStatus)}, nil
}

func ConvertHeadToProto(h evm.Head) *evmcap.Head {
	return &evmcap.Head{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        convertHashToProto(h.Hash),
		ParentHash:  convertHashToProto(h.ParentHash),
	}
}

var errEmptyHead = errors.New("head is nil")

func convertHeadFromProto(h *evmcap.Head) (evm.Head, error) {
	if h == nil {
		return evm.Head{}, errEmptyHead
	}
	return evm.Head{
		Timestamp:  h.Timestamp,
		Hash:       ConvertHashFromProto(h.GetHash()),
		ParentHash: ConvertHashFromProto(h.GetParentHash()),
		Number:     valuespb.NewIntFromBigInt(h.GetBlockNumber()),
	}, nil
}

var errEmptyReceipt = errors.New("receipt is nil")

func ConvertReceiptToProto(r *evm.Receipt) (*evmcap.Receipt, error) {
	if r == nil {
		return nil, errEmptyReceipt
	}

	return &evmcap.Receipt{
		Status:            r.Status,
		Logs:              ConvertLogsToProto(r.Logs),
		TxHash:            convertHashToProto(r.TxHash),
		ContractAddress:   convertAddressToProto(r.ContractAddress),
		GasUsed:           r.GasUsed,
		BlockHash:         convertHashToProto(r.BlockHash),
		BlockNumber:       valuespb.NewBigIntFromInt(r.BlockNumber),
		TxIndex:           r.TransactionIndex,
		EffectiveGasPrice: valuespb.NewBigIntFromInt(r.EffectiveGasPrice),
	}, nil
}

func ConvertReceiptFromProto(protoReceipt *evmcap.Receipt) (*evm.Receipt, error) {
	if protoReceipt == nil {
		return nil, errEmptyReceipt
	}

	return &evm.Receipt{
		Status:            protoReceipt.GetStatus(),
		Logs:              ConvertLogsFromProto(protoReceipt.GetLogs()),
		TxHash:            ConvertHashFromProto(protoReceipt.GetTxHash()),
		ContractAddress:   ConvertAddressFromProto(protoReceipt.GetContractAddress()),
		GasUsed:           protoReceipt.GetGasUsed(),
		BlockHash:         ConvertHashFromProto(protoReceipt.GetBlockHash()),
		BlockNumber:       valuespb.NewIntFromBigInt(protoReceipt.GetBlockNumber()),
		TransactionIndex:  protoReceipt.GetTxIndex(),
		EffectiveGasPrice: valuespb.NewIntFromBigInt(protoReceipt.GetEffectiveGasPrice()),
	}, nil
}

var errEmptyTx = errors.New("transaction is nil")

func ConvertTransactionToProto(tx *evm.Transaction) (*evmcap.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evmcap.Transaction{
		To:       convertAddressToProto(tx.To),
		Data:     ConvertABIPayloadToProto(tx.Data),
		Hash:     convertHashToProto(tx.Hash),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: valuespb.NewBigIntFromInt(tx.GasPrice),
		Value:    valuespb.NewBigIntFromInt(tx.Value),
	}, nil
}

func ConvertTransactionFromProto(protoTx *evmcap.Transaction) (*evm.Transaction, error) {
	if protoTx == nil {
		return nil, errEmptyTx
	}

	var data []byte
	if protoTx.GetData() != nil {
		data = protoTx.GetData().GetAbi()
	}

	return &evm.Transaction{
		To:       ConvertAddressFromProto(protoTx.GetTo()),
		Data:     data,
		Hash:     ConvertHashFromProto(protoTx.GetHash()),
		Nonce:    protoTx.GetNonce(),
		Gas:      protoTx.GetGas(),
		GasPrice: valuespb.NewIntFromBigInt(protoTx.GetGasPrice()),
		Value:    valuespb.NewIntFromBigInt(protoTx.GetValue()),
	}, nil
}

var errEmptyMsg = errors.New("call msg can't be nil")

func ConvertCallMsgToProto(msg *evm.CallMsg) (*evmcap.CallMsg, error) {
	if msg == nil {
		return nil, errEmptyMsg
	}

	return &evmcap.CallMsg{
		From: convertAddressToProto(msg.From),
		To:   convertAddressToProto(msg.To),
		Data: ConvertABIPayloadToProto(msg.Data),
	}, nil
}

func ConvertCallMsgFromProto(protoMsg *evmcap.CallMsg) (*evm.CallMsg, error) {
	if protoMsg == nil {
		return nil, errEmptyMsg
	}

	return &evm.CallMsg{
		From: ConvertAddressFromProto(protoMsg.GetFrom()),
		Data: protoMsg.GetData().GetAbi(),
		To:   ConvertAddressFromProto(protoMsg.GetTo()),
	}, nil
}

var errEmptyFilter = errors.New("filter cant be nil")

func convertLPFilterToProto(filter evm.LPFilterQuery) *evmcap.LPFilter {
	return &evmcap.LPFilter{
		Name:          filter.Name,
		RetentionTime: int64(filter.Retention),
		Addresses:     convertAddressesToProto(filter.Addresses),
		EventSigs:     convertHashesToProto(filter.EventSigs),
		Topic2:        convertHashesToProto(filter.Topic2),
		Topic3:        convertHashesToProto(filter.Topic3),
		Topic4:        convertHashesToProto(filter.Topic4),
		MaxLogsKept:   filter.MaxLogsKept,
		LogsPerBlock:  filter.LogsPerBlock,
	}
}

func ConvertLPFilterFromProto(protoFilter *evmcap.LPFilter) (evm.LPFilterQuery, error) {
	if protoFilter == nil {
		return evm.LPFilterQuery{}, errEmptyFilter
	}

	return evm.LPFilterQuery{
		Name:         protoFilter.GetName(),
		Retention:    time.Duration(protoFilter.GetRetentionTime()),
		Addresses:    ConvertAddressesFromProto(protoFilter.GetAddresses()),
		EventSigs:    convertHashesFromProto(protoFilter.GetEventSigs()),
		Topic2:       convertHashesFromProto(protoFilter.GetTopic2()),
		Topic3:       convertHashesFromProto(protoFilter.GetTopic3()),
		Topic4:       convertHashesFromProto(protoFilter.GetTopic4()),
		MaxLogsKept:  protoFilter.GetMaxLogsKept(),
		LogsPerBlock: protoFilter.GetLogsPerBlock(),
	}, nil
}

func ConvertFilterToProto(filter evm.FilterQuery) *evmcap.FilterQuery {
	return &evmcap.FilterQuery{
		BlockHash: convertHashToProto(filter.BlockHash),
		FromBlock: valuespb.NewBigIntFromInt(filter.FromBlock),
		ToBlock:   valuespb.NewBigIntFromInt(filter.ToBlock),
		Addresses: convertAddressesToProto(filter.Addresses),
		Topics:    convertTopicsToProto(filter.Topics),
	}
}

func ConvertFilterFromProto(protoFilter *evmcap.FilterQuery) (evm.FilterQuery, error) {
	if protoFilter == nil {
		return evm.FilterQuery{}, errEmptyFilter
	}
	return evm.FilterQuery{
		BlockHash: ConvertHashFromProto(protoFilter.GetBlockHash()),
		FromBlock: valuespb.NewIntFromBigInt(protoFilter.GetFromBlock()),
		ToBlock:   valuespb.NewIntFromBigInt(protoFilter.GetToBlock()),
		Addresses: ConvertAddressesFromProto(protoFilter.GetAddresses()),
		Topics:    convertTopicsFromProto(protoFilter.GetTopics()),
	}, nil
}

func ConvertLogsToProto(logs []*evm.Log) []*evmcap.Log {
	ret := make([]*evmcap.Log, 0, len(logs))
	for _, l := range logs {
		ret = append(ret, ConvertLogToProto(l))
	}
	return ret
}

func ConvertLogsFromProto(protoLogs []*evmcap.Log) []*evm.Log {
	logs := make([]*evm.Log, 0, len(protoLogs))
	for _, protoLog := range protoLogs {
		logs = append(logs, convertLogFromProto(protoLog))
	}
	return logs
}

func ConvertLogToProto(log *evm.Log) *evmcap.Log {
	return &evmcap.Log{
		Index:       log.LogIndex,
		BlockHash:   convertHashToProto(log.BlockHash),
		BlockNumber: valuespb.NewBigIntFromInt(log.BlockNumber),
		Topics:      convertHashesToProto(log.Topics),
		EventSig:    convertHashToProto(log.EventSig),
		Address:     convertAddressToProto(log.Address),
		TxHash:      convertHashToProto(log.TxHash),
		Data:        ConvertABIPayloadToProto(log.Data),
		Removed:     log.Removed,
	}
}

func convertLogFromProto(l *evmcap.Log) *evm.Log {
	var data []byte
	if l.GetData() != nil {
		data = l.GetData().GetAbi()
	}

	return &evm.Log{
		LogIndex:    l.GetIndex(),
		BlockHash:   ConvertHashFromProto(l.GetBlockHash()),
		BlockNumber: valuespb.NewIntFromBigInt(l.GetBlockNumber()),
		Topics:      convertHashesFromProto(l.GetTopics()),
		EventSig:    ConvertHashFromProto(l.GetEventSig()),
		Address:     ConvertAddressFromProto(l.GetAddress()),
		TxHash:      ConvertHashFromProto(l.GetTxHash()),
		Data:        data,
		Removed:     l.GetRemoved(),
	}
}

func convertHashesToProto(hash []evm.Hash) []*evmcap.Hash {
	protoHash := make([]*evmcap.Hash, 0, len(hash))
	for _, s := range hash {
		protoHash = append(protoHash, convertHashToProto(s))
	}
	return protoHash
}

func convertHashesFromProto(protoHashes []*evmcap.Hash) []evm.Hash {
	hashes := make([]evm.Hash, 0, len(protoHashes))
	for _, h := range protoHashes {
		hashes = append(hashes, ConvertHashFromProto(h))
	}
	return hashes
}

func convertHashToProto(hash evm.Hash) *evmcap.Hash {
	return &evmcap.Hash{Hash: hash[:]}
}

func ConvertHashFromProto(protoHash *evmcap.Hash) evm.Hash {
	var h evm.Hash
	if protoHash != nil {
		copy(h[:], protoHash.GetHash())
	}
	return h
}

func convertTopicsToProto(topics [][]evm.Hash) []*evmcap.Topics {
	protoTopics := make([]*evmcap.Topics, 0, len(topics))
	for _, topic := range topics {
		protoTopics = append(protoTopics, &evmcap.Topics{Topic: convertHashesToProto(topic)})
	}
	return protoTopics
}

func convertTopicsFromProto(protoTopics []*evmcap.Topics) [][]evm.Hash {
	topics := make([][]evm.Hash, 0, len(protoTopics))
	for _, topic := range protoTopics {
		topics = append(topics, convertHashesFromProto(topic.GetTopic()))
	}
	return topics
}

func convertAddressesToProto(addresses []evm.Address) []*evmcap.Address {
	protoAddresses := make([]*evmcap.Address, 0, len(addresses))
	for _, s := range addresses {
		protoAddresses = append(protoAddresses, convertAddressToProto(s))
	}
	return protoAddresses
}

func ConvertAddressesFromProto(protoAddresses []*evmcap.Address) []evm.Address {
	addresses := make([]evm.Address, 0, len(protoAddresses))
	for _, protoAddress := range protoAddresses {
		addresses = append(addresses, ConvertAddressFromProto(protoAddress))
	}

	return addresses
}

func convertAddressToProto(address evm.Address) *evmcap.Address {
	return &evmcap.Address{Address: address[:]}
}

func ConvertAddressFromProto(protoAddress *evmcap.Address) evm.Address {
	if protoAddress != nil {
		return evm.Address(protoAddress.GetAddress()[:])
	}
	return evm.Address{}
}

func ConvertABIPayloadToProto(payload []byte) *evmcap.ABIPayload {
	return &evmcap.ABIPayload{Abi: payload}
}

func ConvertHashedValueComparatorsToProto(hashedValueComparators []evmprimitives.HashedValueComparator) []*evmcap.HashValueComparator {
	protoHashedValueComparators := make([]*evmcap.HashValueComparator, 0, len(hashedValueComparators))
	for _, hvc := range hashedValueComparators {
		protoHashedValueComparators = append(protoHashedValueComparators,
			&evmcap.HashValueComparator{
				Operator: pb.ComparisonOperator(hvc.Operator),
				Values:   convertHashesToProto(hvc.Values),
			})
	}
	return protoHashedValueComparators
}

func ConvertHashedValueComparatorsFromProto(protoHashedValueComparators []*evmcap.HashValueComparator) []evmprimitives.HashedValueComparator {
	hashedValueComparators := make([]evmprimitives.HashedValueComparator, 0, len(protoHashedValueComparators))
	for _, pbHvc := range protoHashedValueComparators {
		hashedValueComparators = append(hashedValueComparators,
			evmprimitives.HashedValueComparator{
				Values:   convertHashesFromProto(pbHvc.GetValues()),
				Operator: primitives.ComparisonOperator(pbHvc.GetOperator()),
			})
	}
	return hashedValueComparators
}

func convertExpressionsToProto(expressions []query.Expression) ([]*evmcap.Expression, error) {
	protoExpressions := make([]*evmcap.Expression, 0, len(expressions))
	for _, expr := range expressions {
		protoExpression, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, net.WrapRPCErr(err)
		}
		protoExpressions = append(protoExpressions, protoExpression)
	}
	return protoExpressions, nil
}

func ConvertExpressionsFromProto(protoExpressions []*evmcap.Expression) ([]query.Expression, error) {
	expressions := make([]query.Expression, 0, len(protoExpressions))
	for _, protoExpression := range protoExpressions {
		expr, err := convertExpressionFromProto(protoExpression)
		if err != nil {
			return nil, fmt.Errorf("failed to convert expressions, err: %w", err)
		}

		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func convertExpressionToProto(expression query.Expression) (*evmcap.Expression, error) {
	pbExpression := &evmcap.Expression{}
	if expression.IsPrimitive() {
		p := &pb.Primitive{}
		ep := &evmcap.Primitive{}
		switch primitive := expression.Primitive.(type) {
		case *primitives.Comparator:
			return nil, errors.New("comparator primitive is not supported for evm service")
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
				return nil, net.WrapRPCErr(err)
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
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByWord:
			ep.Primitive = &evmcap.Primitive_EventByWord{
				EventByWord: &evmcap.EventByWord{
					WordIndex:            uint32(primitive.WordIndex),
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
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
			return nil, status.Errorf(codes.InvalidArgument, "unknown primitive type: %T", primitive)
		}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &evmcap.Expression_BooleanExpression{BooleanExpression: &evmcap.BooleanExpression{}}
	var expressions []*evmcap.Expression
	for _, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, net.WrapRPCErr(err)
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

func convertExpressionFromProto(protoExpression *evmcap.Expression) (query.Expression, error) {
	switch protoEvaluatedExpr := protoExpression.GetEvaluator().(type) {
	case *evmcap.Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range protoEvaluatedExpr.BooleanExpression.GetExpression() {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}
		if protoEvaluatedExpr.BooleanExpression.GetBooleanOperator() == pb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *evmcap.Expression_Primitive:
		switch primitive := protoEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *evmcap.Primitive_GeneralPrimitive:
			return convertGeneralExpressionToProto(primitive.GeneralPrimitive)
		default:
			return convertEVMExpressionToProto(protoEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "unknown expression type: %T", protoEvaluatedExpr)
	}
}

func convertGeneralExpressionToProto(pbEvaluatedExpr *pb.Primitive) (query.Expression, error) {
	switch primitive := pbEvaluatedExpr.GetPrimitive().(type) {
	case *pb.Primitive_Comparator:
		return query.Expression{}, errors.New("comparator primitive is not supported")
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

func convertEVMExpressionToProto(pbEvaluatedExpr *evmcap.Primitive) (query.Expression, error) {
	switch primitive := pbEvaluatedExpr.GetPrimitive().(type) {
	case *evmcap.Primitive_ContractAddress:
		address := ConvertAddressFromProto(primitive.ContractAddress.GetAddress())
		return evmprimitives.NewAddressFilter(address), nil
	case *evmcap.Primitive_EventSig:
		hash := ConvertHashFromProto(primitive.EventSig.GetEventSig())
		return evmprimitives.NewEventSigFilter(hash), nil
	case *evmcap.Primitive_EventByTopic:
		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(),
			ConvertHashedValueComparatorsFromProto(primitive.EventByTopic.GetHashedValueComparers())), nil
	case *evmcap.Primitive_EventByWord:
		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()),
			ConvertHashedValueComparatorsFromProto(primitive.EventByWord.GetHashedValueComparers())), nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *evmcap.Expression, p *pb.Primitive) {
	exp.Evaluator = &evmcap.Expression_Primitive{Primitive: &evmcap.Primitive{Primitive: &evmcap.Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *evmcap.Expression, p *evmcap.Primitive) {
	exp.Evaluator = &evmcap.Expression_Primitive{Primitive: &evmcap.Primitive{Primitive: p.Primitive}}
}
