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

	testCases := []struct {
		name    string
		payload *aptoscap.ViewPayload
		wantErr string
	}{
		{
			name:    "missing payload",
			payload: nil,
			wantErr: "payload is required",
		},
		{
			name:    "missing module",
			payload: &aptoscap.ViewPayload{Function: "name"},
			wantErr: "payload.module is required",
		},
		{
			name: "missing module address",
			payload: &aptoscap.ViewPayload{
				Module:   &aptoscap.ModuleID{Name: "coin"},
				Function: "name",
			},
			wantErr: "payload.module.address is required",
		},
		{
			name: "missing function",
			payload: &aptoscap.ViewPayload{
				Module: &aptoscap.ModuleID{Address: []byte{0x01}, Name: "coin"},
			},
			wantErr: "payload.function is required",
		},
		{
			name: "missing module name",
			payload: &aptoscap.ViewPayload{
				Module:   &aptoscap.ModuleID{Address: []byte{0x01}},
				Function: "name",
			},
			wantErr: "payload.module.name is required",
		},
		{
			name: "oversized module address",
			payload: &aptoscap.ViewPayload{
				Module:   &aptoscap.ModuleID{Address: make([]byte, typesaptos.AccountAddressLength+1), Name: "coin"},
				Function: "name",
			},
			wantErr: "module address too long",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := aptoscap.ConvertViewPayloadFromProto(tc.payload)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestConvertTypeTagFromProto_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		tag     *aptoscap.TypeTag
		wantErr string
	}{
		{
			name:    "nil type tag",
			tag:     nil,
			wantErr: "type tag is nil",
		},
		{
			name:    "unsupported kind",
			tag:     &aptoscap.TypeTag{Kind: aptoscap.TypeTagKind(255)},
			wantErr: "unsupported type tag kind",
		},
		{
			name: "missing struct address",
			tag: &aptoscap.TypeTag{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
				Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
					Module: "coin",
					Name:   "Coin",
				}},
			},
			wantErr: "struct address is required",
		},
		{
			name: "missing struct module",
			tag: &aptoscap.TypeTag{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
				Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
					Address: []byte{0x01},
					Name:    "Coin",
				}},
			},
			wantErr: "struct module is required",
		},
		{
			name: "missing struct name",
			tag: &aptoscap.TypeTag{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
				Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
					Address: []byte{0x01},
					Module:  "coin",
				}},
			},
			wantErr: "struct name is required",
		},
		{
			name: "oversized struct address",
			tag: &aptoscap.TypeTag{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
				Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
					Address: make([]byte, typesaptos.AccountAddressLength+1),
					Module:  "coin",
					Name:    "Coin",
				}},
			},
			wantErr: "struct address too long",
		},
		{
			name: "invalid vector element type",
			tag: &aptoscap.TypeTag{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_VECTOR,
				Value: &aptoscap.TypeTag_Vector{Vector: &aptoscap.VectorTag{
					ElementType: &aptoscap.TypeTag{
						Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
						Value: &aptoscap.TypeTag_Struct{Struct: &aptoscap.StructTag{
							Address: make([]byte, typesaptos.AccountAddressLength+1),
							Module:  "coin",
							Name:    "Coin",
						}},
					},
				}},
			},
			wantErr: "invalid vector element type: struct address too long",
		},
		{
			name: "generic index out of range",
			tag: &aptoscap.TypeTag{
				Kind:  aptoscap.TypeTagKind_TYPE_TAG_KIND_GENERIC,
				Value: &aptoscap.TypeTag_Generic{Generic: &aptoscap.GenericTag{Index: 1 << 16}},
			},
			wantErr: "generic type index out of range",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := aptoscap.ConvertTypeTagFromProto(tc.tag)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
