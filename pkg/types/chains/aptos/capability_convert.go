package aptos

import (
	"fmt"

	aptoscap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/aptos"
)

// ViewPayloadFromCapability converts Aptos capability request payloads to Aptos relayer payloads.
func ViewPayloadFromCapability(payload *aptoscap.ViewPayload) (*ViewPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("ViewRequest.Payload is required")
	}
	if payload.Module == nil {
		return nil, fmt.Errorf("ViewRequest.Payload.Module is required")
	}
	if payload.Function == "" {
		return nil, fmt.Errorf("ViewRequest.Payload.Function is required")
	}
	if len(payload.Module.Address) > AccountAddressLength {
		return nil, fmt.Errorf("module address too long: %d", len(payload.Module.Address))
	}

	var moduleAddress AccountAddress
	copy(moduleAddress[AccountAddressLength-len(payload.Module.Address):], payload.Module.Address)

	argTypes := make([]TypeTag, 0, len(payload.ArgTypes))
	for i, tag := range payload.ArgTypes {
		converted, err := typeTagFromCapability(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid arg type at index %d: %w", i, err)
		}
		argTypes = append(argTypes, converted)
	}

	return &ViewPayload{
		Module: ModuleID{
			Address: moduleAddress,
			Name:    payload.Module.Name,
		},
		Function: payload.Function,
		ArgTypes: argTypes,
		Args:     payload.Args,
	}, nil
}

func typeTagFromCapability(tag *aptoscap.TypeTag) (TypeTag, error) {
	if tag == nil {
		return TypeTag{}, fmt.Errorf("type tag is nil")
	}

	switch tag.Kind {
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_BOOL:
		return TypeTag{Value: BoolTag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_U8:
		return TypeTag{Value: U8Tag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_U16:
		return TypeTag{Value: U16Tag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_U32:
		return TypeTag{Value: U32Tag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_U64:
		return TypeTag{Value: U64Tag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_U128:
		return TypeTag{Value: U128Tag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_U256:
		return TypeTag{Value: U256Tag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_ADDRESS:
		return TypeTag{Value: AddressTag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_SIGNER:
		return TypeTag{Value: SignerTag{}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_VECTOR:
		vector := tag.GetVector()
		if vector == nil {
			return TypeTag{}, fmt.Errorf("vector tag missing vector value")
		}
		elementType, err := typeTagFromCapability(vector.ElementType)
		if err != nil {
			return TypeTag{}, err
		}
		return TypeTag{Value: VectorTag{ElementType: elementType}}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT:
		structTag := tag.GetStruct()
		if structTag == nil {
			return TypeTag{}, fmt.Errorf("struct tag missing struct value")
		}
		if len(structTag.Address) > AccountAddressLength {
			return TypeTag{}, fmt.Errorf("struct address too long: %d", len(structTag.Address))
		}
		var structAddress AccountAddress
		copy(structAddress[AccountAddressLength-len(structTag.Address):], structTag.Address)

		typeParams := make([]TypeTag, 0, len(structTag.TypeParams))
		for i, tp := range structTag.TypeParams {
			converted, err := typeTagFromCapability(tp)
			if err != nil {
				return TypeTag{}, fmt.Errorf("invalid struct type param at index %d: %w", i, err)
			}
			typeParams = append(typeParams, converted)
		}

		return TypeTag{
			Value: StructTag{
				Address:    structAddress,
				Module:     structTag.Module,
				Name:       structTag.Name,
				TypeParams: typeParams,
			},
		}, nil
	case aptoscap.TypeTagKind_TYPE_TAG_KIND_GENERIC:
		generic := tag.GetGeneric()
		if generic == nil {
			return TypeTag{}, fmt.Errorf("generic tag missing generic value")
		}
		if generic.Index > 0xFFFF {
			return TypeTag{}, fmt.Errorf("generic type index out of range: %d", generic.Index)
		}
		return TypeTag{Value: GenericTag{Index: uint16(generic.Index)}}, nil
	default:
		return TypeTag{}, fmt.Errorf("unsupported type tag kind: %v", tag.Kind)
	}
}
