package beholder

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAuthHeadersV1(t *testing.T) {
	csaPrivKey, err := generateTestCSAPrivateKey()
	require.NoError(t, err)

	expectedHeaders := map[string]string{
		"X-Beholder-Node-Auth-Token": "1:cab39509e63cfaa81c70e2c907391f96803aacb00db5619a5ace5588b4b08159:4403178e299e9acc5b48ae97de617d3975c5d431b794cfab1d23eda01c194119b2360f5f74cfb3e4f706237ab57a0ba88ffd3f8addbc1e5197b3d3e13a1fc409",
	}

	assert.Equal(t, expectedHeaders, BuildAuthHeaders(csaPrivKey))
}

func TestBuildAuthHeadersV2(t *testing.T) {
	csaPrivKey, err := generateTestCSAPrivateKey()
	require.NoError(t, err)
	timestamp := time.Now().UnixMilli()

	authHeaderMap := BuildAuthHeadersV2(csaPrivKey, &AuthHeaderConfig{
		timestamp: timestamp,
	})

	authHeaderValue, ok := authHeaderMap[authHeaderKey]
	require.True(t, ok, "auth header should be present")

	parts := strings.Split(authHeaderValue, ":")
	assert.Len(t, parts, 4, "auth header v2 should have 4 parts")
	// Check the parts
	version, pubKeyHex, timestampStr, signatureHex := parts[0], parts[1], parts[2], parts[3]
	assert.Equal(t, authHeaderVersion2, version, "BuildAuthHeadersV2 should should have version 2")
	assert.Equal(t, hex.EncodeToString(csaPrivKey.Public().(ed25519.PublicKey)), pubKeyHex)
	assert.Equal(t, strconv.FormatInt(timestamp, 10), timestampStr)

	// Decode the public key and signature
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	require.NoError(t, err)
	assert.Equal(t, csaPrivKey.Public().(ed25519.PublicKey), ed25519.PublicKey(pubKeyBytes))

	// Parse the timestamp
	timestampParsed, err := strconv.ParseInt(timestampStr, 10, 64)
	require.NoError(t, err)
	assert.Equal(t, timestamp, timestampParsed)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestampParsed))

	// Reconstruct the message bytes
	messageBytes := append(pubKeyBytes, timestampBytes...)

	// Verify the signature
	signatureBytes, err := hex.DecodeString(signatureHex)
	require.NoError(t, err)
	assert.True(t, ed25519.Verify(pubKeyBytes, messageBytes, signatureBytes))
}

func TestBuildAuthHeadersV2WithDefaults(t *testing.T) {
	csaPrivKey, err := generateTestCSAPrivateKey()
	require.NoError(t, err)

	authHeaderMap := BuildAuthHeadersV2(csaPrivKey, nil)
	authHeaderValue, ok := authHeaderMap[authHeaderKey]
	require.True(t, ok, "auth header should be present")

	parts := strings.Split(authHeaderValue, ":")
	assert.Len(t, parts, 4, "auth header v2 should have 4 parts")
	// Check the parts
	version, pubKeyHex, timestampStr, signatureHex := parts[0], parts[1], parts[2], parts[3]
	assert.Equal(t, "2", version, "using WithAuthHeaderV2 should should have version 2")
	assert.Equal(t, hex.EncodeToString(csaPrivKey.Public().(ed25519.PublicKey)), pubKeyHex)

	// Decode the public key and signature
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	require.NoError(t, err)
	assert.Equal(t, csaPrivKey.Public().(ed25519.PublicKey), ed25519.PublicKey(pubKeyBytes))

	// Parse the timestamp
	timestampParsed, err := strconv.ParseInt(timestampStr, 10, 64)
	require.NoError(t, err)

	// Verify the timestamp is within the last 5 seconds
	// This verifies that default configuration is to use the current time
	now := time.Now().UnixMilli()
	assert.InDelta(t, now, timestampParsed, 5000, "timestamp should be within the last 5 seconds")

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestampParsed))

	// Reconstruct the message bytes
	messageBytes := append(pubKeyBytes, timestampBytes...)

	// Verify the signature
	signatureBytes, err := hex.DecodeString(signatureHex)
	require.NoError(t, err)
	assert.True(t, ed25519.Verify(pubKeyBytes, messageBytes, signatureBytes))
}

func TestBuildAuthHeadersV2WithNegativeTimestamp(t *testing.T) {
	csaPrivKey, err := generateTestCSAPrivateKey()
	require.NoError(t, err)
	timestamp := int64(-111)

	authHeaderMap := BuildAuthHeadersV2(csaPrivKey, &AuthHeaderConfig{
		timestamp: timestamp,
	})

	authHeaderValue, ok := authHeaderMap[authHeaderKey]
	require.True(t, ok, "auth header should be present")

	parts := strings.Split(authHeaderValue, ":")
	assert.Len(t, parts, 4, "auth header v2 should have 4 parts")
	// Check the the returned timestamp is 0
	_, _, timestampStr, _ := parts[0], parts[1], parts[2], parts[3]
	timestampParsed, err := strconv.ParseInt(timestampStr, 10, 64)
	require.NoError(t, err)
	assert.Zero(t, timestampParsed)
}

func generateTestCSAPrivateKey() (ed25519.PrivateKey, error) {
	csaPrivKeyHex := "1ac84741fa51c633845fa65c06f37a700303619135630a01f2d22fb98eb1c54ecab39509e63cfaa81c70e2c907391f96803aacb00db5619a5ace5588b4b08159"
	csaPrivKeyBytes, err := hex.DecodeString(csaPrivKeyHex)
	if err != nil {
		return nil, err
	}
	return ed25519.PrivateKey(csaPrivKeyBytes), nil
}
