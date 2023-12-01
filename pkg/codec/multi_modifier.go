package codec

import (
	"reflect"
)

type MultiModifier []Modifier

func (c MultiModifier) RetypeForOffChain(onChainType reflect.Type) (reflect.Type, error) {
	return forEach(c, onChainType, Modifier.RetypeForOffChain)
}

func (c MultiModifier) TransformForOnChain(offChainValue any) (any, error) {
	onChainValue := offChainValue
	for i := len(c) - 1; i >= 0; i-- {
		var err error
		if onChainValue, err = c[i].TransformForOnChain(onChainValue); err != nil {
			return nil, err
		}
	}

	return onChainValue, nil
}

func (c MultiModifier) TransformForOffChain(onChainValue any) (any, error) {
	return forEach(c, onChainValue, Modifier.TransformForOffChain)
}

func forEach[T any](c MultiModifier, input T, fn func(Modifier, T) (T, error)) (T, error) {
	output := input
	for _, m := range c {
		var err error
		if output, err = fn(m, output); err != nil {
			return output, err
		}
	}
	return output, nil
}
