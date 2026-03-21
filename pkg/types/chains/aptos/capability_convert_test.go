package aptos

import (
	"testing"

	"github.com/stretchr/testify/require"

	aptoscap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/aptos"
)

func TestViewPayloadFromCapability_Success(t *testing.T) {
	t.Parallel()

	payload, err := ViewPayloadFromCapability(&aptoscap.ViewPayload{
		Module: &aptoscap.ModuleID{
			Address: []byte{0x01},
			Name:    "coin",
		},
		Function: "name",
		ArgTypes: []*aptoscap.TypeTag{
			{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_VECTOR,
				Value: &aptoscap.TypeTag_Vector{
					Vector: &aptoscap.VectorTag{
						ElementType: &aptoscap.TypeTag{
							Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_U8,
						},
					},
				},
			},
			{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_STRUCT,
				Value: &aptoscap.TypeTag_Struct{
					Struct: &aptoscap.StructTag{
						Address: []byte{0x02},
						Module:  "m",
						Name:    "n",
						TypeParams: []*aptoscap.TypeTag{
							{
								Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_GENERIC,
								Value: &aptoscap.TypeTag_Generic{
									Generic: &aptoscap.GenericTag{Index: 7},
								},
							},
						},
					},
				},
			},
		},
		Args: [][]byte{{0xAA, 0xBB}},
	})
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, "coin", payload.Module.Name)
	require.Equal(t, "name", payload.Function)
	require.Len(t, payload.ArgTypes, 2)
	require.Equal(t, [][]byte{{0xAA, 0xBB}}, payload.Args)
}

func TestViewPayloadFromCapability_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	_, err := ViewPayloadFromCapability(nil)
	require.ErrorContains(t, err, "ViewRequest.Payload is required")

	_, err = ViewPayloadFromCapability(&aptoscap.ViewPayload{})
	require.ErrorContains(t, err, "ViewRequest.Payload.Module is required")

	_, err = ViewPayloadFromCapability(&aptoscap.ViewPayload{
		Module: &aptoscap.ModuleID{Address: []byte{1}, Name: "coin"},
	})
	require.ErrorContains(t, err, "ViewRequest.Payload.Function is required")

	_, err = ViewPayloadFromCapability(&aptoscap.ViewPayload{
		Module: &aptoscap.ModuleID{
			Address: make([]byte, AccountAddressLength+1),
			Name:    "coin",
		},
		Function: "name",
	})
	require.ErrorContains(t, err, "module address too long")
}

func TestViewPayloadFromCapability_RejectsBadTypeTags(t *testing.T) {
	t.Parallel()

	_, err := ViewPayloadFromCapability(&aptoscap.ViewPayload{
		Module:   &aptoscap.ModuleID{Address: []byte{1}, Name: "coin"},
		Function: "name",
		ArgTypes: []*aptoscap.TypeTag{nil},
	})
	require.ErrorContains(t, err, "invalid arg type at index 0")
	require.ErrorContains(t, err, "type tag is nil")

	_, err = ViewPayloadFromCapability(&aptoscap.ViewPayload{
		Module:   &aptoscap.ModuleID{Address: []byte{1}, Name: "coin"},
		Function: "name",
		ArgTypes: []*aptoscap.TypeTag{
			{
				Kind: aptoscap.TypeTagKind_TYPE_TAG_KIND_GENERIC,
				Value: &aptoscap.TypeTag_Generic{
					Generic: &aptoscap.GenericTag{Index: 1 << 20},
				},
			},
		},
	})
	require.ErrorContains(t, err, "generic type index out of range")
}
