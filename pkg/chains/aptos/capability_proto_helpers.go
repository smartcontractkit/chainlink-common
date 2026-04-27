package aptos

import (
	"fmt"
	"math"

	capaptos "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/aptos"
	typeaptos "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

// ConvertCapabilityViewPayloadFromProto converts Aptos capability proto input into the
// shared Aptos domain payload used by relayers and loop clients.
func ConvertCapabilityViewPayloadFromProto(payload *capaptos.ViewPayload) (*typeaptos.ViewPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("viewRequest.Payload is required")
	}
	if payload.Module == nil {
		return nil, fmt.Errorf("viewRequest.Payload.Module is required")
	}
	if payload.Function == "" {
		return nil, fmt.Errorf("viewRequest.Payload.Function is required")
	}
	if len(payload.Module.Address) > typeaptos.AccountAddressLength {
		return nil, fmt.Errorf("module address too long: %d", len(payload.Module.Address))
	}

	var moduleAddress typeaptos.AccountAddress
	copy(moduleAddress[typeaptos.AccountAddressLength-len(payload.Module.Address):], payload.Module.Address)

	argTypes := make([]typeaptos.TypeTag, 0, len(payload.ArgTypes))
	for i, tag := range payload.ArgTypes {
		converted, err := ConvertCapabilityTypeTagFromProto(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid arg type at index %d: %w", i, err)
		}
		argTypes = append(argTypes, *converted)
	}

	return &typeaptos.ViewPayload{
		Module: typeaptos.ModuleID{
			Address: moduleAddress,
			Name:    payload.Module.Name,
		},
		Function: payload.Function,
		ArgTypes: argTypes,
		Args:     payload.Args,
	}, nil
}

// ConvertCapabilityTypeTagFromProto converts Aptos capability proto type tags into
// the shared Aptos domain type tags used by relayers and loop clients.
func ConvertCapabilityTypeTagFromProto(tag *capaptos.TypeTag) (*typeaptos.TypeTag, error) {
	if tag == nil {
		return nil, fmt.Errorf("type tag is nil")
	}

	var impl typeaptos.TypeTagImpl
	switch tag.Kind {
	case capaptos.TypeTagKind_TYPE_TAG_KIND_BOOL:
		impl = typeaptos.BoolTag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_U8:
		impl = typeaptos.U8Tag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_U16:
		impl = typeaptos.U16Tag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_U32:
		impl = typeaptos.U32Tag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_U64:
		impl = typeaptos.U64Tag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_U128:
		impl = typeaptos.U128Tag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_U256:
		impl = typeaptos.U256Tag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_ADDRESS:
		impl = typeaptos.AddressTag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_SIGNER:
		impl = typeaptos.SignerTag{}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_VECTOR:
		vector := tag.GetVector()
		if vector == nil {
			return nil, fmt.Errorf("vector tag missing vector value")
		}
		elementType, err := ConvertCapabilityTypeTagFromProto(vector.ElementType)
		if err != nil {
			return nil, err
		}
		impl = typeaptos.VectorTag{ElementType: *elementType}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_STRUCT:
		structTag := tag.GetStruct()
		if structTag == nil {
			return nil, fmt.Errorf("struct tag missing struct value")
		}
		if len(structTag.Address) > typeaptos.AccountAddressLength {
			return nil, fmt.Errorf("struct address too long: %d", len(structTag.Address))
		}

		var structAddress typeaptos.AccountAddress
		copy(structAddress[typeaptos.AccountAddressLength-len(structTag.Address):], structTag.Address)

		typeParams := make([]typeaptos.TypeTag, 0, len(structTag.TypeParams))
		for i, tp := range structTag.TypeParams {
			converted, err := ConvertCapabilityTypeTagFromProto(tp)
			if err != nil {
				return nil, fmt.Errorf("invalid struct type param at index %d: %w", i, err)
			}
			typeParams = append(typeParams, *converted)
		}

		impl = typeaptos.StructTag{
			Address:    structAddress,
			Module:     structTag.Module,
			Name:       structTag.Name,
			TypeParams: typeParams,
		}
	case capaptos.TypeTagKind_TYPE_TAG_KIND_GENERIC:
		generic := tag.GetGeneric()
		if generic == nil {
			return nil, fmt.Errorf("generic tag missing generic value")
		}
		if generic.Index > math.MaxUint16 {
			return nil, fmt.Errorf("generic type index out of range: %d", generic.Index)
		}
		impl = typeaptos.GenericTag{Index: uint16(generic.Index)}
	default:
		return nil, fmt.Errorf("unsupported type tag kind: %v", tag.Kind)
	}

	return &typeaptos.TypeTag{Value: impl}, nil
}
