package aptos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	capaptos "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/aptos"
	conv "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	typeaptos "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

func TestConvertCapabilityViewPayloadFromProto(t *testing.T) {
	t.Run("converts nested vector struct and generic tags", func(t *testing.T) {
		payload, err := conv.ConvertCapabilityViewPayloadFromProto(&capaptos.ViewPayload{
			Module:   &capaptos.ModuleID{Address: []byte{0x01}, Name: "coin"},
			Function: "name",
			ArgTypes: []*capaptos.TypeTag{
				{
					Kind: capaptos.TypeTagKind_TYPE_TAG_KIND_VECTOR,
					Value: &capaptos.TypeTag_Vector{Vector: &capaptos.VectorTag{
						ElementType: &capaptos.TypeTag{
							Kind: capaptos.TypeTagKind_TYPE_TAG_KIND_STRUCT,
							Value: &capaptos.TypeTag_Struct{Struct: &capaptos.StructTag{
								Address: []byte{0x02},
								Module:  "aptos_coin",
								Name:    "Coin",
								TypeParams: []*capaptos.TypeTag{
									{
										Kind:  capaptos.TypeTagKind_TYPE_TAG_KIND_GENERIC,
										Value: &capaptos.TypeTag_Generic{Generic: &capaptos.GenericTag{Index: 7}},
									},
								},
							}},
						},
					}},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, payload)
		require.Equal(t, "name", payload.Function)
		require.Len(t, payload.ArgTypes, 1)

		vectorTag, ok := payload.ArgTypes[0].Value.(typeaptos.VectorTag)
		require.True(t, ok)
		structTag, ok := vectorTag.ElementType.Value.(typeaptos.StructTag)
		require.True(t, ok)
		require.Equal(t, "aptos_coin", structTag.Module)
		require.Equal(t, "Coin", structTag.Name)
		require.Len(t, structTag.TypeParams, 1)
		genericTag, ok := structTag.TypeParams[0].Value.(typeaptos.GenericTag)
		require.True(t, ok)
		require.EqualValues(t, 7, genericTag.Index)
	})

	t.Run("rejects nil payload", func(t *testing.T) {
		_, err := conv.ConvertCapabilityViewPayloadFromProto(nil)
		require.ErrorContains(t, err, "viewRequest.Payload is required")
	})

	t.Run("rejects nil module", func(t *testing.T) {
		_, err := conv.ConvertCapabilityViewPayloadFromProto(&capaptos.ViewPayload{
			Function: "name",
		})
		require.ErrorContains(t, err, "viewRequest.Payload.Module is required")
	})

	t.Run("rejects empty function", func(t *testing.T) {
		_, err := conv.ConvertCapabilityViewPayloadFromProto(&capaptos.ViewPayload{
			Module: &capaptos.ModuleID{Address: []byte{0x01}, Name: "coin"},
		})
		require.ErrorContains(t, err, "viewRequest.Payload.Function is required")
	})

	t.Run("rejects oversized module address", func(t *testing.T) {
		_, err := conv.ConvertCapabilityViewPayloadFromProto(&capaptos.ViewPayload{
			Module:   &capaptos.ModuleID{Address: make([]byte, typeaptos.AccountAddressLength+1), Name: "coin"},
			Function: "name",
		})
		require.ErrorContains(t, err, "module address too long")
	})
}

func TestConvertCapabilityTypeTagFromProto(t *testing.T) {
	t.Run("rejects nil tag", func(t *testing.T) {
		_, err := conv.ConvertCapabilityTypeTagFromProto(nil)
		require.ErrorContains(t, err, "type tag is nil")
	})

	t.Run("rejects unsupported kind", func(t *testing.T) {
		_, err := conv.ConvertCapabilityTypeTagFromProto(&capaptos.TypeTag{Kind: capaptos.TypeTagKind(255)})
		require.ErrorContains(t, err, "unsupported type tag kind")
	})

	t.Run("rejects oversized struct address", func(t *testing.T) {
		_, err := conv.ConvertCapabilityTypeTagFromProto(&capaptos.TypeTag{
			Kind: capaptos.TypeTagKind_TYPE_TAG_KIND_STRUCT,
			Value: &capaptos.TypeTag_Struct{Struct: &capaptos.StructTag{
				Address: make([]byte, typeaptos.AccountAddressLength+1),
			}},
		})
		require.ErrorContains(t, err, "struct address too long")
	})

	t.Run("rejects overflowing generic index", func(t *testing.T) {
		_, err := conv.ConvertCapabilityTypeTagFromProto(&capaptos.TypeTag{
			Kind:  capaptos.TypeTagKind_TYPE_TAG_KIND_GENERIC,
			Value: &capaptos.TypeTag_Generic{Generic: &capaptos.GenericTag{Index: 1 << 16}},
		})
		require.ErrorContains(t, err, "generic type index out of range")
	})
}
