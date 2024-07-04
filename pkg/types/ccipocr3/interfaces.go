package ccipocr3

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type CommitPluginCodec interface {
	Encode(context.Context, CommitPluginReport) ([]byte, error)
	Decode(context.Context, []byte) (CommitPluginReport, error)
}

type ExecutePluginCodec interface {
	Encode(context.Context, ExecutePluginReport) ([]byte, error)
	Decode(context.Context, []byte) (ExecutePluginReport, error)
}

type MessageHasher interface {
	Hash(context.Context, CCIPMsg) (Bytes32, error)
}
