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
		modByItemType: modByItemType,
		enableNesting: false,
	}, nil
}

// NewNestableByItemTypeModifier returns a Modifier that uses modByItemType to determine which Modifier to use for a
// given itemType. If itemType is structured as a dot-separated string like 'A.B.C', the first part 'A' will be used to
// match in the mod map and the remaining list will be provided to the found Modifier 'B.C'.
func NewNestableByItemTypeModifier(modByItemType map[string]Modifier) (Modifier, error) {
	if modByItemType == nil {
		modByItemType = map[string]Modifier{}
	}

	return &byItemTypeModifier{
		modByItemType: modByItemType,
		enableNesting: true,
	}, nil
}

type byItemTypeModifier struct {
	modByItemType map[string]Modifier
	enableNesting bool
}

// RetypeToOffChain attempts to apply a modifier using the provided itemType. To allow access to nested fields, this
// function returns an error if a modifier by the specified name is not found. If nesting is enabled, the itemType can
// be of the form `Path.To.Type` and this modifier will attempt to only match on `Path` to find a valid modifier.
func (m *byItemTypeModifier) RetypeToOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error) {
	head := itemType
	tail := itemType

	if m.enableNesting {
		head, tail = ItemTyper(itemType).Next()
	}

	mod, ok := m.modByItemType[head]
	if !ok {
		return nil, fmt.Errorf("%w: cannot find modifier for %s", types.ErrInvalidType, itemType)
	}

	return mod.RetypeToOffChain(onChainType, tail)
}

func (m *byItemTypeModifier) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	return m.transform(offChainValue, itemType, Modifier.TransformToOnChain)
}

func (m *byItemTypeModifier) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	return m.transform(onChainValue, itemType, Modifier.TransformToOffChain)
}

func (m *byItemTypeModifier) transform(
	val any,
	itemType string,
	transform func(Modifier, any, string) (any, error),
) (any, error) {
	head := itemType
	tail := itemType

	if m.enableNesting {
		head, tail = ItemTyper(itemType).Next()
	}

	mod, ok := m.modByItemType[head]
	if !ok {
		return nil, fmt.Errorf("%w: cannot find modifier for %s", types.ErrInvalidType, itemType)
	}

	return transform(mod, val, tail)
}

var _ Modifier = &byItemTypeModifier{}
