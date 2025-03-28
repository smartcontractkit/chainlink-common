package sdk

import (
	"sync"
)

type Promise[T any] interface {
	Await() (T, error)
	CapabilityPromise
	Subscribe(then func(t T, e error))
}

type CapabilityPromise interface {
	promise()
}

func NewBasicPromise[T any](await func() (T, error)) Promise[T] {
	return &basicPromise[T]{await: await}
}

func PromiseFromResult[T any](result T, err error) Promise[T] {
	return &basicPromise[T]{resolved: result, err: err, isResolved: true}
}

type basicPromise[T any] struct {
	sync.Mutex
	resolved   T
	err        error
	isResolved bool
	callbacks  []func(T, error)
	await      func() (T, error)
}

func (t *basicPromise[T]) Subscribe(then func(t T, e error)) {
	if t.isResolved {
		then(t.resolved, t.err)
		return
	}

	t.Lock()
	defer t.Unlock()

	t.callbacks = append(t.callbacks, then)
}

func (t *basicPromise[T]) Await() (T, error) {
	if t.isResolved {
		return t.resolved, t.err
	}

	t.Lock()
	defer t.Unlock()

	t.resolved, t.err = t.await()
	t.isResolved = true

	for _, cb := range t.callbacks {
		cb(t.resolved, t.err)
	}

	return t.resolved, t.err
}

func (t *basicPromise[T]) promise() {}

func Then[I, O any](p Promise[I], fn func(I) (O, error)) Promise[O] {
	resolver := sync.Mutex{}
	then := NewBasicPromise[O](func() (O, error) {
		resolver.Lock()
		defer resolver.Unlock()
		underlyingResult, err := p.Await()
		if err != nil {
			var o O
			return o, err
		}
		
		return fn(underlyingResult)
	})

	p.Subscribe(func(t I, e error) {
		// if we're already in an await, we can't await again
		if resolver.TryLock() {
			defer resolver.Unlock()
			_, _ = then.Await()
		}
	})
	return then
}
