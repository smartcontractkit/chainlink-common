package types

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

type CCIPCommitProviderOCR3 interface {
	PluginProvider

	ReportCodec(ctx context.Context) (ccipocr3.CommitPluginCodec, error)
	MsgHasher(ctx context.Context) (ccipocr3.MessageHasher, error)
}

type CCIPExecuteProviderOCR3 interface {
	PluginProvider

	ReportCodec(ctx context.Context) (ccipocr3.ExecutePluginCodec, error)
	MsgHasher(ctx context.Context) (ccipocr3.MessageHasher, error)
}
