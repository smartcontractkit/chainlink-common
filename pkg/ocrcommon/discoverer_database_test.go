package ocrcommon_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/ocrcommon"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/sqltest"
)

const testDiscovererTable = "test_discoverer_announcements"

func Test_DiscovererDatabase(t *testing.T) {
	sqltest.SkipInMemory(t)
	db := sqltest.NewDB(t, sqltest.TestURL(t))

	ctx := t.Context()
	_, err := db.ExecContext(ctx, `CREATE TABLE `+testDiscovererTable+` (
		local_peer_id text NOT NULL,
		remote_peer_id text NOT NULL,
		ann bytea NOT NULL,
		created_at timestamptz not null,
		updated_at timestamptz not null,
		PRIMARY KEY(local_peer_id, remote_peer_id)
	);`)
	require.NoError(t, err)

	localPeerID1 := mustRandomPeerID(t)
	localPeerID2 := mustRandomPeerID(t)

	dd1 := ocrcommon.NewDiscovererDatabase(db, localPeerID1, testDiscovererTable)
	dd2 := ocrcommon.NewDiscovererDatabase(db, localPeerID2, testDiscovererTable)

	t.Run("StoreAnnouncement writes a value", func(t *testing.T) {
		ann := []byte{1, 2, 3}
		err := dd1.StoreAnnouncement(ctx, "remote1", ann)
		assert.NoError(t, err)

		// test upsert
		ann = []byte{4, 5, 6}
		err = dd1.StoreAnnouncement(ctx, "remote1", ann)
		assert.NoError(t, err)

		// write a different value
		ann = []byte{7, 8, 9}
		err = dd1.StoreAnnouncement(ctx, "remote2", ann)
		assert.NoError(t, err)
	})

	t.Run("ReadAnnouncements reads values filtered by given peerIDs", func(t *testing.T) {
		announcements, err := dd1.ReadAnnouncements(ctx, []string{"remote1", "remote2"})
		require.NoError(t, err)

		assert.Len(t, announcements, 2)
		assert.Equal(t, []byte{4, 5, 6}, announcements["remote1"])
		assert.Equal(t, []byte{7, 8, 9}, announcements["remote2"])

		announcements, err = dd1.ReadAnnouncements(ctx, []string{"remote1"})
		require.NoError(t, err)

		assert.Len(t, announcements, 1)
		assert.Equal(t, []byte{4, 5, 6}, announcements["remote1"])
	})

	t.Run("is scoped to local peer ID", func(t *testing.T) {
		ann := []byte{10, 11, 12}
		err := dd2.StoreAnnouncement(ctx, "remote1", ann)
		assert.NoError(t, err)

		announcements, err := dd2.ReadAnnouncements(ctx, []string{"remote1"})
		require.NoError(t, err)
		assert.Len(t, announcements, 1)
		assert.Equal(t, []byte{10, 11, 12}, announcements["remote1"])

		announcements, err = dd1.ReadAnnouncements(ctx, []string{"remote1"})
		require.NoError(t, err)
		assert.Len(t, announcements, 1)
		assert.Equal(t, []byte{4, 5, 6}, announcements["remote1"])
	})

	t.Run("persists data across restarts", func(t *testing.T) {
		dd3 := ocrcommon.NewDiscovererDatabase(db, localPeerID1, testDiscovererTable)

		announcements, err := dd3.ReadAnnouncements(ctx, []string{"remote1"})
		require.NoError(t, err)
		assert.Len(t, announcements, 1)
		assert.Equal(t, []byte{4, 5, 6}, announcements["remote1"])
	})
}

func mustRandomPeerID(t *testing.T) string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return hex.EncodeToString(b)
}
