package codec

import (
	"reflect"
)

// MultiModifier is a Modifier that applies each element for the slice in-order (reverse order for TransformForOnChain).
type MultiModifier []Modifier

func (c MultiModifier) RetypeForOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	return forEach(c, onChainType, itemType, Modifier.RetypeForOffChain)
}

func (c MultiModifier) TransformForOnChain(offChainValue any, itemType string) (any, error) {
	onChainValue := offChainValue
	for i := len(c) - 1; i >= 0; i-- {
		var err error
		if onChainValue, err = c[i].TransformForOnChain(onChainValue, itemType); err != nil {
			return nil, err
		}
	}

	return onChainValue, nil
}

func (c MultiModifier) TransformForOffChain(onChainValue any, itemType string) (any, error) {
	return forEach(c, onChainValue, itemType, Modifier.TransformForOffChain)
}

func forEach[T any](c MultiModifier, input T, itemType string, fn func(Modifier, T, string) (T, error)) (T, error) {
	output := input
	for _, m := range c {
		var err error
		if output, err = fn(m, output, itemType); err != nil {
			return output, err
		}
	}
	return output, nil
}
