package registry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type mockCapability struct {
	capabilities.CapabilityInfo
}

func (m *mockCapability) Execute(_ context.Context, _ capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	return capabilities.CapabilityResponse{}, nil
}

func (m *mockCapability) RegisterToWorkflow(_ context.Context, _ capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (m *mockCapability) UnregisterFromWorkflow(_ context.Context, _ capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func TestRegistry(t *testing.T) {
	r := registry.NewBase(logger.Test(t))
	ctx := t.Context()

	id := "capability-1@1.0.0"
	ci, err := capabilities.NewCapabilityInfo(
		id,
		capabilities.CapabilityTypeAction,
		"capability-1-description",
	)
	require.NoError(t, err)

	c := &mockCapability{CapabilityInfo: ci}
	err = r.Add(ctx, c)
	require.NoError(t, err)

	gc, err := r.Get(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, c, gc)

	cs, err := r.List(ctx)
	require.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, c, cs[0])
}

func TestRegistry_NoDuplicateIDs(t *testing.T) {
	r := registry.NewBase(logger.Test(t))
	ctx := t.Context()

	id := "capability-1@1.0.0"
	ci, err := capabilities.NewCapabilityInfo(
		id,
		capabilities.CapabilityTypeAction,
		"capability-1-description",
	)
	require.NoError(t, err)

	c := &mockCapability{CapabilityInfo: ci}
	err = r.Add(ctx, c)
	require.NoError(t, err)

	ci, err = capabilities.NewCapabilityInfo(
		id,
		capabilities.CapabilityTypeConsensus,
		"capability-2-description",
	)
	require.NoError(t, err)
	c2 := &mockCapability{CapabilityInfo: ci}

	err = r.Add(ctx, c2)
	assert.ErrorIs(t, err, registry.ErrCapabilityAlreadyExists)
}

func TestRegistry_ChecksExecutionAPIByType(t *testing.T) {
	tcs := []struct {
		name          string
		newCapability func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error)
		getCapability func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error
		errContains   string
	}{
		{
			name: "action",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				id := fmt.Sprintf("%s@%s", uuid.New().String(), "1.0.0")
				ci, err := capabilities.NewCapabilityInfo(
					id,
					capabilities.CapabilityTypeAction,
					"capability-1-description",
				)
				require.NoError(t, err)

				c := &mockCapability{CapabilityInfo: ci}
				return id, reg.Add(ctx, c)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetExecutable(ctx, id)
				return err
			},
		},
		{
			name: "target",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				id := fmt.Sprintf("%s@%s", uuid.New().String(), "1.0.0")
				ci, err := capabilities.NewCapabilityInfo(
					id,
					capabilities.CapabilityTypeTarget,
					"capability-1-description",
				)
				require.NoError(t, err)

				c := &mockCapability{CapabilityInfo: ci}
				return id, reg.Add(ctx, c)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetExecutable(ctx, id)
				return err
			},
		},
		{
			name: "trigger",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				odt := triggers.NewOnDemand(logger.Test(t))
				info, err := odt.Info(ctx)
				require.NoError(t, err)
				return info.ID, reg.Add(ctx, odt)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetTrigger(ctx, id)
				return err
			},
		},
		{
			name: "consensus",
			newCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase) (string, error) {
				id := fmt.Sprintf("%s@%s", uuid.New().String(), "1.0.0")
				ci, err := capabilities.NewCapabilityInfo(
					id,
					capabilities.CapabilityTypeConsensus,
					"capability-1-description",
				)
				require.NoError(t, err)

				c := &mockCapability{CapabilityInfo: ci}
				return id, reg.Add(ctx, c)
			},
			getCapability: func(ctx context.Context, reg core.CapabilitiesRegistryBase, id string) error {
				_, err := reg.GetExecutable(ctx, id)
				return err
			},
		},
	}

	ctx := t.Context()
	reg := registry.NewBase(logger.Test(t))
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.newCapability(ctx, reg)
			require.NoError(t, err)

			err = tc.getCapability(ctx, reg, id)
			require.NoError(t, err)
		})
	}
}
