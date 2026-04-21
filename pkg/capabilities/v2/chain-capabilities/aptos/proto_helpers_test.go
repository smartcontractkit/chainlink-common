package aptos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	aptoscap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/aptos"
	typesaptos "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

func TestConvertViewPayloadFromProto_ConvertsNestedVectorStructAndGenericTags(t *testing.T) {
	t.Parallel()

	payload, err := aptoscap.ConvertViewPayloadFromProto(&aptoscap.ViewPayload{
		Module:   &aptoscap.ModuleID{Address: []byte{0x01}, Name: "coin"},
		Function: "name",
		ArgTypes: []*aptoscap.TypeTag{
			{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_VECTOR,
				Value: &aptoscap.TypeTag_Vector{Vector: &aptoscap.VectorTag{
					ElementType: &aptoscap.TypeTag{
						Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
						Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
							Address: []byte{0x02},
							Module:  "aptos_coin",
							Name:    "Coin",
							TypeParams: []*aptoscap.TypeTag{
								{
									Kind:  aptoscap.TypeTagKind_TYPE_TAG_KIND_GENERIC,
									Value: &aptoscap.TypeTag_Generic{Generic: &aptoscap.GenericTag{Index: 7}},
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

	vectorTag, ok := payload.ArgTypes[0].Value.(typesaptos.VectorTag)
	require.True(t, ok)
	structTag, ok := vectorTag.ElementType.Value.(typesaptos.StructTag)
	require.True(t, ok)
	require.Equal(t, "aptos_coin", structTag.Module)
	require.Equal(t, "Coin", structTag.Name)
	require.Len(t, structTag.TypeParams, 1)
	genericTag, ok := structTag.TypeParams[0].Value.(typesaptos.GenericTag)
	require.True(t, ok)
	require.EqualValues(t, 7, genericTag.Index)
}

func TestConvertViewPayloadFromProto_RejectsInvalidPayloadInputs(t *testing.T) {
	t.Parallel()

	_, err := aptoscap.ConvertViewPayloadFromProto(nil)
	require.ErrorContains(t, err, "viewRequest.Payload is required")

	_, err = aptoscap.ConvertViewPayloadFromProto(&aptoscap.ViewPayload{Function: "name"})
	require.ErrorContains(t, err, "viewRequest.Payload.Module is required")

	_, err = aptoscap.ConvertViewPayloadFromProto(&aptoscap.ViewPayload{
		Module: &aptoscap.ModuleID{Address: []byte{0x01}, Name: "coin"},
	})
	require.ErrorContains(t, err, "viewRequest.Payload.Function is required")

	_, err = aptoscap.ConvertViewPayloadFromProto(&aptoscap.ViewPayload{
		Module:   &aptoscap.ModuleID{Address: make([]byte, typesaptos.AccountAddressLength+1), Name: "coin"},
		Function: "name",
	})
	require.ErrorContains(t, err, "module address too long")
}

func TestConvertTypeTagFromProto_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	_, err := aptoscap.ConvertTypeTagFromProto(nil)
	require.ErrorContains(t, err, "type tag is nil")

	_, err = aptoscap.ConvertTypeTagFromProto(&aptoscap.TypeTag{Kind: aptoscap.TypeTagKind(255)})
	require.ErrorContains(t, err, "unsupported type tag kind")

	_, err = aptoscap.ConvertTypeTagFromProto(&aptoscap.TypeTag{
		Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
		Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
			Address: make([]byte, typesaptos.AccountAddressLength+1),
		}},
	})
	require.ErrorContains(t, err, "struct address too long")

	_, err = aptoscap.ConvertTypeTagFromProto(&aptoscap.TypeTag{
		Kind:  aptoscap.TypeTagKind_TYPE_TAG_KIND_GENERIC,
		Value: &aptoscap.TypeTag_Generic{Generic: &aptoscap.GenericTag{Index: 1 << 16}},
	})
	require.ErrorContains(t, err, "generic type index out of range")
}
