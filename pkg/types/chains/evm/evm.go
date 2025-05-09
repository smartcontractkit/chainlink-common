package evm

import (
	"math/big"
	"time"
)

const AddressLength = 20
const HashLength = 32

// represents evm-style address 40-character hexadecimal string prefixed by 0x making it 42 characters total
// lower case, upper case or checksummed case (EIP-55) inputs are allowed.
// output is checksummed case
type Address = [AddressLength]byte

// represents evm-style hash, 64-character hexadecimal string prefixed by 0x making it 66 characters total
// input is case-insensetive
// output is lower case
type Hash = [HashLength]byte

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

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// {} or nil          matches any topic list
	// {{A}}              matches topic A in first position
	// {{}, {B}}          matches any topic in first position AND B in second position
	// {{A}, {B}}         matches topic A in first position AND B in second position
	// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
	Topics [][]Hash // filter log by event sigs and indexed args
}

// matches cache-filter
// this filter defines what logs should be cached
// cached logs can be retrieved with [types.EVMService.QueryLogsFromCache]
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
	Timestamp  uint64 // time in seconds
	Hash       Hash
	ParentHash Hash
	Number     *big.Int
}

type TransactionFee struct {
	TransactionFee *big.Int // Cost of transaction in wei
}
