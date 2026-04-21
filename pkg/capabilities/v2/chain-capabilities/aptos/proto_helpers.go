package aptos

import (
	"fmt"
	"math"

	typesaptos "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

// ConvertViewPayloadFromProto converts a capability ViewPayload into Aptos domain types.
// Capability requests currently accept shortened Aptos addresses, so this helper left-pads
// addresses up to 32 bytes instead of requiring exact-length address bytes.
func ConvertViewPayloadFromProto(payload *ViewPayload) (*typesaptos.ViewPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is required")
	}
	if payload.Module == nil {
		return nil, fmt.Errorf("payload.module is required")
	}
	if len(payload.Module.Address) == 0 {
		return nil, fmt.Errorf("payload.module.address is required")
	}
	if payload.Module.Name == "" {
		return nil, fmt.Errorf("payload.module.name is required")
	}
	if payload.Function == "" {
		return nil, fmt.Errorf("payload.function is required")
	}

	moduleAddress, err := convertAccountAddressFromProto(payload.Module.Address, "module")
	if err != nil {
		return nil, err
	}

	argTypes := make([]typesaptos.TypeTag, 0, len(payload.ArgTypes))
	for i, tag := range payload.ArgTypes {
		converted, err := ConvertTypeTagFromProto(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid arg type at index %d: %w", i, err)
		}
		argTypes = append(argTypes, *converted)
	}

	return &typesaptos.ViewPayload{
		Module: typesaptos.ModuleID{
			Address: moduleAddress,
			Name:    payload.Module.Name,
		},
		Function: payload.Function,
		ArgTypes: argTypes,
		Args:     payload.Args,
	}, nil
}

// ConvertTypeTagFromProto converts a capability TypeTag into Aptos domain types.
func ConvertTypeTagFromProto(tag *TypeTag) (*typesaptos.TypeTag, error) {
	if tag == nil {
		return nil, fmt.Errorf("type tag is nil")
	}

	switch tag.Kind {
	case TypeTagKind_TYPE_TAG_KIND_BOOL:
		return &typesaptos.TypeTag{Value: typesaptos.BoolTag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_U8:
		return &typesaptos.TypeTag{Value: typesaptos.U8Tag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_U16:
		return &typesaptos.TypeTag{Value: typesaptos.U16Tag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_U32:
		return &typesaptos.TypeTag{Value: typesaptos.U32Tag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_U64:
		return &typesaptos.TypeTag{Value: typesaptos.U64Tag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_U128:
		return &typesaptos.TypeTag{Value: typesaptos.U128Tag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_U256:
		return &typesaptos.TypeTag{Value: typesaptos.U256Tag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_ADDRESS:
		return &typesaptos.TypeTag{Value: typesaptos.AddressTag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_SIGNER:
		return &typesaptos.TypeTag{Value: typesaptos.SignerTag{}}, nil
	case TypeTagKind_TYPE_TAG_KIND_VECTOR:
		vector := tag.GetVector()
		if vector == nil {
			return nil, fmt.Errorf("vector tag missing vector value")
		}
		elementType, err := ConvertTypeTagFromProto(vector.ElementType)
		if err != nil {
			return nil, fmt.Errorf("invalid vector element type: %w", err)
		}
		return &typesaptos.TypeTag{Value: typesaptos.VectorTag{ElementType: *elementType}}, nil
	case TypeTagKind_TYPE_TAG_KIND_STRUCT:
		structTag := tag.GetStruct()
		if structTag == nil {
			return nil, fmt.Errorf("struct tag missing struct value")
		}
		if len(structTag.Address) == 0 {
			return nil, fmt.Errorf("struct address is required")
		}
		if structTag.Module == "" {
			return nil, fmt.Errorf("struct module is required")
		}
		if structTag.Name == "" {
			return nil, fmt.Errorf("struct name is required")
		}

		structAddress, err := convertAccountAddressFromProto(structTag.Address, "struct")
		if err != nil {
			return nil, err
		}

		typeParams := make([]typesaptos.TypeTag, 0, len(structTag.TypeParams))
		for i, tp := range structTag.TypeParams {
			converted, err := ConvertTypeTagFromProto(tp)
			if err != nil {
				return nil, fmt.Errorf("invalid struct type param at index %d: %w", i, err)
			}
			typeParams = append(typeParams, *converted)
		}

		return &typesaptos.TypeTag{Value: typesaptos.StructTag{
			Address:    structAddress,
			Module:     structTag.Module,
			Name:       structTag.Name,
			TypeParams: typeParams,
		}}, nil
	case TypeTagKind_TYPE_TAG_KIND_GENERIC:
		generic := tag.GetGeneric()
		if generic == nil {
			return nil, fmt.Errorf("generic tag missing generic value")
		}
		if generic.Index > math.MaxUint16 {
			return nil, fmt.Errorf("generic type index out of range: %d", generic.Index)
		}
		return &typesaptos.TypeTag{Value: typesaptos.GenericTag{Index: uint16(generic.Index)}}, nil
	default:
		return nil, fmt.Errorf("unsupported type tag kind: %v", tag.Kind)
	}
}

func convertAccountAddressFromProto(address []byte, field string) (typesaptos.AccountAddress, error) {
	if len(address) > typesaptos.AccountAddressLength {
		return typesaptos.AccountAddress{}, fmt.Errorf("%s address too long: %d", field, len(address))
	}

	var converted typesaptos.AccountAddress
	copy(converted[typesaptos.AccountAddressLength-len(address):], address)
	return converted, nil
}
