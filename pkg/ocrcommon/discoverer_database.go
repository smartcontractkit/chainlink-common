// Package ocrcommon contains reusable building blocks shared by OCR-based
// services. It currently provides DiscovererDatabase, a key-value store for
// RageP2P announcements backed by Postgres.
package ocrcommon

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"

	ocrnetworking "github.com/smartcontractkit/libocr/networking/types"

	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
)

var _ ocrnetworking.DiscovererDatabase = &DiscovererDatabase{}

// DiscovererDatabase is a key-value store for p2p announcements
// that are based on the RageP2P library and bootstrap nodes.
//
// It is generic over the backing table so that callers can scope different
// announcement use cases to separate tables while sharing the same logic.
type DiscovererDatabase struct {
	ds        sqlutil.DataSource
	peerID    string
	tableName string
}

// NewDiscovererDatabase creates a DiscovererDatabase that reads and writes
// announcements in the given table. The table must have the schema produced by
// the migration generator in ./setup (local_peer_id, remote_peer_id, ann,
// created_at, updated_at).
func NewDiscovererDatabase(ds sqlutil.DataSource, peerID, tableName string) *DiscovererDatabase {
	return &DiscovererDatabase{
		ds:        ds,
		peerID:    peerID,
		tableName: tableName,
	}
}

// StoreAnnouncement has key-value-store semantics and stores a peerID (key) and an associated serialized
// announcement (value).
func (d *DiscovererDatabase) StoreAnnouncement(ctx context.Context, peerID string, ann []byte) error {
	q := fmt.Sprintf(`
INSERT INTO %s (local_peer_id, remote_peer_id, ann, created_at, updated_at)
VALUES ($1,$2,$3,NOW(),NOW()) ON CONFLICT (local_peer_id, remote_peer_id) DO UPDATE SET
ann = EXCLUDED.ann,
updated_at = EXCLUDED.updated_at
;`, d.tableName)

	_, err := d.ds.ExecContext(ctx,
		q, d.peerID, peerID, ann)
	if err != nil {
		return fmt.Errorf("DiscovererDatabase failed to StoreAnnouncement: %w", err)
	}
	return nil
}

// ReadAnnouncements returns one serialized announcement (if available) for each of the peerIDs in the form of a map
// keyed by each announcement's corresponding peer ID.
func (d *DiscovererDatabase) ReadAnnouncements(ctx context.Context, peerIDs []string) (results map[string][]byte, err error) {
	q := fmt.Sprintf(`SELECT remote_peer_id, ann FROM %s WHERE remote_peer_id = ANY($1) AND local_peer_id = $2`, d.tableName)

	rows, err := d.ds.QueryContext(ctx, q, pq.Array(peerIDs), d.peerID)
	if err != nil {
		return nil, fmt.Errorf("DiscovererDatabase failed to ReadAnnouncements: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	results = make(map[string][]byte)
	for rows.Next() {
		var peerID string
		var ann []byte
		err = rows.Scan(&peerID, &ann)
		if err != nil {
			return
		}
		results[peerID] = ann
	}
	if err = rows.Err(); err != nil {
		return
	}
	return results, nil
}