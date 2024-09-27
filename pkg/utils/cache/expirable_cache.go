package cache

import (
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type StatsCollector interface {
	OnCacheHit()
	OnCacheMiss()
	OnCacheEviction(int)
	OnCacheAddition()
}

// ExpirableCache is a cache that evicts entries after a configurable time of inactivity once a minimum size is reached.
// It is safe for concurrent use.
type ExpirableCache[K comparable, V any] struct {
	m  map[K]*cachedValue[V]
	mu sync.RWMutex

	wg       sync.WaitGroup
	stopChan services.StopChan

	tickInterval   time.Duration
	timeout        time.Duration
	evictAfterSize int

	statsCollector StatsCollector

	clock clockwork.Clock
}

type cachedValue[V any] struct {
	value         V
	lastFetchedAt time.Time
}

type noopStatsCollector struct{}

func (n *noopStatsCollector) OnCacheHit()         {}
func (n *noopStatsCollector) OnCacheMiss()        {}
func (n *noopStatsCollector) OnCacheEviction(int) {}
func (n *noopStatsCollector) OnCacheAddition()    {}

// NewExpirableCache creates a new ExpirableCache with the given tick interval, expiry time, and minimum size.  An optional
// statsCollector can be provided to collect cache stats.
func NewExpirableCache[K comparable, V any](clock clockwork.Clock, tick, expiryAfter time.Duration, evictAfterSize int,
	statsCollector StatsCollector) *ExpirableCache[K, V] {

	if statsCollector == nil {
		statsCollector = &noopStatsCollector{}
	}

	return &ExpirableCache[K, V]{
		m:              map[K]*cachedValue[V]{},
		tickInterval:   tick,
		timeout:        expiryAfter,
		evictAfterSize: evictAfterSize,
		clock:          clock,
		stopChan:       make(chan struct{}),
		statsCollector: statsCollector,
	}
}

func (ec *ExpirableCache[K, V]) Start() {
	ec.wg.Add(1)
	go func() {
		defer ec.wg.Done()
		ec.reapLoop()
	}()
}

func (ec *ExpirableCache[K, V]) Close() {
	close(ec.stopChan)
	ec.wg.Wait()
}

func (ec *ExpirableCache[K, V]) reapLoop() {
	ticker := ec.clock.NewTicker(ec.tickInterval)
	for {
		select {
		case <-ticker.Chan():
			ec.evictOlderThan(ec.timeout)
		case <-ec.stopChan:
			return
		}
	}
}

func (ec *ExpirableCache[K, V]) Add(id K, v V) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.m[id] = &cachedValue[V]{
		value:         v,
		lastFetchedAt: time.Now(),
	}
	ec.statsCollector.OnCacheAddition()
}

func (ec *ExpirableCache[K, V]) Get(id K) (V, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	fetchdValue, ok := ec.m[id]
	if !ok {
		ec.statsCollector.OnCacheMiss()
		var zero V
		return zero, false
	}

	ec.statsCollector.OnCacheHit()
	fetchdValue.lastFetchedAt = ec.clock.Now()
	return fetchdValue.value, true
}

func (ec *ExpirableCache[K, V]) evictOlderThan(duration time.Duration) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	evicted := 0

	if len(ec.m) > ec.evictAfterSize {
		for id, m := range ec.m {
			if ec.clock.Now().Sub(m.lastFetchedAt) > duration {
				delete(ec.m, id)
				evicted++
			}

			if len(ec.m) <= ec.evictAfterSize {
				break
			}
		}
	}

	ec.statsCollector.OnCacheEviction(evicted)
}
