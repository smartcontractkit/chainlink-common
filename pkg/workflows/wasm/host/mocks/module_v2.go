package mocks

import (
	"github.com/stretchr/testify/mock"

	hostmocks "github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
)

// ModuleV2 is a backward-compatible alias for hostmocks.Module.
// The ModuleV2 interface now lives in pkg/workflows/host as Module;
// this alias keeps existing consumers compiling without changes.
type ModuleV2 = hostmocks.Module

// NewModuleV2 creates a new instance of ModuleV2 (alias for hostmocks.NewModule).
func NewModuleV2(t interface {
	mock.TestingT
	Cleanup(func())
}) *ModuleV2 {
	return hostmocks.NewModule(t)
}
