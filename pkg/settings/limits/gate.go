package limits

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

type GateLimiter interface {
	Limiter[bool]
	AllowErr(context.Context) error
}

func NewGateLimiter(open bool) GateLimiter {
	return &simpleGateLimiter{open: open}
}

type simpleGateLimiter struct {
	open   bool
	closed atomic.Bool
}

func (s *simpleGateLimiter) Close() error { s.closed.Store(true); return nil }

func (s *simpleGateLimiter) Limit(ctx context.Context) (bool, error) {
	return s.open, nil
}

func (s *simpleGateLimiter) AllowErr(ctx context.Context) error {
	if ok, err := s.Limit(ctx); err != nil {
		return err
	} else if !ok {
		return ErrorNotAllowed{}
	}
	return nil
}

func newGateLimiter(f Factory, limit settings.Setting[bool]) (GateLimiter, error) {
	//TODO
	return &gateLimiter{
		//TODO
	}, nil
}

type gateLimiter struct {
	*updater[bool]
	defaultOpen settings.Setting[bool]

	key   string // optional
	scope settings.Scope

	//TODO record func

	// opt: reap after period of non-use
	updaters sync.Map           // map[string]*updater[N]
	wg       services.WaitGroup // tracks and blocks updaters background routines
}

func (g *gateLimiter) Limit(ctx context.Context) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (g *gateLimiter) AllowErr(ctx context.Context) error {
	//TODO tenant
	if ok, err := g.Limit(ctx); err != nil {
		return err
	} else if !ok {
		return ErrorNotAllowed{Key: g.key, Scope: g.scope, Tenant: tenant}
	}
	return nil
}
