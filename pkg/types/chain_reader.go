package types

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/status"
)

// Errors exposed to product plugins

type errCodecAndChainReader string

func (e errCodecAndChainReader) Error() string { return string(e) }

const (
	ErrInvalidType           = errCodecAndChainReader("invalid type")
	ErrFieldNotFound         = errCodecAndChainReader("field not found")
	ErrInvalidEncoding       = errCodecAndChainReader("invalid encoding")
	ErrWrongNumberOfElements = errCodecAndChainReader("wrong number of elements in slice")
	ErrNotASlice             = errCodecAndChainReader("element is not a slice")
	ErrUnknown               = errCodecAndChainReader("unknown error")
)

func UnwrapClientError(err error) error {
	if s, ok := status.FromError(err); ok {
		return errCodecAndChainReader(s.String())
	}
	return err
}

var ErrInvalidConfig = errors.New("invalid configuration")

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
