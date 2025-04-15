package sdk

import (
	"sync"
)

type Promise[T any] interface {
	Await() (T, error)
	CapabilityPromise
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
	await      func() (T, error)
}

func (t *basicPromise[T]) Await() (T, error) {
	if t.isResolved {
		return t.resolved, t.err
	}

	t.Lock()
	defer t.Unlock()

	t.resolved, t.err = t.await()
	t.isResolved = true

	return t.resolved, t.err
}

func (t *basicPromise[T]) promise() {}

func Then[I, O any](p Promise[I], fn func(I) (O, error)) Promise[O] {
	return NewBasicPromise[O](func() (O, error) {
		underlyingResult, err := p.Await()
		if err != nil {
			var o O
			return o, err
		}

		return fn(underlyingResult)
	})
}
