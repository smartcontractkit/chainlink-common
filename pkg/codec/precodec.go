package codec

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// PreCodec creates a modifier that will run a preliminary encoding/decoding step.
// This is useful when wanting to move nested data as generic bytes.
func NewPreCodec(fields map[string]string, codecs map[string]types.RemoteCodec) (Modifier, error) {
	m := &preCodec{
		modifierBase: modifierBase[string]{
			fields:           fields,
			onToOffChainType: map[reflect.Type]reflect.Type{},
			offToOnChainType: map[reflect.Type]reflect.Type{},
		},
		codecs: codecs,
	}

	// validate that there is a codec for each unique type definition
	for _, typeDef := range fields {
		if _, ok := m.codecs[typeDef]; ok {
			continue
		}
		return nil, fmt.Errorf("codec not supplied for: %s", typeDef)
	}

	m.modifyFieldForInput = func(_ string, field *reflect.StructField, _ string, typeDef string) error {
		if field.Type != reflect.SliceOf(reflect.TypeFor[uint8]()) {
			return fmt.Errorf("can only decode []byte from on-chain: %s", field.Type)
		}

		codec, ok := m.codecs[typeDef]
		if !ok || codec == nil {
			return fmt.Errorf("codec not found for type definition: '%s'", typeDef)
		}

		newType, err := codec.CreateType("", false)
		if err != nil {
			return err
		}
		field.Type = reflect.TypeOf(newType)

		return nil
	}

	return m, nil
}

type preCodec struct {
	modifierBase[string]
	codecs map[string]types.RemoteCodec
}

func (pc *preCodec) TransformToOffChain(onChainValue any, _ string) (any, error) {
	allHooks := make([]mapstructure.DecodeHookFunc, 1)
	allHooks[0] = hardCodeManyHook

	return transformWithMaps(onChainValue, pc.onToOffChainType, pc.fields, pc.decodeFieldMapAction, allHooks...)
}

func (pc *preCodec) decodeFieldMapAction(extractMap map[string]any, key string, typeDef string) error {
	_, exists := extractMap[key]
	if !exists {
		return fmt.Errorf("field %s does not exist", key)
	}

	codec, ok := pc.codecs[typeDef]
	if !ok || codec == nil {
		return fmt.Errorf("codec not found for type definition: '%s'", typeDef)
	}

	to, err := codec.CreateType("", false)
	if err != nil {
		return err
	}
	err = codec.Decode(context.Background(), extractMap[key].([]byte), &to, "")
	if err != nil {
		return err
	}
	extractMap[key] = to
	return nil
}

func (pc *preCodec) TransformToOnChain(offChainValue any, _ string) (any, error) {
	allHooks := make([]mapstructure.DecodeHookFunc, 1)
	allHooks[0] = hardCodeManyHook

	return transformWithMaps(offChainValue, pc.offToOnChainType, pc.fields, pc.encodeFieldMapAction, allHooks...)
}

func (pc *preCodec) encodeFieldMapAction(extractMap map[string]any, key string, typeDef string) error {
	_, exists := extractMap[key]
	if !exists {
		return fmt.Errorf("field %s does not exist", key)
	}

	codec, ok := pc.codecs[typeDef]
	if !ok || codec == nil {
		return fmt.Errorf("codec not found for type definition: '%s'", typeDef)
	}

	encoded, err := codec.Encode(context.Background(), extractMap[key], "")
	if err != nil {
		return err
	}
	extractMap[key] = encoded
	return nil
}
