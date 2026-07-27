package resourcemanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventID(t *testing.T) {
	// Deterministic: identical inputs -> identical id (the cross-node contract).
	a := EventID("register", "wf-1", "trigger-1")
	b := EventID("register", "wf-1", "trigger-1")
	assert.Equal(t, a, b)
	assert.True(t, len(a) > len("register:"))
	assert.Equal(t, "register:", a[:len("register:")])

	// Distinct parts -> distinct id (e.g. re-register with a disambiguator).
	assert.NotEqual(t, a, EventID("register", "wf-1", "trigger-1", "reactivation-2"))
	// Distinct namespace (paired +N/-N) -> distinct id.
	assert.NotEqual(t, a, EventID("unregister", "wf-1", "trigger-1"))
	// Part boundaries are unambiguous: ("a","b") != ("ab").
	assert.NotEqual(t, EventID("ns", "a", "b"), EventID("ns", "ab"))
}

func TestSnapshotEventID(t *testing.T) {
	got := SnapshotEventID("node-1", "cron-trigger", "trigger_registrations", "operations", "trigger-1", 1704067230)
	assert.Equal(t, "snapshot:node-1:cron-trigger:trigger_registrations:operations:trigger-1:1704067230", got)

	// Stable across retransmits of the same bucket, distinct across buckets and nodes.
	assert.Equal(t, got, SnapshotEventID("node-1", "cron-trigger", "trigger_registrations", "operations", "trigger-1", 1704067230))
	assert.NotEqual(t, got, SnapshotEventID("node-1", "cron-trigger", "trigger_registrations", "operations", "trigger-1", 1704067260))
	assert.NotEqual(t, got, SnapshotEventID("node-2", "cron-trigger", "trigger_registrations", "operations", "trigger-1", 1704067230))
}
