package beholder_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func defaultTestingConfig() beholder.Config {
	return beholder.Config{
		CSAAuthEnabled:     true,
		CSAPublicKey:       []byte("test-public-key"),
		CSASigner:          func([]byte) []byte { return []byte("test-signature") },
		InsecureConnection: false,
	}
}

func TestNodeAuthenticator_HappyPath(t *testing.T) {
	config := defaultTestingConfig()
	na, err := beholder.NewNodeAuthenticator(config)
	assert.NoError(t, err)

	// Test GetRequestMetadata
	expectedMessage := hex.EncodeToString([]byte("test-public-key"))
	expectedSignature := hex.EncodeToString([]byte("test-signature"))
	expectedRequestMetadata := map[string]string{
		"X-Beholder-Node-Auth-Token": fmt.Sprintf("1:%s:%s", expectedMessage, expectedSignature),
	}
	requestMetadata, err := na.GetRequestMetadata(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedRequestMetadata, requestMetadata)

	// Test RequireTransportSecurity
	assert.True(t, na.RequireTransportSecurity())
}

func TestNodeAuthenticator_NodeAuthConfig(t *testing.T) {
	// Should error on nil public CSA key
	c1 := defaultTestingConfig()
	c1.CSAPublicKey = nil
	_, err := beholder.NewNodeAuthenticator(c1)
	assert.Error(t, err, "CSA auth is enabled but no CSA public key was provided")

	// Should error on nil CSA signer
	c2 := defaultTestingConfig()
	c2.CSASigner = nil
	_, err = beholder.NewNodeAuthenticator(c2)
	assert.Error(t, err, "CSA auth is enabled but no CSA signer was provided")
}

func TestNodeAuthenticator_CSAAuthDisabled(t *testing.T) {
	// CSA Auth disabled should accept nil values and return empty map for request metadata
	c := defaultTestingConfig()
	c.CSAAuthEnabled = false
	c.CSAPublicKey = nil
	c.CSASigner = nil

	na, err := beholder.NewNodeAuthenticator(c)
	assert.NoError(t, err)

	expectedRequestMetadata := map[string]string{}
	requestMetadata, err := na.GetRequestMetadata(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedRequestMetadata, requestMetadata)
}
