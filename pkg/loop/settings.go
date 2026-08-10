package loop

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ settings.Getter = (*AtomicSettings)(nil)
var _ core.SettingsBroadcaster = (*AtomicSettings)(nil)

// AtomicSettings implements settings.Getter, and supports atomic updates with subscriptions.
type AtomicSettings struct {
	Lggr logger.Logger // optional

	mu      sync.RWMutex
	current *core.SettingsUpdate
	getter  settings.Getter
	subs    []chan core.SettingsUpdate
}

// Deprecated: instantiate AtomicSettings directly, then use AtomicSettings.SetGetter.
func NewAtomicSettings(initial settings.Getter) *AtomicSettings {
	var as AtomicSettings
	as.SetGetter(initial)
	return &as
}

func (a *AtomicSettings) SetGetter(getter settings.Getter) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.getter = getter
}

func (a *AtomicSettings) Subscribe(ctx context.Context) (<-chan core.SettingsUpdate, error) {
	from := make(chan core.SettingsUpdate)
	to := make(chan core.SettingsUpdate)
	go func() { // launch a helper so we can have non-blocking sends but avoid stale values
		defer close(to)
	outter:
		for v := range from {
			// buffered loop
			for {
				var ok bool
				select {
				case to <- v:
					// continue the unbuffered loop
					continue outter
				case v, ok = <-from:
					if !ok {
						return
					}
					// continue buffered with new v
				}
			}
		}
	}()
	if a.current != nil {
		from <- *a.current // seed current value
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.subs = append(a.subs, from)
	return to, nil
}

func (a *AtomicSettings) Unsubscribe(ch <-chan core.SettingsUpdate) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if i := slices.IndexFunc(a.subs, func(e chan core.SettingsUpdate) bool { return e == ch }); i >= 0 {
		close(a.subs[i])
		l := len(a.subs)
		if l > 1 { // replace with last
			a.subs[i] = a.subs[l-1]
		}
		a.subs = a.subs[:l-1] // remove last
	}
}

func (a *AtomicSettings) Load() (core.SettingsUpdate, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.current == nil {
		return core.SettingsUpdate{}, nil
	}
	return *a.current, nil
}

func (a *AtomicSettings) Store(update core.SettingsUpdate) error {
	cfg := settings.GetterConfig{Logger: a.Lggr}
	if cfg.Logger == nil {
		cfg.Logger = logger.Nop()
	}
	getter, err := cfg.NewTOMLGetter([]byte(update.Settings))
	if err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.current = &update
	a.getter = getter
	for _, ch := range a.subs {
		ch <- update // non-blocking due to Subscribe goroutine
	}
	return nil
}

func (a *AtomicSettings) GetScoped(ctx context.Context, scope settings.Scope, key string) (value string, err error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.getter == nil {
		return "", nil
	}
	return a.getter.GetScoped(ctx, scope, key)
}
