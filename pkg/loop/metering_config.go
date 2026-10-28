package loop

import (
	"github.com/smartcontractkit/chainlink-common/pkg/resourcemanager"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
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
// fail-open, so metering never blocks startup. settings is the CRE settings
// getter the host resolved (Server passes its own, so runtime updates apply);
// nil disables usage fault injection. A nil EnvConfig yields a zero Config
// (metering disabled).
func (e *EnvConfig) MeteringConfig(emitter resourcemanager.Emitter, settings settings.Getter) resourcemanager.Config {
	if e == nil {
		return resourcemanager.Config{}
	}
	return resourcemanager.Config{
		MeterRecordsEnabled:   e.MeterRecordsEnabled,
		MeterSnapshotsEnabled: e.MeterSnapshotsEnabled,
		Emitter:               emitter,
		SnapshotInterval:      resourcemanager.DefaultSnapshotInterval,
		Settings:              settings,
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
