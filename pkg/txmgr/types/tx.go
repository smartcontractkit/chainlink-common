package types

import (
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gopkg.in/guregu/null.v4"

	feetypes "github.com/smartcontractkit/chainlink-relay/pkg/fee/types"
	clnull "github.com/smartcontractkit/chainlink-relay/pkg/null"
	"github.com/smartcontractkit/chainlink-relay/pkg/pg/datatypes"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type TxAttemptState int8

type TxState string

const (
	TxAttemptInProgress TxAttemptState = iota + 1
	TxAttemptInsufficientFunds
	TxAttemptBroadcast
	txAttemptStateCount // always at end to calculate number of states
)

var txAttemptStateStrings = []string{
	"unknown_attempt_state",    // default 0 value
	TxAttemptInProgress:        "in_progress",
	TxAttemptInsufficientFunds: "insufficient_funds",
	TxAttemptBroadcast:         "broadcast",
}

func NewTxAttemptState(state string) (s TxAttemptState) {
	if index := slices.Index(txAttemptStateStrings, state); index != -1 {
		s = TxAttemptState(index)
	}
	return s
}

// String returns string formatted states for logging
func (s TxAttemptState) String() (str string) {
	if s < txAttemptStateCount {
		return txAttemptStateStrings[s]
	}
	return txAttemptStateStrings[0]
}

type TxAttempt[
	CHAIN_ID types.ID,
	ADDR types.Hashable,
	TX_HASH, BLOCK_HASH types.Hashable,
	SEQ types.Sequence,
	FEE feetypes.Fee,
] struct {
	ID    int64
	TxID  int64
	Tx    Tx[CHAIN_ID, ADDR, TX_HASH, BLOCK_HASH, SEQ, FEE]
	TxFee FEE
	// ChainSpecificFeeLimit on the TxAttempt is always the same as the on-chain encoded value for fee limit
	ChainSpecificFeeLimit   uint32
	SignedRawTx             []byte
	Hash                    TX_HASH
	CreatedAt               time.Time
	BroadcastBeforeBlockNum *int64
	State                   TxAttemptState
	Receipts                []ChainReceipt[TX_HASH, BLOCK_HASH] `json:"-"`
	TxType                  int
}

func (a *TxAttempt[CHAIN_ID, ADDR, TX_HASH, BLOCK_HASH, SEQ, FEE]) String() string {
	return fmt.Sprintf("TxAttempt(ID:%d,TxID:%d,Fee:%s,TxType:%d", a.ID, a.TxID, a.TxFee, a.TxType)
}

type Tx[
	CHAIN_ID types.ID,
	ADDR types.Hashable,
	TX_HASH, BLOCK_HASH types.Hashable,
	SEQ types.Sequence,
	FEE feetypes.Fee,
] struct {
	ID             int64
	Sequence       *SEQ
	FromAddress    ADDR
	ToAddress      ADDR
	EncodedPayload []byte
	Value          big.Int
	// FeeLimit on the Tx is always the conceptual gas limit, which is not
	// necessarily the same as the on-chain encoded value (i.e. Optimism)
	FeeLimit uint32
	Error    null.String
	// BroadcastAt is updated every time an attempt for this tx is re-sent
	// In almost all cases it will be within a second or so of the actual send time.
	BroadcastAt *time.Time
	// InitialBroadcastAt is recorded once, the first ever time this tx is sent
	InitialBroadcastAt *time.Time
	CreatedAt          time.Time
	State              TxState
	TxAttempts         []TxAttempt[CHAIN_ID, ADDR, TX_HASH, BLOCK_HASH, SEQ, FEE] `json:"-"`
	// Marshalled TxMeta
	// Used for additional context around transactions which you want to log
	// at send time.
	Meta    *datatypes.JSON
	Subject uuid.NullUUID
	ChainID CHAIN_ID

	PipelineTaskRunID uuid.NullUUID
	MinConfirmations  clnull.Uint32

	// TransmitChecker defines the check that should be performed before a transaction is submitted on
	// chain.
	TransmitChecker *datatypes.JSON
}
