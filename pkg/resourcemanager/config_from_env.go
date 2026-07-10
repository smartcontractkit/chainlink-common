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

// ConfigFromEnv maps a resolved loop.EnvConfig to a metering Config. It is the
// single, canonical source of the loop-env -> metering mapping: the enable
// flags (MeterRecordsEnabled / MeterSnapshotsEnabled), the process-global
// durable ChIP emitter (durableemitter.GetGlobalEmitter()), DefaultSnapshotInterval,
// and the deployment identity dimensions (product/tenant/numeric_tenant_id/
// environment/zone/node_id). Producers must call this rather than re-deriving
// from os.Getenv.
//
// Metering uses the durable emitter — not the standard beholder emitter — so
// each MeterRecord/MeterSnapshot is persisted to the durable queue and delivered
// at-least-once (records are counter-semantic billing events that must survive a
// transport outage; the consumer dedups by event_id). The durable emitter is
// initialized once per process by durableemitter.Setup (the LOOP server does
// this at startup). If it has not been initialized, GetGlobalEmitter returns nil
// and this leaves Config.Emitter nil, which makes the ResourceManager a no-op
// emitter (records/snapshots are not emitted) — consistent with the fail-open
// posture that metering must never block startup or the operation being metered.
//
// The CL_METER_* environment variables it reads are populated exclusively by
// loop.EnvConfig.AsCmdEnv from the core node's TOML [Metering] config; env is
// only the LOOP child-process transport.
//
// A nil envCfg yields a zero Config (metering disabled, no emitter).
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
