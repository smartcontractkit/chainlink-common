package evmpb

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

func ConvertHeadToProto(h evmtypes.Head) *Head {
	return &Head{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        convertHashToProto(h.Hash),
		ParentHash:  convertHashToProto(h.ParentHash),
	}
}

var errEmptyHead = errors.New("head is nil")

func ConvertHeadFromProto(head *Head) (evmtypes.Head, error) {
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

func ConvertReceiptToProto(receipt *evmtypes.Receipt) (*Receipt, error) {
	if receipt == nil {
		return nil, errEmptyReceipt
	}

	return &Receipt{
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

func ConvertReceiptFromProto(protoReceipt *Receipt) (*evmtypes.Receipt, error) {
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

func ConvertTransactionToProto(tx *evmtypes.Transaction) (*Transaction, error) {
	if tx == nil {
		return nil, errEmptyTx
	}
	return &Transaction{
		To:       convertAddressToProto(tx.To),
		Data:     convertABIPayloadToProto(tx.Data),
		Hash:     convertHashToProto(tx.Hash),
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

func ConvertCallMsgToProto(msg *evmtypes.CallMsg) (*CallMsg, error) {
	if msg == nil {
		return nil, errEmptyMsg
	}

	return &CallMsg{
		From: convertAddressToProto(msg.From),
		To:   convertAddressToProto(msg.To),
		Data: convertABIPayloadToProto(msg.Data),
	}, nil
}

func ConvertCallMsgFromProto(protoMsg *CallMsg) (*evmtypes.CallMsg, error) {
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

func ConvertLPFilterToProto(filter evmtypes.LPFilterQuery) *LPFilter {
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
		Addresses:    ConvertAddressesFromProto(protoFilter.GetAddresses()),
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
		BlockHash: convertHashToProto(filter.BlockHash),
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
		BlockHash: ConvertHashFromProto(protoFilter.GetBlockHash()),
		FromBlock: valuespb.NewIntFromBigInt(protoFilter.GetFromBlock()),
		ToBlock:   valuespb.NewIntFromBigInt(protoFilter.GetToBlock()),
		Addresses: ConvertAddressesFromProto(protoFilter.GetAddresses()),
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

func convertHashesFromProto(protoHashes []*Hash) []evmtypes.Hash {
	hashes := make([]evmtypes.Hash, 0, len(protoHashes))
	for _, h := range protoHashes {
		hashes = append(hashes, ConvertHashFromProto(h))
	}
	return hashes
}

func ConvertHashFromProto(protoHash *Hash) evmtypes.Hash {
	var hash evmtypes.Hash
	if protoHash != nil {
		copy(hash[:], protoHash.GetHash())
	}
	return hash
}

func ConvertTopicsFromProto(protoTopics []*Topics) [][]evmtypes.Hash {
	topics := make([][]evmtypes.Hash, 0, len(protoTopics))
	for _, topic := range protoTopics {
		topics = append(topics, convertHashesFromProto(topic.GetTopic()))
	}
	return topics
}
func ConvertAddressesFromProto(protoAddresses []*Address) []evmtypes.Address {
	addresses := make([]evmtypes.Address, 0, len(protoAddresses))
	for _, protoAddress := range protoAddresses {
		addresses = append(addresses, ConvertAddressFromProto(protoAddress))
	}

	return addresses
}

func ConvertLogToProto(log *evmtypes.Log) *Log {
	return &Log{
		Index:       log.LogIndex,
		BlockHash:   convertHashToProto(log.BlockHash),
		BlockNumber: valuespb.NewBigIntFromInt(log.BlockNumber),
		Topics:      convertHashesToProto(log.Topics),
		EventSig:    convertHashToProto(log.EventSig),
		Address:     convertAddressToProto(log.Address),
		TxHash:      convertHashToProto(log.TxHash),
		Data:        convertABIPayloadToProto(log.Data),
		Removed:     log.Removed,
	}
}

func convertHashesToProto(hashes []evmtypes.Hash) []*Hash {
	protoHash := make([]*Hash, 0, len(hashes))
	for _, hash := range hashes {
		protoHash = append(protoHash, convertHashToProto(hash))
	}
	return protoHash
}

func convertHashToProto(hash evmtypes.Hash) *Hash {
	return &Hash{Hash: hash[:]}
}

func convertTopicsToProto(topics [][]evmtypes.Hash) []*Topics {
	protoTopics := make([]*Topics, 0, len(topics))
	for _, topic := range topics {
		protoTopics = append(protoTopics, &Topics{Topic: convertHashesToProto(topic)})
	}
	return protoTopics
}

func convertAddressesToProto(addresses []evmtypes.Address) []*Address {
	protoAddresses := make([]*Address, 0, len(addresses))
	for _, s := range addresses {
		protoAddresses = append(protoAddresses, convertAddressToProto(s))
	}
	return protoAddresses
}

func convertAddressToProto(address evmtypes.Address) *Address {
	return &Address{Address: address[:]}
}

func ConvertAddressFromProto(protoAddress *Address) evmtypes.Address {
	if protoAddress != nil {
		return evmtypes.Address(protoAddress.GetAddress()[:])
	}
	return evmtypes.Address{}
}

func convertABIPayloadToProto(payload []byte) *ABIPayload {
	return &ABIPayload{Abi: payload}
}

func ConvertHashedValueComparatorsToProto(hashedValueComparators []evmprimitives.HashedValueComparator) []*HashValueComparator {
	protoHashedValueComparators := make([]*HashValueComparator, 0, len(hashedValueComparators))
	for _, hvc := range hashedValueComparators {
		protoHashedValueComparators = append(protoHashedValueComparators,
			&HashValueComparator{
				Operator: pb.ComparisonOperator(hvc.Operator),
				Values:   convertHashesToProto(hvc.Values),
			})
	}
	return protoHashedValueComparators
}

func ConvertHashedValueComparatorsFromProto(protoHashedValueComparators []*HashValueComparator) []evmprimitives.HashedValueComparator {
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
			return nil, fmt.Errorf("err to convert expr idx %d err: %v", idx, err)
		}

		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func convertExpressionToProto(expression query.Expression) (*Expression, error) {
	pbExpression := &Expression{}
	if expression.IsPrimitive() {
		p := &pb.Primitive{}
		ep := &Primitive{}
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
			ep.Primitive = &Primitive_ContractAddress{ContractAddress: &ContractAddress{
				Address: &Address{Address: primitive.Address[:]},
			}}

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
					WordIndex:            uint32(primitive.WordIndex),
					HashedValueComparers: ConvertHashedValueComparatorsToProto(primitive.HashedValueComparers),
				},
			}

			putEVMPrimitive(pbExpression, ep)
		case *evmprimitives.EventSig:
			ep.Primitive = &Primitive_EventSig{
				EventSig: &EventSig{
					EventSig: &Hash{Hash: primitive.EventSig[:]},
				},
			}

			putEVMPrimitive(pbExpression, ep)
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Unknown primitive type: %T", primitive)
		}
		return pbExpression, nil
	}

	pbExpression.Evaluator = &Expression_BooleanExpression{BooleanExpression: &BooleanExpression{}}
	var expressions []*Expression
	for _, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pbExpr)
	}
	pbExpression.Evaluator = &Expression_BooleanExpression{
		BooleanExpression: &BooleanExpression{
			BooleanOperator: pb.BooleanOperator(expression.BoolExpression.BoolOperator),
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
		if protoEvaluatedExpr.BooleanExpression.GetBooleanOperator() == pb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil
	case *Expression_Primitive:
		switch primitive := protoEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *Primitive_GeneralPrimitive:
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

func convertEVMExpressionToProto(protoPrimitive *Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *Primitive_ContractAddress:
		address := ConvertAddressFromProto(primitive.ContractAddress.GetAddress())
		return evmprimitives.NewAddressFilter(address), nil
	case *Primitive_EventSig:
		hash := ConvertHashFromProto(primitive.EventSig.GetEventSig())
		return evmprimitives.NewEventSigFilter(hash), nil
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

func putGeneralPrimitive(exp *Expression, p *pb.Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: &Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *Expression, p *Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: p.Primitive}}
}
