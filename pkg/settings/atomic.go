package settings

import (
	"context"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var _ Getter = (*Atomic)(nil)
var _ Broadcaster = (*Atomic)(nil)
var _ LocalBroadcaster = (*Atomic)(nil)

// Atomic implements Getter, and supports atomic updates with subscriptions.
type Atomic struct {
	Lggr logger.Logger // optional

	mu      sync.RWMutex
	current *Snapshot
	getter  Getter
	// subs maps the channel handed to each subscriber to the channel Store sends on.
	subs map[<-chan Snapshot]chan Snapshot
}

func (a *Atomic) SetGetter(getter Getter) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.getter = getter
}

func (a *Atomic) Subscribe(ctx context.Context) (<-chan Snapshot, error) {
	from := make(chan Snapshot)
	to := make(chan Snapshot)
	go func() { // launch a helper so we can have non-blocking sends but avoid stale values
		defer close(to)
	outer:
		for v := range from {
			for {
				var ok bool
				select {
				case to <- v:
					continue outer
				case v, ok = <-from:
					if !ok {
						return
					}
				}
			}
		}
	}()
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.current != nil {
		from <- *a.current // seed current value
	}
	if a.subs == nil {
		a.subs = make(map[<-chan Snapshot]chan Snapshot)
	}
	a.subs[to] = from
	return to, nil
}

func (a *Atomic) Unsubscribe(ch <-chan Snapshot) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if from, ok := a.subs[ch]; ok {
		close(from)
		delete(a.subs, ch)
	}
}

func (a *Atomic) Load() (Snapshot, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.current == nil {
		return Snapshot{}, nil
	}
	return *a.current, nil
}

func (a *Atomic) Store(update Snapshot) error {
	cfg := GetterConfig{Logger: a.Lggr}
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

func (a *Atomic) GetScoped(ctx context.Context, scope Scope, key string) (value string, err error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.getter == nil {
		return "", nil
	}
	return a.getter.GetScoped(ctx, scope, key)
}
