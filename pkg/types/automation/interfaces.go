package automation

import (
	"context"
	"io"
)

type UpkeepTypeGetter func(UpkeepIdentifier) UpkeepType
type WorkIDGenerator func(UpkeepIdentifier, Trigger) string

// UpkeepStateStore is the interface for managing upkeeps final state in a local store.
type UpkeepStateStore interface {
	UpkeepStateUpdater
	UpkeepStateReader
	Start(context.Context) error
	io.Closer
}

type Registry interface {
	CheckUpkeeps(ctx context.Context, keys ...UpkeepPayload) ([]CheckResult, error)
	Name() string
	Start(ctx context.Context) error
	Close() error
	HealthReport() map[string]error
}

type EventProvider interface {
	Name() string
	Start(_ context.Context) error
	Close() error
	Ready() error
	HealthReport() map[string]error
	GetLatestEvents(ctx context.Context) ([]TransmitEvent, error)
}

type LogRecoverer interface {
	RecoverableProvider
	GetProposalData(context.Context, CoordinatedBlockProposal) ([]byte, error)

	Start(context.Context) error
	io.Closer
}

// UpkeepStateReader is the interface for reading the current state of upkeeps.
type UpkeepStateReader interface {
	SelectByWorkIDs(ctx context.Context, workIDs ...string) ([]UpkeepState, error)
}

type Encoder interface {
	Encode(...CheckResult) ([]byte, error)
	Extract([]byte) ([]ReportedUpkeep, error)
}

type LogEventProvider interface {
	GetLatestPayloads(context.Context) ([]UpkeepPayload, error)
	Start(context.Context) error
	Close() error
}

type RecoverableProvider interface {
	GetRecoveryProposals(context.Context) ([]UpkeepPayload, error)
}

type TransmitEventProvider interface {
	GetLatestEvents(context.Context) ([]TransmitEvent, error)
}

type ConditionalUpkeepProvider interface {
	GetActiveUpkeeps(context.Context) ([]UpkeepPayload, error)
}

type PayloadBuilder interface {
	// Can get payloads for a subset of proposals along with an error
	BuildPayloads(context.Context, ...CoordinatedBlockProposal) ([]UpkeepPayload, error)
}

type Runnable interface {
	// Can get results for a subset of payloads along with an error
	CheckUpkeeps(context.Context, ...UpkeepPayload) ([]CheckResult, error)
}

type BlockSubscriber interface {
	// Subscribe provides an identifier integer, a new channel, and potentially an error
	Subscribe() (int, chan BlockHistory, error)
	// Unsubscribe requires an identifier integer and indicates the provided channel should be closed
	Unsubscribe(int) error
	Start(context.Context) error
	Close() error
}

type UpkeepStateUpdater interface {
	SetUpkeepState(context.Context, CheckResult, UpkeepState) error
}

type RetryQueue interface {
	// Enqueue adds new items to the queue
	Enqueue(items ...RetryRecord) error
	// Dequeue returns the next n items in the queue, considering retry time schedules
	Dequeue(n int) ([]UpkeepPayload, error)
}

type ProposalQueue interface {
	// Enqueue adds new items to the queue
	Enqueue(items ...CoordinatedBlockProposal) error
	// Dequeue returns the next n items in the queue, considering retry time schedules
	Dequeue(t UpkeepType, n int) ([]CoordinatedBlockProposal, error)
}

type ResultStore interface {
	Add(...CheckResult)
	Remove(...string)
	View() ([]CheckResult, error)
}

type Coordinator interface {
	PreProcess(_ context.Context, payloads []UpkeepPayload) ([]UpkeepPayload, error)

	Accept(ReportedUpkeep) bool
	ShouldTransmit(ReportedUpkeep) bool
	FilterResults([]CheckResult) ([]CheckResult, error)
	FilterProposals([]CoordinatedBlockProposal) ([]CoordinatedBlockProposal, error)
}

type MetadataStore interface {
	SetBlockHistory(BlockHistory)
	GetBlockHistory() BlockHistory

	AddProposals(proposals ...CoordinatedBlockProposal)
	ViewProposals(utype UpkeepType) []CoordinatedBlockProposal
	RemoveProposals(proposals ...CoordinatedBlockProposal)

	Start(context.Context) error
	Close() error
}

type Ratio interface {
	// OfInt should return n out of x such that n/x ~ r (ratio)
	OfInt(int) int
}
