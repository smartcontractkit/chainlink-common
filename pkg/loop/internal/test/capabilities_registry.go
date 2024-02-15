package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.CapabilitiesRegistry = (*StaticCapabilitiesRegistry)(nil)

type StaticCapabilitiesRegistry struct{}

func (c *StaticCapabilitiesRegistry) Get(_ context.Context, ID string) (capabilities.BaseCapability, error) {
	if ID != GetID {
		return nil, fmt.Errorf("expected %s but got %s", GetID, ID)
	}
	return nil, nil
}

func (c *StaticCapabilitiesRegistry) GetTrigger(_ context.Context, ID string) (capabilities.TriggerCapability, error) {
	if ID != GetTriggerID {
		return nil, fmt.Errorf("expected %s but got %s", GetTriggerID, ID)
	}
	return nil, nil
}

func (c *StaticCapabilitiesRegistry) GetAction(_ context.Context, ID string) (capabilities.ActionCapability, error) {
	if ID != GetActionID {
		return nil, fmt.Errorf("expected %s but got %s", GetActionID, ID)
	}
	return nil, nil
}

func (c *StaticCapabilitiesRegistry) GetConsensus(_ context.Context, ID string) (capabilities.ConsensusCapability, error) {
	if ID != GetConsensusID {
		return nil, fmt.Errorf("expected %s but got %s", GetConsensusID, ID)
	}
	return nil, nil
}

func (c *StaticCapabilitiesRegistry) GetTarget(_ context.Context, ID string) (capabilities.TargetCapability, error) {
	if ID != GetTargetID {
		return nil, fmt.Errorf("expected %s but got %s", GetTargetID, ID)
	}
	return nil, nil
}

func (c *StaticCapabilitiesRegistry) List(_ context.Context) ([]capabilities.BaseCapability, error) {
	return nil, nil
}

func (c *StaticCapabilitiesRegistry) Add(ctx context.Context, cap capabilities.BaseCapability) error {
	i, _ := cap.Info(ctx)
	if i.ID != CapabilityInfo.ID {
		return fmt.Errorf("expected %s but got %s", CapabilityInfo.ID, i.ID)
	}
	if i.CapabilityType != CapabilityInfo.CapabilityType {
		return fmt.Errorf("expected %d but got %d", CapabilityInfo.CapabilityType, i.CapabilityType)
	}
	if i.Description != CapabilityInfo.Description {
		return fmt.Errorf("expected %s but got %s", CapabilityInfo.Description, i.Description)
	}
	if i.Version != CapabilityInfo.Version {
		return fmt.Errorf("expected %s but got %s", CapabilityInfo.Version, i.Version)
	}
	return nil
}

func CapabilitiesRegistry(t *testing.T, cr types.CapabilitiesRegistry) {
	t.Run("CapabilitiesRegistry", func(t *testing.T) {
		_, err := cr.Get(context.Background(), GetID)
		require.NoError(t, err)
		_, err = cr.GetTrigger(context.Background(), GetTriggerID)
		require.NoError(t, err)
		_, err = cr.GetAction(context.Background(), GetActionID)
		require.NoError(t, err)
		_, err = cr.GetConsensus(context.Background(), GetConsensusID)
		require.NoError(t, err)
		_, err = cr.GetTarget(context.Background(), GetTargetID)
		require.NoError(t, err)
		err = cr.Add(context.Background(), baseCapability{})
		require.NoError(t, err)
	})
}
