package resourcemanager

import (
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
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
// flags (MeterRecordsEnabled / MeterSnapshotsEnabled), the production beholder
// emitter (beholder.GetEmitter()), DefaultSnapshotInterval, and the deployment
// identity dimensions (product/tenant/numeric_tenant_id/environment/zone/
// node_id). Producers must call this rather than re-deriving from os.Getenv.
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
	return Config{
		ResourceManagerConfig: ResourceManagerConfig{
			MeterRecordsEnabled:   envCfg.MeterRecordsEnabled,
			MeterSnapshotsEnabled: envCfg.MeterSnapshotsEnabled,
			Emitter:               beholder.GetEmitter(),
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
