package types

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

var (
	ErrNotExist        = errors.New("data does not exist for key")
	ErrTypeMismatch    = errors.New("data type mismatch")
	ErrInvalidStrategy = errors.New("invalid cache strategy")
)

const (
	CachedContractReaderName = "CachedContractReader"

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

var _ ContractReader = &CachedContractReader{}
var _ services.Service = &CachedContractReader{}

type CachedContractReader struct {
	UnimplementedContractReader

	// injected dependencies
	base ContractReader

	// internal configuration properties
	strategies map[string]any
	polling    []cacheKey
	triggers   []cacheKey

	// active data
	data      map[string]*cachedValue
	keyLookup map[string]cacheKey
	subprocs  []chan struct{}
	state     services.StateMachine

	// properties for thread safety
	strategyLock sync.RWMutex
	dataLock     sync.RWMutex
	mu           sync.RWMutex
}

func NewCachedContractReader(base ContractReader) *CachedContractReader {
	return &CachedContractReader{
		base:       base,
		strategies: make(map[string]any),
		data:       make(map[string]*cachedValue),
		keyLookup:  make(map[string]cacheKey),
		polling:    make([]cacheKey, 0),
		triggers:   make([]cacheKey, 0),
		subprocs:   make([]chan struct{}, 0, 2),
	}
}

// GetLatestValue implements the ContractReader interface and caches values.
func (c *CachedContractReader) GetLatestValue(
	ctx context.Context,
	readIdentifier string,
	confidenceLevel primitives.ConfidenceLevel,
	params, returnVal any,
) error {
	strategy, hasConfig := c.getStrategy(readIdentifier)
	if !hasConfig {
		return c.fallThrough(ctx, readIdentifier, confidenceLevel, params, returnVal)
	}

	switch typed := strategy.(type) {
	case PollAndCache:
		// the strategy params should match the incoming params
		if !reflect.DeepEqual(typed.Params, params) {
			return c.fallThrough(ctx, readIdentifier, confidenceLevel, params, returnVal)
		}

		key := newCacheKey(readIdentifier, confidenceLevel, params, 0)
		if err := c.getFromCache(key, returnVal); err == nil || !errors.Is(err, ErrNotExist) {
			return err
		}

		return c.fallThroughAndCache(ctx, key, returnVal)
	case FallthroughOnStale:
		key := newCacheKey(readIdentifier, confidenceLevel, params, typed.Staleness)
		if err := c.getFromCache(key, returnVal); err == nil || !errors.Is(err, ErrNotExist) {
			return err
		}

		return c.fallThroughAndCache(ctx, key, returnVal)
	case NoCache:
		return c.fallThrough(ctx, readIdentifier, confidenceLevel, params, returnVal)
	}

	return c.base.GetLatestValue(ctx, readIdentifier, confidenceLevel, params, returnVal)
}

func (c *CachedContractReader) BatchGetLatestValues(
	ctx context.Context,
	request BatchGetLatestValuesRequest,
) (BatchGetLatestValuesResult, error) {
	return c.base.BatchGetLatestValues(ctx, request)
}

// QueryKey method implements the ContractReader interface and passes through the cache.
func (c *CachedContractReader) QueryKey(
	ctx context.Context,
	contract BoundContract,
	filter query.KeyFilter,
	limitAndSort query.LimitAndSort,
	sequenceDataType any,
) ([]Sequence, error) {
	return c.base.QueryKey(ctx, contract, filter, limitAndSort, sequenceDataType)
}

// Bind implements the ChainReader interface and passes through values to the base ChainReader. No caching applied.
func (c *CachedContractReader) Bind(ctx context.Context, bindings []BoundContract) error {
	return c.base.Bind(ctx, bindings)
}

// Unbind implements the ChainReader interface and passes through values to the base ChainReader. No caching applied.
func (c *CachedContractReader) Unbind(ctx context.Context, bindings []BoundContract) error {
	return c.base.Bind(ctx, bindings)
}

func (c *CachedContractReader) Start(ctx context.Context) error {
	return c.state.StartOnce(CachedContractReaderName, func() error {
		c.startSubproc(c.clean, cleanInterval)
		c.startSubproc(c.pollAndCache, pollInterval)

		return nil
	})
}

func (c *CachedContractReader) Close() error {
	return c.state.StopOnce(CachedContractReaderName, func() error {
		for _, chClose := range c.subprocs {
			chClose <- struct{}{}
		}

		return nil
	})
}

func (c *CachedContractReader) Ready() error { return nil }

func (c *CachedContractReader) Name() string { return CachedContractReaderName }

func (c *CachedContractReader) HealthReport() map[string]error {
	return map[string]error{c.Name(): nil}
}

func (c *CachedContractReader) SetCacheStrategy(readIdentifier string, confidence primitives.ConfidenceLevel, strategy any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch typed := strategy.(type) {
	case PollAndCache:
		c.polling = append(c.polling, newCacheKey(readIdentifier, confidence, typed.Params, 0))
	case TriggerAndCache:
		c.triggers = append(c.triggers, newCacheKey(readIdentifier, confidence, typed.Params, 0))
	case FallthroughOnStale, NoCache:
		// nothing to set up
	default:
		return ErrInvalidStrategy
	}

	c.setStrategy(readIdentifier, strategy)

	return nil
}

func (c *CachedContractReader) TriggerWithContext(ctx context.Context) {
	c.mu.RLock()
	defer c.mu.Unlock()

	for _, key := range c.triggers {
		if ctx.Err() != nil {
			break
		}

		config, ok := c.getStrategy(key.readIdentifier)
		if !ok {
			continue
		}

		triggerConfig := config.(TriggerAndCache)

		_ = c.fallThroughAndCache(ctx, key, triggerConfig.Value)
	}
}

func (c *CachedContractReader) getFromCache(key cacheKey, returnVal any) error {
	c.dataLock.RLock()
	data, hasData := c.data[key.String()]
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

func (c *CachedContractReader) fallThroughAndCache(ctx context.Context, key cacheKey, returnVal any) error {
	data := c.mustGetData(key)

	data.SetLastUpdate(time.Now())

	retBaseType := reflect.Indirect(reflect.New(reflect.Indirect(reflect.ValueOf(returnVal)).Type()))
	savedVal := reflect.New(retBaseType.Type())

	reflect.Indirect(savedVal).Set(retBaseType)

	cachedValue := savedVal.Interface()

	if err := c.fallThrough(ctx, key.readIdentifier, key.confidence, key.params, cachedValue); err != nil {
		return err
	}

	reflect.Indirect(reflect.ValueOf(returnVal)).Set(reflect.Indirect(reflect.ValueOf(cachedValue)))
	data.SetValue(cachedValue)

	return nil
}

func (c *CachedContractReader) fallThrough(
	ctx context.Context,
	readIdentifier string,
	confidenceLevel primitives.ConfidenceLevel,
	params, returnVal any,
) error {
	return c.base.GetLatestValue(ctx, readIdentifier, confidenceLevel, params, returnVal)
}

func (c *CachedContractReader) startSubproc(fn func(), frequency time.Duration) {
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

func (c *CachedContractReader) clean() {
	forRemoval := make([]string, 0)

	c.dataLock.RLock()

	for keyStr, data := range c.data {
		key, keyExists := c.keyLookup[keyStr]
		if keyExists {
			if key.staleness == 0 {
				continue
			}

			if time.Since(data.LastUpdate()) > key.staleness {
				forRemoval = append(forRemoval, keyStr)
			}
		}
	}

	c.dataLock.RUnlock()

	c.dataLock.Lock()

	for _, key := range forRemoval {
		delete(c.data, key)
		delete(c.keyLookup, key)
	}

	c.dataLock.Unlock()
}

func (c *CachedContractReader) pollAndCache() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), pollInterval)

	for _, key := range c.polling {
		if ctx.Err() != nil {
			break
		}

		config, ok := c.getStrategy(key.readIdentifier)
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

func (c *CachedContractReader) getStrategy(readIdentifier string) (any, bool) {
	c.strategyLock.RLock()
	defer c.strategyLock.RUnlock()

	strategy, ok := c.strategies[readIdentifier]

	return strategy, ok
}

func (c *CachedContractReader) setStrategy(readIdentifier string, strategy any) {
	c.strategyLock.Lock()
	defer c.strategyLock.Unlock()

	c.strategies[readIdentifier] = strategy
}

func (c *CachedContractReader) mustGetData(key cacheKey) *cachedValue {
	c.dataLock.RLock()
	data, hasData := c.data[key.String()]
	c.dataLock.RUnlock()

	if !hasData {
		data = &cachedValue{}

		c.dataLock.Lock()
		c.data[key.String()] = data
		c.keyLookup[key.String()] = key
		c.dataLock.Unlock()
	}

	return data
}

func newCacheKey(readIdentifier string, confidence primitives.ConfidenceLevel, params any, staleness time.Duration) cacheKey {
	return cacheKey{
		readIdentifier: readIdentifier,
		confidence:     confidence,
		params:         params,
		staleness:      staleness,
	}
}

type cacheKey struct {
	readIdentifier string
	confidence     primitives.ConfidenceLevel
	params         any
	staleness      time.Duration
}

func (k cacheKey) String() string {
	return fmt.Sprintf("%s::%s::%+v", k.readIdentifier, reflect.TypeOf(k.params), k.params)
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
