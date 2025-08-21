package utils

import (
	"context"
	"sync"
)

type LazyLoad[T any] struct {
	f     func() (T, error)
	state T
	ok    bool
	lock  sync.Mutex
}

func NewLazyLoad[T any](f func() (T, error)) *LazyLoad[T] {
	return &LazyLoad[T]{f: f}
}

func (l *LazyLoad[T]) Get() (out T, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.ok {
		return l.state, nil
	}
	l.state, err = l.f()
	l.ok = err == nil
	return l.state, err
}

func (l *LazyLoad[T]) Reset() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.ok = false
}

type LazyLoadCtx[T any] struct {
	f     func(context.Context) (T, error)
	state T
	ok    bool
	mu    chan struct{}
}

func NewLazyLoadCtx[T any](f func(context.Context) (T, error)) *LazyLoadCtx[T] {
	return &LazyLoadCtx[T]{f: f, mu: make(chan struct{}, 1)}
}

func (l *LazyLoadCtx[T]) Get(ctx context.Context) (out T, err error) {
	select {
	case <-ctx.Done():
		return out, ctx.Err()
	case l.mu <- struct{}{}: // lock
	}
	defer func() { <-l.mu }() // unlock

	if l.ok {
		return l.state, nil
	}
	l.state, err = l.f(ctx)
	l.ok = err == nil
	return l.state, err
}

func (l *LazyLoadCtx[T]) Reset() {
	select {
	case l.mu <- struct{}{}: // lock
	}
	defer func() { <-l.mu }() // unlock
	l.ok = false
}
