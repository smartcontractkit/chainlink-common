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
	// ResourceIdentity returns the producer's base identity: the six coarse
	// dimensions (product, environment, zone, don_id, node_id, service) plus
	// the service-level resource / resource_type. The per-resource resource_id
	// is left empty here and populated per active resource by GetUtilization.
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
// Identity is the full per-resource identity (the producer's base dimensions
// with resource / resource_type / resource_id populated for this specific
// resource), and Value is its current level. The resource is identified
// entirely by Identity; there is no separate label metadata.
type SnapshotEntry struct {
	Identity ResourceIdentity
	Value    int64
}

// ResourceIdentity is the structured, first-class identity of a durable
// resource. Its nine fields map 1:1 to metering.v1.ResourceIdentity so every
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

	// Resource is the resource pool the record applies to, e.g.
	// "trigger_registrations", "log_filters".
	Resource string

	// ResourceType is the billing unit used to convert the utilization value
	// to universal credits, e.g. "operations", "log_filter_addresses".
	ResourceType string

	// ResourceID is the physical/logical resource identity, workflow-
	// independent where a shared physical resource exists. For EVM log filters
	// it is the content hash of (chain_selector + canonicalized addresses +
	// event sigs + positional topics), so identical filters from different
	// workflows share one ResourceID. For cron/http/syncer (no shared physical
	// resource) it is the workflow-scoped trigger_id / workflow_id, from which
	// the workflow (and, downstream, the owner) is recoverable. ResourceIdentity
	// is the sole identity of a metered resource; there is no label metadata.
	ResourceID string
}

// WithResourceID returns a copy of r with ResourceID set to id, leaving the
// receiver unchanged. Producers build a base identity once and derive a
// per-resource identity with this helper.
func (r ResourceIdentity) WithResourceID(id string) ResourceIdentity {
	r.ResourceID = id
	return r
}

// toProto converts r to its wire form. Field order mirrors the proto.
func (r ResourceIdentity) toProto() *meteringpb.ResourceIdentity {
	return &meteringpb.ResourceIdentity{
		Product:      r.Product,
		Environment:  r.Environment,
		Zone:         r.Zone,
		DonId:        r.DONID,
		NodeId:       r.NodeID,
		Service:      r.Service,
		Resource:     r.Resource,
		ResourceType: r.ResourceType,
		ResourceId:   r.ResourceID,
	}
}
