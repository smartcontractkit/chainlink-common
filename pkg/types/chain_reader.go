package types

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type chainReaderError string

func (e chainReaderError) Error() string { return string(e) }

const (
	ErrInvalidType   = chainReaderError("invalid type")
	ErrFieldNotFound = chainReaderError("field not found")
	ErrInvalidConfig = chainReaderError("invalid configuration")
)

func UnwrapClientError(err error) error {
	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Unimplemented {
			return errors.ErrUnsupported
		}
		return chainReaderError(s.String())
	}
	return err
}

type ChainReader interface {
	// The params argument of GetLatestValue() can be any object which maps a set of generic parameters into chain specific parameters defined in RelayConfig. It must be a pointer which can be passed into
	// json.Marshal(). Typically would be either an anonymous map such as `map[string]any{"baz": 42, "test": true}}` or something which implements the `MarshalJSON()` method (satisfying `Marshaller` interface).
	//
	// returnVal should be a pointer which can be passed to json.Marshal()
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
