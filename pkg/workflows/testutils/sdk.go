package testutils

import "github.com/smartcontractkit/chainlink-common/pkg/workflows"

type Sdk struct{}

var _ workflows.Sdk = &Sdk{}
