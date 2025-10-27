package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	dontimepb "github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
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

// TestCalculateRequestDigest_DeterministicWithSimpleMap tests that hashing is deterministic
// for protobuf messages with simple string maps, regardless of insertion order
func TestCalculateRequestDigest_DeterministicWithSimpleMap(t *testing.T) {
	// Create two messages with the same content but different insertion order
	msg1 := &pb.BaseMessage{
		Msg:       "test message",
		Timestamp: "2024-01-01T00:00:00Z",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	msg2 := &pb.BaseMessage{
		Msg:       "test message",
		Timestamp: "2024-01-01T00:00:00Z",
		Labels: map[string]string{
			"key3": "value3", // Different insertion order
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Calculate digests
	digest1 := CalculateRequestDigest(msg1)
	digest2 := CalculateRequestDigest(msg2)

	// Assert: Same content should produce the same digest
	require.Equal(t, digest1, digest2, "Digests should be identical for same content with different map insertion order")

	// Verify digest format (should be 64 character hex string for SHA256)
	assert.Len(t, digest1, 64)
	assert.Regexp(t, "^[0-9a-f]{64}$", digest1)
}

// TestCalculateRequestDigest_DeterministicWithComplexNestedMap tests that hashing is deterministic
// for protobuf messages with complex nested maps
func TestCalculateRequestDigest_DeterministicWithComplexNestedMap(t *testing.T) {
	// Create two messages with the same content but different insertion order
	msg1 := &dontimepb.Outcome{
		Timestamp: 1234567890,
		ObservedDonTimes: map[string]*dontimepb.ObservedDonTimes{
			"workflow1": {
				Timestamps: []int64{100, 200, 300},
			},
			"workflow2": {
				Timestamps: []int64{400, 500, 600},
			},
			"workflow3": {
				Timestamps: []int64{700, 800, 900},
			},
		},
	}

	msg2 := &dontimepb.Outcome{
		Timestamp: 1234567890,
		ObservedDonTimes: map[string]*dontimepb.ObservedDonTimes{
			"workflow3": { // Different insertion order
				Timestamps: []int64{700, 800, 900},
			},
			"workflow1": {
				Timestamps: []int64{100, 200, 300},
			},
			"workflow2": {
				Timestamps: []int64{400, 500, 600},
			},
		},
	}

	// Calculate digests
	digest1 := CalculateRequestDigest(msg1)
	digest2 := CalculateRequestDigest(msg2)

	// Assert: Same content should produce the same digest
	require.Equal(t, digest1, digest2, "Digests should be identical for same content with different nested map insertion order")

	// Verify digest format
	assert.Len(t, digest1, 64)
	assert.Regexp(t, "^[0-9a-f]{64}$", digest1)
}

// TestCalculateRequestDigest_ConsistentMultipleCalls verifies that calling the function
// multiple times on the same message produces consistent results
func TestCalculateRequestDigest_ConsistentMultipleCalls(t *testing.T) {
	msg := &pb.BaseMessage{
		Msg:       "test consistency",
		Timestamp: "2024-01-01T00:00:00Z",
		Labels: map[string]string{
			"env":     "production",
			"service": "node-auth",
			"version": "1.0.0",
		},
	}

	// Call digest function multiple times
	digest1 := CalculateRequestDigest(msg)
	digest2 := CalculateRequestDigest(msg)
	digest3 := CalculateRequestDigest(msg)

	// All digests should be identical
	assert.Equal(t, digest1, digest2)
	assert.Equal(t, digest2, digest3)
}

// TestCalculateRequestDigest_DifferentContentDifferentDigest ensures that different
// content produces different digests
func TestCalculateRequestDigest_DifferentContentDifferentDigest(t *testing.T) {
	msg1 := &pb.BaseMessage{
		Msg: "message1",
		Labels: map[string]string{
			"key": "value1",
		},
	}

	msg2 := &pb.BaseMessage{
		Msg: "message2",
		Labels: map[string]string{
			"key": "value1",
		},
	}

	msg3 := &pb.BaseMessage{
		Msg: "message1",
		Labels: map[string]string{
			"key": "value2",
		},
	}

	digest1 := CalculateRequestDigest(msg1)
	digest2 := CalculateRequestDigest(msg2)
	digest3 := CalculateRequestDigest(msg3)

	// All digests should be different
	assert.NotEqual(t, digest1, digest2, "Different message content should produce different digests")
	assert.NotEqual(t, digest1, digest3, "Different label values should produce different digests")
	assert.NotEqual(t, digest2, digest3, "Different messages should produce different digests")
}

// TestCalculateRequestDigest_EmptyMap tests that empty maps are handled correctly
func TestCalculateRequestDigest_EmptyMap(t *testing.T) {
	msg1 := &pb.BaseMessage{
		Msg:    "test",
		Labels: map[string]string{},
	}

	msg2 := &pb.BaseMessage{
		Msg:    "test",
		Labels: nil,
	}

	digest1 := CalculateRequestDigest(msg1)
	digest2 := CalculateRequestDigest(msg2)

	// Empty map and nil map should produce the same digest
	assert.Equal(t, digest1, digest2, "Empty and nil maps should produce the same digest")
}
