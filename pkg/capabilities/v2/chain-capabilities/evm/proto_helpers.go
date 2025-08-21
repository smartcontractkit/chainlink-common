package evm

import (
	"errors"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

func ConvertHeaderToProto(h *evmtypes.Header) (*Header, error) {
	if h == nil {
		return nil, evm.ErrEmptyHead
	}

	return &Header{
		Timestamp:   h.Timestamp,
		BlockNumber: valuespb.NewBigIntFromInt(h.Number),
		Hash:        h.Hash[:],
		ParentHash:  h.ParentHash[:],
	}, nil
}

func ConvertHeaderFromProto(protoHeader *Header) (evmtypes.Header, error) {
	if protoHeader == nil {
		return evmtypes.Header{}, evm.ErrEmptyHead
	}

	hash, err := evm.ConvertHashFromProto(protoHeader.GetHash())
	if err != nil {
		return evmtypes.Header{}, fmt.Errorf("failed to convert hash: %w", err)
	}

	parentHash, err := evm.ConvertHashFromProto(protoHeader.GetParentHash())
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

func ConvertReceiptToProto(receipt *evmtypes.Receipt) (*Receipt, error) {
	if receipt == nil {
		return nil, evm.ErrEmptyReceipt
	}

	logs, err := ConvertLogsToProto(receipt.Logs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert logs: %w", err)
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
		return nil, evm.ErrEmptyReceipt
	}

	logs, err := ConvertLogsFromProto(protoReceipt.GetLogs())
	if err != nil {
		return nil, err
	}

	txHash, err := evm.ConvertHashFromProto(protoReceipt.GetTxHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	// can be empty on contract creation
	contractAddress, err := evm.ConvertOptionalAddressFromProto(protoReceipt.GetContractAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to convert contract address: %w", err)
	}

	blockHash, err := evm.ConvertHashFromProto(protoReceipt.GetBlockHash())
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

func ConvertTransactionToProto(tx *evmtypes.Transaction) (*Transaction, error) {
	if tx == nil {
		return nil, evm.ErrEmptyTx
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
		return nil, evm.ErrEmptyTx
	}

	var data []byte
	if protoTx.GetData() != nil {
		data = protoTx.GetData()
	}

	toAddress, err := evm.ConvertOptionalAddressFromProto(protoTx.GetTo())
	if err != nil {
		return nil, fmt.Errorf("failed to convert 'to' address: %w", err)
	}

	txHash, err := evm.ConvertHashFromProto(protoTx.GetHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert tx hash: %w", err)
	}

	return &evmtypes.Transaction{
		To:       toAddress,
		Data:     data,
		Hash:     txHash,
		Nonce:    protoTx.GetNonce(),
		Gas:      protoTx.GetGas(),
		GasPrice: valuespb.NewIntFromBigInt(protoTx.GetGasPrice()),
		Value:    valuespb.NewIntFromBigInt(protoTx.GetValue()),
	}, nil
}

func ConvertCallMsgToProto(msg *evmtypes.CallMsg) (*CallMsg, error) {
	if msg == nil {
		return nil, evm.ErrEmptyMsg
	}

	return &CallMsg{
		From: msg.From[:],
		To:   msg.To[:],
		Data: msg.Data,
	}, nil
}

func ConvertCallMsgFromProto(protoMsg *CallMsg) (*evmtypes.CallMsg, error) {
	if protoMsg == nil {
		return nil, evm.ErrEmptyMsg
	}

	fromAddress, err := evm.ConvertAddressFromProto(protoMsg.GetFrom())
	if err != nil {
		return nil, fmt.Errorf("failed to convert 'from' address: %w", err)
	}

	toAddress, err := evm.ConvertOptionalAddressFromProto(protoMsg.GetTo())
	if err != nil {
		return nil, fmt.Errorf("failed to convert 'to' address: %w", err)
	}

	return &evmtypes.CallMsg{
		From: fromAddress,
		Data: protoMsg.GetData(),
		To:   toAddress,
	}, nil
}

func ConvertLPFilterToProto(filter evmtypes.LPFilterQuery) *LPFilter {
	return &LPFilter{
		Name:          filter.Name,
		RetentionTime: int64(filter.Retention),
		Addresses:     evm.ConvertAddressesToProto(filter.Addresses),
		EventSigs:     evm.ConvertHashesToProto(filter.EventSigs),
		Topic2:        evm.ConvertHashesToProto(filter.Topic2),
		Topic3:        evm.ConvertHashesToProto(filter.Topic3),
		Topic4:        evm.ConvertHashesToProto(filter.Topic4),
		MaxLogsKept:   filter.MaxLogsKept,
		LogsPerBlock:  filter.LogsPerBlock,
	}
}

func ConvertLPFilterFromProto(protoFilter *LPFilter) (evmtypes.LPFilterQuery, error) {
	if protoFilter == nil {
		return evmtypes.LPFilterQuery{}, evm.ErrEmptyFilter
	}

	var addresses []evmtypes.Address
	for i, protoAddress := range protoFilter.GetAddresses() {
		address, err := evm.ConvertOptionalAddressFromProto(protoAddress)
		if err != nil {
			return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert address[%d]: %w", i, err)
		}
		addresses = append(addresses, address)
	}

	sigs, err := evm.ConvertHashesFromProto(protoFilter.GetEventSigs())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert event sigs: %w", err)
	}

	t2, err := evm.ConvertHashesFromProto(protoFilter.GetTopic2())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic2: %w", err)
	}

	t3, err := evm.ConvertHashesFromProto(protoFilter.GetTopic3())
	if err != nil {
		return evmtypes.LPFilterQuery{}, fmt.Errorf("failed to convert topic3: %w", err)
	}

	t4, err := evm.ConvertHashesFromProto(protoFilter.GetTopic4())
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

func ConvertFilterToProto(filter evmtypes.FilterQuery) (*FilterQuery, error) {
	topics, err := convertTopicsToProto(filter.Topics)
	if err != nil {
		return nil, errors.Join(evm.ErrTopicsConversion, err)
	}

	return &FilterQuery{
		BlockHash: filter.BlockHash[:],
		FromBlock: valuespb.NewBigIntFromInt(filter.FromBlock),
		ToBlock:   valuespb.NewBigIntFromInt(filter.ToBlock),
		Addresses: evm.ConvertAddressesToProto(filter.Addresses),
		Topics:    topics,
	}, nil
}

func ConvertLogsToProto(logs []*evmtypes.Log) ([]*Log, error) {
	if logs == nil {
		return nil, fmt.Errorf("logs are nil")
	}

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
		return evmtypes.FilterQuery{}, evm.ErrEmptyFilter
	}

	blockHash, err := evm.ConvertOptionalHashFromProto(protoFilter.GetBlockHash())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("failed to convert blockHash: %w", err)
	}

	addresses, err := evm.ConvertAddressesFromProto(protoFilter.GetAddresses())
	if err != nil {
		return evmtypes.FilterQuery{}, fmt.Errorf("failed to convert addresses: %w", err)
	}

	topics, err := ConvertTopicsFromProto(protoFilter.GetTopics())
	if err != nil {
		return evmtypes.FilterQuery{}, errors.Join(evm.ErrTopicsConversion, err)
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
		return nil, fmt.Errorf("logs can't be nil")
	}

	logs := make([]*evmtypes.Log, 0, len(protoLogs))
	for i, protoLog := range protoLogs {
		if protoLog == nil {
			return nil, fmt.Errorf("log at index %d can't be nil", i)
		}

		l, err := convertLogFromProto(protoLog)
		if err != nil {
			return nil, fmt.Errorf("failed to convert log at index %d: %w", i, err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func ConvertTopicsFromProto(protoTopics []*Topics) ([][]evmtypes.Hash, error) {
	if protoTopics == nil {
		return nil, fmt.Errorf("topics can't be nil")
	}

	topics := make([][]evmtypes.Hash, 0, len(protoTopics))
	for i, protoTopic := range protoTopics {
		if protoTopic == nil {
			return nil, fmt.Errorf("topic[%d] can't be nil", i)
		}

		hashes, err := evm.ConvertHashesFromProto(protoTopic.GetTopic())
		if err != nil {
			return nil, fmt.Errorf("failed to convert topics[%d]: %w", i, err)
		}

		topics = append(topics, hashes)
	}
	return topics, nil
}

func ConvertLogToProto(log evmtypes.Log) *Log {
	return &Log{
		Index:       log.LogIndex,
		BlockHash:   log.BlockHash[:],
		BlockNumber: valuespb.NewBigIntFromInt(log.BlockNumber),
		Topics:      evm.ConvertHashesToProto(log.Topics),
		EventSig:    log.EventSig[:],
		Address:     log.Address[:],
		TxHash:      log.TxHash[:],
		Data:        log.Data[:],
		// TODO tx index
		//TxIndex: log.TxIndex
		Removed: log.Removed,
	}
}

func convertTopicsToProto(topics [][]evmtypes.Hash) ([]*Topics, error) {
	if topics == nil {
		return nil, fmt.Errorf("topics can't be nil")
	}

	protoTopics := make([]*Topics, 0, len(topics))
	for i, topic := range topics {
		if topic == nil {
			return nil, fmt.Errorf("topic[%d] can't be nil", i)
		}

		protoTopics = append(protoTopics, &Topics{Topic: evm.ConvertHashesToProto(topic)})
	}
	return protoTopics, nil
}

func convertLogFromProto(protoLog *Log) (*evmtypes.Log, error) {
	if protoLog == nil {
		return nil, fmt.Errorf("log can't be nil")
	}

	blockHash, err := evm.ConvertHashFromProto(protoLog.GetBlockHash())
	if err != nil {
		return nil, fmt.Errorf("failed to convert block hash: %w", err)
	}

	topics, err := evm.ConvertHashesFromProto(protoLog.GetTopics())
	if err != nil {
		return nil, errors.Join(evm.ErrTopicsConversion, err)
	}

	eventSigs, err := evm.ConvertHashFromProto(protoLog.GetEventSig())
	if err != nil {
		return nil, fmt.Errorf("failed to convert event sig: %w", err)
	}

	address, err := evm.ConvertAddressFromProto(protoLog.GetAddress())
	if err != nil {
		return nil, err
	}

	txHash, err := evm.ConvertHashFromProto(protoLog.GetTxHash())
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
		// TODO TxIndex
	}, nil
}
