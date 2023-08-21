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
	"github.com/smartcontractkit/chainlink-relay/pkg/services/datatypes"
	"github.com/smartcontractkit/chainlink-relay/pkg/services/pg"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type TxAttemptState int8

type TxState string

// TxStrategy controls how txes are queued and sent
//
//go:generate mockery --quiet --name TxStrategy --output ./mocks/ --case=underscore --structname TxStrategy --filename tx_strategy.go
type TxStrategy interface {
	// Subject will be saved txes.subject if not null
	Subject() uuid.NullUUID
	// PruneQueue is called after tx insertion
	// It accepts the service responsible for deleting
	// unstarted txs and deletion options
	PruneQueue(pruneService UnstartedTxQueuePruner, qopt pg.QOpt) (n int64, err error)
}

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

type TxRequest[ADDR types.Hashable, TX_HASH types.Hashable] struct {
	FromAddress      ADDR
	ToAddress        ADDR
	EncodedPayload   []byte
	Value            big.Int
	FeeLimit         uint32
	Meta             *TxMeta[ADDR, TX_HASH]
	ForwarderAddress ADDR

	// Pipeline variables - if you aren't calling this from chain tx task within
	// the pipeline, you don't need these variables
	MinConfirmations  clnull.Uint32
	PipelineTaskRunID *uuid.UUID

	Strategy TxStrategy

	// Checker defines the check that should be run before a transaction is submitted on chain.
	Checker TransmitCheckerSpec[ADDR]
}

// TransmitCheckerSpec defines the check that should be performed before a transaction is submitted
// on chain.
type TransmitCheckerSpec[ADDR types.Hashable] struct {
	// CheckerType is the type of check that should be performed. Empty indicates no check.
	CheckerType TransmitCheckerType `json:",omitempty"`

	// VRFCoordinatorAddress is the address of the VRF coordinator that should be used to perform
	// VRF transmit checks. This should be set iff CheckerType is TransmitCheckerTypeVRFV2.
	VRFCoordinatorAddress *ADDR `json:",omitempty"`

	// VRFRequestBlockNumber is the block number in which the provided VRF request has been made.
	// This should be set iff CheckerType is TransmitCheckerTypeVRFV2.
	VRFRequestBlockNumber *big.Int `json:",omitempty"`
}

// TransmitCheckerType describes the type of check that should be performed before a transaction is
// executed on-chain.
type TransmitCheckerType string

// TxMeta contains fields of the transaction metadata
// Not all fields are guaranteed to be present
type TxMeta[ADDR types.Hashable, TX_HASH types.Hashable] struct {
	JobID *int32 `json:"JobID,omitempty"`

	// Pipeline fields
	FailOnRevert null.Bool `json:"FailOnRevert,omitempty"`

	// VRF-only fields
	RequestID     *TX_HASH `json:"RequestID,omitempty"`
	RequestTxHash *TX_HASH `json:"RequestTxHash,omitempty"`
	// Batch variants of the above
	RequestIDs      []TX_HASH `json:"RequestIDs,omitempty"`
	RequestTxHashes []TX_HASH `json:"RequestTxHashes,omitempty"`
	// Used for the VRFv2 - max link this tx will bill
	// should it get bumped
	MaxLink *string `json:"MaxLink,omitempty"`
	// Used for the VRFv2 - the subscription ID of the
	// requester of the VRF.
	SubID *uint64 `json:"SubId,omitempty"`
	// Used for the VRFv2Plus - the uint256 subscription ID of the
	// requester of the VRF.
	GlobalSubID *string `json:"GlobalSubId,omitempty"`
	// Used for VRFv2Plus - max native token this tx will bill
	// should it get bumped
	MaxEth *string `json:"MaxEth,omitempty"`

	// Used for keepers
	UpkeepID *string `json:"UpkeepID,omitempty"`

	// Used only for forwarded txs, tracks the original destination address.
	// When this is set, it indicates tx is forwarded through To address.
	FwdrDestAddress *ADDR `json:"ForwarderDestAddress,omitempty"`

	// MessageIDs is used by CCIP for tx to executed messages correlation in logs
	MessageIDs []string `json:"MessageIDs,omitempty"`
	// SeqNumbers is used by CCIP for tx to committed sequence numbers correlation in logs
	SeqNumbers []uint64 `json:"SeqNumbers,omitempty"`
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
