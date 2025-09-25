package beholder_test

import (
	"fmt"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewChipClient(t *testing.T) {
	t.Run("returns error when client is nil", func(t *testing.T) {
		registry, err := beholder.NewChipIngressClient(nil)
		assert.Nil(t, registry)
		assert.EqualError(t, err, "chip ingress client is nil")
	})

	t.Run("returns schema registry when client is valid", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		registry, err := beholder.NewChipIngressClient(mockClient)
		require.NoError(t, err)
		assert.NotNil(t, registry)
	})
}

func TestRegisterSchema(t *testing.T) {
	t.Run("successfully registers schemas", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.
			On("RegisterSchema", mock.Anything, mock.Anything).
			Return(&pb.RegisterSchemaResponse{
				Registered: []*pb.RegisteredSchema{
					{Subject: "schema1", Version: 1},
					{Subject: "schema2", Version: 2},
				},
			}, nil)

		registry, err := beholder.NewChipIngressClient(mockClient)
		require.NoError(t, err)

		schemas := []*pb.Schema{
			{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"field1"}]}`, Format: 1},
			{Subject: "schema2", Schema: `{"type":"record","name":"Test2","fields":[{"name":"field2"}]}`, Format: 2},
		}

		result, err := registry.RegisterSchema(t.Context(), schemas...)
		require.NoError(t, err)
		assert.Equal(t, map[string]int{"schema1": 1, "schema2": 2}, result)
	})

	t.Run("returns error when registration fails", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.
			On("RegisterSchema", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("registration failed"))

		registry, err := beholder.NewChipIngressClient(mockClient)
		require.NoError(t, err)

		schemas := []*pb.Schema{
			{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"field1"}]}`, Format: 1},
		}

		result, err := registry.RegisterSchema(t.Context(), schemas...)
		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to register schema: registration failed")
	})
}
