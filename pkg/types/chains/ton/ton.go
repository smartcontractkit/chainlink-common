package ton

import (
	"context"
	"math/big"
	"time"
)

const AddressLength = 20

// represents evm-style address
type Address = [AddressLength]byte

type ExitCode = int32

type BOC = []byte

type BlockIDExt struct {
	Workchain int32
	Shard     int64
	SeqNo     uint32
}
type Block struct {
	GlobalID int32 // Represents EVM-equivalent ChainID for TVM
}

type Message struct {
	Mode      uint8  // TON send mode
	ToAddress string // TON address (raw or user-friendly)
	Amount    string // Amount in tons
	Bounce    bool   // Bounce flag
	Body      BOC    // BOC-encoded message body cell
	StateInit BOC    // BOC-encoded state init cell
}

type LPFilterQuery struct {
	ID          int64
	Name        string
	Address     string
	EventName   string
	EventTopic  uint64
	StartingSeq uint32
	Retention   time.Duration
}

type Log struct {
	ID         int64
	FilterID   int64
	SeqNo      uint32
	Address    string
	EventTopic uint64
	Data       []byte // raw BOC of the body cell
	ReceivedAt time.Time
	ExpiresAt  *time.Time
	Error      *string
}

type TransactionFee struct {
	TransactionFee *big.Int // Cost of transaction in NanoTONs
}

type Balance struct {
	Balance *big.Int // Balance in NanoTONs
}

type LiteClient interface {
	GetMasterchainInfo(ctx context.Context) (*BlockIDExt, error)
	GetBlockData(ctx context.Context, block *BlockIDExt) (*Block, error)
	GetAccountBalance(ctx context.Context, address string, block *BlockIDExt) (*Balance, error)
}
