package resourcemanager

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// IdempotencyKey returns the canonical deduplication key for a MeterRecord:
// the lowercase hex SHA-256 over the full structured identity (including
// node_id), the action, and an event identity, joined with "|":
//
//	product|environment|zone|don_id|node_id|service|resource|resource_type|action|resource_id|eventIdentity
//
// where action is the MeterAction enum name (e.g. "METER_ACTION_RESERVE").
//
// node_id is included, so keys are unique per node: billing dedups a node's
// retries and counts distinct nodes for quorum, while cross-node grouping and
// convergence is the consumer's job on resource_id + dimensions, independent
// of the key.
//
// Keys are level-triggered by design. Producers intentionally reuse the same
// key whenever they re-emit the same resource lifecycle edge — for example a
// node restart re-registering an existing trigger. A repeated key asserts
// "this lifecycle edge happened" rather than "a new event happened".
//
// Consumers therefore use the key for exact-duplicate suppression only.
// Lifecycle state must not be ordered by key arrival: consumers order
// lifecycle state by record timestamp per resource_id, last write wins.
//
// RESERVE and RELEASE are emitted only for genuine allocation/deallocation,
// never for process-lifecycle cleanup of leaked/orphaned resources; a RELEASE
// carries the same value as its paired RESERVE. A reservation lost without a
// RELEASE (e.g. a node crash) is reconciled by its absence from subsequent
// Snapshots, not by a synthetic release.
//
// This helper is the only MeterRecord key generator; producers must not derive
// keys themselves. Inputs are identifiers and must not contain "|".
func IdempotencyKey(id ResourceIdentity, action meteringpb.MeterAction, eventIdentity string) string {
	preimage := strings.Join([]string{
		id.Product,
		id.Environment,
		id.Zone,
		id.DONID,
		id.NodeID,
		id.Service,
		id.Resource,
		id.ResourceType,
		action.String(),
		id.ResourceID,
		eventIdentity,
	}, "|")
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

// SnapshotIdempotencyKey returns the canonical deduplication key for one
// MeterSnapshot: the lowercase hex SHA-256 over the literal "snapshot", the
// node-scoped dimensions, the resource and resource_id, and the interval
// bucket, joined with "|":
//
//	snapshot|product|environment|zone|don_id|node_id|service|resource|resource_id|interval-bucket
//
// intervalBucket is the snapshot timestamp truncated to the snapshot interval
// (e.g. unix seconds of the interval boundary). It makes each interval's
// snapshot of a resource distinct — so per-interval increments are not
// collapsed — while deduping retries of the same interval. Snapshots are
// action-less, so no action is mixed into the key.
//
// This helper is the only MeterSnapshot key generator; producers must not
// derive keys themselves. Inputs are identifiers and must not contain "|".
func SnapshotIdempotencyKey(id ResourceIdentity, intervalBucket int64) string {
	preimage := strings.Join([]string{
		"snapshot",
		id.Product,
		id.Environment,
		id.Zone,
		id.DONID,
		id.NodeID,
		id.Service,
		id.Resource,
		id.ResourceID,
		strconv.FormatInt(intervalBucket, 10),
	}, "|")
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

// NewUtilization builds a Utilization for EmitMeterRecord with its idempotency
// key derived via IdempotencyKey, so producers never construct keys by hand.
// The resource being metered is identified entirely by id (with its ResourceID
// set, typically via ResourceIdentity.WithResourceID); Utilization carries no
// labels.
func NewUtilization(id ResourceIdentity, action meteringpb.MeterAction, value int64, eventIdentity string) *meteringpb.Utilization {
	return &meteringpb.Utilization{
		Value:          value,
		IdempotencyKey: IdempotencyKey(id, action, eventIdentity),
	}
}
