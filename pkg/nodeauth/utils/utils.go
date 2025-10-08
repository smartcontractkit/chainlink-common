package utils

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
	"google.golang.org/protobuf/proto"
)

// CalculateRequestDigest creates a SHA256 digest of the request for integrity verification
// This function is shared between client (JWT generation) and server (JWT validation)
func CalculateRequestDigest(req any) string {
	var data []byte
	if m, ok := req.(proto.Message); ok {
		// Use protobuf canonical serialization
		serialized, err := proto.Marshal(m)
		if err == nil {
			data = serialized
		} else {
			// fallback to string representation if marshal fails
			data = fmt.Appendf(nil, "%v", req)
		}
	} else if s, ok := req.(fmt.Stringer); ok {
		data = []byte(s.String())
	} else {
		data = fmt.Appendf(nil, "%v", req)
	}

	hash := sha256.Sum256(data)
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
