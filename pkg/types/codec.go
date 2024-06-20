package types

import (
	"context"
)

const (
	ErrFieldNotFound   = InvalidArgumentError("field not found")
	ErrInvalidEncoding = InvalidArgumentError("invalid encoding")
	ErrSliceWrongLen   = InvalidArgumentError("slice is wrong length")
	ErrNotASlice       = InvalidArgumentError("element is not a slice")
)

//go:generate mockery --quiet --name Encoder --output ./mocks/ --case=underscore
type Encoder interface {
	Encode(ctx context.Context, item any, itemType string) ([]byte, error)
	// GetMaxEncodingSize returns the max size in bytes if n elements are supplied for all top level dynamically sized elements.
	// If no elements are dynamically sized, the returned value will be the same for all n.
	// If there are multiple levels of dynamically sized elements, or itemType cannot be found,
	// ErrInvalidType will be returned.
	GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error)
}

//go:generate mockery --quiet --name Decoder --output ./mocks/ --case=underscore
type Decoder interface {
	Decode(ctx context.Context, raw []byte, into any, itemType string) error
	// GetMaxDecodingSize returns the max size in bytes if n elements are supplied for all top level dynamically sized elements.
	// If no elements are dynamically sized, the returned value will be the same for all n.
	// If there are multiple levels of dynamically sized elements, or itemType cannot be found,
	// ErrInvalidType will be returned.
	GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error)
}

//go:generate mockery --quiet --name Codec --output ./mocks/ --case=underscore
type Codec interface {
	Encoder
	Decoder
}

//go:generate mockery --quiet --name TypeProvider --output ./mocks/ --case=underscore
type TypeProvider interface {
	CreateType(itemType string, forEncoding bool) (any, error)
}

//go:generate mockery --quiet --name ContractTypeProvider --output ./mocks/ --case=underscore
type ContractTypeProvider interface {
	CreateContractType(contractName, itemType string, forEncoding bool) (any, error)
}

//go:generate mockery --quiet --name RemoteCodec --output ./mocks/ --case=underscore
type RemoteCodec interface {
	Codec
	TypeProvider
}
