package codec

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// NewPreCodec creates a modifier that will run a preliminary encoding/decoding step. This is useful when wanting to
// move nested data as generic bytes.
func NewPreCodec(
	fields map[string]string,
	codecs map[string]types.RemoteCodec,
) (Modifier, error) {
	return NewPathTraversePreCodec(fields, codecs, false)
}

// NewPathTraversePreCodec creates a PreCodec modifier with itemType path traversal enabled or disabled. The standard
// constructor. NewPreCodec has path traversal off by default.
func NewPathTraversePreCodec(
	fields map[string]string,
	codecs map[string]types.RemoteCodec,
	enablePathTraverse bool,
) (Modifier, error) {
	m := &preCodec{
		modifierBase: modifierBase[string]{
			enablePathTraverse: enablePathTraverse,
			fields:             fields,
			onToOffChainType:   map[reflect.Type]reflect.Type{},
			offToOnChainType:   map[reflect.Type]reflect.Type{},
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
		if field.Type != reflect.SliceOf(reflect.TypeFor[uint8]()) && field.Type != reflect.PointerTo(reflect.SliceOf(reflect.TypeFor[uint8]())) {
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

func (pc *preCodec) TransformToOffChain(onChainValue any, itemType string) (any, error) {
	// set itemType to an ignore value if path traversal is not enabled
	if !pc.modifierBase.enablePathTraverse {
		itemType = ""
	}

	allHooks := make([]mapstructure.DecodeHookFunc, 1)
	allHooks[0] = hardCodeManyHook

	// the offChainValue might be a subfield value; get the true offChainStruct type already stored and set the value
	onChainStructValue := onChainValue

	// path traversal is expected, but offChainValue is the value of a field, not the actual struct
	// create a new struct from the stored offChainStruct with the provided value applied and all other fields set to
	// their zero value.
	if itemType != "" {
		into := reflect.New(pc.onChainStructType)

		if err := SetValueAtPath(into, reflect.ValueOf(onChainValue), itemType); err != nil {
			return nil, err
		}

		onChainStructValue = reflect.Indirect(into).Interface()
	}

	modified, err := transformWithMaps(onChainStructValue, pc.onToOffChainType, pc.fields, pc.decodeFieldMapAction, allHooks...)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return valueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
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
	raw, ok := extractMap[key].([]byte)
	if !ok {
		return fmt.Errorf("expected field %s to be []byte but got %T", key, extractMap[key])
	}
	err = codec.Decode(context.Background(), raw, &to, "")
	if err != nil {
		return err
	}
	extractMap[key] = to
	return nil
}

func (pc *preCodec) TransformToOnChain(offChainValue any, itemType string) (any, error) {
	allHooks := make([]mapstructure.DecodeHookFunc, 1)
	allHooks[0] = hardCodeManyHook

	// set itemType to an ignore value if path traversal is not enabled
	if !pc.modifierBase.enablePathTraverse {
		itemType = ""
	}

	// the offChainValue might be a subfield value; get the true offChainStruct type already stored and set the value
	offChainStructValue := offChainValue

	// path traversal is expected, but offChainValue is the value of a field, not the actual struct
	// create a new struct from the stored offChainStruct with the provided value applied and all other fields set to
	// their zero value.
	if itemType != "" {
		into := reflect.New(pc.offChainStructType)

		if err := SetValueAtPath(into, reflect.ValueOf(offChainValue), itemType); err != nil {
			return nil, err
		}

		offChainStructValue = reflect.Indirect(into).Interface()
	}

	modified, err := transformWithMaps(offChainStructValue, pc.offToOnChainType, pc.fields, pc.encodeFieldMapAction, allHooks...)
	if err != nil {
		return nil, err
	}

	if itemType != "" {
		return valueForPath(reflect.ValueOf(modified), itemType)
	}

	return modified, nil
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
