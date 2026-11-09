package loop

import (
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// Deprecated: use [settings.Atomic].
type AtomicSettings = settings.Atomic

// Deprecated: instantiate [settings.Atomic] directly, then use [settings.Atomic.SetGetter].
func NewAtomicSettings(initial settings.Getter) *AtomicSettings {
	var as AtomicSettings
	as.SetGetter(initial)
	return &as
}
