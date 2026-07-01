package resourcemanager

import (
	"context"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// Meterable is implemented by producers that manage durable billable
// resources (trigger registrations, workflow specs, log filters). A producer
// registers itself with a ResourceManager (see ResourceManager.Register) so it
// is polled once per snapshot tick for the absolute state of its currently
// active resources, in addition to emitting lifecycle edges inline via
// EmitMeterRecord.
type Meterable interface {
	// ResourceIdentity returns the producer's base identity: the coarse
	// dimensions (product, tenant, environment, zone, don_identifier, service) plus
	// the service-level resource_pool / resource_pool_id. Per-resource billing
	// fields (resource_type/resource_id/org_id/event_id/value) are carried by
	// Utilizations on MeterRecord and MeterSnapshot.
	ResourceIdentity() ResourceIdentity

	// GetUtilization returns the current level of the producer's currently
	// active resources, one SnapshotEntry per resource. The manager emits one
	// MeterSnapshot per entry.
	//
	// It is called on the snapshot tick and MUST be a cheap, non-blocking
	// read-snapshot of in-memory state: no network, no disk, no lock held
	// across I/O. It must tolerate ctx cancellation (returning promptly, and
	// nil/empty is acceptable) and tolerate concurrent registration of new
	// resources. An empty or nil return is valid and means nothing is currently
	// active: no snapshots are emitted, and billing zeroes the resource out by
	// its absence from subsequent snapshots.
	GetUtilization(ctx context.Context) []SnapshotEntry
}

// SnapshotEntry is the current level of one active resource at a snapshot tick.
// Identity is the base resource identity, and Utilizations carries one or more
// billed dimensions for that resource (resource_type/resource_id/org_id/event_id/value).
type SnapshotEntry struct {
	Identity     ResourceIdentity
	Utilizations []*meteringpb.Utilization
}

// DeploymentIdentity carries the static deployment + node identity dimensions
// that are fixed for a LOOP plugin process. They are resolved once from node
// config by the host and delivered to every LOOP plugin over the environment
// (loop.EnvConfig), not the standard-capabilities boundary, so any LOOP plugin
// can populate the coarse metering rollup dimensions. Any field may be empty if
// the host did not provide it.
type DeploymentIdentity struct {
	// Product is the deployment product, e.g. "cre".
	Product string
	// Tenant is the deployment tenant, e.g. "mainline" or "enterprise".
	Tenant string
	// Environment is the deployment environment, e.g. "production", "staging".
	Environment string
	// Zone is the deployment zone, e.g. "wf-zone-a".
	Zone string
	// NodeID is the node's logical name, e.g. "clp-cre-wf-zone-a-1". It is NOT
	// the CSA public key; it is a stable name the billing service can use to
	// look up the node's CSA key in the workflow registry. The CSA key itself is
	// carried separately as the node_csa_key event attribute.
	NodeID string
}

// DonIdentifier captures DON-specific identity dimensions as one unit.
type DonIdentifier struct {
	// DonID is the DON identifier the emitting service belongs to.
	DonID string
	// NodeID is the node identifier (the node's CSA public key).
	NodeID string
}

// ResourceIdentity is the structured, first-class identity of a durable
// resource. Its fields map 1:1 to metering.v1.ResourceIdentity so every
// emitted record carries each dimension as a discrete column rather than a
// parsed dotted string or out-of-band telemetry attribute.
type ResourceIdentity struct {
	// Product is the deployment product, e.g. "cre".
	Product string

	// Tenant is the deployment tenant, e.g. "mainline" or "enterprise".
	Tenant string

	// Environment is the deployment environment, e.g. "production",
	// "staging".
	Environment string

	// Zone is the deployment zone, e.g. "wf-zone-a".
	Zone string

	// DonIdentifier groups DON-specific identity dimensions so consumers can
	// branch on one struct instead of handling don/node permutations.
	DonIdentifier *DonIdentifier

	// Service is the stable service constant identifying the emitting service,
	// e.g. "cron-trigger", "http-trigger", "evm-log-trigger",
	// "workflow-syncer-v2". It must not
	// encode deployment environment or zone.
	Service string

	// ResourcePool is the service-level resource pool the record applies to,
	// e.g. "trigger_registrations", "log_filters", "workflow_storage".
	ResourcePool string

	// ResourcePoolID optionally scopes identity further within the resource pool.
	ResourcePoolID string
}

// toProto converts r to its wire form. Field order mirrors the proto.
func (r ResourceIdentity) toProto() *meteringpb.ResourceIdentity {
	pb := &meteringpb.ResourceIdentity{
		Product:        r.Product,
		Tenant:         r.Tenant,
		Environment:    r.Environment,
		Zone:           r.Zone,
		Service:        r.Service,
		ResourcePool:   r.ResourcePool,
		ResourcePoolId: r.ResourcePoolID,
	}
	if r.DonIdentifier != nil {
		pb.DonIdentifier = &meteringpb.DonIdentifier{
			DonId:  r.DonIdentifier.DonID,
			NodeId: r.DonIdentifier.NodeID,
		}
	}
	return pb
}

// DonID returns the DON identifier when present.
func (r ResourceIdentity) DonID() string {
	if r.DonIdentifier == nil {
		return ""
	}
	return r.DonIdentifier.DonID
}

// NodeID returns the node identifier when present.
func (r ResourceIdentity) NodeID() string {
	if r.DonIdentifier == nil {
		return ""
	}
	return r.DonIdentifier.NodeID
}
