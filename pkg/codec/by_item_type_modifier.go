package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewByItemTypeModifier(modByitemType map[string]Modifier) (Modifier, error) {
	if modByitemType == nil {
		modByitemType = map[string]Modifier{}
	}

	return &byItemTypeModifier{
		modByitemType: modByitemType,
	}, nil
}

type byItemTypeModifier struct {
	modByitemType map[string]Modifier
}

func (b *byItemTypeModifier) RetypeForOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	mod, ok := b.modByitemType[itemType]
	if !ok {
		return nil, fmt.Errorf("%w: cannot find modifier for %s", types.ErrInvalidType, itemType)
	}

	return mod.RetypeForOffChain(onChainType, itemType)
}

func (b *byItemTypeModifier) TransformForOnChain(offChainValue any, itemType string) (any, error) {
	return b.transform(offChainValue, itemType, Modifier.TransformForOnChain)
}

func (b *byItemTypeModifier) TransformForOffChain(onChainValue any, itemType string) (any, error) {
	return b.transform(onChainValue, itemType, Modifier.TransformForOffChain)
}

func (b *byItemTypeModifier) transform(
	val any, itemType string, transform func(Modifier, any, string) (any, error)) (any, error) {
	mod, ok := b.modByitemType[itemType]
	if !ok {
		return nil, fmt.Errorf("%w: cannot find modifier for %s", types.ErrInvalidType, itemType)
	}

	return transform(mod, val, itemType)
}

var _ Modifier = &byItemTypeModifier{}
