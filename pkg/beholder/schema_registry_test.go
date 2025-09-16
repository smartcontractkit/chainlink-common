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

func TestNewSchemaRegistry(t *testing.T) {
	t.Run("returns error when client is nil", func(t *testing.T) {
		registry, err := beholder.NewSchemaRegistry(nil)
		assert.Nil(t, registry)
		assert.EqualError(t, err, "chip ingress client is nil")
	})

	t.Run("returns schema registry when client is valid", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		registry, err := beholder.NewSchemaRegistry(mockClient)
		require.NoError(t, err)
		assert.NotNil(t, registry)
	})
}

func TestSchemaRegistry_Register(t *testing.T) {
	t.Run("successfully registers schemas", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.
			On("RegisterSchema", mock.Anything, mock.Anything).
			Return(&pb.RegisterSchemaResponse{}, nil)
		registry, err := beholder.NewSchemaRegistry(mockClient)
		require.NoError(t, err)

		schemas := []*pb.Schema{
			{Subject: "schema1", Schema: `{"type":"record","name":"Test","fields":[{"name":"jeff"}]}`, Format: 1},
			{Subject: "schema1", Schema: `{"name":"jeff"}`, Format: 2},
		}
		err = registry.Register(t.Context(), schemas...)
		assert.NoError(t, err)
	})

	t.Run("returns error when registration fails", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.
			On("RegisterSchema", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("registration failed"))
		registry, err := beholder.NewSchemaRegistry(mockClient)
		require.NoError(t, err)

		schemas := []*pb.Schema{
			{Subject: "schema1", Schema: `{"name":"jeff"}`, Format: 2},
		}
		err = registry.Register(t.Context(), schemas...)
		assert.EqualError(t, err, "failed to register schema: registration failed")
	})
}
