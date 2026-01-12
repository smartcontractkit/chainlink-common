package kms_test

import (
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	kmsinternal "github.com/smartcontractkit/chainlink-common/keystore/kms/internal"
	"github.com/stretchr/testify/require"
)

func TestSEC1ToASN1PublicKey(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Geth library uses SEC1 format.
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)
	require.Len(t, sec1PubKey, 65)
	require.Equal(t, byte(0x04), sec1PubKey[0])

	// Convert to ASN.1
	asn1PubKey, err := kmsinternal.SEC1ToASN1PublicKey(sec1PubKey)
	require.NoError(t, err)
	log.Println("asn1PubKey", len(asn1PubKey))

	// Convert back to SEC1
	sec1PubKey2, err := kmsinternal.ASN1ToSEC1PublicKey(asn1PubKey)
	require.NoError(t, err)
	require.Len(t, sec1PubKey2, 65)
	require.Equal(t, byte(0x04), sec1PubKey2[0])
	require.Equal(t, privateKey.PublicKey.X.Bytes(), sec1PubKey2[1:33])
	require.Equal(t, privateKey.PublicKey.Y.Bytes(), sec1PubKey2[33:65])
}

func TestASN1SignatureToSEC1Signature(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	sec1PubKey := crypto.FromECDSAPub(&privateKey.PublicKey)

	hash := crypto.Keccak256Hash([]byte("test"))

	sig, err := crypto.Sign(hash[:], privateKey)
	require.NoError(t, err)

	asn1Sig, err := kmsinternal.SEC1ToASN1Sig(sig)
	require.NoError(t, err)

	// We pass the expected SEC1 public key for verification.
	sec1Sig, err := kmsinternal.ASN1ToSEC1Sig(asn1Sig, sec1PubKey, hash[:])
	require.NoError(t, err)
	require.Len(t, sec1Sig, 65)
	require.Equal(t, sig, sec1Sig)
}
