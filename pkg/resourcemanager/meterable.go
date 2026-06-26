package resourcemanager

import (
	"context"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
	"google.golang.org/protobuf/proto"
)

// Meterable is implemented by producers that manage durable billable
// resources (trigger registrations, workflow specs, log filters). A producer
// registers itself with a ResourceManager (see ResourceManager.Register) so it
// is polled once per snapshot tick for the absolute state of its currently
// active resources, in addition to emitting lifecycle edges inline via
// EmitMeterRecord.
type Meterable interface {
	// ResourceIdentity returns the producer's base identity: the six coarse
	// dimensions (product, environment, zone, don_id, node_id, service) plus
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

// ResourceIdentity is the structured, first-class identity of a durable
// resource. Its fields map 1:1 to metering.v1.ResourceIdentity so every
// emitted record carries each dimension as a discrete column rather than a
// parsed dotted string or out-of-band telemetry attribute.
type ResourceIdentity struct {
	// Product is the deployment product, e.g. "cre-mainline". A coarse
	// billing-rollup dimension.
	Product string

	// Environment is the deployment environment, e.g. "production",
	// "staging". A coarse billing-rollup dimension.
	Environment string

	// Zone is the deployment zone, e.g. "wf-zone-a". A coarse billing-rollup
	// dimension.
	Zone string

	// DONID is the DON identifier the emitting service belongs to. A coarse
	// billing-rollup dimension; used with NodeID to count distinct nodes for
	// quorum.
	DONID string

	// NodeID is the node identifier (the node's CSA public key). A coarse
	// billing-rollup dimension; lets billing dedup a node's retries and count
	// distinct nodes.
	NodeID string

	// Service is the stable service constant identifying the emitting service,
	// e.g. "cron-trigger", "http-trigger", "evm-log-trigger",
	// "workflow-syncer-v2". A coarse billing-rollup dimension. It must not
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
		Environment:    r.Environment,
		Zone:           r.Zone,
		Service:        r.Service,
		ResourcePool:   r.ResourcePool,
		ResourcePoolId: r.ResourcePoolID,
	}
	if r.DONID != "" {
		pb.DonId = proto.String(r.DONID)
	}
	if r.NodeID != "" {
		pb.NodeId = proto.String(r.NodeID)
	}
	return pb
}
