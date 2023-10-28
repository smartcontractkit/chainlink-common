package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

// Service represents a long-running service inside the Application.
//
// The simplest way to implement a Service is via NewService.
//
// For other cases, consider embedding a services.StateMachine to implement these
// calls in a safe manner.
type Service interface {
	// Start the service.
	//  - Must return promptly if the context is cancelled.
	//  - Must not retain the context after returning (only applies to start-up)
	//  - Must not depend on external resources (no blocking network calls)
	Start(context.Context) error
	// Close stops the Service.
	// Invariants: Usually after this call the Service cannot be started
	// again, you need to build a new Service to do so.
	//
	// See MultiCloser
	Close() error

	HealthReporter
}

// Engine manages service internals like health, goroutine tracking, and shutdown signals.
type Engine struct {
	StopChan
	logger.SugaredLogger

	wg sync.WaitGroup

	healthEvent func(error)
	conds       map[string]error
	condsMu     sync.RWMutex
}

// Go runs fn in a tracked goroutine that will block closing the service.
func (e *Engine) Go(fn func(context.Context)) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		ctx, cancel := e.StopChan.NewCtx()
		defer cancel()
		fn(ctx)
	}()
}

// GoTick is like Go but calls fn for each tick.
//
//	v.e.GoTick(services.NewTicker(time.Minute), v.method)
func (e *Engine) GoTick(ticker *timeutil.Ticker, fn func(context.Context)) {
	e.Go(func(ctx context.Context) {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fn(ctx)
			}
		}
	})
}

// HealthEvent records an error to be reported via the next call to Healthy().
func (e *Engine) HealthEvent(err error) { e.healthEvent(err) }

// SetUnhealthy records a condition key and an error, which causes an unhealthy report, until SetHealthy(condition) is called.
func (e *Engine) SetUnhealthy(condition string, err error) {
	e.condsMu.Lock()
	defer e.condsMu.Unlock()
	e.conds[condition] = fmt.Errorf("%s: %e", condition, err)
}

// SetHealthy removes a condition and error recorded by SetUnhealthy.
func (e *Engine) SetHealthy(condition string) {
	e.condsMu.Lock()
	defer e.condsMu.Unlock()
	delete(e.conds, condition)
}

// Unhealthy causes an unhealthy report, until the returned func() is called.
// Use this for simple cases where the func() can be kept in scope, and prefer to defer it inline if possible:
//
//	defer Unhealthy(fmt.Errorf("foo bar: %i", err))()
//
// See SetUnhealthy for an alternative API.
func (e *Engine) Unhealthy(err error) func() {
	cond := uuid.NewString()
	e.SetUnhealthy(cond, err)
	return func() { e.SetHealthy(cond) }
}

func (e *Engine) healthy() error {
	e.condsMu.RLock()
	errs := maps.Values(e.conds)
	e.condsMu.RUnlock()
	return errors.Join(errs...)
}

// Config is a service configuration.
type Config struct {
	// Name is required. It will be logged shorthand on Start and Close, and appended to the logger name.
	// It must be unique among services sharing the same logger, in order to ensure uniqueness of the fully qualified name.
	Name string
	// NewSubServices is an optional constructor for dependent Services to Start and Close along with this one.
	NewSubServices func(logger.Logger) []Service
	// Start is an optional hook called after starting SubServices.
	Start func(context.Context) error
	// Close is an optional hook called before closing SubServices.
	Close func() error //TODO examples
}

// NewEngine returns a new Service defined by Config, and an Engine for managing health, goroutines, and logging.
//   - You *should* embed the Service, in order to inherit the methods.
//   - You *should not* embed the Engine. Use an unexported field instead.
func (c Config) NewEngine(lggr logger.Logger) (Service, *Engine) {
	s := c.new(lggr)
	return s, &s.eng
}

// NewService returns a new Service defined by Config.
//   - You *should* embed the Service, in order to inherit the methods.
func (c Config) NewService(lggr logger.Logger) Service {
	return c.new(lggr)
}

func (c Config) new(lggr logger.Logger) *service {
	lggr = logger.Sugared(logger.Named(lggr, c.Name))
	s := &service{
		cfg: c,
		eng: Engine{
			StopChan:      make(StopChan),
			SugaredLogger: logger.Sugared(lggr),
			conds:         make(map[string]error),
		},
	}
	s.eng.healthEvent = s.StateMachine.SvcErrBuffer.Append
	if c.NewSubServices != nil {
		s.subs = c.NewSubServices(lggr)
	}
	return s
}

type service struct {
	StateMachine
	cfg  Config
	eng  Engine
	subs []Service
}

// Ready implements [HealthReporter.Ready] and overrides and extends [utils.StartStopOnce.Ready()] to include [Config.SubServices]
// readiness as well.
func (s *service) Ready() (err error) {
	err = s.StateMachine.Ready()
	for _, sub := range s.subs {
		err = errors.Join(err, sub.Ready())
	}
	return
}

// Healthy overrides [StateMachine.Healthy] and extends it to include Engine errors as well.
// Do not override this method in your service. Instead, report errors via the Engine.
func (s *service) Healthy() (err error) {
	err = s.StateMachine.Healthy()
	if err == nil {
		err = s.eng.healthy()
	}
	return
}

func (s *service) HealthReport() map[string]error {
	m := map[string]error{s.Name(): s.Healthy()}
	for _, sub := range s.subs {
		CopyHealth(m, sub.HealthReport())
	}
	return m
}

func (s *service) Name() string { return s.eng.SugaredLogger.Name() }

func (s *service) Start(ctx context.Context) error {
	return s.StartOnce(s.cfg.Name, func() error {
		s.eng.Info("Starting")
		if len(s.subs) > 0 {
			var ms MultiStart
			s.eng.Infof("Starting %d sub-services", len(s.subs))
			for _, sub := range s.subs {
				if err := ms.Start(ctx, sub); err != nil {
					s.eng.Errorw("Failed to start sub-service", "error", err)
					return fmt.Errorf("failed to start sub-service of %s: %w", s.cfg.Name, err)
				}
			}
		}
		if s.cfg.Start != nil {
			if err := s.cfg.Start(ctx); err != nil {
				return fmt.Errorf("failed to start service %s: %w", s.cfg.Name, err)
			}
		}
		s.eng.Info("Started")
		return nil
	})
}

func (s *service) Close() error {
	return s.StopOnce(s.cfg.Name, func() (err error) {
		s.eng.Info("Closing")
		defer s.eng.Info("Closed")

		close(s.eng.StopChan)
		s.eng.wg.Wait()

		if s.cfg.Close != nil {
			err = s.cfg.Close()
		}

		if len(s.subs) == 0 {
			return
		}
		s.eng.Infof("Closing %d sub-services", len(s.subs))
		err = errors.Join(err, MultiCloser(s.subs).Close())
		return
	})
}
