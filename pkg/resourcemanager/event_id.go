package resourcemanager

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// EventID derives a deterministic, cross-node-identical MeterRecord event_id
// from an action namespace and ordered identifier parts.
//
// It is the canonical helper every producer must use so that all DON nodes
// fielding the same logical delta compute the identical event_id — the
// consumer's cross-node dedup/aggregation key (see
// metering.v1.Utilization.event_id). Correctness therefore depends entirely on
// the CALLER: every part must be DON-consistent, i.e. identical on every node
// that fields the same request. Safe inputs are fields of a DON-aggregated
// request (the remote capability layer runs mode consensus over workflow-DON
// senders, so the request delivered to every capability node is byte-identical —
// see core/capabilities/remote/trigger_publisher.go) or a reconciliation
// signature. NEVER pass a locally-generated value (a per-node random id,
// os.Hostname, time.Now, ...): that reintroduces per-node event_ids and silently
// breaks cross-node dedup.
//
// The namespace (e.g. "register", "unregister", "evm-activate") both scopes the
// action so paired +N/-N deltas do not collide and makes the id
// human-recognizable; the parts are hashed so the id has bounded length and no
// delimiter-injection ambiguity. The result is "<namespace>:<hex sha256>".
func EventID(namespace string, parts ...string) string {
	h := sha256.New()
	h.Write([]byte(namespace))
	h.Write([]byte{0})
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return namespace + ":" + hex.EncodeToString(h.Sum(nil))
}

// SnapshotEventID derives the deterministic event_id the ResourceManager stamps
// on a MeterSnapshot utilization. Unlike MeterRecord event_ids, snapshots are
// RM-derived (never producer-supplied): the key is
//
//	snapshot:{node_id}:{service}:{resource_pool}:{resource_type}:{resource_id}:{bucket_unix}
//
// where bucket_unix is the snapshot tick time truncated to the snapshot interval,
// as unix seconds. It is stable across a node's retransmits of the same tick and
// distinct across buckets. node_id is intentionally part of the key: snapshots
// are reduced by median across nodes, so the dedup scope is a single node's own
// tick retries, not cross-node convergence.
func SnapshotEventID(nodeID, service, resourcePool, resourceType, resourceID string, bucketUnix int64) string {
	return strings.Join([]string{
		"snapshot",
		nodeID,
		service,
		resourcePool,
		resourceType,
		resourceID,
		strconv.FormatInt(bucketUnix, 10),
	}, ":")
}
