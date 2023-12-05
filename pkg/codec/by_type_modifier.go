package codec

import (
	"fmt"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func NewModifierByType(onChainTypeToMod map[reflect.Type]Modifier) (Modifier, error) {
	if onChainTypeToMod == nil {
		onChainTypeToMod = map[reflect.Type]Modifier{}
	}

	offChainTypeToMod := map[reflect.Type]Modifier{}
	for onChainType, mod := range onChainTypeToMod {
		offChainType, err := mod.RetypeForOffChain(onChainType)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", types.ErrInvalidConfig, err)
		}
		offChainTypeToMod[offChainType] = mod
	}

	return &byTypeModifier{
		onChainTypeToMod:  onChainTypeToMod,
		offChainTypeToMod: offChainTypeToMod,
	}, nil
}

type byTypeModifier struct {
	onChainTypeToMod  map[reflect.Type]Modifier
	offChainTypeToMod map[reflect.Type]Modifier
}

func (b *byTypeModifier) RetypeForOffChain(onChainType reflect.Type) (reflect.Type, error) {
	mod, ok := b.onChainTypeToMod[onChainType]
	if !ok {
		return nil, types.ErrInvalidType
	}

	return mod.RetypeForOffChain(onChainType)
}

func (b *byTypeModifier) TransformForOnChain(offChainValue any) (any, error) {
	mod, ok := b.offChainTypeToMod[reflect.TypeOf(offChainValue)]
	if !ok {
		return nil, types.ErrInvalidType
	}

	return mod.TransformForOnChain(offChainValue)
}

func (b *byTypeModifier) TransformForOffChain(onChainValue any) (any, error) {
	mod, ok := b.onChainTypeToMod[reflect.TypeOf(onChainValue)]
	if !ok {
		return nil, types.ErrInvalidType
	}

	return mod.TransformForOffChain(onChainValue)
}

var _ Modifier = &byTypeModifier{}
