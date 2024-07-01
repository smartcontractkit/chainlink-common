package types

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
)

// Errors exposed to product plugins
const (
	ErrInvalidType              = InvalidArgumentError("invalid type")
	ErrInvalidConfig            = InvalidArgumentError("invalid configuration")
	ErrChainReaderConfigMissing = UnimplementedError("ChainReader entry missing from RelayConfig")
	ErrInternal                 = InternalError("internal error")
	ErrNotFound                 = NotFoundError("not found")
)

// Deprecated: Use ContractReader. New naming should clear up confusion around the usage of this interface which should strictly be contract reading related.
type ChainReader = ContractReader

// ContractReader defines essential read operations a chain should implement for reading contract values and events.
type ContractReader interface {
	services.Service
	// GetLatestValue gets the latest value....
	// The params argument can be any object which maps a set of generic parameters into chain specific parameters defined in RelayConfig.
	// It must encode as an object via [json.Marshal] and [github.com/fxamacker/cbor/v2.Marshal].
	// Typically, would be either a struct with field names mapping to arguments, or anonymous map such as `map[string]any{"baz": 42, "test": true}}`
	//
	// returnVal must [json.Unmarshal] and and [github.com/fxamacker/cbor/v2.Marshal] as an object.
	//
	// Example use:
	//  type ProductParams struct {
	// 		ID int `json:"id"`
	//  }
	//  type ProductReturn struct {
	// 		Foo string `json:"foo"`
	// 		Bar *big.Int `json:"bar"`
	//  }
	//  func do(ctx context.Context, cr ChainReader) (resp ProductReturn, err error) {
	// 		err = cr.GetLatestValue(ctx, "FooContract", "GetProduct", ProductParams{ID:1}, &resp)
	// 		return
	//  }
	//
	// Note that implementations should ignore extra fields in params that are not expected in the call to allow easier
	// use across chains and contract versions.
	// Similarly, when using a struct for returnVal, fields in the return value that are not on-chain will not be set.
	GetLatestValue(ctx context.Context, contract BoundContract, params, returnVal any) error

	// Bind will add provided bindings and will return an error if the contract is not known by the ContractReader, or if
	// the Address is invalid. Any provided binding that already exists should result in a noop.
	Bind(ctx context.Context, bindings []BoundContract) error

	// Unbind will remove all bindings provided.
	Unbind(ctx context.Context, bindings []BoundContract) error

	// QueryKey provides fetching chain agnostic events (Sequence) with general querying capability.
	QueryKey(ctx context.Context, contract BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]Sequence, error)
}

type Head struct {
	Identifier string
	Hash       []byte
	Timestamp  uint64
}

type Sequence struct {
	// This way we can retrieve past/future sequences (EVM log events) very granularly, but still hide the chain detail.
	Cursor string
	Head
	Data any
}

type BoundContract struct {
	Address  string
	Contract string
	Method   string
	Pending  bool
}

func (bc BoundContract) Key() string {
	return bc.Address + "-" + bc.Contract + "-" + bc.Method
}
