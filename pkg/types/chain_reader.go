package types

import (
	"context"
	"time"
)

// Errors exposed to product plugins

type InvalidTypeError struct{}

func (InvalidTypeError) Error() string {
	return "invalid type"
}

type FieldNotFoundError struct{}

func (FieldNotFoundError) Error() string {
	return "field not found"
}

// Errors used only by relay plugins

type ErrorChainReaderUnsupported struct{}

func (e ErrorChainReaderUnsupported) Error() string {
	return "ChainReader is not supported by the relay"
}

type ErrorNoChainReaderInJobSpec struct{}

func (e ErrorNoChainReaderInJobSpec) Error() string {
	return "There is no ChainReader configuration defined in the job spec"
}

type ErrorChainReaderInvalidConfig struct{}

func (e ErrorChainReaderInvalidConfig) Error() string {
	return "Invalid ChainReader configuration"
}

type ChainReader interface {
	// returnVal should satisfy Marshaller interface
	GetLatestValue(ctx context.Context, bc BoundContract, method string, params, returnVal any) error
}

type BoundContract struct {
	Address string
	Name    string
	Pending bool
}

type Event struct {
	ChainID           string
	EventIndexInBlock string
	Address           string
	TxHash            string
	BlockHash         string
	BlockNumber       int64
	BlockTimestamp    time.Time
	CreatedAt         time.Time
	Keys              []string
	Data              []byte
}

type EventFilter struct {
	AddressList []string   // contract address
	KeysList    [][]string // 2D list of indexed search keys, outer dim = AND, inner dim = OR.  Params[0] is the name of the event (or "event type"), rest are any narrowing parameters that may help further specify the event
	Retention   time.Duration
}
