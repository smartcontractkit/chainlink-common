package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type runtime struct{}

var _ sdk.Runtime = &runtime{}
