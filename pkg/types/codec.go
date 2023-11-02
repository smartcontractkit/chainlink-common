package types

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type Encoder interface {
	Encode(ctx context.Context, item any, itemType string) (types.Report, error)
}

type Decoder interface {
	Decode(ctx context.Context, raw []byte, into any, itemType string) error
}

type Codec interface {
	Encoder
	Decoder
}

type TypeProvider interface {
	CreateType(itemType string, forceSlice, forEncoding bool) (any, error)
}

type RemoteCodec interface {
	Codec
	TypeProvider
}
