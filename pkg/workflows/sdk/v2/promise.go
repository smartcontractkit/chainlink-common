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
	return &basicPromise[T]{await: sync.OnceValues(await)}
}

func PromiseFromResult[T any](result T, err error) Promise[T] {
	return &basicPromise[T]{await: func() (T, error) { return result, err }}
}

type basicPromise[T any] struct {
	await func() (T, error)
}

func (t *basicPromise[T]) Await() (T, error) {
	return t.await()
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
