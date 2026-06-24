package utils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// StartStopOnce can be embedded in a struct to help implement types.Service.
//
// Deprecated: use services.StateMachine
//
//go:fix inline
type StartStopOnce = services.StateMachine
