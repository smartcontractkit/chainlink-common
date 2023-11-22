package codec

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func NewHardCoder(onChain map[string]any, offChain map[string]any, hooks ...mapstructure.DecodeHookFunc) Modifier {
	m := &onChainHardCoder{
		modifierBase: modifierBase[any]{
			fields:            offChain,
			onToOffChainType:  map[reflect.Type]reflect.Type{},
			offToOneChainType: map[reflect.Type]reflect.Type{},
		},
		onChain: onChain,
		hooks:   hooks,
	}
	m.modifyFieldForInput = func(field *reflect.StructField, v any) { field.Type = reflect.TypeOf(v) }
	m.addFieldForInput = func(key string, value any) reflect.StructField {
		return reflect.StructField{
			Name: key,
			Type: reflect.TypeOf(value),
		}
	}
	return m
}

type onChainHardCoder struct {
	modifierBase[any]
	onChain map[string]any
	hooks   []mapstructure.DecodeHookFunc
}

func (o onChainHardCoder) TransformForOnChain(offChainValue any) (any, error) {
	return nil, nil
}

func (o onChainHardCoder) TransformForOffChain(onChainValue any) (any, error) {
	return nil, nil
}
