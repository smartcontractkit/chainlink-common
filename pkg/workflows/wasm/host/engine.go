package host

import (
	"sync"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v47"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type Engine struct {
	*wasmtime.Engine
	cfg                *wasmtime.Config
	stopCh             chan struct{}
	wg                 sync.WaitGroup
	engineTickInterval time.Duration
}

func newEngine(lggr logger.Logger) *Engine {
	cfg := wasmtime.NewConfig()
	cfg.SetEpochInterruption(true)
	if err := cfg.CacheConfigLoadDefault(); err != nil {
		lggr.Warnw("failed to load cache config, continuing without cache", "error", err)
	}
	cfg.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)
	SetUnwinding(cfg) // Handled differently based on host OS.

	return &Engine{
		Engine:             wasmtime.NewEngineWithConfig(cfg),
		cfg:                cfg,
		stopCh:             make(chan struct{}),
		engineTickInterval: engineTickInterval,
	}
}

func (e *Engine) Start() {
	e.wg.Go(func() {
		ticker := time.NewTicker(e.engineTickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-e.stopCh:
				return
			case <-ticker.C:
				e.IncrementEpoch()
			}
		}
	})
}

func (e *Engine) Close() {
	// Wait for the ticker to exit before deallocating, so it can't increment
	// the epoch of an engine that is no longer there.
	close(e.stopCh)
	e.wg.Wait()
	e.Engine.Close()
	e.cfg.Close()
}

const engineTickInterval = 100 * time.Millisecond

var (
	globalEngineMu sync.Mutex
	globalEngine   *Engine
)

func GetEngine(lggr logger.Logger) *Engine {
	globalEngineMu.Lock()
	defer globalEngineMu.Unlock()

	if globalEngine != nil {
		return globalEngine
	}

	e := newEngine(lggr)
	globalEngine = e
	e.Start()

	return e
}

func CloseEngine() {
	globalEngineMu.Lock()
	defer globalEngineMu.Unlock()

	if globalEngine == nil {
		return
	}

	globalEngine.Close()
	globalEngine = nil
}
