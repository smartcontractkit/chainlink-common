package kms_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	kms "github.com/smartcontractkit/chainlink-common/keystore/kms"
)

func TestSEC1ToASN1PublicKey(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Geth library uses SEC1 format.
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)
	require.Len(t, sec1PubKey, 65)
	require.Equal(t, byte(0x04), sec1PubKey[0])

	// Convert to ASN.1
	asn1PubKey, err := kms.SEC1ToASN1PublicKey(sec1PubKey)
	require.NoError(t, err)

	// Convert back to SEC1
	sec1PubKey2, err := kms.ASN1ToSEC1PublicKey(asn1PubKey)
	require.NoError(t, err)
	require.Len(t, sec1PubKey2, 65)
	require.Equal(t, byte(0x04), sec1PubKey2[0])
	pubKey := privateKey.PublicKey
	require.Equal(t, pubKey.X.Bytes(), sec1PubKey2[1:33])
	require.Equal(t, pubKey.Y.Bytes(), sec1PubKey2[33:65])
}

func TestASN1SignatureToSEC1Signature(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)

	hash := crypto.Keccak256Hash([]byte("test"))

	sig, err := crypto.Sign(hash[:], privateKey)
	require.NoError(t, err)

	asn1Sig, err := kms.SEC1ToASN1Sig(sig)
	require.NoError(t, err)

	// We pass the expected SEC1 public key for verification.
	sec1Sig, err := kms.ASN1ToSEC1Sig(asn1Sig, sec1PubKey, hash[:])
	require.NoError(t, err)
	require.Len(t, sec1Sig, 65)
	require.Equal(t, sig, sec1Sig)
}
