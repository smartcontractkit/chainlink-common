package evm

import (
	"math/big"
	"time"
)

// matches evm-style logs
type Log struct {
	LogIndex    uint32
	BlockHash   string
	BlockNumber *big.Int
	Topics      []string
	EventSig    string
	Address     string
	TxHash      string
	Data        []byte
	Removed     bool
}

// EVMFilter Query matches evm-style filterQuery
type EVMFilterQuery struct {
	BlockHash string
	FromBlock *big.Int
	ToBlock   *big.Int
	Addresses []string

	Topics []string
}

// matches LP-filter
type FilterQuery struct {
	Name         string
	Addresses    []string
	EventSigs    []string
	Topic2       []string
	Topic3       []string
	Topic4       []string
	Retention    time.Duration
	MaxLogsKept  uint64 // maximum number of logs to retain ( 0 = unlimited )
	LogsPerBlock uint64 // rate limit ( maximum # of logs per block, 0 = unlimited )
}

// matches simplifie evm-style callMsg for reads/EstimateGas
type CallMsg struct {
	To   string
	From string // from field is needed if contract read depends on msg.sender
	Data []byte
}

// matches evm-style transaction
type Transaction struct {
	To       string
	Data     []byte
	Hash     string
	Nonce    uint64
	Gas      uint64
	GasPrice *big.Int
	Value    *big.Int
}

// matches evm-style receipt
type Receipt struct {
	PostState         []byte
	Status            uint64
	Logs              []*Log
	TxHash            string
	ContractAddress   string
	GasUsed           uint64
	BlockHash         string
	BlockNumber       *big.Int
	TransactionIndex  uint64
	EffectiveGasPrice *big.Int
}

// matches simplified evm-style head
type Head struct {
	Timestamp  uint64 // time in nanoseconds
	Hash       string
	ParentHash string
	Number     *big.Int
}

// TransactionStatus are the status TXM  supports and that can be returned by Tx idempotency key.
type TransactionStatus int

const (
	Unknown TransactionStatus = iota
	Pending
	Unconfirmed
	Finalized
	Failed
	Fatal
)
