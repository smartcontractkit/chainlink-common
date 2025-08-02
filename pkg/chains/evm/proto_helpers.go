package evm

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
)

var (
	ErrInvalidAddressLength = errors.New("invalid address length: expected 20 bytes")
	ErrInvalidHashLength    = errors.New("invalid hash length: expected 32 bytes")
)

func ConvertAddressesFromProto(addresses [][]byte) ([]evmtypes.Address, error) {
	if addresses == nil {
		return nil, nil
	}

	evmAddresses := make([]evmtypes.Address, 0, len(addresses))
	for i, address := range addresses {
		if len(address) != evmtypes.AddressLength {
			return nil, fmt.Errorf("address at index %d: %w (got %d bytes)", i, ErrInvalidAddressLength, len(address))
		}
		evmAddresses = append(evmAddresses, evmtypes.Address(address))
	}
	return evmAddresses, nil
}

func convertAddressesToProto(addresses []evmtypes.Address) [][]byte {
	if addresses == nil {
		return nil
	}
	protoAddresses := make([][]byte, 0, len(addresses))
	for _, address := range addresses {
		protoAddresses = append(protoAddresses, address[:])
	}
	return protoAddresses
}

func ConvertHashesFromProto(hashes [][]byte) ([]evmtypes.Hash, error) {
	if hashes == nil {
		return nil, nil
	}

	hashesList := make([]evmtypes.Hash, 0, len(hashes))
	for i, hash := range hashes {
		if len(hash) != evmtypes.HashLength {
			return nil, fmt.Errorf("hash at index %d: %w (got %d bytes)", i, ErrInvalidHashLength, len(hash))
		}
		hashesList = append(hashesList, evmtypes.Hash(hash))
	}
	return hashesList, nil
}

func convertHashesToProto(hashes []evmtypes.Hash) [][]byte {
	if hashes == nil {
		return nil
	}
	protoHashes := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		protoHashes = append(protoHashes, hash[:])
	}
	return protoHashes
}

func convertTopicsToProto(topics [][]evmtypes.Hash) []*Topics {
	if topics == nil {
		return nil
	}
	protoTopics := make([]*Topics, 0, len(topics))
	for _, topic := range topics {
		topicProto := &Topics{Topic: convertHashesToProto(topic)}
		protoTopics = append(protoTopics, topicProto)
	}
	return protoTopics
}

func ConvertHeaderToProto(h *evmtypes.Header) *Header {
	if h == nil {
		return nil
	}
	return &Header{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        h.Hash[:],
		ParentHash:  h.ParentHash[:],
	}
}

var errEmptyHead = errors.New("head is nil")

func ConvertHeaderFromProto(header *Header) (*evmtypes.Header, error) {
	if header == nil {
		return nil, errEmptyHead
	}

	hashBytes := header.GetHash()
	if hashBytes == nil {
		return nil, errors.New("header hash cannot be nil")
	}
	hash, err := ConvertHashFromProto(hashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hash: %w", err)
	}
	// Header hash must not be empty
	if len(hash) == 0 {
		return nil, errors.New("header hash cannot be empty")
	}

	parentHashBytes := header.GetParentHash()
	if parentHashBytes == nil {
		return nil, errors.New("header parent hash cannot be nil")
	}
	parentHash, err := ConvertHashFromProto(parentHashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert parent hash: %w", err)
	}
	// Parent hash must not be empty
	if len(parentHash) == 0 {
		return nil, errors.New("header parent hash cannot be empty")
	}

	blockNumber := header.GetBlockNumber()
	if blockNumber == nil {
		return nil, errors.New("header block number cannot be nil")
	}

	return &evmtypes.Header{
		Timestamp:  header.GetTimestamp(),
		Hash:       hash,
		ParentHash: parentHash,
		Number:     valuespb.NewIntFromBigInt(blockNumber),
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

	logs, err := ConvertLogsFromProto(protoReceipt.GetLogs())
	if err != nil {
		return nil, fmt.Errorf("failed to convert receipt logs: %w", err)
	}

	txHashBytes := protoReceipt.GetTxHash()
	if txHashBytes == nil {
		return nil, errors.New("receipt transaction hash cannot be nil")
	}
	txHash, err := ConvertHashFromProto(txHashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction hash: %w", err)
	}
	// Transaction hash must not be empty
	if len(txHash) == 0 {
		return nil, errors.New("receipt transaction hash cannot be empty")
	}

	// Contract address can be empty for non-contract-creation transactions - convert directly
	var contractAddress evmtypes.Address
	contractBytes := protoReceipt.GetContractAddress()
	if contractBytes != nil {
		if len(contractBytes) != 0 && len(contractBytes) != evmtypes.AddressLength {
			return nil, fmt.Errorf("invalid contract address length: expected %d or 0, got %d", evmtypes.AddressLength, len(contractBytes))
		}
		contractAddress = evmtypes.Address(contractBytes)
	}

	blockHashBytes := protoReceipt.GetBlockHash()
	if blockHashBytes == nil {
		return nil, errors.New("receipt block hash cannot be nil")
	}
	blockHash, err := ConvertHashFromProto(blockHashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}
	// Block hash must not be empty
	if len(blockHash) == 0 {
		return nil, errors.New("receipt block hash cannot be empty")
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

	// Transaction 'to' can be empty for contract creation - convert directly
	var to evmtypes.Address
	toBytes := protoTx.GetTo()
	if toBytes != nil {
		if len(toBytes) != 0 && len(toBytes) != evmtypes.AddressLength {
			return nil, fmt.Errorf("invalid 'to' address length: expected %d or 0, got %d", evmtypes.AddressLength, len(toBytes))
		}
		to = evmtypes.Address(toBytes)
	}

	hashBytes := protoTx.GetHash()
	if hashBytes == nil {
		return nil, errors.New("transaction hash cannot be nil")
	}
	hash, err := ConvertHashFromProto(hashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction hash: %w", err)
	}
	// Transaction hash must not be empty
	if len(hash) == 0 {
		return nil, errors.New("transaction hash cannot be empty")
	}

	var data []byte
	if protoTx.GetData() != nil {
		data = protoTx.GetData()
	}

	return &evmtypes.Transaction{
		To:       to,
		Data:     data,
		Hash:     hash,
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

	// Both from and to can be empty in call contexts - convert directly
	var from, to evmtypes.Address

	fromBytes := protoMsg.GetFrom()
	if fromBytes != nil {
		if len(fromBytes) != 0 && len(fromBytes) != evmtypes.AddressLength {
			return nil, fmt.Errorf("invalid 'from' address length: expected %d or 0, got %d", evmtypes.AddressLength, len(fromBytes))
		}
		from = evmtypes.Address(fromBytes)
	}

	toBytes := protoMsg.GetTo()
	if toBytes != nil {
		if len(toBytes) != 0 && len(toBytes) != evmtypes.AddressLength {
			return nil, fmt.Errorf("invalid 'to' address length: expected %d or 0, got %d", evmtypes.AddressLength, len(toBytes))
		}
		to = evmtypes.Address(toBytes)
	}

	return &evmtypes.CallMsg{
		From: from,
		Data: protoMsg.GetData(),
		To:   to,
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

	addresses, err := ConvertAddressesFromProto(protoFilter.GetAddresses())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert filter addresses: %w", err)
	}

	eventSigs, err := ConvertHashesFromProto(protoFilter.GetEventSigs())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert event signatures: %w", err)
	}

	topic2, err := ConvertHashesFromProto(protoFilter.GetTopic2())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic2: %w", err)
	}

	topic3, err := ConvertHashesFromProto(protoFilter.GetTopic3())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic3: %w", err)
	}

	topic4, err := ConvertHashesFromProto(protoFilter.GetTopic4())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic4: %w", err)
	}

	return evmtypes.LPFilterQuery{
		Name:         protoFilter.GetName(),
		Retention:    time.Duration(protoFilter.GetRetentionTime()),
		Addresses:    addresses,
		EventSigs:    eventSigs,
		Topic2:       topic2,
		Topic3:       topic3,
		Topic4:       topic4,
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
	if logs == nil {
		return nil
	}
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

	// Block hash can be empty in filters - convert directly
	var blockHash evmtypes.Hash
	blockHashBytes := protoFilter.GetBlockHash()
	if blockHashBytes != nil {
		if len(blockHashBytes) != 0 && len(blockHashBytes) != evmtypes.HashLength {
			return evmtypes.FilterQuery{}, fmt.Errorf("invalid block hash length: expected %d or 0, got %d", evmtypes.HashLength, len(blockHashBytes))
		}
		blockHash = evmtypes.Hash(blockHashBytes)
	}

	addresses, err := ConvertAddressesFromProto(protoFilter.GetAddresses())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("failed to convert addresses: %w", err)
	}

	topics, err := ConvertTopicsFromProto(protoFilter.GetTopics())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("failed to convert topics: %w", err)
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
	if protoLogs == nil {
		return nil, nil
	}

	logs := make([]*evmtypes.Log, 0, len(protoLogs))
	for i, protoLog := range protoLogs {
		log, err := convertLogFromProto(protoLog)
		if err != nil {
			return nil, fmt.Errorf("failed to convert log at index %d: %w", i, err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func convertLogFromProto(protoLog *Log) (*evmtypes.Log, error) {
	if protoLog == nil {
		return nil, errors.New("proto log cannot be nil")
	}

	topics, err := ConvertHashesFromProto(protoLog.GetTopics())
	if err != nil {
		return nil, fmt.Errorf("failed to convert topics: %w", err)
	}

	blockHashBytes := protoLog.GetBlockHash()
	if blockHashBytes == nil {
		return nil, errors.New("log block hash cannot be nil")
	}
	blockHash, err := ConvertHashFromProto(blockHashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}
	// Log block hash must not be empty
	if len(blockHash) == 0 {
		return nil, errors.New("log block hash cannot be empty")
	}

	eventSig, err := ConvertHashFromProto(protoLog.GetEventSig())
	if err != nil {
		return nil, fmt.Errorf("failed to convert event signature: %w", err)
	}

	address, err := ConvertAddressFromProto(protoLog.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	txHashBytes := protoLog.GetTxHash()
	if txHashBytes == nil {
		return nil, errors.New("log transaction hash cannot be nil")
	}
	txHash, err := ConvertHashFromProto(txHashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction hash: %w", err)
	}
	// Log transaction hash must not be empty
	if len(txHash) == 0 {
		return nil, errors.New("log transaction hash cannot be empty")
	}

	return &evmtypes.Log{
		LogIndex:    protoLog.GetIndex(),
		BlockHash:   blockHash,
		BlockNumber: valuespb.NewIntFromBigInt(protoLog.GetBlockNumber()),
		Topics:      topics,
		EventSig:    eventSig,
		Address:     address,
		TxHash:      txHash,
		Data:        protoLog.GetData(),
		Removed:     protoLog.GetRemoved(),
	}, nil
}

func ConvertTopicsFromProto(protoTopics []*Topics) ([][]evmtypes.Hash, error) {
	if protoTopics == nil {
		return nil, nil
	}

	topics := make([][]evmtypes.Hash, 0, len(protoTopics))
	for i, topic := range protoTopics {
		if topic == nil {
			return nil, fmt.Errorf("topic at index %d cannot be nil", i)
		}

		hashes, err := ConvertHashesFromProto(topic.GetTopic())
		if err != nil {
			return nil, fmt.Errorf("failed to convert topic at index %d: %w", i, err)
		}
		topics = append(topics, hashes)
	}
	return topics, nil
}

func ConvertLogToProto(log *evmtypes.Log) *Log {
	if log == nil {
		return nil
	}

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
	if hashedValueComparators == nil {
		return nil
	}
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
	if protoHashedValueComparators == nil {
		return nil, nil
	}

	hashedValueComparators := make([]evmprimitives.HashedValueComparator, 0, len(protoHashedValueComparators))
	for i, protoHvc := range protoHashedValueComparators {
		if protoHvc == nil {
			return nil, fmt.Errorf("hashed value comparator at index %d cannot be nil", i)
		}

		values, err := ConvertHashesFromProto(protoHvc.GetValues())
		if err != nil {
			return nil, fmt.Errorf("failed to convert values for comparator at index %d: %w", i, err)
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
	if expressions == nil {
		return nil, nil
	}
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
	if protoExpressions == nil {
		return nil, nil
	}

	expressions := make([]query.Expression, 0, len(protoExpressions))
	for idx, protoExpression := range protoExpressions {
		if protoExpression == nil {
			return nil, fmt.Errorf("expression at index %d cannot be nil", idx)
		}

		expr, err := convertExpressionFromProto(protoExpression)
		if err != nil {
			return nil, fmt.Errorf("failed to convert expression at index %d: %w", idx, err)
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
				return nil, fmt.Errorf("failed to convert general primitive: %w", err)
			}
			putGeneralPrimitive(pbExpression, generalPrimitive)
		}
		return pbExpression, nil
	}

	if len(expression.BoolExpression.Expressions) == 0 {
		return nil, errors.New("boolean expression must have at least one sub-expression")
	}

	pbExpression.Evaluator = &Expression_BooleanExpression{BooleanExpression: &BooleanExpression{}}
	expressions := make([]*Expression, 0)
	for i, expr := range expression.BoolExpression.Expressions {
		pbExpr, err := convertExpressionToProto(expr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert sub-expression at index %d: %w", i, err)
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
		return query.Expression{}, errors.New("proto expression cannot be nil")
	}

	switch protoEvaluatedExpr := protoExpression.GetEvaluator().(type) {
	case *Expression_BooleanExpression:
		if protoEvaluatedExpr.BooleanExpression == nil {
			return query.Expression{}, errors.New("boolean expression cannot be nil")
		}

		var expressions []query.Expression
		for i, expression := range protoEvaluatedExpr.BooleanExpression.GetExpression() {
			convertedExpression, err := convertExpressionFromProto(expression)
			if err != nil {
				return query.Expression{}, fmt.Errorf("failed to convert sub-expression at index %d: %w", i, err)
			}
			expressions = append(expressions, convertedExpression)
		}
		if protoEvaluatedExpr.BooleanExpression.GetBooleanOperator() == chaincommonpb.BooleanOperator_AND {
			return query.And(expressions...), nil
		}
		return query.Or(expressions...), nil

	case *Expression_Primitive:
		if protoEvaluatedExpr.Primitive == nil {
			return query.Expression{}, errors.New("primitive expression cannot be nil")
		}

		switch primitive := protoEvaluatedExpr.Primitive.GetPrimitive().(type) {
		case *Primitive_GeneralPrimitive:
			return chaincommonpb.ConvertPrimitiveFromProto(primitive.GeneralPrimitive, func(_ string, _ bool) (any, error) {
				return nil, fmt.Errorf("unsupported primitive type: %T", primitive)
			})
		default:
			return convertEVMExpressionToProto(protoEvaluatedExpr.Primitive)
		}
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "unknown expression type: %T", protoExpression)
	}
}

func convertEVMExpressionToProto(protoPrimitive *Primitive) (query.Expression, error) {
	if protoPrimitive == nil {
		return query.Expression{}, errors.New("primitive cannot be nil")
	}

	switch primitive := protoPrimitive.GetPrimitive().(type) {
	case *Primitive_ContractAddress:
		address, err := ConvertAddressFromProto(primitive.ContractAddress)
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert contract address: %w", err)
		}
		return evmprimitives.NewAddressFilter(address), nil
	case *Primitive_EventSig:
		eventSig, err := ConvertHashFromProto(primitive.EventSig)
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert event signature: %w", err)
		}
		return evmprimitives.NewEventSigFilter(eventSig), nil
	case *Primitive_EventByTopic:
		if primitive.EventByTopic == nil {
			return query.Expression{}, errors.New("event by topic cannot be nil")
		}

		comparers, err := ConvertHashedValueComparatorsFromProto(primitive.EventByTopic.GetHashedValueComparers())
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert hashed value comparers: %w", err)
		}

		return evmprimitives.NewEventByTopicFilter(primitive.EventByTopic.GetTopic(), comparers), nil
	case *Primitive_EventByWord:
		if primitive.EventByWord == nil {
			return query.Expression{}, errors.New("event by word cannot be nil")
		}

		comparers, err := ConvertHashedValueComparatorsFromProto(primitive.EventByWord.GetHashedValueComparers())
		if err != nil {
			return query.Expression{}, fmt.Errorf("failed to convert hashed value comparers: %w", err)
		}

		return evmprimitives.NewEventByWordFilter(int(primitive.EventByWord.GetWordIndex()), comparers), nil
	default:
		return query.Expression{}, status.Errorf(codes.InvalidArgument, "unknown primitive type: %T", primitive)
	}
}

func putGeneralPrimitive(exp *Expression, p *chaincommonpb.Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: &Primitive_GeneralPrimitive{GeneralPrimitive: p}}}
}

func putEVMPrimitive(exp *Expression, p *Primitive) {
	exp.Evaluator = &Expression_Primitive{Primitive: &Primitive{Primitive: p.Primitive}}
}

func ConvertGasConfigToProto(gasConfig *evmtypes.GasConfig) *GasConfig {
	if gasConfig == nil {
		return nil
	}
	return &GasConfig{
		GasLimit: *gasConfig.GasLimit,
	}
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
		return evmtypes.SubmitTransactionRequest{}, errors.New("transaction request cannot be nil")
	}

	// 'to' can be empty for contract creation transactions - convert directly
	var to evmtypes.Address
	toBytes := txRequest.To
	if toBytes != nil {
		if len(toBytes) != 0 && len(toBytes) != evmtypes.AddressLength {
			return evmtypes.SubmitTransactionRequest{}, fmt.Errorf("invalid 'to' address length: expected %d or 0, got %d", evmtypes.AddressLength, len(toBytes))
		}
		to = evmtypes.Address(toBytes)
	}

	return evmtypes.SubmitTransactionRequest{
		To:        to,
		Data:      evmtypes.ABIPayload(txRequest.Data),
		GasConfig: ConvertGasConfigFromProto(txRequest.GasConfig),
	}, nil
}

func ConvertAddressFromProto(b []byte) (evmtypes.Address, error) {
	if b == nil {
		return evmtypes.Address{}, errors.New("address bytes cannot be nil")
	}
	if len(b) == 0 {
		return evmtypes.Address{}, errors.New("address bytes cannot be empty")
	}
	if len(b) != evmtypes.AddressLength {
		return evmtypes.Address{}, fmt.Errorf("invalid address length: expected %d, got %d", evmtypes.AddressLength, len(b))
	}

	return evmtypes.Address(b), nil
}

func ConvertHashFromProto(b []byte) (evmtypes.Hash, error) {
	if b == nil {
		return evmtypes.Hash{}, errors.New("hash bytes cannot be nil")
	}
	if len(b) == 0 {
		return evmtypes.Hash{}, errors.New("hash bytes cannot be empty")
	}
	if len(b) != evmtypes.HashLength {
		return evmtypes.Hash{}, fmt.Errorf("invalid hash length: expected %d, got %d", evmtypes.HashLength, len(b))
	}

	return evmtypes.Hash(b), nil
}
