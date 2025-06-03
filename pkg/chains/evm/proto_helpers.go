package evm

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

func convertAddressesFromProto(addresses [][]byte) []evmtypes.Address {
	evmAddresses := make([]evmtypes.Address, 0, len(addresses))
	for _, address := range addresses {
		if len(address) != 20 {
			continue // Invalid address length
		}
		evmAddresses = append(evmAddresses, evmtypes.Address(address))
	}
	return evmAddresses
}

func convertAddressesToProto(addresses []evmtypes.Address) [][]byte {
	protoAddresses := make([][]byte, 0, len(addresses))
	for _, address := range addresses {
		protoAddresses = append(protoAddresses, address[:])
	}
	return protoAddresses
}

func convertHashesFromProto(hashes [][]byte) []evmtypes.Hash {
	hashesList := make([]evmtypes.Hash, 0, len(hashes))
	for _, hash := range hashes {
		if len(hash) != 32 {
			continue // Invalid hash length
		}
		hashesList = append(hashesList, evmtypes.Hash(hash))
	}
	return hashesList
}
func convertHashesToProto(hashes []evmtypes.Hash) [][]byte {
	protoHashes := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		protoHashes = append(protoHashes, hash[:])
	}
	return protoHashes
}

func convertTopicsToProto(topics [][]evmtypes.Hash) []*Topics {
	protoTopics := make([]*Topics, 0, len(topics))
	for _, topic := range topics {
		topicProto := &Topics{Topic: convertHashesToProto(topic)}
		protoTopics = append(protoTopics, topicProto)
	}
	return protoTopics
}

func ConvertHeadToProto(h evmtypes.Head) *Head {
	return &Head{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        h.Hash[:],
		ParentHash:  h.ParentHash[:],
	}
}

var errEmptyHead = errors.New("head is nil")

func ConvertHeadFromProto(head *Head) (evmtypes.Head, error) {
	if head == nil {
		return evmtypes.Head{}, errEmptyHead
	}
	return evmtypes.Head{
		Timestamp:  head.GetTimestamp(),
		Hash:       evmtypes.Hash(head.GetHash()[:]),
		ParentHash: evmtypes.Hash(head.GetParentHash()[:]),
		Number:     valuespb.NewIntFromBigInt(head.GetBlockNumber()),
	}, nil
}

var errEmptyReceipt = errors.New("receipt is nil")

func ConvertReceiptToProto(receipt *evmtypes.Receipt) (*Receipt, error) {
	if receipt == nil {
		return nil, errEmptyReceipt
	}

	return &Receipt{
		Status:            receipt.Status,
		Logs:              ConvertLogsToProto(receipt.Logs),
		TxHash:            receipt.TxHash[:],
		ContractAddress:   receipt.ContractAddress[:],
		GasUsed:           receipt.GasUsed,
		BlockHash:         receipt.BlockHash[:],
		BlockNumber:       valuespb.NewBigIntFromInt(receipt.BlockNumber),
		TxIndex:           receipt.TransactionIndex,
		EffectiveGasPrice: valuespb.NewBigIntFromInt(receipt.EffectiveGasPrice),
	}, nil
}

func ConvertReceiptFromProto(protoReceipt *Receipt) (*evmtypes.Receipt, error) {
	if protoReceipt == nil {
		return nil, errEmptyReceipt
	}
	return &evmtypes.Receipt{
		Status:            protoReceipt.GetStatus(),
		Logs:              ConvertLogsFromProto(protoReceipt.GetLogs()),
		TxHash:            evmtypes.Hash(protoReceipt.GetTxHash()),
		ContractAddress:   evmtypes.Address(protoReceipt.GetContractAddress()),
		GasUsed:           protoReceipt.GetGasUsed(),
		BlockHash:         evmtypes.Hash(protoReceipt.GetBlockHash()),
		BlockNumber:       valuespb.NewIntFromBigInt(protoReceipt.GetBlockNumber()),
		TransactionIndex:  protoReceipt.GetTxIndex(),
		EffectiveGasPrice: valuespb.NewIntFromBigInt(protoReceipt.GetEffectiveGasPrice()),
	}, nil
}

var errEmptyTx = errors.New("transaction is nil")

func ConvertTransactionToProto(tx *evmtypes.Transaction) (*Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &Transaction{
		To:       tx.To[:],
		Data:     tx.Data,
		Hash:     tx.Hash[:],
		Nonce:    tx.Nonce,
		Gas:      tx.Gas,
		GasPrice: valuespb.NewBigIntFromInt(tx.GasPrice),
		Value:    valuespb.NewBigIntFromInt(tx.Value),
	}, nil
}

func ConvertTransactionFromProto(protoTx *Transaction) (*evmtypes.Transaction, error) {
	if protoTx == nil {
		return nil, errEmptyTx
	}

	var data []byte
	if protoTx.GetData() != nil {
		data = protoTx.GetData()
	}

	return &evmtypes.Transaction{
		To:       evmtypes.Address(protoTx.GetTo()),
		Data:     data,
		Hash:     evmtypes.Hash(protoTx.GetHash()),
		Nonce:    protoTx.GetNonce(),
		Gas:      protoTx.GetGas(),
		GasPrice: valuespb.NewIntFromBigInt(protoTx.GetGasPrice()),
		Value:    valuespb.NewIntFromBigInt(protoTx.GetValue()),
	}, nil
}

var errEmptyMsg = errors.New("call msg can't be nil")

func ConvertCallMsgToProto(msg *evmtypes.CallMsg) (*CallMsg, error) {
	if msg == nil {
		return nil, errEmptyMsg
	}

	return &CallMsg{
		From: msg.From[:],
		To:   msg.To[:],
		Data: msg.Data,
	}, nil
}

func ConvertCallMsgFromProto(protoMsg *CallMsg) (*evmtypes.CallMsg, error) {
	if protoMsg == nil {
		return nil, errEmptyMsg
	}

	return &evmtypes.CallMsg{
		From: evmtypes.Address(protoMsg.GetFrom()),
		Data: protoMsg.GetData(),
		To:   evmtypes.Address(protoMsg.GetTo()),
	}, nil
}

var errEmptyFilter = errors.New("filter can't be nil")

func ConvertLPFilterToProto(filter evmtypes.LPFilterQuery) *LPFilter {
	convertAddressesToProto := func(addresses []evmtypes.Address) [][]byte {
		protoAddresses := make([][]byte, 0, len(addresses))
		for _, address := range addresses {
			protoAddresses = append(protoAddresses, address[:])
		}
		return protoAddresses
	}
	return &LPFilter{
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

func ConvertLPFilterFromProto(protoFilter *LPFilter) (evmtypes.LPFilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.LPFilterQuery{}, errEmptyFilter
	}

	return evmtypes.LPFilterQuery{
		Name:         protoFilter.GetName(),
		Retention:    time.Duration(protoFilter.GetRetentionTime()),
		Addresses:    convertAddressesFromProto(protoFilter.GetAddresses()),
		EventSigs:    convertHashesFromProto(protoFilter.GetEventSigs()),
		Topic2:       convertHashesFromProto(protoFilter.GetTopic2()),
		Topic3:       convertHashesFromProto(protoFilter.GetTopic3()),
		Topic4:       convertHashesFromProto(protoFilter.GetTopic4()),
		MaxLogsKept:  protoFilter.GetMaxLogsKept(),
		LogsPerBlock: protoFilter.GetLogsPerBlock(),
	}, nil
}

func ConvertFilterToProto(filter evmtypes.FilterQuery) *FilterQuery {
	return &FilterQuery{
		BlockHash: filter.BlockHash[:],
		FromBlock: valuespb.NewBigIntFromInt(filter.FromBlock),
		ToBlock:   valuespb.NewBigIntFromInt(filter.ToBlock),
		Addresses: convertAddressesToProto(filter.Addresses),
		Topics:    convertTopicsToProto(filter.Topics),
	}
}

func ConvertLogsToProto(logs []*evmtypes.Log) []*Log {
	protoLogs := make([]*Log, 0, len(logs))
	for _, l := range logs {
		protoLogs = append(protoLogs, ConvertLogToProto(l))
	}
	return protoLogs
}

func ConvertFilterFromProto(protoFilter *FilterQuery) (evmtypes.FilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.FilterQuery{}, errEmptyFilter
	}
	return evmtypes.FilterQuery{
		BlockHash: evmtypes.Hash(protoFilter.GetBlockHash()),
		FromBlock: valuespb.NewIntFromBigInt(protoFilter.GetFromBlock()),
		ToBlock:   valuespb.NewIntFromBigInt(protoFilter.GetToBlock()),
		Addresses: convertAddressesFromProto(protoFilter.GetAddresses()),
		Topics:    ConvertTopicsFromProto(protoFilter.GetTopics()),
	}, nil
}

func ConvertLogsFromProto(protoLogs []*Log) []*evmtypes.Log {
	logs := make([]*evmtypes.Log, 0, len(protoLogs))
	for _, protoLog := range protoLogs {
		logs = append(logs, convertLogFromProto(protoLog))
	}
	return logs
}

func convertLogFromProto(protoLog *Log) *evmtypes.Log {
	return &evmtypes.Log{
		LogIndex:    protoLog.GetIndex(),
		BlockHash:   evmtypes.Hash(protoLog.GetBlockHash()),
		BlockNumber: valuespb.NewIntFromBigInt(protoLog.GetBlockNumber()),
		Topics:      convertHashesFromProto(protoLog.GetTopics()),
		EventSig:    evmtypes.Hash(protoLog.GetEventSig()),
		Address:     evmtypes.Address(protoLog.GetAddress()),
		TxHash:      evmtypes.Hash(protoLog.GetTxHash()),
		Data:        protoLog.GetData(),
		Removed:     protoLog.GetRemoved(),
	}
}

func ConvertTopicsFromProto(protoTopics []*Topics) [][]evmtypes.Hash {
	topics := make([][]evmtypes.Hash, 0, len(protoTopics))
	for _, topic := range protoTopics {
		hash := make([]evmtypes.Hash, 0, len(topic.GetTopic()))
		for _, t := range topic.GetTopic() {
			hash = append(hash, evmtypes.Hash(t))
		}
		topics = append(topics, hash)
	}
	return topics
}

func ConvertLogToProto(log *evmtypes.Log) *Log {
	var topics [][]byte
	for _, topic := range log.Topics {
		topics = append(topics, topic[:])
	}

	return &Log{
		Index:       log.LogIndex,
		BlockHash:   log.BlockHash[:],
		BlockNumber: valuespb.NewBigIntFromInt(log.BlockNumber),
		Topics:      topics,
		EventSig:    log.EventSig[:],
		Address:     log.Address[:],
		TxHash:      log.TxHash[:],
		Data:        log.Data[:],
		// TODO tx index
		//TxIndex: log.TxIndex
		Removed: log.Removed,
	}
}

func ConvertHashedValueComparatorsToProto(hashedValueComparators []evmprimitives.HashedValueComparator) []*HashValueComparator {
	protoHashedValueComparators := make([]*HashValueComparator, 0, len(hashedValueComparators))
	for _, hvc := range hashedValueComparators {
		var values [][]byte
		for _, value := range hvc.Values {
			values = append(values, value[:])
		}
		protoHashedValueComparators = append(protoHashedValueComparators,
			&HashValueComparator{
				//nolint: gosec // G115
				Operator: chaincommonpb.ComparisonOperator(hvc.Operator),
				Values:   values,
			})
	}
	return protoHashedValueComparators
}

func ConvertHashedValueComparatorsFromProto(protoHashedValueComparators []*HashValueComparator) []evmprimitives.HashedValueComparator {
	hashedValueComparators := make([]evmprimitives.HashedValueComparator, 0, len(protoHashedValueComparators))
	for _, protoHvc := range protoHashedValueComparators {
		values := make([]evmtypes.Hash, 0, len(protoHvc.GetValues()))
		for _, value := range protoHvc.GetValues() {
			values = append(values, evmtypes.Hash(value))
		}
		hashedValueComparators = append(hashedValueComparators,
			evmprimitives.HashedValueComparator{
				Values:   values,
				Operator: primitives.ComparisonOperator(protoHvc.GetOperator()),
			})
	}
	return hashedValueComparators
}

func ConvertExpressionsToProto(expressions []query.Expression) ([]*Expression, error) {
	protoExpressions := make([]*Expression, 0, len(expressions))
	for _, expr := range expressions {
		protoExpression, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		protoExpressions = append(protoExpressions, protoExpression)
	}
	return protoExpressions, nil
}

func ConvertExpressionsFromProto(protoExpressions []*Expression) ([]query.Expression, error) {
	expressions := make([]query.Expression, 0, len(protoExpressions))
	for idx, protoExpression := range protoExpressions {
		expr, err := convertExpressionFromProto(protoExpression)
		if err != nil {
			return nil, fmt.Errorf("err to convert expr idx %d err: %w", idx, err)
		}

		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func convertExpressionToProto(expression query.Expression) (*Expression, error) {
	pbExpression := &Expression{}
	if expression.IsPrimitive() {
		ep := &Primitive{}
		switch primitive := expression.Primitive.(type) {
		case *evmprimitives.Address:
			ep.Primitive = &Primitive_ContractAddress{ContractAddress: primitive.Address[:]}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByTopic:
			ep.Primitive = &Primitive_EventByTopic{
				EventByTopic: &EventByTopic{
					Topic:                primitive.Topic,
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventByWord:
			ep.Primitive = &Primitive_EventByWord{
				EventByWord: &EventByWord{
					//nolint: gosec // G115
					WordIndex:            uint32(primitive.WordIndex),
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventSig:
			ep.Primitive = &Primitive_EventSig{
				EventSig: primitive.EventSig[:],
			}

			putEVMPrimitive(pbExpression, ep)
		default:
			generalPrimitive, err := chaincommonpb.ConvertPrimitiveToProto(primitive, func(value any) (*codecpb.VersionedBytes, error) {
				return nil, fmt.Errorf("unsupported primitive type: %T", value)
			})
			if err != nil {
				return nil, err
			}
			putGeneralPrimitive(pbExpression, generalPrimitive)
		}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &Expression_BooleanExpression{BooleanExpression: &BooleanExpression{}}
	expressions := make([]*Expression, 0)
	for _, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pbExpr)
	}
	pbExpression.Evaluator = &Expression_BooleanExpression{
		BooleanExpression: &BooleanExpression{
			//nolint: gosec // G115
			BooleanOperator: chaincommonpb.BooleanOperator(expression.BoolExpression.BoolOperator),
			Expression:      expressions,
		}}

	return pbExpression, nil
}

func convertExpressionFromProto(protoExpression *Expression) (query.Expression, error) {
	switch protoEvaluatedExpr := protoExpression.GetEvaluator().(type) {
	case *Expression_BooleanExpression:
		var expressions []query.Expression
		for _, expression := range protoEvaluatedExpr.BooleanExpression.GetExpression() {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, err
			}
			expressions = append(expressions, convertedExpression)
		}
		if protoEvaluatedExpr.BooleanExpression.GetBooleanOperator() == chaincommonpb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil

	case *Expression_Primitive:
		switch primitive := protoEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *Primitive_GeneralPrimitive:
			return chaincommonpb.ConvertPrimitiveFromProto(primitive.GeneralPrimitive, func(_ string, _ bool) (any, error) {
				return nil, fmt.Errorf("unsupported primitive type: %T", primitive)
			})
		default:
			return convertEVMExpressionToProto(protoEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown expression type: %T", protoExpression)
	}
}

func convertEVMExpressionToProto(protoPrimitive *Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *Primitive_ContractAddress:
		address := evmtypes.Address(primitive.ContractAddress)
		return evmprimitives.NewAddressFilter(address), nil
	case *Primitive_EventSig:
		return evmprimitives.NewEventSigFilter(evmtypes.Hash(primitive.EventSig)), nil
	case *Primitive_EventByTopic:
		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(),
			ConvertHashedValueComparatorsFromProto(primitive.EventByTopic.GetHashedValueComparers())), nil
	case *Primitive_EventByWord:
		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()),
			ConvertHashedValueComparatorsFromProto(primitive.EventByWord.GetHashedValueComparers())), nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *Expression, p *chaincommonpb.Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: &Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *Expression, p *Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: p.Primitive}}
}
