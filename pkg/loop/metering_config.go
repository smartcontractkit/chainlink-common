package loop

import (
	"github.com/smartcontractkit/chainlink-common/pkg/resourcemanager"
)

// MeteringConfig maps this resolved EnvConfig to a resourcemanager.Config, so
// the loop-env -> metering mapping lives here instead of being copy-pasted
// into every producer main. The CL_METER_* fields it reads are populated by
// EnvConfig.AsCmdEnv from the core node's TOML [Metering] config; env is only
// the LOOP transport.
//
// emitter is injected by the caller (can be process global durableemitter.GetGlobalEmitter())
// rather than looked up here, so this mapping stays a pure function of its
// inputs. A nil emitter is valid: the ResourceManager treats it as a no-op —
// fail-open, so metering never blocks startup. A nil EnvConfig yields a zero
// Config (metering disabled).
func (e *EnvConfig) MeteringConfig(emitter resourcemanager.Emitter) resourcemanager.Config {
	if e == nil {
		return resourcemanager.Config{}
	}
	return resourcemanager.Config{
		ResourceManagerConfig: resourcemanager.ResourceManagerConfig{
			MeterRecordsEnabled:   e.MeterRecordsEnabled,
			MeterSnapshotsEnabled: e.MeterSnapshotsEnabled,
			Emitter:               emitter,
			SnapshotInterval:      resourcemanager.DefaultSnapshotInterval,
		},
		DeploymentIdentity: resourcemanager.DeploymentIdentity{
			Product:         e.MeterProduct,
			Tenant:          e.MeterTenant,
			NumericTenantID: e.MeterNumericTenantID,
			Environment:     e.MeterEnvironment,
			Zone:            e.MeterZone,
			NodeID:          e.MeterNodeID,
		},
	}
}
