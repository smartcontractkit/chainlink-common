package resourcemanager

import (
	"github.com/smartcontractkit/chainlink-common/pkg/durableemitter"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
)

// Config is everything a LOOP producer needs to wire metering from its resolved
// process environment: the ResourceManagerConfig for NewResourceManager, plus
// the DeploymentIdentity used to build base ResourceIdentities via
// NewBaseIdentity. It exists to kill the loop-env -> metering-config mapping
// that would otherwise be copy-pasted into every producer main.
type Config struct {
	ResourceManagerConfig

	// DeploymentIdentity carries the static deployment + node identity
	// dimensions resolved from the LOOP environment.
	DeploymentIdentity DeploymentIdentity
}

// ConfigFromEnv maps a resolved loop.EnvConfig to a metering Config, so the
// loop-env -> metering mapping lives here instead of in every producer main.
// The CL_METER_* fields it reads are populated by loop.EnvConfig.AsCmdEnv from
// the core node's TOML [Metering] config; env is only the LOOP transport.
//
// Metering uses the process-global durable emitter so records/snapshots survive
// a transport outage (at-least-once; the consumer dedups by event_id). When the
// durable emitter has not been initialized the Emitter is left nil, which the
// ResourceManager treats as a no-op — fail-open, so metering never blocks
// startup. A nil envCfg yields a zero Config (metering disabled).
func ConfigFromEnv(envCfg *loop.EnvConfig) Config {
	if envCfg == nil {
		return Config{}
	}
	// Source the durable emitter. GetGlobalEmitter returns a typed nil pointer
	// when Setup has not run; assign through a local Emitter so an unconfigured
	// durable emitter yields a true nil interface (not a non-nil interface
	// wrapping a nil pointer), which the ResourceManager treats as no-op.
	var emitter Emitter
	if de := durableemitter.GetGlobalEmitter(); de != nil {
		emitter = de
	}
	return Config{
		ResourceManagerConfig: ResourceManagerConfig{
			MeterRecordsEnabled:   envCfg.MeterRecordsEnabled,
			MeterSnapshotsEnabled: envCfg.MeterSnapshotsEnabled,
			Emitter:               emitter,
			SnapshotInterval:      DefaultSnapshotInterval,
		},
		DeploymentIdentity: DeploymentIdentity{
			Product:         envCfg.MeterProduct,
			Tenant:          envCfg.MeterTenant,
			NumericTenantID: envCfg.MeterNumericTenantID,
			Environment:     envCfg.MeterEnvironment,
			Zone:            envCfg.MeterZone,
			NodeID:          envCfg.MeterNodeID,
		},
	}
}
