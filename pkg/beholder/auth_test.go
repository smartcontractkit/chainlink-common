package beholder_test

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestBuildAuthHeaders(t *testing.T) {
	csaPrivKeyHex := "1ac84741fa51c633845fa65c06f37a700303619135630a01f2d22fb98eb1c54ecab39509e63cfaa81c70e2c907391f96803aacb00db5619a5ace5588b4b08159"
	csaPrivKeyBytes, err := hex.DecodeString(csaPrivKeyHex)
	assert.NoError(t, err)
	csaPrivKey := ed25519.PrivateKey(csaPrivKeyBytes)

	expectedHeaders := map[string]string{
		"X-Beholder-Node-Auth-Token": "1:cab39509e63cfaa81c70e2c907391f96803aacb00db5619a5ace5588b4b08159:4403178e299e9acc5b48ae97de617d3975c5d431b794cfab1d23eda01c194119b2360f5f74cfb3e4f706237ab57a0ba88ffd3f8addbc1e5197b3d3e13a1fc409",
	}

	headers := beholder.BuildAuthHeaders(csaPrivKey)
	assert.Equal(t, expectedHeaders, headers)

	headers, err = beholder.NewAuthHeaders(csaPrivKey)
	require.NoError(t, err)
	assert.Equal(t, expectedHeaders, headers)
}

func TestStaticAuthHeaderProvider(t *testing.T) {
	// Create test headers
	testHeaders := map[string]string{
		"header1": "value1",
		"header2": "value2",
	}

	// Create new header provider
	provider := beholder.NewStaticAuthHeaderProvider(testHeaders)

	// Get headers and verify they match
	headers, err := provider.Headers(t.Context())
	require.NoError(t, err)
	assert.Equal(t, testHeaders, headers)
}
