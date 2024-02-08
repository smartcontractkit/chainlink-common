package ccip

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Address string

func (a Address) Equals(addr2 Address) bool {
	if common.IsHexAddress(string(a)) && common.IsHexAddress(string(addr2)) {
		return common.HexToAddress(string(a)) == common.HexToAddress(string(addr2))
	}
	return a == addr2
}

type Hash [32]byte

type BlockMeta struct {
	BlockTimestamp time.Time
	BlockNumber    int64
	TxHash         string
	LogIndex       uint
}
