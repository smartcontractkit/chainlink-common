package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

// Helper function to create keyparis
func createTestKeys() (ed25519.PublicKey, p2ptypes.PeerID) {
	csaPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 key pair: " + err.Error())
	}

	p2pIdKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("Failed to generate Ed25519 p2pId: " + err.Error())
	}

	p2pId, err := p2ptypes.PeerIDFromPublicKey(p2pIdKey)
	if err != nil {
		panic("Failed to create PeerID from public key: " + err.Error())
	}

	return csaPubKey, p2pId
}

func TestDecodePublicKey_Success(t *testing.T) {
	csaPubKey, _ := createTestKeys()
	csaPubKeyHex := hex.EncodeToString(csaPubKey)

	// Test
	decodedKey, err := DecodePublicKey(csaPubKeyHex)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, csaPubKey, decodedKey)
}

func TestDecodePublicKey_InvalidHex(t *testing.T) {
	// Test with invalid hex
	_, err := DecodePublicKey("invalid-hex")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex encoding")
}

func TestDecodePublicKey_InvalidSize(t *testing.T) {
	// Test with wrong size
	_, err := DecodePublicKey("1234")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

func TestDecodeP2PId_Success(t *testing.T) {
	_, p2pId := createTestKeys()
	p2pIdString := p2pId.String()

	// Test
	decodedP2PId, err := DecodeP2PId(p2pIdString)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, p2pId, decodedP2PId)
}

func TestDecodeP2PId_InvalidHex(t *testing.T) {
	// Test with invalid hex
	_, err := DecodeP2PId("invalid-hex")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal PeerID")
}

func TestDecodeP2PId_InvalidSize(t *testing.T) {
	// Test with wrong size
	_, err := DecodeP2PId("1234")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal PeerID")
}
