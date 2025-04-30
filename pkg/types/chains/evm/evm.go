package evm

import (
	"math/big"
	"time"
)

// represents evm-style address 40-character hexadecimal string prefixed by 0x making it 42 characters total
// lower case, upper case or checksummed case (EIP-55) are allowed.
type Address = string

// represents evm-style hash, 64-character hexadecimal string prefixed by 0x making it 66 characters total
// case-insensetive
type Hash = string

// represents solidity-spec abi encoded bytes
type ABIPayload = []byte

// matches evm-style logs
type Log struct {
	LogIndex    uint32     // index of the log inside of the block
	BlockHash   Hash       // hash of the block containing this log
	BlockNumber *big.Int   // number of the block containing this log
	Topics      []Hash     // indexed fields of the log
	EventSig    Hash       // keccak256 hash of log event signature
	Address     Address    // address of the contract that emmited the log
	TxHash      Hash       // hash of the transaction this log is produced by
	Data        ABIPayload // abi encoded data of the log
	Removed     bool       // flag if log was removed during reorg
}

// matches evm-style eth_getLogs filterQuery
type FilterQuery struct {
	BlockHash Hash      // for filter by exact block, if not empty can't use from/to
	FromBlock *big.Int  // start block range
	ToBlock   *big.Int  // end block range
	Addresses []Address // contract(s) to filter logs from

	Topics []Hash // filter log by event sigs and indexed args
}

// matches LP-filter
type LPFilterQuery struct {
	Name         string        // filter identifier, used to remove filter
	Addresses    []Address     // list of addresses to include
	EventSigs    []Hash        // list of possible signatures
	Topic2       []Hash        // list of possible values for topic2
	Topic3       []Hash        // list of possible values for topic3
	Topic4       []Hash        // list of possible values for topic3
	Retention    time.Duration // maximum amount of time to retain
	MaxLogsKept  uint64        // maximum number of logs to retain ( 0 = unlimited )
	LogsPerBlock uint64        // rate limit ( maximum # of logs per block, 0 = unlimited )
}

// matches simplifie evm-style callMsg for reads/EstimateGas
type CallMsg struct {
	To   Address
	From Address // from field is needed if contract read depends on msg.sender
	Data ABIPayload
}

// matches evm-style transaction
type Transaction struct {
	To       Address    // receipient address
	Data     ABIPayload // input data for func call payload
	Hash     Hash       // derived from transaction structure
	Nonce    uint64     // number of txs sent from sender
	Gas      uint64     // max gas allowed per execution (in gas units)
	GasPrice *big.Int   // price for a single gas unit in wei
	Value    *big.Int   // amount of eth sent in wei
}

// matches evm-style receipt
type Receipt struct {
	Status            uint64   // 1 for success 0 for revert
	Logs              []*Log   // logs emmited by the transaction
	TxHash            Hash     // hash of the transaction this receipt is for
	ContractAddress   Address  // Address of the contract if one was created by this transaction
	GasUsed           uint64   // actual gas used during execution in gas units
	BlockHash         Hash     // hash of the block containing this receipt
	BlockNumber       *big.Int // number of the block containing this receipt
	TransactionIndex  uint64   // index of the transaction inside of the block
	EffectiveGasPrice *big.Int // actual price in wei paid per gas unit
}

// matches simplified evm-style head
type Head struct {
	Timestamp  uint64 // time in nanoseconds
	Hash       Hash
	ParentHash Hash
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
