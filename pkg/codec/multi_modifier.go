package codec

import (
	"reflect"
	"slices"
)

// MultiModifier is a Modifier that applies each element for the slice in-order (reverse order for TransformForOnChain).
type MultiModifier []Modifier

func (c MultiModifier) RetypeToOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	return forEach(c, onChainType, itemType, Modifier.RetypeToOffChain)
}

func (c MultiModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	onChainValue := offChainValue
	for _, v := range slices.Backward(c) {
		var err error
		if onChainValue, err = v.TransformToOnChain(onChainValue, itemType); err != nil {
			return nil, err
		}
	}

	return onChainValue, nil
}

func (c MultiModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return forEach(c, onChainValue, itemType, Modifier.TransformToOffChain)
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
