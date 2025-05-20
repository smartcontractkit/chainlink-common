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

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

type EVMClient struct {
	grpcClient evmpb.EVMClient
}

func NewEVMCClient(grpcClient evmpb.EVMClient) *EVMClient {
	return &EVMClient{
		grpcClient: grpcClient,
	}
}

var _ types.EVMService = (*EVMClient)(nil)

func (e *EVMClient) GetTransactionFee(ctx context.Context, transactionID string) (*evmtypes.TransactionFee, error) {
	reply, err := e.grpcClient.GetTransactionFee(ctx, &evmpb.GetTransactionFeeRequest{TransactionId: transactionID})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &evmtypes.TransactionFee{TransactionFee: valuespb.NewIntFromBigInt(reply.GetTransationFee())}, nil
}

func (e *EVMClient) CallContract(ctx context.Context, msg *evmtypes.CallMsg, blockNumber *big.Int) ([]byte, error) {
	protoCallMsg, err := ConvertCallMsgToProto(msg)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.CallContract(ctx, &evmpb.CallContractRequest{
		Call:        protoCallMsg,
		BlockNumber: valuespb.NewBigIntFromInt(blockNumber),
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply.GetData().GetAbi(), nil
}

func (e *EVMClient) FilterLogs(ctx context.Context, filterQuery evmtypes.FilterQuery) ([]*evmtypes.Log, error) {
	reply, err := e.grpcClient.FilterLogs(ctx, &evmpb.FilterLogsRequest{FilterQuery: ConvertFilterToProto(filterQuery)})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertLogsFromProto(reply.GetLogs()), nil
}

func (e *EVMClient) BalanceAt(ctx context.Context, account evmtypes.Address, blockNumber *big.Int) (*big.Int, error) {
	reply, err := e.grpcClient.BalanceAt(ctx, &evmpb.BalanceAtRequest{
		Account:     &evmpb.Address{Address: account[:]},
		BlockNumber: valuespb.NewBigIntFromInt(blockNumber),
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return valuespb.NewIntFromBigInt(reply.GetBalance()), nil
}

func (e *EVMClient) EstimateGas(ctx context.Context, msg *evmtypes.CallMsg) (uint64, error) {
	protoCallMsg, err := ConvertCallMsgToProto(msg)
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.EstimateGas(ctx, &evmpb.EstimateGasRequest{Msg: protoCallMsg})
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	return reply.GetGas(), nil
}

func (e *EVMClient) TransactionByHash(ctx context.Context, hash evmtypes.Hash) (*evmtypes.Transaction, error) {
	reply, err := e.grpcClient.TransactionByHash(ctx, &evmpb.TransactionByHashRequest{Hash: &evmpb.Hash{Hash: hash[:]}})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertTransactionFromProto(reply.GetTransaction())
}

func (e *EVMClient) TransactionReceipt(ctx context.Context, txHash evmtypes.Hash) (*evmtypes.Receipt, error) {
	reply, err := e.grpcClient.TransactionReceipt(ctx, &evmpb.TransactionReceiptRequest{Hash: &evmpb.Hash{Hash: txHash[:]}})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertReceiptFromProto(reply.GetReceipt())
}

func (e *EVMClient) LatestAndFinalizedHead(ctx context.Context) (latest evmtypes.Head, finalized evmtypes.Head, err error) {
	reply, err := e.grpcClient.LatestAndFinalizedHead(ctx, &emptypb.Empty{})
	if err != nil {
		return evmtypes.Head{}, evmtypes.Head{}, net.WrapRPCErr(err)
	}

	latest, err = convertHeadFromProto(reply.GetLatest())
	if err != nil {
		return evmtypes.Head{}, evmtypes.Head{}, net.WrapRPCErr(err)
	}

	finalized, err = convertHeadFromProto(reply.GetFinalized())
	if err != nil {
		return evmtypes.Head{}, evmtypes.Head{}, net.WrapRPCErr(err)
	}

	return latest, finalized, nil

}
func (e *EVMClient) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evmtypes.Log, error) {
	protoExpressions, err := convertExpressionsToProto(filterQuery)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoLimitAndSort, err := contractreader.ConvertLimitAndSortToProto(limitAndSort)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	protoConfidenceLevel, err := contractreader.ConvertConfidenceToProto(confidenceLevel)
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	reply, err := e.grpcClient.QueryTrackedLogs(ctx, &evmpb.QueryTrackedLogsRequest{
		Expression:      protoExpressions,
		LimitAndSort:    protoLimitAndSort,
		ConfidenceLevel: protoConfidenceLevel,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return ConvertLogsFromProto(reply.GetLogs()), nil
}

func (e *EVMClient) RegisterLogTracking(ctx context.Context, filter evmtypes.LPFilterQuery) error {
	_, err := e.grpcClient.RegisterLogTracking(ctx, &evmpb.RegisterLogTrackingRequest{Filter: convertLPFilterToProto(filter)})
	return net.WrapRPCErr(err)
}

func (e *EVMClient) UnregisterLogTracking(ctx context.Context, filterName string) error {
	_, err := e.grpcClient.UnregisterLogTracking(ctx, &evmpb.UnregisterLogTrackingRequest{FilterName: filterName})
	return net.WrapRPCErr(err)
}

func (e *EVMClient) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := e.grpcClient.GetTransactionStatus(ctx, &pb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, net.WrapRPCErr(err)
	}

	return types.TransactionStatus(reply.GetTransactionStatus()), nil
}

type evmServer struct {
	evmpb.UnimplementedEVMServer

	*net.BrokerExt

	impl types.EVMService
}

var _ evmpb.EVMServer = (*evmServer)(nil)

func newEVMServer(impl types.EVMService, b *net.BrokerExt) *evmServer {
	return &evmServer{impl: impl, BrokerExt: b.WithName("EVMServer")}
}

func (e *evmServer) GetTransactionFee(ctx context.Context, request *evmpb.GetTransactionFeeRequest) (*evmpb.GetTransactionFeeReply, error) {
	txFee, err := e.impl.GetTransactionFee(ctx, request.GetTransactionId())
	if err != nil {
		return nil, err
	}

	return &evmpb.GetTransactionFeeReply{TransationFee: valuespb.NewBigIntFromInt(txFee.TransactionFee)}, nil
}

func (e *evmServer) CallContract(ctx context.Context, request *evmpb.CallContractRequest) (*evmpb.CallContractReply, error) {
	callMsg, err := ConvertCallMsgFromProto(request.GetCall())
	if err != nil {
		return nil, err
	}

	data, err := e.impl.CallContract(ctx, callMsg, valuespb.NewIntFromBigInt(request.GetBlockNumber()))
	if err != nil {
		return nil, err
	}

	return &evmpb.CallContractReply{Data: &evmpb.ABIPayload{Abi: data}}, nil
}
func (e *evmServer) FilterLogs(ctx context.Context, request *evmpb.FilterLogsRequest) (*evmpb.FilterLogsReply, error) {
	filter, err := ConvertFilterFromProto(request.GetFilterQuery())
	if err != nil {
		return nil, err
	}

	logs, err := e.impl.FilterLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &evmpb.FilterLogsReply{Logs: ConvertLogsToProto(logs)}, nil
}
func (e *evmServer) BalanceAt(ctx context.Context, request *evmpb.BalanceAtRequest) (*evmpb.BalanceAtReply, error) {
	balance, err := e.impl.BalanceAt(ctx, ConvertAddressFromProto(request.GetAccount()), valuespb.NewIntFromBigInt(request.GetBlockNumber()))
	if err != nil {
		return nil, err
	}

	return &evmpb.BalanceAtReply{Balance: valuespb.NewBigIntFromInt(balance)}, nil
}

func (e *evmServer) EstimateGas(ctx context.Context, request *evmpb.EstimateGasRequest) (*evmpb.EstimateGasReply, error) {
	callMsg, err := ConvertCallMsgFromProto(request.GetMsg())
	if err != nil {
		return nil, err
	}

	gas, err := e.impl.EstimateGas(ctx, callMsg)
	if err != nil {
		return nil, err
	}

	return &evmpb.EstimateGasReply{Gas: gas}, nil
}

func (e *evmServer) TransactionByHash(ctx context.Context, request *evmpb.TransactionByHashRequest) (*evmpb.TransactionByHashReply, error) {
	tx, err := e.impl.TransactionByHash(ctx, ConvertHashFromProto(request.GetHash()))
	if err != nil {
		return nil, err
	}

	protoTx, err := ConvertTransactionToProto(tx)
	if err != nil {
		return nil, err
	}

	return &evmpb.TransactionByHashReply{Transaction: protoTx}, nil
}

func (e *evmServer) TransactionReceipt(ctx context.Context, request *evmpb.TransactionReceiptRequest) (*evmpb.TransactionReceiptReply, error) {
	receipt, err := e.impl.TransactionReceipt(ctx, ConvertHashFromProto(request.GetHash()))
	if err != nil {
		return nil, err
	}

	protoReceipt, err := ConvertReceiptToProto(receipt)
	if err != nil {
		return nil, err
	}

	return &evmpb.TransactionReceiptReply{Receipt: protoReceipt}, nil
}

func (e *evmServer) LatestAndFinalizedHead(ctx context.Context, _ *emptypb.Empty) (*evmpb.LatestAndFinalizedHeadReply, error) {
	latest, finalized, err := e.impl.LatestAndFinalizedHead(ctx)
	if err != nil {
		return nil, err
	}

	return &evmpb.LatestAndFinalizedHeadReply{
		Latest:    ConvertHeadToProto(latest),
		Finalized: ConvertHeadToProto(finalized),
	}, nil
}

func (e *evmServer) QueryTrackedLogs(ctx context.Context, request *evmpb.QueryTrackedLogsRequest) (*evmpb.QueryTrackedLogsReply, error) {
	expressions, err := ConvertExpressionsFromProto(request.GetExpression())
	if err != nil {
		return nil, err
	}

	limitAndSort, err := contractreader.ConvertLimitAndSortFromProto(request.GetLimitAndSort())
	if err != nil {
		return nil, err
	}

	conf, err := contractreader.ConfidenceFromProto(request.GetConfidenceLevel())
	if err != nil {
		return nil, err
	}

	logs, err := e.impl.QueryTrackedLogs(ctx, expressions, limitAndSort, conf)
	if err != nil {
		return nil, err
	}

	return &evmpb.QueryTrackedLogsReply{Logs: ConvertLogsToProto(logs)}, nil
}

func (e *evmServer) RegisterLogTracking(ctx context.Context, request *evmpb.RegisterLogTrackingRequest) (*emptypb.Empty, error) {
	lpFilter, err := ConvertLPFilterFromProto(request.GetFilter())
	if err != nil {
		return nil, err
	}
	return nil, e.impl.RegisterLogTracking(ctx, lpFilter)
}

func (e *evmServer) UnregisterLogTracking(ctx context.Context, request *evmpb.UnregisterLogTrackingRequest) (*emptypb.Empty, error) {
	return nil, e.impl.UnregisterLogTracking(ctx, request.GetFilterName())
}

func (e *evmServer) GetTransactionStatus(ctx context.Context, request *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusReply, error) {
	txStatus, err := e.impl.GetTransactionStatus(ctx, request.GetTransactionId())
	if err != nil {
		return nil, err
	}

	return &pb.GetTransactionStatusReply{TransactionStatus: pb.TransactionStatus(txStatus)}, nil
}

func ConvertHeadToProto(h evmtypes.Head) *evmpb.Head {
	return &evmpb.Head{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        convertHashToProto(h.Hash),
		ParentHash:  convertHashToProto(h.ParentHash),
	}
}

var errEmptyHead = errors.New("head is nil")

func convertHeadFromProto(head *evmpb.Head) (evmtypes.Head, error) {
	if head == nil {
		return evmtypes.Head{}, errEmptyHead
	}
	return evmtypes.Head{
		Timestamp:  head.GetTimestamp(),
		Hash:       ConvertHashFromProto(head.GetHash()),
		ParentHash: ConvertHashFromProto(head.GetParentHash()),
		Number:     valuespb.NewIntFromBigInt(head.GetBlockNumber()),
	}, nil
}

var errEmptyReceipt = errors.New("receipt is nil")

func ConvertReceiptToProto(receipt *evmtypes.Receipt) (*evmpb.Receipt, error) {
	if receipt == nil {
		return nil, errEmptyReceipt
	}

	return &evmpb.Receipt{
		Status:            receipt.Status,
		Logs:              ConvertLogsToProto(receipt.Logs),
		TxHash:            convertHashToProto(receipt.TxHash),
		ContractAddress:   convertAddressToProto(receipt.ContractAddress),
		GasUsed:           receipt.GasUsed,
		BlockHash:         convertHashToProto(receipt.BlockHash),
		BlockNumber:       valuespb.NewBigIntFromInt(receipt.BlockNumber),
		TxIndex:           receipt.TransactionIndex,
		EffectiveGasPrice: valuespb.NewBigIntFromInt(receipt.EffectiveGasPrice),
	}, nil
}

func ConvertReceiptFromProto(protoReceipt *evmpb.Receipt) (*evmtypes.Receipt, error) {
	if protoReceipt == nil {
		return nil, errEmptyReceipt
	}
	return &evmtypes.Receipt{
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

func ConvertTransactionToProto(tx *evmtypes.Transaction) (*evmpb.Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &evmpb.Transaction{
		To:       convertAddressToProto(tx.To),
		Data:     ConvertABIPayloadToProto(tx.Data),
		Hash:     convertHashToProto(tx.Hash),
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: valuespb.NewBigIntFromInt(tx.GasPrice),
		Value:    valuespb.NewBigIntFromInt(tx.Value),
	}, nil
}

func ConvertTransactionFromProto(protoTx *evmpb.Transaction) (*evmtypes.Transaction, error) {
	if protoTx == nil {
		return nil, errEmptyTx
	}

	var data []byte
	if protoTx.GetData() != nil {
		data = protoTx.GetData().GetAbi()
	}

	return &evmtypes.Transaction{
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

func ConvertCallMsgToProto(msg *evmtypes.CallMsg) (*evmpb.CallMsg, error) {
	if msg == nil {
		return nil, errEmptyMsg
	}

	return &evmpb.CallMsg{
		From: convertAddressToProto(msg.From),
		To:   convertAddressToProto(msg.To),
		Data: ConvertABIPayloadToProto(msg.Data),
	}, nil
}

func ConvertCallMsgFromProto(protoMsg *evmpb.CallMsg) (*evmtypes.CallMsg, error) {
	if protoMsg == nil {
		return nil, errEmptyMsg
	}

	return &evmtypes.CallMsg{
		From: ConvertAddressFromProto(protoMsg.GetFrom()),
		Data: protoMsg.GetData().GetAbi(),
		To:   ConvertAddressFromProto(protoMsg.GetTo()),
	}, nil
}

var errEmptyFilter = errors.New("filter can't be nil")

func convertLPFilterToProto(filter evmtypes.LPFilterQuery) *evmpb.LPFilter {
	return &evmpb.LPFilter{
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

func ConvertLPFilterFromProto(protoFilter *evmpb.LPFilter) (evmtypes.LPFilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.LPFilterQuery{}, errEmptyFilter
	}

	return evmtypes.LPFilterQuery{
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

func ConvertFilterToProto(filter evmtypes.FilterQuery) *evmpb.FilterQuery {
	return &evmpb.FilterQuery{
		BlockHash: convertHashToProto(filter.BlockHash),
		FromBlock: valuespb.NewBigIntFromInt(filter.FromBlock),
		ToBlock:   valuespb.NewBigIntFromInt(filter.ToBlock),
		Addresses: convertAddressesToProto(filter.Addresses),
		Topics:    convertTopicsToProto(filter.Topics),
	}
}

func ConvertFilterFromProto(protoFilter *evmpb.FilterQuery) (evmtypes.FilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.FilterQuery{}, errEmptyFilter
	}
	return evmtypes.FilterQuery{
		BlockHash: ConvertHashFromProto(protoFilter.GetBlockHash()),
		FromBlock: valuespb.NewIntFromBigInt(protoFilter.GetFromBlock()),
		ToBlock:   valuespb.NewIntFromBigInt(protoFilter.GetToBlock()),
		Addresses: ConvertAddressesFromProto(protoFilter.GetAddresses()),
		Topics:    convertTopicsFromProto(protoFilter.GetTopics()),
	}, nil
}

func ConvertLogsToProto(logs []*evmtypes.Log) []*evmpb.Log {
	protoLogs := make([]*evmpb.Log, 0, len(logs))
	for _, l := range logs {
		protoLogs = append(protoLogs, ConvertLogToProto(l))
	}
	return protoLogs
}

func ConvertLogsFromProto(protoLogs []*evmpb.Log) []*evmtypes.Log {
	logs := make([]*evmtypes.Log, 0, len(protoLogs))
	for _, protoLog := range protoLogs {
		logs = append(logs, convertLogFromProto(protoLog))
	}
	return logs
}

func ConvertLogToProto(log *evmtypes.Log) *evmpb.Log {
	return &evmpb.Log{
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

func convertLogFromProto(protoLog *evmpb.Log) *evmtypes.Log {
	var data []byte
	if protoLog.GetData() != nil {
		data = protoLog.GetData().GetAbi()
	}

	return &evmtypes.Log{
		LogIndex:    protoLog.GetIndex(),
		BlockHash:   ConvertHashFromProto(protoLog.GetBlockHash()),
		BlockNumber: valuespb.NewIntFromBigInt(protoLog.GetBlockNumber()),
		Topics:      convertHashesFromProto(protoLog.GetTopics()),
		EventSig:    ConvertHashFromProto(protoLog.GetEventSig()),
		Address:     ConvertAddressFromProto(protoLog.GetAddress()),
		TxHash:      ConvertHashFromProto(protoLog.GetTxHash()),
		Data:        data,
		Removed:     protoLog.GetRemoved(),
	}
}

func convertHashesToProto(hashes []evmtypes.Hash) []*evmpb.Hash {
	protoHash := make([]*evmpb.Hash, 0, len(hashes))
	for _, hash := range hashes {
		protoHash = append(protoHash, convertHashToProto(hash))
	}
	return protoHash
}

func convertHashesFromProto(protoHashes []*evmpb.Hash) []evmtypes.Hash {
	hashes := make([]evmtypes.Hash, 0, len(protoHashes))
	for _, h := range protoHashes {
		hashes = append(hashes, ConvertHashFromProto(h))
	}
	return hashes
}

func convertHashToProto(hash evmtypes.Hash) *evmpb.Hash {
	return &evmpb.Hash{Hash: hash[:]}
}

func ConvertHashFromProto(protoHash *evmpb.Hash) evmtypes.Hash {
	var hash evmtypes.Hash
	if protoHash != nil {
		copy(hash[:], protoHash.GetHash())
	}
	return hash
}

func convertTopicsToProto(topics [][]evmtypes.Hash) []*evmpb.Topics {
	protoTopics := make([]*evmpb.Topics, 0, len(topics))
	for _, topic := range topics {
		protoTopics = append(protoTopics, &evmpb.Topics{Topic: convertHashesToProto(topic)})
	}
	return protoTopics
}

func convertTopicsFromProto(protoTopics []*evmpb.Topics) [][]evmtypes.Hash {
	topics := make([][]evmtypes.Hash, 0, len(protoTopics))
	for _, topic := range protoTopics {
		topics = append(topics, convertHashesFromProto(topic.GetTopic()))
	}
	return topics
}

func convertAddressesToProto(addresses []evmtypes.Address) []*evmpb.Address {
	protoAddresses := make([]*evmpb.Address, 0, len(addresses))
	for _, s := range addresses {
		protoAddresses = append(protoAddresses, convertAddressToProto(s))
	}
	return protoAddresses
}

func ConvertAddressesFromProto(protoAddresses []*evmpb.Address) []evmtypes.Address {
	addresses := make([]evmtypes.Address, 0, len(protoAddresses))
	for _, protoAddress := range protoAddresses {
		addresses = append(addresses, ConvertAddressFromProto(protoAddress))
	}

	return addresses
}

func convertAddressToProto(address evmtypes.Address) *evmpb.Address {
	return &evmpb.Address{Address: address[:]}
}

func ConvertAddressFromProto(protoAddress *evmpb.Address) evmtypes.Address {
	if protoAddress != nil {
		return evmtypes.Address(protoAddress.GetAddress()[:])
	}
	return evmtypes.Address{}
}

func ConvertABIPayloadToProto(payload []byte) *evmpb.ABIPayload {
	return &evmpb.ABIPayload{Abi: payload}
}

func ConvertHashedValueComparatorsToProto(hashedValueComparators []evmprimitives.HashedValueComparator) []*evmpb.HashValueComparator {
	protoHashedValueComparators := make([]*evmpb.HashValueComparator, 0, len(hashedValueComparators))
	for _, hvc := range hashedValueComparators {
		protoHashedValueComparators = append(protoHashedValueComparators,
			&evmpb.HashValueComparator{
				Operator: pb.ComparisonOperator(hvc.Operator),
				Values:   convertHashesToProto(hvc.Values),
			})
	}
	return protoHashedValueComparators
}

func ConvertHashedValueComparatorsFromProto(protoHashedValueComparators []*evmpb.HashValueComparator) []evmprimitives.HashedValueComparator {
	hashedValueComparators := make([]evmprimitives.HashedValueComparator, 0, len(protoHashedValueComparators))
	for _, protoHvc := range protoHashedValueComparators {
		hashedValueComparators = append(hashedValueComparators,
			evmprimitives.HashedValueComparator{
				Values:   convertHashesFromProto(protoHvc.GetValues()),
				Operator: primitives.ComparisonOperator(protoHvc.GetOperator()),
			})
	}
	return hashedValueComparators
}

func convertExpressionsToProto(expressions []query.Expression) ([]*evmpb.Expression, error) {
	protoExpressions := make([]*evmpb.Expression, 0, len(expressions))
	for _, expr := range expressions {
		protoExpression, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		protoExpressions = append(protoExpressions, protoExpression)
	}
	return protoExpressions, nil
}

func ConvertExpressionsFromProto(protoExpressions []*evmpb.Expression) ([]query.Expression, error) {
	expressions := make([]query.Expression, 0, len(protoExpressions))
	for idx, protoExpression := range protoExpressions {
		expr, err := convertExpressionFromProto(protoExpression)
		if err != nil {
			return nil, fmt.Errorf("err to convert expr idx %d err: %v", idx, err)
		}

		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func convertExpressionToProto(expression query.Expression) (*evmpb.Expression, error) {
	pbExpression := &evmpb.Expression{}
	if expression.IsPrimitive() {
		p := &pb.Primitive{}
		ep := &evmpb.Primitive{}
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
			pbConfidence, err := contractreader.ConvertConfidenceToProto(primitive.ConfidenceLevel)
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
			ep.Primitive = &evmpb.Primitive_ContractAddress{ContractAddress: &evmpb.ContractAddress{
				Address: &evmpb.Address{Address: primitive.Address[:]},
			}}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByTopic:
			ep.Primitive = &evmpb.Primitive_EventByTopic{
				EventByTopic: &evmpb.EventByTopic{
					Topic:                primitive.Topic,
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByWord:
			ep.Primitive = &evmpb.Primitive_EventByWord{
				EventByWord: &evmpb.EventByWord{
					WordIndex:            uint32(primitive.WordIndex),
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
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
		pbExpr, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pbExpr)
	}
	pbExpression.Evaluator = &evmpb.Expression_BooleanExpression{
		BooleanExpression: &evmpb.BooleanExpression{
			BooleanOperator: pb.BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func convertExpressionFromProto(protoExpression *evmpb.Expression) (query.Expression, error) {
	switch protoEvaluatedExpr := protoExpression.GetEvaluator().(type) {
	case *evmpb.Expression_BooleanExpression:
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
	case *evmpb.Expression_Primitive:
		switch primitive := protoEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *evmpb.Primitive_GeneralPrimitive:
			return convertGeneralExpressionToProto(primitive.GeneralPrimitive)
		default:
			return convertEVMExpressionToProto(protoEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", protoExpression)
	}
}

func convertGeneralExpressionToProto(protoPrimitive *pb.Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *pb.Primitive_Comparator:
		return query.Expression{}, errors.New("comparator primitive is not supported for EVMService")
	case *pb.Primitive_Confidence:
		confidence, err := contractreader.ConfidenceFromProto(primitive.Confidence)
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

func convertEVMExpressionToProto(protoPrimitive *evmpb.Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *evmpb.Primitive_ContractAddress:
		address := ConvertAddressFromProto(primitive.ContractAddress.GetAddress())
		return evmprimitives.NewAddressFilter(address), nil
	case *evmpb.Primitive_EventSig:
		hash := ConvertHashFromProto(primitive.EventSig.GetEventSig())
		return evmprimitives.NewEventSigFilter(hash), nil
	case *evmpb.Primitive_EventByTopic:
		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(),
			ConvertHashedValueComparatorsFromProto(primitive.EventByTopic.GetHashedValueComparers())), nil
	case *evmpb.Primitive_EventByWord:
		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()),
			ConvertHashedValueComparatorsFromProto(primitive.EventByWord.GetHashedValueComparers())), nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *evmpb.Expression, p *pb.Primitive) {
	exp.Evaluator = &evmpb.Expression_Primitive{Primitive: &evmpb.Primitive{Primitive: &evmpb.Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *evmpb.Expression, p *evmpb.Primitive) {
	exp.Evaluator = &evmpb.Expression_Primitive{Primitive: &evmpb.Primitive{Primitive: p.Primitive}}
}
