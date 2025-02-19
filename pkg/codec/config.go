package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-viper/mapstructure/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// ModifiersConfig unmarshalls as a list of [ModifierConfig] by using a field called Type
// The values available for Type are case-insensitive and the config they require are below:
// - rename -> [RenameModifierConfig]
// - drop -> [DropModifierConfig]
// - hard code -> [HardCodeModifierConfig]
// - extract element -> [ElementExtractorModifierConfig]
// - extract element from onchain slice or array -> [ElementExtractorFromOnchainModifierConfig]
// - epoch to time -> [EpochToTimeModifierConfig]
// - bytes to boolean -> [ByteToBooleanModifierConfig]
// - address to string -> [AddressBytesToStringModifierConfig]
// - field wrapper -> [WrapperModifierConfig]
// - precodec -> [PreCodecModifierConfig]
type ModifiersConfig []ModifierConfig

func (m *ModifiersConfig) UnmarshalJSON(data []byte) error {
	var rawDeserialized []json.RawMessage
	if err := decode(data, &rawDeserialized); err != nil {
		return err
	}

	*m = make([]ModifierConfig, len(rawDeserialized))

	for i, d := range rawDeserialized {
		t := typer{}
		if err := decode(d, &t); err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidConfig, err)
		}

		mType := ModifierType(strings.ToLower(t.Type))
		switch mType {
		case ModifierRename:
			(*m)[i] = &RenameModifierConfig{}
		case ModifierDrop:
			(*m)[i] = &DropModifierConfig{}
		case ModifierHardCode:
			(*m)[i] = &HardCodeModifierConfig{}
		case ModifierExtractElement:
			(*m)[i] = &ElementExtractorModifierConfig{}
		case ModifierEpochToTime:
			(*m)[i] = &EpochToTimeModifierConfig{}
		case ModifierExtractProperty:
			(*m)[i] = &PropertyExtractorConfig{}
		case ModifierAddressToString:
			(*m)[i] = &AddressBytesToStringModifierConfig{}
		case ModifierBytesToString:
			(*m)[i] = &ConstrainedBytesToStringModifierConfig{}
		case ModifierWrapper:
			(*m)[i] = &WrapperModifierConfig{}
		case ModifierPreCodec:
			(*m)[i] = &PreCodecModifierConfig{}
		case ModifierByteToBoolean:
			(*m)[i] = &ByteToBooleanModifierConfig{}
		case ModifierExtractElementFromOnchain:
			(*m)[i] = &ElementExtractorFromOnchainModifierConfig{}
		default:
			return fmt.Errorf("%w: unknown modifier type: %s", types.ErrInvalidConfig, mType)
		}

		if err := decode(d, (*m)[i]); err != nil {
			return fmt.Errorf("%w: %w", types.ErrInvalidConfig, err)
		}
	}
	return nil
}

func (m *ModifiersConfig) ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error) {
	modifier := make(MultiModifier, len(*m))
	for i, c := range *m {
		mod, err := c.ToModifier(onChainHooks...)
		if err != nil {
			return nil, err
		}
		modifier[i] = mod
	}
	return modifier, nil
}

type ModifierType string

const (
	ModifierPreCodec                  ModifierType = "precodec"
	ModifierRename                    ModifierType = "rename"
	ModifierDrop                      ModifierType = "drop"
	ModifierHardCode                  ModifierType = "hard code"
	ModifierExtractElement            ModifierType = "extract element"
	ModifierExtractElementFromOnchain ModifierType = "extract element from onchain"
	ModifierByteToBoolean             ModifierType = "byte to boolean"
	ModifierEpochToTime               ModifierType = "epoch to time"
	ModifierExtractProperty           ModifierType = "extract property"
	ModifierAddressToString           ModifierType = "address to string"
	ModifierBytesToString             ModifierType = "constrained bytes to string"
	ModifierWrapper                   ModifierType = "wrapper"
)

type ModifierConfig interface {
	ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error)
}

// RenameModifierConfig renames all fields in the map from the key to the value
// The casing of the first character is ignored to allow compatibility
// of go convention for public fields and on-chain names.
type RenameModifierConfig struct {
	Fields             map[string]string
	EnablePathTraverse bool
}

func (r *RenameModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	for k, v := range r.Fields {
		delete(r.Fields, k)
		r.Fields[upperFirstCharacter(k)] = upperFirstCharacter(v)
	}

	return NewPathTraverseRenamer(r.Fields, r.EnablePathTraverse), nil
}

func (r *RenameModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[RenameModifierConfig]{
		Type: ModifierRename,
		T:    r,
	})
}

// DropModifierConfig drops all fields listed.  The casing of the first character is ignored to allow compatibility.
// Note that unused fields are ignored by [types.Codec].
// This is only required if you want to rename a field to an already used name.
// For example, if a struct has fields A and B, and you want to rename A to B,
// then you need to either also rename B or drop it.
type DropModifierConfig struct {
	Fields             []string
	EnablePathTraverse bool
}

func (d *DropModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	fields := map[string]string{}
	for i, f := range d.Fields {
		// using a private variable will make the field not serialize, essentially dropping the field
		fields[upperFirstCharacter(f)] = fmt.Sprintf("dropFieldPrivateName%d", i)
	}

	return NewPathTraverseRenamer(fields, d.EnablePathTraverse), nil
}

func (d *DropModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[DropModifierConfig]{
		Type: ModifierDrop,
		T:    d,
	})
}

// ByteToBooleanModifierConfig converts onchain uint8 fields to offchain bool fields and vice versa.
type ByteToBooleanModifierConfig struct {
	Fields []string
}

func (d *ByteToBooleanModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewByteToBooleanModifier(d.Fields), nil
}

func (d *ByteToBooleanModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[ByteToBooleanModifierConfig]{
		Type: ModifierByteToBoolean,
		T:    d,
	})
}

// ElementExtractorModifierConfig is used to extract an element from a slice or array
type ElementExtractorModifierConfig struct {
	// Key is the name of the field to extract from and the value is which element to extract.
	Extractions map[string]*ElementExtractorLocation
}

func (e *ElementExtractorModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	mapKeyToUpperFirst(e.Extractions)
	return NewElementExtractor(e.Extractions), nil
}

func (e *ElementExtractorModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[ElementExtractorModifierConfig]{
		Type: ModifierExtractElement,
		T:    e,
	})
}

// ElementExtractorFromOnchainModifierConfig is used to extract an element from an onchain slice or array.
type ElementExtractorFromOnchainModifierConfig struct {
	// Key is the name of the field to extract from and the value is which element to extract.
	Extractions map[string]*ElementExtractorLocation
}

func (e *ElementExtractorFromOnchainModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	mapKeyToUpperFirst(e.Extractions)
	return NewElementExtractorFromOnchain(e.Extractions), nil
}

func (e *ElementExtractorFromOnchainModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[ElementExtractorFromOnchainModifierConfig]{
		Type: ModifierExtractElementFromOnchain,
		T:    e,
	})
}

// HardCodeModifierConfig is used to hard code values into the map.
// Note that hard-coding values will override other values.
type HardCodeModifierConfig struct {
	OnChainValues      map[string]any
	OffChainValues     map[string]any
	EnablePathTraverse bool
}

func (h *HardCodeModifierConfig) ToModifier(onChainHooks ...mapstructure.DecodeHookFunc) (Modifier, error) {
	for key, value := range h.OnChainValues {
		number, ok := value.(json.Number)
		if ok {
			h.OnChainValues[key] = Number(number)
		}
	}

	for key, value := range h.OffChainValues {
		number, ok := value.(json.Number)
		if ok {
			h.OffChainValues[key] = Number(number)
		}
	}

	mapKeyToUpperFirst(h.OnChainValues)
	mapKeyToUpperFirst(h.OffChainValues)

	return NewPathTraverseHardCoder(h.OnChainValues, h.OffChainValues, h.EnablePathTraverse, onChainHooks...)
}

func (h *HardCodeModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[HardCodeModifierConfig]{
		Type: ModifierHardCode,
		T:    h,
	})
}

// PreCodec creates a modifier that will transform data using a preliminary encoding/decoding step.
// 'Off-chain' values will be overwritten with the encoded data as a byte array.
// 'On-chain' values will be typed using the optimistic types from the codec.
// This is useful when wanting to move the data as generic bytes.
//
//				Example:
//
//				Based on this input struct:
//					type example struct {
//						A []B
//					}
//
//					type B struct {
//						C string
//						D string
//					}
//
//				And the fields config defined as:
//			 		{"A": "string C, string D"}
//
//				The codec config gives a map of strings (the values from fields config map) to implementation for encoding/decoding
//
//		           RemoteCodec {
//		              func (types.TypeProvider) CreateType(itemType string, forEncoding bool) (any, error)
//		              func (types.Decoder) Decode(ctx context.Context, raw []byte, into any, itemType string) error
//		              func (types.Encoder) Encode(ctx context.Context, item any, itemType string) ([]byte, error)
//		              func (types.Decoder) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error)
//		              func (types.Encoder) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error)
//		           }
//
//		  		   {"string C, string D": RemoteCodec}
//
//				Result:
//					type example struct {
//						A [][]bytes
//					}
//
//	             Where []bytes are the encoded input struct B
type PreCodecModifierConfig struct {
	// A map of a path of properties to encoding scheme.
	// If the path leads to an array, encoding will occur on every entry.
	//
	// Example: "a.b" -> "uint256 Value"
	Fields             map[string]string
	EnablePathTraverse bool
	// Codecs is skipped in JSON serialization, it will be injected later.
	// The map should be keyed using the value from "Fields" to a corresponding Codec that can encode/decode for it
	// This allows encoding and decoding implementations to be handled outside of the modifier.
	//
	// Example: "uint256 Value" -> a chain specific encoder for "uint256 Value"
	Codecs map[string]types.RemoteCodec `json:"-"`
}

func (c *PreCodecModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewPathTraversePreCodec(c.Fields, c.Codecs, c.EnablePathTraverse)
}

func (c *PreCodecModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[PreCodecModifierConfig]{
		Type: ModifierPreCodec,
		T:    c,
	})
}

// EpochToTimeModifierConfig is used to convert epoch seconds as uint64 fields on-chain to time.Time
type EpochToTimeModifierConfig struct {
	Fields             []string
	EnablePathTraverse bool
}

func (e *EpochToTimeModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	for i, f := range e.Fields {
		e.Fields[i] = upperFirstCharacter(f)
	}
	return NewPathTraverseEpochToTimeModifier(e.Fields, e.EnablePathTraverse), nil
}

func (e *EpochToTimeModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[EpochToTimeModifierConfig]{
		Type: ModifierEpochToTime,
		T:    e,
	})
}

type PropertyExtractorConfig struct {
	FieldName string
}

func (c *PropertyExtractorConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewPropertyExtractor(upperFirstCharacter(c.FieldName)), nil
}

func (c *PropertyExtractorConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[PropertyExtractorConfig]{
		Type: ModifierExtractProperty,
		T:    c,
	})
}

// AddressBytesToStringModifierConfig is used to transform address byte fields into string fields.
// It holds the list of fields that should be modified and the chain-specific logic to do the modifications.
type AddressBytesToStringModifierConfig struct {
	Fields             []string
	EnablePathTraverse bool
	// Modifier is skipped in JSON serialization, will be injected later.
	Modifier AddressModifier `json:"-"`
}

func (c *AddressBytesToStringModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewPathTraverseAddressBytesToStringModifier(c.Fields, c.Modifier, c.EnablePathTraverse), nil
}

func (c *AddressBytesToStringModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[AddressBytesToStringModifierConfig]{
		Type: ModifierAddressToString,
		T:    c,
	})
}

type ConstrainedBytesToStringModifierConfig struct {
	Fields             []string
	MaxLen             int
	EnablePathTraverse bool
}

func (c *ConstrainedBytesToStringModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	return NewPathTraverseConstrainedLengthBytesToStringModifier(c.Fields, c.MaxLen, c.EnablePathTraverse), nil
}

func (c *ConstrainedBytesToStringModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[ConstrainedBytesToStringModifierConfig]{
		Type: ModifierBytesToString,
		T:    c,
	})
}

// WrapperModifierConfig replaces each field based on cfg map keys with a struct containing one field with the value of the original field which has is named based on map values.
// Wrapper modifier does not maintain the original pointers.
// Wrapper modifier config shouldn't edit fields that affect each other since the results are not deterministic.
//
//		Example #1:
//
//		Based on this input struct:
//			type example struct {
//				A string
//			}
//
//		And the wrapper config defined as:
//	 		{"D": "W"}
//
//		Result:
//			type example struct {
//				D
//			}
//
//		where D is a struct that contains the original value of D under the name W:
//			type D struct {
//				W string
//			}
//
//
//		Example #2:
//		Wrapper modifier works on any type of field, including nested fields or nested fields in slices etc.!
//
//		Based on this input struct:
//			type example struct {
//				A []B
//			}
//
//			type B struct {
//				C string
//				D string
//			}
//
//		And the wrapper config defined as:
//	 		{"A.C": "E", "A.D": "F"}
//
//		Result:
//			type example struct {
//				A []B
//			}
//
//			type B struct {
//				C type struct { E string }
//				D type struct { F string }
//			}
//
//		Where each element of slice A under fields C.E and D.F retains the values of their respective input slice elements A.C and A.D .
type WrapperModifierConfig struct {
	// Fields key defines the fields to be wrapped and the name of the wrapper struct.
	// The field becomes a subfield of the wrapper struct where the name of the subfield is map value.
	Fields             map[string]string
	EnablePathTraverse bool
}

func (r *WrapperModifierConfig) ToModifier(_ ...mapstructure.DecodeHookFunc) (Modifier, error) {
	fields := map[string]string{}
	for i, f := range r.Fields {
		// using a private variable will make the field not serialize, essentially dropping the field
		fields[upperFirstCharacter(f)] = fmt.Sprintf("dropFieldPrivateName-%s", i)
	}
	return NewPathTraverseWrapperModifier(r.Fields, r.EnablePathTraverse), nil
}

func (r *WrapperModifierConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&modifierMarshaller[WrapperModifierConfig]{
		Type: ModifierWrapper,
		T:    r,
	})
}

type typer struct {
	Type string
}

func upperFirstCharacter(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}

	return strings.Join(parts, ".")
}

func mapKeyToUpperFirst[T any](m map[string]T) {
	for k, v := range m {
		delete(m, k)
		m[upperFirstCharacter(k)] = v
	}
}

type modifierMarshaller[T any] struct {
	Type ModifierType
	T    *T
}

func (h *modifierMarshaller[T]) MarshalJSON() ([]byte, error) {
	v := reflect.Indirect(reflect.ValueOf(h.T))
	t := v.Type()

	m := map[string]interface{}{
		"Type": h.Type,
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		m[field.Name] = value
	}

	return json.Marshal(m)
}

func decode(bts []byte, val any) error {
	decoder := json.NewDecoder(bytes.NewBuffer(bts))
	decoder.UseNumber()

	return decoder.Decode(val)
}

type Number string

func (n Number) Float64() (float64, error) {
	return json.Number(n).Float64()
}

func (n Number) Int64() (int64, error) {
	return json.Number(n).Int64()
}

func (n Number) MarshalCBOR() ([]byte, error) {
	if strings.Contains(string(n), ".") {
		// parse as float64 and encode
		floatVal, err := strconv.ParseFloat(string(n), 64)
		if err != nil {
			return nil, err
		}

		return cbor.Marshal(floatVal)
	}

	// parse as int64 and encode
	intVal, err := strconv.ParseInt(string(n), 10, 64)
	if err != nil {
		return nil, err
	}

	return cbor.Marshal(intVal)
}

func (n *Number) UnmarshalCBOR(data []byte) error {
	var value string
	if err := cbor.Unmarshal(data, &value); err != nil {
		return err
	}

	*n = Number(value)

	return nil
}

func (n Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.Number(n))
}

func (n *Number) UnmarshalJSON(data []byte) error {
	var value json.Number
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*n = Number(value)

	return nil
}
