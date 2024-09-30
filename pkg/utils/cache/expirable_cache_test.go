package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	A int
}

func TestExpirableCache(t *testing.T) {
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 1 * time.Second

	cache := NewExpirableCache[string, testStruct](clock, tick, timeout, 0, nil)
	cache.Start()
	defer cache.Close()

	id := uuid.New().String()
	value := testStruct{A: 42}
	cache.Add(id, value)

	got, ok := cache.Get(id)
	assert.True(t, ok)
	assert.Equal(t, value, got)

	assert.Eventually(t, func() bool {
		clock.Advance(15 * time.Second)
		_, ok := cache.Get(id)
		return !ok
	}, 100*time.Second, 100*time.Millisecond)
}

func TestExpirableCache_DoesNotEvictIfBelowMinimumSize(t *testing.T) {
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 60 * time.Second

	cache := NewExpirableCache[string, testStruct](clock, tick, timeout, 1, nil)
	cache.Start()
	defer cache.Close()

	id := uuid.New().String()
	value := testStruct{A: 42}
	cache.Add(id, value)

	got, ok := cache.Get(id)
	assert.True(t, ok)
	assert.Equal(t, value, got)

	clock.Advance(120 * time.Second)
	_, ok = cache.Get(id)
	assert.True(t, ok)
}

func TestExpirableCache_DoesNotEvictBelowMinimumSize(t *testing.T) {
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 60 * time.Second

	cache := NewExpirableCache[string, testStruct](clock, tick, timeout, 1, nil)
	cache.Start()
	defer cache.Close()

	id1 := uuid.New().String()
	value1 := testStruct{A: 43}
	cache.Add(id1, value1)

	id2 := uuid.New().String()
	value2 := testStruct{A: 44}
	cache.Add(id2, value2)

	// Advance time to check eviction behavior
	assert.Eventually(t, func() bool {
		clock.Advance(120 * time.Second)
		_, ok1 := cache.Get(id1)
		_, ok2 := cache.Get(id2)
		return ok1 != ok2
	}, 100*time.Second, 100*time.Millisecond)
}

func TestExpirableCache_ExpiryTimeResetAfterFetch(t *testing.T) {
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 100 * time.Second

	cache := NewExpirableCache[string, testStruct](clock, tick, timeout, 0, nil)
	cache.Start()
	defer cache.Close()

	id := uuid.New().String()
	value := testStruct{A: 42}
	cache.Add(id, value)

	clock.Advance(timeout / 2)

	// Fetch the item to reset its expiry time
	_, ok := cache.Get(id)
	assert.True(t, ok)

	clock.Advance(timeout)

	_, ok = cache.Get(id)
	assert.True(t, ok)
}

func TestExpirableCache_StatsCollector(t *testing.T) {
	clock := clockwork.NewFakeClock()
	tick := 1 * time.Second
	timeout := 10 * time.Second

	stats := &mockStatsCollector{}
	cache := NewExpirableCache[string, testStruct](clock, tick, timeout, 0, stats)
	cache.Start()
	defer cache.Close()

	id := uuid.New().String()
	value := testStruct{A: 42}
	cache.Add(id, value)

	// Check addition count
	assert.Equal(t, 1, stats.Additions())

	// Fetch the item to increment hit counter
	_, ok := cache.Get(id)
	assert.True(t, ok)
	assert.Equal(t, 1, stats.Hits())

	// Fetch a non-existent item to increment miss counter
	_, ok = cache.Get("non-existent")
	assert.False(t, ok)
	assert.Equal(t, 1, stats.Misses())

	// Advance time to trigger eviction
	assert.Eventually(t, func() bool {
		clock.Advance(15 * time.Second)
		_, ok := cache.Get(id)
		return !ok
	}, 100*time.Second, 100*time.Millisecond)

	// Check eviction count
	assert.Equal(t, 1, stats.Evictions())
}

type mockStatsCollector struct {
	mu        sync.Mutex
	hits      int
	misses    int
	evictions int
	additions int
}

func (m *mockStatsCollector) OnCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hits++
}

func (m *mockStatsCollector) OnCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.misses++
}

func (m *mockStatsCollector) OnCacheEviction(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evictions += count
}

func (m *mockStatsCollector) OnCacheAddition() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.additions++
}

func (m *mockStatsCollector) Hits() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hits
}

func (m *mockStatsCollector) Misses() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.misses
}

func (m *mockStatsCollector) Evictions() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.evictions
}

func (m *mockStatsCollector) Additions() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.additions
}
