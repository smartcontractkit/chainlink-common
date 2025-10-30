package ragep2p_test

import (
	"testing"

	commonks "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/ragep2p"
	"github.com/stretchr/testify/require"
)

func TestPeerKeyring(t *testing.T) {
	storage := commonks.NewMemoryStorage()
	ctx := t.Context()
	ks, err := commonks.LoadKeystore(ctx, storage, commonks.EncryptionParams{
		Password:     "test-password",
		ScryptParams: commonks.FastScryptParams,
	})
	require.NoError(t, err)
	peerKeyring, err := ragep2p.CreatePeerKeyring(ctx, ks, "test-peer-keyring")
	require.NoError(t, err)
	msg := []byte("test-message")
	signature, err := peerKeyring.Sign(msg)
	require.NoError(t, err)
	require.NotNil(t, signature)

	peerKeyrings, err := ragep2p.GetPeerKeyrings(ctx, ks, []string{"test-peer-keyring"})
	require.NoError(t, err)
	require.Equal(t, 1, len(peerKeyrings))
	require.Equal(t, peerKeyring.PublicKey(), peerKeyrings[0].PublicKey())
	require.Equal(t, peerKeyring.KeyPath(), peerKeyrings[0].KeyPath())

	// List all works
	peerKeyRings, err := ragep2p.GetPeerKeyrings(ctx, ks, []string{})
	require.NoError(t, err)
	require.Equal(t, 1, len(peerKeyRings))

	// List non-existent errors.
	peerKeyRings, err = ragep2p.GetPeerKeyrings(ctx, ks, []string{"non-existent-peer-keyring"})
	require.Error(t, err)

	// Can create multiple.
	peerKeyring2, err := ragep2p.CreatePeerKeyring(ctx, ks, "test-peer-keyring-2")
	require.NoError(t, err)
	msg2 := []byte("test-message-2")
	signature2, err := peerKeyring2.Sign(msg2)
	require.NoError(t, err)
	require.NotNil(t, signature2)

	// List by name works.
	peerKeyRings, err = ragep2p.GetPeerKeyrings(ctx, ks, []string{"test-peer-keyring-2"})
	require.NoError(t, err)
	require.Equal(t, 1, len(peerKeyRings))
	require.Equal(t, peerKeyring2.PublicKey(), peerKeyRings[0].PublicKey())
	require.Equal(t, peerKeyring2.KeyPath(), peerKeyRings[0].KeyPath())

	// List all works with multiple.
	peerKeyRings, err = ragep2p.GetPeerKeyrings(ctx, ks, []string{})
	require.NoError(t, err)
	require.Equal(t, 2, len(peerKeyRings))
}
