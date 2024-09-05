package testutils

import "github.com/smartcontractkit/chainlink-common/pkg/workflows"

type SDK struct{}

var _ workflows.SDK = &SDK{}
