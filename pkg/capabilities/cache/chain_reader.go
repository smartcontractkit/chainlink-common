package cache

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

var (
	ErrNotExist        = errors.New("data does not exist for key")
	ErrTypeMismatch    = errors.New("data type mismatch")
	ErrInvalidStrategy = errors.New("invalid cache strategy")
)

const (
	ChainReaderCacheName = "ChainReaderCache"

	cleanInterval = time.Second
	pollInterval  = time.Second
)

type TriggerAndCache struct {
	Params any
	Value  any
}

type PollAndCache struct {
	TriggerAndCache
	Interval time.Duration
}

type FallthroughOnStale struct {
	Staleness time.Duration
}

type NoCache struct{}

var _ types.ContractReader = &ChainReader{}
var _ services.Service = &ChainReader{}

type ChainReader struct {
	services.StateMachine

	// injected dependencies
	base types.ContractReader

	// internal configuration properties
	strategies map[string]any
	polling    []cacheKey
	triggers   []cacheKey

	// active data
	data     map[cacheKey]*cachedValue
	subprocs []chan struct{}

	// properties for thread safety
	strategyLock sync.RWMutex
	dataLock     sync.RWMutex
	mu           sync.RWMutex
}

func NewChainReader(base types.ContractReader) *ChainReader {
	return &ChainReader{
		base:       base,
		strategies: make(map[string]any),
		data:       make(map[cacheKey]*cachedValue),
		polling:    make([]cacheKey, 0),
		triggers:   make([]cacheKey, 0),
		subprocs:   make([]chan struct{}, 0, 2),
	}
}

// GetLatestValue implements the ContractReader interface and caches values.
func (c *ChainReader) GetLatestValue(
	ctx context.Context,
	contractName,
	method string,
	params, returnVal any,
) error {
	strategy, hasConfig := c.getStrategy(contractName, method)
	if !hasConfig {
		return c.fallThrough(ctx, contractName, method, params, returnVal)
	}

	switch typed := strategy.(type) {
	case PollAndCache:
		// the strategy params should match the incoming params
		if !reflect.DeepEqual(typed.Params, params) {
			return c.fallThrough(ctx, contractName, method, params, returnVal)
		}

		key := newCacheKey(contractName, method, params, 0)
		if err := c.getFromCache(key, returnVal); err == nil || !errors.Is(err, ErrNotExist) {
			return err
		}

		return c.fallThroughAndCache(ctx, key, returnVal)
	case FallthroughOnStale:
		key := newCacheKey(contractName, method, params, typed.Staleness)
		if err := c.getFromCache(key, returnVal); err == nil || !errors.Is(err, ErrNotExist) {
			return err
		}

		return c.fallThroughAndCache(ctx, key, returnVal)
	case NoCache:
		return c.fallThrough(ctx, contractName, method, params, returnVal)
	}

	return c.base.GetLatestValue(ctx, contractName, method, params, returnVal)
}

// QueryKey method implements the ContractReader interface and passes through the cache.
func (c *ChainReader) QueryKey(
	ctx context.Context,
	contractName string,
	filter query.KeyFilter,
	limitAndSort query.LimitAndSort,
	sequenceDataType any,
) ([]types.Sequence, error) {
	return c.base.QueryKey(ctx, contractName, filter, limitAndSort, sequenceDataType)
}

// Bind implements the ChainReader interface and passes through values to the base ChainReader. No caching applied.
func (c *ChainReader) Bind(ctx context.Context, bindings []types.BoundContract) error {
	return c.base.Bind(ctx, bindings)
}

func (c *ChainReader) Start(ctx context.Context) error {
	return c.StartOnce(ChainReaderCacheName, func() error {
		c.startSubproc(c.clean, cleanInterval)
		c.startSubproc(c.pollAndCache, pollInterval)

		return nil
	})
}

func (c *ChainReader) Close() error {
	return c.StopOnce(ChainReaderCacheName, func() error {
		for _, chClose := range c.subprocs {
			chClose <- struct{}{}
		}

		return nil
	})
}

func (c *ChainReader) Ready() error { return nil }

func (c *ChainReader) Name() string { return ChainReaderCacheName }

func (c *ChainReader) HealthReport() map[string]error {
	return map[string]error{c.Name(): nil}
}

func (c *ChainReader) SetCacheStrategy(contractName, method string, strategy any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch typed := strategy.(type) {
	case PollAndCache:
		c.polling = append(c.polling, newCacheKey(contractName, method, typed.Params, 0))
	case TriggerAndCache:
		c.triggers = append(c.triggers, newCacheKey(contractName, method, typed.Params, 0))
	case FallthroughOnStale, NoCache:
		// nothing to set up
	default:
		return ErrInvalidStrategy
	}

	c.setStrategy(contractName, method, strategy)

	return nil
}

func (c *ChainReader) TriggerWithContext(ctx context.Context) {
	c.mu.RLock()
	defer c.mu.Unlock()

	for _, key := range c.triggers {
		if ctx.Err() != nil {
			break
		}

		config, ok := c.getStrategy(key.contractName, key.method)
		if !ok {
			continue
		}

		triggerConfig := config.(TriggerAndCache)

		_ = c.fallThroughAndCache(ctx, key, triggerConfig.Value)
	}
}

func (c *ChainReader) getFromCache(key cacheKey, returnVal any) error {
	c.dataLock.RLock()
	data, hasData := c.data[key]
	c.dataLock.RUnlock()

	if !hasData {
		return ErrNotExist
	}

	if key.staleness > 0 {
		// do staleness check
		if time.Since(data.LastUpdate()) > key.staleness {
			return ErrNotExist
		}
	}

	retVal := reflect.ValueOf(returnVal)
	dataVal := reflect.ValueOf(data.Value())

	if retVal.Type() != dataVal.Type() {
		return fmt.Errorf("%w: cached value and return value do not have the same type", ErrTypeMismatch)
	}

	reflect.Indirect(retVal).Set(reflect.Indirect(dataVal))

	return nil
}

func (c *ChainReader) fallThroughAndCache(ctx context.Context, key cacheKey, returnVal any) error {
	data := c.mustGetData(key)

	data.SetLastUpdate(time.Now())

	retBaseType := reflect.Indirect(reflect.New(reflect.Indirect(reflect.ValueOf(returnVal)).Type()))
	savedVal := reflect.New(retBaseType.Type())

	reflect.Indirect(savedVal).Set(retBaseType)

	cachedValue := savedVal.Interface()

	if err := c.fallThrough(ctx, key.contractName, key.method, key.params, cachedValue); err != nil {
		return err
	}

	reflect.Indirect(reflect.ValueOf(returnVal)).Set(reflect.Indirect(reflect.ValueOf(cachedValue)))
	data.SetValue(cachedValue)

	return nil
}

func (c *ChainReader) fallThrough(ctx context.Context, contractName, method string, params, returnVal any) error {
	return c.base.GetLatestValue(ctx, contractName, method, params, returnVal)
}

func (c *ChainReader) startSubproc(fn func(), frequency time.Duration) {
	c.subprocs = append(c.subprocs, make(chan struct{}))

	go func(chClose <-chan struct{}, fn func(), freq time.Duration) {
		for {
			timer := time.NewTimer(utils.WithJitter(freq))

			select {
			case <-timer.C:
				timer.Stop()
				fn()
			case <-chClose:
				timer.Stop()

				return
			}
		}
	}(c.subprocs[len(c.subprocs)-1], fn, frequency)
}

func (c *ChainReader) clean() {
	forRemoval := make([]cacheKey, 0)

	c.dataLock.RLock()

	for key, data := range c.data {
		if key.staleness == 0 {
			continue
		}

		if time.Since(data.LastUpdate()) > key.staleness {
			forRemoval = append(forRemoval, key)
		}
	}

	c.dataLock.RUnlock()

	c.dataLock.Lock()

	for _, key := range forRemoval {
		delete(c.data, key)
	}

	c.dataLock.Unlock()
}

func (c *ChainReader) pollAndCache() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), pollInterval)

	for _, key := range c.polling {
		if ctx.Err() != nil {
			break
		}

		config, ok := c.getStrategy(key.contractName, key.method)
		if !ok {
			continue
		}

		pollConfig := config.(PollAndCache)
		data := c.mustGetData(key)

		if time.Since(data.LastUpdate()) < pollConfig.Interval {
			continue
		}

		_ = c.fallThroughAndCache(ctx, key, pollConfig.Value)
	}

	cancel()
}

func (c *ChainReader) getStrategy(contract, method string) (any, bool) {
	c.strategyLock.RLock()
	defer c.strategyLock.RUnlock()

	strategy, ok := c.strategies[contract+method]

	return strategy, ok
}

func (c *ChainReader) setStrategy(contract, method string, strategy any) {
	c.strategyLock.Lock()
	defer c.strategyLock.Unlock()

	c.strategies[contract+method] = strategy
}

func (c *ChainReader) mustGetData(key cacheKey) *cachedValue {
	c.dataLock.RLock()
	data, hasData := c.data[key]
	c.dataLock.RUnlock()

	if !hasData {
		data = &cachedValue{}

		c.dataLock.Lock()
		c.data[key] = data
		c.dataLock.Unlock()
	}

	return data
}

func newCacheKey(contractName, method string, params any, staleness time.Duration) cacheKey {
	return cacheKey{
		contractName: contractName,
		method:       method,
		params:       params,
		staleness:    staleness,
	}
}

type cacheKey struct {
	contractName, method string
	params               any
	staleness            time.Duration
}

func (k cacheKey) String() string {
	return fmt.Sprintf("%s::%s::%s::%+v", k.contractName, k.method, reflect.TypeOf(k.params), k.params)
}

type cachedValue struct {
	value      any
	lastUpdate time.Time
	mu         sync.RWMutex
}

func (v *cachedValue) LastUpdate() time.Time {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.lastUpdate
}

func (v *cachedValue) SetLastUpdate(updated time.Time) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.lastUpdate = updated
}

func (v *cachedValue) Value() any {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.value
}

func (v *cachedValue) SetValue(value any) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.value = value
}
