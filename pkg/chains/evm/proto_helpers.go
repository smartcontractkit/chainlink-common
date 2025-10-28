package evm

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"

	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
)

func bytesToHex(b []byte) string {
	return fmt.Sprintf("0x%x", b)
}

func ConvertAddressesFromProto(protoAddresses [][]byte) ([]evmtypes.Address, error) {
	addresses := make([]evmtypes.Address, 0, len(protoAddresses))
	var errs []error

	for i, protoAddress := range protoAddresses {
		address, err := ConvertAddressFromProto(protoAddress)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to convert address at index %d: %w", i, err))
			continue
		}
		addresses = append(addresses, address)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return addresses, nil
}

func ConvertAddressesToProto(addresses []evmtypes.Address) [][]byte {
	protoAddresses := make([][]byte, 0, len(addresses))
	for _, address := range addresses {
		protoAddresses = append(protoAddresses, address[:])
	}
	return protoAddresses
}

func ConvertHashesFromProto(protoHashes [][]byte) ([]evmtypes.Hash, error) {
	hashes := make([]evmtypes.Hash, 0, len(protoHashes))
	var errs []error

	for i, protoHash := range protoHashes {
		hash, err := ConvertHashFromProto(protoHash)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to convert hash at index %d: %w", i, err))
			continue
		}
		hashes = append(hashes, hash)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return hashes, nil
}

func ConvertHashesToProto(hashes []evmtypes.Hash) [][]byte {
	protoHashes := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		protoHashes = append(protoHashes, hash[:])
	}
	return protoHashes
}

func convertTopicsToProto(topics [][]evmtypes.Hash) ([]*Topics, error) {
	protoTopics := make([]*Topics, 0, len(topics))
	for i, topic := range topics {
		if topic == nil {
			return nil, fmt.Errorf("topic[%d] can't be nil", i)
		}

		protoTopics = append(protoTopics, &Topics{Topic: ConvertHashesToProto(topic)})
	}
	return protoTopics, nil
}

func ConvertHeaderToProto(h *evmtypes.Header) (*Header, error) {
	if h == nil {
		return nil, ErrEmptyHead
	}
	return &Header{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        h.Hash[:],
		ParentHash:  h.ParentHash[:],
	}, nil
}

var ErrEmptyHead = errors.New("head is nil")

func ConvertHeaderFromProto(protoHeader *Header) (evmtypes.Header, error) {
	if protoHeader == nil {
		return evmtypes.Header{}, ErrEmptyHead
	}

	hash, err := ConvertHashFromProto(protoHeader.GetHash())
	if err != nil {
		return evmtypes.Header{}, fmt.Errorf("failed to convert hash: %w", err)
	}

	parentHash, err := ConvertHashFromProto(protoHeader.GetParentHash())
	if err != nil {
		return evmtypes.Header{}, fmt.Errorf("failed to convert parent hash: %w", err)
	}

	return evmtypes.Header{
		Timestamp:  protoHeader.GetTimestamp(),
		Hash:       hash,
		ParentHash: parentHash,
		Number:     valuespb.NewIntFromBigInt(protoHeader.GetBlockNumber()),
	}, nil
}

var ErrEmptyReceipt = errors.New("receipt is nil")

func ConvertReceiptToProto(receipt *evmtypes.Receipt) (*Receipt, error) {
	if receipt == nil {
		return nil, ErrEmptyReceipt
	}

	logs, err := ConvertLogsToProto(receipt.Logs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert logs err: %w", err)
	}

	return &Receipt{
		Status:            receipt.Status,
		Logs:              logs,
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
		return nil, ErrEmptyReceipt
	}

	logs, err := ConvertLogsFromProto(protoReceipt.GetLogs())
	if err != nil {
		return nil, err
	}

	txHash, err := ConvertHashFromProto(protoReceipt.GetTxHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	// can be empty on contract creation
	contractAddress, err := ConvertOptionalAddressFromProto(protoReceipt.GetContractAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to convert contract address: %w", err)
	}

	blockHash, err := ConvertHashFromProto(protoReceipt.GetBlockHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}

	return &evmtypes.Receipt{
		Status:            protoReceipt.GetStatus(),
		Logs:              logs,
		TxHash:            txHash,
		ContractAddress:   contractAddress,
		GasUsed:           protoReceipt.GetGasUsed(),
		BlockHash:         blockHash,
		BlockNumber:       valuespb.NewIntFromBigInt(protoReceipt.GetBlockNumber()),
		TransactionIndex:  protoReceipt.GetTxIndex(),
		EffectiveGasPrice: valuespb.NewIntFromBigInt(protoReceipt.GetEffectiveGasPrice()),
	}, nil
}

var ErrEmptyTx = errors.New("transaction is nil")

func ConvertTransactionToProto(tx *evmtypes.Transaction) (*Transaction, error) {
	if tx == nil {
		return nil, ErrEmptyTx
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
		return nil, ErrEmptyTx
	}

	toAddress, err := ConvertOptionalAddressFromProto(protoTx.GetTo())
	if err != nil {
		return nil, fmt.Errorf("failed to convert 'to' address: %w", err)
	}

	txHash, err := ConvertHashFromProto(protoTx.GetHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	return &evmtypes.Transaction{
		To:       toAddress,
		Data:     protoTx.GetData(),
		Hash:     txHash,
		Nonce:    protoTx.GetNonce(),
		Gas:      protoTx.GetGas(),
		GasPrice: valuespb.NewIntFromBigInt(protoTx.GetGasPrice()),
		Value:    valuespb.NewIntFromBigInt(protoTx.GetValue()),
	}, nil
}

var ErrEmptyMsg = errors.New("call msg can't be nil")

func ConvertCallMsgToProto(msg *evmtypes.CallMsg) (*CallMsg, error) {
	if msg == nil {
		return nil, ErrEmptyMsg
	}

	return &CallMsg{
		From: msg.From[:],
		To:   msg.To[:],
		Data: msg.Data,
	}, nil
}

func ConvertCallMsgFromProto(protoMsg *CallMsg) (*evmtypes.CallMsg, error) {
	if protoMsg == nil {
		return nil, ErrEmptyMsg
	}

	toAddress, err := ConvertOptionalAddressFromProto(protoMsg.GetTo())
	if err != nil {
		return nil, fmt.Errorf("failed to convert 'to' address: %w", err)
	}

	callMsg := &evmtypes.CallMsg{
		Data: protoMsg.GetData(),
		To:   toAddress,
	}

	// fromAddress is optional
	if ValidateAddressBytes(protoMsg.GetFrom()) == nil {
		callMsg.From, err = ConvertAddressFromProto(protoMsg.GetFrom())
		if err != nil {
			return nil, fmt.Errorf("failed to convert 'from' address: %w", err)
		}
	}

	return callMsg, nil
}

var ErrEmptyFilter = errors.New("filter can't be nil")

func ConvertLPFilterToProto(filter evmtypes.LPFilterQuery) *LPFilter {
	return &LPFilter{
		Name:          filter.Name,
		RetentionTime: int64(filter.Retention),
		Addresses:     ConvertAddressesToProto(filter.Addresses),
		EventSigs:     ConvertHashesToProto(filter.EventSigs),
		Topic2:        ConvertHashesToProto(filter.Topic2),
		Topic3:        ConvertHashesToProto(filter.Topic3),
		Topic4:        ConvertHashesToProto(filter.Topic4),
		MaxLogsKept:   filter.MaxLogsKept,
		LogsPerBlock:  filter.LogsPerBlock,
	}
}

func ConvertLPFilterFromProto(protoFilter *LPFilter) (evmtypes.LPFilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.LPFilterQuery{}, ErrEmptyFilter
	}

	var addresses []evmtypes.Address
	for i, protoAddress := range protoFilter.GetAddresses() {
		address, err := ConvertOptionalAddressFromProto(protoAddress)
		if err != nil {
			return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert address[%d]: %w", i, err)
		}
		addresses = append(addresses, address)
	}

	sigs, err := ConvertHashesFromProto(protoFilter.GetEventSigs())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert event sigs: %w", err)
	}

	t2, err := ConvertHashesFromProto(protoFilter.GetTopic2())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic2: %w", err)
	}

	t3, err := ConvertHashesFromProto(protoFilter.GetTopic3())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic3: %w", err)
	}

	t4, err := ConvertHashesFromProto(protoFilter.GetTopic4())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic4: %w", err)
	}

	return evmtypes.LPFilterQuery{
		Name:         protoFilter.GetName(),
		Retention:    time.Duration(protoFilter.GetRetentionTime()),
		Addresses:    addresses,
		EventSigs:    sigs,
		Topic2:       t2,
		Topic3:       t3,
		Topic4:       t4,
		MaxLogsKept:  protoFilter.GetMaxLogsKept(),
		LogsPerBlock: protoFilter.GetLogsPerBlock(),
	}, nil
}

var ErrTopicsConversion = errors.New("failed to convert topics")

func ConvertFilterToProto(filter evmtypes.FilterQuery) (*FilterQuery, error) {
	topics, err := convertTopicsToProto(filter.Topics)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTopicsConversion, err)
	}

	return &FilterQuery{
		BlockHash: filter.BlockHash[:],
		FromBlock: valuespb.NewBigIntFromInt(filter.FromBlock),
		ToBlock:   valuespb.NewBigIntFromInt(filter.ToBlock),
		Addresses: ConvertAddressesToProto(filter.Addresses),
		Topics:    topics,
	}, nil
}

func ConvertLogsToProto(logs []*evmtypes.Log) ([]*Log, error) {
	protoLogs := make([]*Log, 0, len(logs))
	for i, log := range logs {
		if log == nil {
			return nil, fmt.Errorf("log[%d] can't be nil", i)
		}
		protoLogs = append(protoLogs, ConvertLogToProto(*log))
	}
	return protoLogs, nil
}

func ConvertFilterFromProto(protoFilter *FilterQuery) (evmtypes.FilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.FilterQuery{}, ErrEmptyFilter
	}

	blockHash, err := ConvertOptionalHashFromProto(protoFilter.GetBlockHash())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("failed to convert blockHash: %w", err)
	}

	addresses, err := ConvertAddressesFromProto(protoFilter.GetAddresses())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("failed to convert addresses: %w", err)
	}

	topics, err := ConvertTopicsFromProto(protoFilter.GetTopics())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("%w: %w", ErrTopicsConversion, err)
	}

	return evmtypes.FilterQuery{
		BlockHash: blockHash,
		FromBlock: valuespb.NewIntFromBigInt(protoFilter.GetFromBlock()),
		ToBlock:   valuespb.NewIntFromBigInt(protoFilter.GetToBlock()),
		Addresses: addresses,
		Topics:    topics,
	}, nil
}

func ConvertLogsFromProto(protoLogs []*Log) ([]*evmtypes.Log, error) {
	logs := make([]*evmtypes.Log, 0, len(protoLogs))
	var errs []error

	for i, protoLog := range protoLogs {
		if protoLog == nil {
			errs = append(errs, fmt.Errorf("log at index %d can't be nil", i))
			continue
		}

		l, err := convertLogFromProto(protoLog)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to convert log at index %d: %w", i, err))
			continue
		}
		logs = append(logs, l)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return logs, nil
}

func convertLogFromProto(protoLog *Log) (*evmtypes.Log, error) {
	if protoLog == nil {
		return nil, fmt.Errorf("log can't be nil")
	}

	blockHash, err := ConvertHashFromProto(protoLog.GetBlockHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}

	topics, err := ConvertHashesFromProto(protoLog.GetTopics())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTopicsConversion, err)
	}

	eventSigs, err := ConvertHashFromProto(protoLog.GetEventSig())
	if err != nil {
		return nil, fmt.Errorf("failed to convert event sig: %w", err)
	}

	address, err := ConvertAddressFromProto(protoLog.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	txHash, err := ConvertHashFromProto(protoLog.GetTxHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	return &evmtypes.Log{
		LogIndex:    protoLog.GetIndex(),
		BlockHash:   blockHash,
		BlockNumber: valuespb.NewIntFromBigInt(protoLog.GetBlockNumber()),
		Topics:      topics,
		EventSig:    eventSigs,
		Address:     address,
		TxHash:      txHash,
		Data:        protoLog.GetData(),
		Removed:     protoLog.GetRemoved(),
		// TODO TxIndex PRODCRE-1709
	}, nil
}

func ConvertTopicsFromProto(protoTopics []*Topics) ([][]evmtypes.Hash, error) {
	topics := make([][]evmtypes.Hash, 0, len(protoTopics))
	var errs []error

	for i, protoTopic := range protoTopics {
		if protoTopic == nil {
			errs = append(errs, fmt.Errorf("topic at index %d can't be nil", i))
			continue
		}

		hashes, err := ConvertHashesFromProto(protoTopic.GetTopic())
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to convert topic at index %d: %w", i, err))
			continue
		}

		topics = append(topics, hashes)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return topics, nil
}

func ConvertLogToProto(log evmtypes.Log) *Log {
	return &Log{
		Index:       log.LogIndex,
		BlockHash:   log.BlockHash[:],
		BlockNumber: valuespb.NewBigIntFromInt(log.BlockNumber),
		Topics:      ConvertHashesToProto(log.Topics),
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

func ConvertHashedValueComparatorsFromProto(protoHashedValueComparators []*HashValueComparator) ([]evmprimitives.HashedValueComparator, error) {
	hashedValueComparators := make([]evmprimitives.HashedValueComparator, 0, len(protoHashedValueComparators))
	for _, protoHvc := range protoHashedValueComparators {
		if protoHvc == nil {
			return nil, errors.New("hashed value comparator can't be nil")
		}
		values := make([]evmtypes.Hash, 0, len(protoHvc.GetValues()))
		for _, value := range protoHvc.GetValues() {
			hashValue, err := ConvertHashFromProto(value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert hash value: %w", err)
			}
			values = append(values, hashValue)
		}
		hashedValueComparators = append(hashedValueComparators,
			evmprimitives.HashedValueComparator{
				Values:   values,
				Operator: primitives.ComparisonOperator(protoHvc.GetOperator()),
			})
	}
	return hashedValueComparators, nil
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
			return nil, status.Errorf(codes.InvalidArgument, "err to convert expr idx %d err: %s", idx, err.Error())
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
	if protoExpression == nil {
		return query.Expression{}, errors.New("expression can not be nil")
	}

	switch protoEvaluatedExpr := protoExpression.GetEvaluator().(type) {
	case *Expression_BooleanExpression:
		var expressions []query.Expression
		for idx, expression := range protoEvaluatedExpr.BooleanExpression.GetExpression() {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, fmt.Errorf("failed to convert sub-expression %d: %w", idx, err)
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
		return query.Expression{}, fmt.Errorf("unknown expression type: %T", protoExpression.GetEvaluator())
	}
}

func convertEVMExpressionToProto(protoPrimitive *Primitive) (query.Expression, error) {
	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *Primitive_ContractAddress:
		address, err := ConvertAddressFromProto(primitive.ContractAddress)
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert contract address: %w", err)
		}
		return evmprimitives.NewAddressFilter(address), nil
	case *Primitive_EventSig:
		hash, err := ConvertHashFromProto(primitive.EventSig)
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert event sig: %w", err)
		}
		return evmprimitives.NewEventSigFilter(hash), nil
	case *Primitive_EventByTopic:
		if primitive.EventByTopic == nil {
			return query.Expression{}, errors.New("EventByTopic can not be nil")
		}
		valueCmp, err := ConvertHashedValueComparatorsFromProto(primitive.EventByTopic.GetHashedValueComparers())
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert EventByTopic hashed value comparators: %w", err)
		}
		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(), valueCmp), nil
	case *Primitive_EventByWord:
		if primitive.EventByWord == nil {
			return query.Expression{}, errors.New("EventByWord can not be nil")
		}
		valueCmp, err := ConvertHashedValueComparatorsFromProto(primitive.EventByWord.GetHashedValueComparers())
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert EventByWord hashed value comparators: %w", err)
		}
		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()), valueCmp), nil
	default:
		return query.Expression{}, fmt.Errorf("unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *Expression, p *chaincommonpb.Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: &Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *Expression, p *Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: p.Primitive}}
}

func ConvertGasConfigToProto(gasConfig evmtypes.GasConfig) (*GasConfig, error) {
	if gasConfig.GasLimit == nil {
		return nil, fmt.Errorf("gas limit can't be nil")
	}

	return &GasConfig{
		GasLimit: *gasConfig.GasLimit,
	}, nil
}

func ConvertGasConfigFromProto(gasConfig *GasConfig) *evmtypes.GasConfig {
	if gasConfig == nil {
		return nil
	}
	return &evmtypes.GasConfig{
		GasLimit: &gasConfig.GasLimit,
	}
}

func ConvertTxStatusFromProto(txStatus TxStatus) evmtypes.TransactionStatus {
	switch txStatus {
	case TxStatus_TX_SUCCESS:
		return evmtypes.TxSuccess
	case TxStatus_TX_REVERTED:
		return evmtypes.TxReverted
	default:
		return evmtypes.TxFatal
	}
}

func ConvertTxStatusToProto(txStatus evmtypes.TransactionStatus) TxStatus {
	switch txStatus {
	case evmtypes.TxSuccess:
		return TxStatus_TX_SUCCESS
	case evmtypes.TxReverted:
		return TxStatus_TX_REVERTED
	default:
		return TxStatus_TX_FATAL
	}
}

func ConvertSubmitTransactionRequestFromProto(txRequest *SubmitTransactionRequest) (evmtypes.SubmitTransactionRequest, error) {
	if txRequest == nil {
		return evmtypes.SubmitTransactionRequest{}, fmt.Errorf("tx request can't be nil")
	}

	return evmtypes.SubmitTransactionRequest{
		To:        evmtypes.Address(txRequest.To),
		Data:      txRequest.Data,
		GasConfig: ConvertGasConfigFromProto(txRequest.GasConfig),
	}, nil
}

func ConvertAddressFromProto(protoAddress []byte) (evmtypes.Address, error) {
	if err := ValidateAddressBytes(protoAddress); err != nil {
		return evmtypes.Address{}, err
	}

	return evmtypes.Address(protoAddress), nil
}

func ConvertOptionalAddressFromProto(b []byte) (evmtypes.Address, error) {
	if len(b) == 0 {
		return evmtypes.Address{}, nil
	}

	return ConvertAddressFromProto(b)
}

func ConvertOptionalHashFromProto(b []byte) (evmtypes.Hash, error) {
	if len(b) == 0 {
		return evmtypes.Hash{}, nil
	}

	return ConvertHashFromProto(b)
}

func ConvertHashFromProto(b []byte) (evmtypes.Hash, error) {
	if err := validateHashBytes(b); err != nil {
		return evmtypes.Hash{}, err
	}
	return evmtypes.Hash(b), nil
}

func ValidateAddressBytes(b []byte) error {
	if b == nil {
		return fmt.Errorf("address can't be nil")
	}

	if len(b) != evmtypes.AddressLength {
		return fmt.Errorf("invalid address: got %d bytes, expected %d, value=%s", len(b), evmtypes.AddressLength, bytesToHex(b))
	}
	return nil
}

func validateHashBytes(b []byte) error {
	if b == nil {
		return fmt.Errorf("hash can't be nil")
	}

	if len(b) != evmtypes.HashLength {
		return fmt.Errorf("invalid hash: got %d bytes, expected %d, value=%s", len(b), evmtypes.HashLength, bytesToHex(b))
	}
	return nil
}
