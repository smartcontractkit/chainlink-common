package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewByItemTypeModifier returns a Modifier that uses modByItemType to determine which Modifier to use for a given itemType.
func NewByItemTypeModifier(modByItemType map[string]Modifier) (Modifier, error) {
	if modByItemType == nil {
		modByItemType = map[string]Modifier{}
	}

	return &byItemTypeModifier{
		modByitemType: modByItemType,
	}, nil
}

type byItemTypeModifier struct {
	modByitemType map[string]Modifier
}

// RetypeToOffChain attempts to apply a modifier using the provided itemType. To allow access to nested fields, this
// function applies no modifications if a modifier by the specified name is not found.
func (b *byItemTypeModifier) RetypeToOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	head, tail := extendedItemType(itemType).next()

	mod, ok := b.modByitemType[head]
	if !ok {
		return nil, fmt.Errorf("%w: cannot find modifier for %s", types.ErrInvalidType, itemType)
	}

	return mod.RetypeToOffChain(onChainType, tail)
}

func (b *byItemTypeModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	return b.transform(offChainValue, itemType, Modifier.TransformToOnChain)
}

func (b *byItemTypeModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return b.transform(onChainValue, itemType, Modifier.TransformToOffChain)
}

func (b *byItemTypeModifier) transform(
	val any,
	itemType string,
	transform func(Modifier, any, string) (any, error),
) (any, error) {
	head, tail := extendedItemType(itemType).next()

	if mod, ok := b.modByitemType[head]; ok {
		return transform(mod, val, tail)
	}

	return val, nil
}

var _ Modifier = &byItemTypeModifier{}
