package utils

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

// CalculateRequestDigest creates a SHA256 digest of the request for integrity verification
// This function is shared between client (JWT generation) and server (JWT validation)
func CalculateRequestDigest(req any) string {
	// Create canonical string representation
	var canonical string
	if s, ok := req.(fmt.Stringer); ok {
		canonical = s.String()
	} else {
		canonical = fmt.Sprintf("%v", req)
	}

	// Hash and encode as hex
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])
}

// DecodePublicKey converts hex-encoded public key to ed25519.PublicKey
func DecodePublicKey(publicKeyHex string) (ed25519.PublicKey, error) {
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex encoding: %w", err)
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d bytes, got %d", ed25519.PublicKeySize, len(publicKeyBytes))
	}

	return ed25519.PublicKey(publicKeyBytes), nil
}

// DecodeP2PId converts base58-encoded p2pId string to p2ptypes.PeerID
func DecodeP2PId(p2pIdString string) (p2ptypes.PeerID, error) {
	var peerID p2ptypes.PeerID
	err := peerID.UnmarshalText([]byte(p2pIdString))
	if err != nil {
		return p2ptypes.PeerID{}, fmt.Errorf("failed to unmarshal PeerID: %w", err)
	}
	return peerID, nil
}
