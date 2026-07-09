package resourcemanager

import (
	"context"
	"strconv"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// Meterable is implemented by producers that manage durable billable
// resources (trigger registrations, workflow specs, log filters). A producer
// registers itself with a ResourceManager (see ResourceManager.Register) so it
// is polled once per snapshot tick for the absolute state of its currently
// active resources, in addition to emitting request-time deltas inline via
// EmitDelta / EmitUsage.
type Meterable interface {
	// ResourceIdentity returns the producer's base identity: the coarse
	// dimensions (product, tenant, numeric_tenant_id, environment, zone, don, service) plus
	// the service-level resource_pool / resource_pool_id. Per-resource billing
	// fields (resource_type/resource_id/org_id/value) are carried by
	// Utilizations on MeterRecord and MeterSnapshot; event_id is stamped by the
	// ResourceManager at emit time.
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
// billed dimensions for that resource (resource_type/resource_id/org_id/value);
// event_id is stamped by the ResourceManager at emit time.
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
	// Tenant is the human-readable deployment tenant name, e.g. "mainline" or
	// "enterprise".
	Tenant string
	// NumericTenantID is the numbered tenant identifier as a string.
	NumericTenantID string
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

// DonIdentity captures DON-specific identity dimensions as one unit.
type DonIdentity struct {
	// DonID is the DON ID the emitting service belongs to.
	DonID string
	// NodeID is the node's logical name within the scope of the DON, e.g.
	// "clp-cre-wf-zone-a-1". It is a human-readable ID, NOT the CSA
	// public key. The prefix can be redundant with other fully-qualified
	// dimensions, but helps readability. The CSA key is emitted separately via
	// the node_csa_key attribute.
	NodeID string
}

// ResourceIdentity is the structured, first-class identity of a durable
// resource. Its fields map 1:1 to metering.v1.ResourceIdentity so every
// emitted record carries each dimension as a discrete column rather than a
// parsed dotted string or out-of-band telemetry attribute.
type ResourceIdentity struct {
	// Product is the deployment product, e.g. "cre".
	Product string

	// Tenant is the human-readable deployment tenant name, e.g. "mainline" or
	// "enterprise".
	Tenant string

	// NumericTenantID is the numbered tenant identifier as a string.
	NumericTenantID string

	// Environment is the deployment environment, e.g. "production",
	// "staging".
	Environment string

	// Zone is the deployment zone, e.g. "wf-zone-a".
	Zone string

	// Don groups DON-specific identity dimensions so consumers can
	// branch on one struct instead of handling don/node permutations.
	Don *DonIdentity

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
		Product:         r.Product,
		Tenant:          r.Tenant,
		NumericTenantId: r.NumericTenantID,
		Environment:     r.Environment,
		Zone:            r.Zone,
		Service:         r.Service,
		ResourcePool:    r.ResourcePool,
		ResourcePoolId:  r.ResourcePoolID,
	}
	if r.Don != nil {
		pb.Don = &meteringpb.DonIdentity{
			DonId:  r.Don.DonID,
			NodeId: r.Don.NodeID,
		}
	}
	return pb
}

// DonID returns the DON ID when present.
func (r ResourceIdentity) DonID() string {
	if r.Don == nil {
		return ""
	}
	return r.Don.DonID
}

// NodeID returns the node ID when present.
func (r ResourceIdentity) NodeID() string {
	if r.Don == nil {
		return ""
	}
	return r.Don.NodeID
}

// DefaultMeteringProduct is the fallback product dimension used when the host
// did not inject one (a legacy node or a boot path that predates the metering
// deployment-identity plumbing).
const DefaultMeteringProduct = "cre"

// NewBaseIdentity builds a producer's base ResourceIdentity from its static
// deployment identity plus the service/resource-pool constants. It centralizes
// three rules that every producer must apply identically:
//
//   - product falls back to DefaultMeteringProduct ("cre") when unset;
//   - capDONID is the authoritative DON ID supplied over the standardcapabilities
//     interface (StandardCapabilitiesDependencies, host-injected at Initialise);
//     it is rendered with strconv.FormatUint, and left empty when 0 (the host
//     has not populated it) so the request's workflow DON can be applied later
//     via WithWorkflowDonFallback;
//   - the Don sub-identity is set only when there is a DON ID or node ID to
//     carry; otherwise it stays nil rather than an empty struct.
//
// Per-resource fields (resource_type/resource_id/org_id/value) and event_id are
// not part of the base identity; they are carried per emission.
func NewBaseIdentity(dep DeploymentIdentity, capDONID uint32, service, resourcePool string) ResourceIdentity {
	product := dep.Product
	if product == "" {
		product = DefaultMeteringProduct
	}

	var donID string
	if capDONID != 0 {
		donID = strconv.FormatUint(uint64(capDONID), 10)
	}

	id := ResourceIdentity{
		Product:         product,
		Tenant:          dep.Tenant,
		NumericTenantID: dep.NumericTenantID,
		Environment:     dep.Environment,
		Zone:            dep.Zone,
		Service:         service,
		ResourcePool:    resourcePool,
	}
	if donID != "" || dep.NodeID != "" {
		id.Don = &DonIdentity{DonID: donID, NodeID: dep.NodeID}
	}
	return id
}

// WithWorkflowDonFallback returns a copy of r stamped with the request's
// workflow DON ID, but only when the base identity has no CapDONID. CapDONID
// (supplied over the standardcapabilities interface) is the authoritative DON
// ID for a capability node; the request's WorkflowDonID is a fallback that
// applies only for hosts that have not populated CapDONID (value 0, so
// NewBaseIdentity left don_id empty). When the base identity already carries a
// CapDONID, or workflowDonID is 0, r is returned unchanged.
func (r ResourceIdentity) WithWorkflowDonFallback(workflowDonID uint32) ResourceIdentity {
	if r.DonID() != "" || workflowDonID == 0 {
		return r
	}
	r.Don = &DonIdentity{
		DonID:  strconv.FormatUint(uint64(workflowDonID), 10),
		NodeID: r.NodeID(),
	}
	return r
}
