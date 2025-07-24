package evm

import (
	"errors"
	"fmt"
	"time"

	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
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
		if address == nil {
			return nil, fmt.Errorf("address at index %d cannot be nil", i)
		}
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
		if hash == nil {
			return nil, fmt.Errorf("hash at index %d cannot be nil", i)
		}
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
		topicProto := convertHashesToProto(topic)
		protoTopics = append(protoTopics, &Topics{Topic: topicProto})
	}
	return protoTopics
}

func ConvertHeaderToProto(h evmtypes.Header) *Header {
	return &Header{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        h.Hash[:],
		ParentHash:  h.ParentHash[:],
	}
}

func ConvertHeadFromProto(head *Header) (evmtypes.Header, error) {
	if head == nil {
		return evmtypes.Header{}, fmt.Errorf("head is nil")
	}

	hashBytes := head.GetHash()
	if hashBytes == nil {
		return evmtypes.Header{}, errors.New("header hash cannot be nil")
	}
	hash, err := ConvertHashFromProto(hashBytes)
	if err != nil {
		return evmtypes.Header{}, fmt.Errorf("failed to convert hash: %w", err)
	}
	// Header hash must not be empty
	if len(hash) == 0 {
		return evmtypes.Header{}, errors.New("header hash cannot be empty")
	}

	parentHashBytes := head.GetParentHash()
	if parentHashBytes == nil {
		return evmtypes.Header{}, errors.New("header parent hash cannot be nil")
	}

	parentHash, err := ConvertHashFromProto(parentHashBytes)
	if err != nil {
		return evmtypes.Header{}, fmt.Errorf("failed to convert parent hash: %w", err)
	}
	// Parent hash must not be empty
	if len(parentHash) == 0 {
		return evmtypes.Header{}, errors.New("header parent hash cannot be empty")
	}

	blockNumber := head.GetBlockNumber()
	if blockNumber == nil {
		return evmtypes.Header{}, errors.New("header block number cannot be nil")
	}

	return evmtypes.Header{
		Timestamp:  head.GetTimestamp(),
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

func ConvertLogsToProto(logs []*evmtypes.Log) ([]*Log, error) {
	if logs == nil {
		return nil, fmt.Errorf("logs cannot be nil")
	}
	protoLogs := make([]*Log, 0, len(logs))
	for _, l := range logs {
		protoLogs = append(protoLogs, ConvertLogToProto(l))
	}
	return protoLogs, nil
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

	topics := make([][]byte, len(log.Topics))
	for i, topic := range log.Topics {
		topics[i] = topic[:]
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

func ConvertHashFromProto(protoHash []byte) (evmtypes.Hash, error) {
	if protoHash == nil {
		return evmtypes.Hash{}, errors.New("hash bytes cannot be nil")
	}
	if len(protoHash) == 0 {
		return evmtypes.Hash{}, nil
	}
	if len(protoHash) != evmtypes.HashLength {
		return evmtypes.Hash{}, fmt.Errorf("invalid hash length: expected %d, got %d", evmtypes.HashLength, len(protoHash))
	}
	return evmtypes.Hash(protoHash), nil
}

func ConvertAddressFromProto(protoAddress []byte) (evmtypes.Address, error) {
	if protoAddress == nil {
		return evmtypes.Address{}, errors.New("address bytes cannot be nil")
	}
	if len(protoAddress) == 0 {
		return evmtypes.Address{}, nil
	}
	if len(protoAddress) != evmtypes.AddressLength {
		return evmtypes.Address{}, fmt.Errorf("invalid address length: expected %d, got %d", evmtypes.AddressLength, len(protoAddress))
	}
	return evmtypes.Address(protoAddress), nil
}
