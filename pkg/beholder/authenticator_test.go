package beholder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func defaultTestingConfig() beholder.Config {
	return beholder.Config{
		AuthenticatorPublicKey: []byte("test-public-key"),
		AuthenticatorSigner:    func([]byte) []byte { return []byte("test-signature") },
		AuthenticatorHeaders:   nil,
	}
}
func TestAuthenticator_SignerAuth(t *testing.T) {
	// Authenticator should derive headers if the signer is set
	c := defaultTestingConfig()
	a, err := beholder.NewAuthenticator(c)
	assert.NoError(t, err)

	expectedHeaders := map[string]string{
		"X-Beholder-Node-Auth-Token": "1:746573742d7075626c69632d6b6579:746573742d7369676e6174757265",
	}
	assert.Equal(t, expectedHeaders, a.GetHeaders())
	assert.Equal(t, c.AuthenticatorPublicKey, a.GetPubKey())
}

func TestAuthenticator_HeadersAuth(t *testing.T) {
	// Authenticator should use the headers if they are set
	expectedHeaders := map[string]string{"test-header-key": "test-header-value"}
	c := beholder.Config{
		AuthenticatorHeaders:   expectedHeaders,
		AuthenticatorPublicKey: []byte("test-public-key"),
	}

	a, err := beholder.NewAuthenticator(c)
	assert.NoError(t, err)
	assert.Equal(t, expectedHeaders, a.GetHeaders())
	assert.Equal(t, c.AuthenticatorPublicKey, a.GetPubKey())
}

func TestAuthenticator_NoAuth(t *testing.T) {
	// Authenticator should not set any headers if neither signer nor headers are set
	c := beholder.Config{}

	a, err := beholder.NewAuthenticator(c)
	assert.NoError(t, err)
	expectedHeaders := map[string]string{}
	assert.Equal(t, expectedHeaders, a.GetHeaders())
	assert.Equal(t, []byte(nil), a.GetPubKey())
}

func TestAuthenticator_Config(t *testing.T) {
	// Configuring both auth signer and auth header should error
	c := defaultTestingConfig()
	c.AuthenticatorHeaders = map[string]string{"test-header-key": "test-header-value"}
	_, err := beholder.NewAuthenticator(c)
	assert.Error(t, err)

	// Configuring signer with no public key should error
	c = defaultTestingConfig()
	c.AuthenticatorPublicKey = nil
	_, err = beholder.NewAuthenticator(c)
	assert.Error(t, err)

	// Configuring headers with no public key should error
	c = defaultTestingConfig()
	c.AuthenticatorSigner = nil
	c.AuthenticatorHeaders = map[string]string{"test-header-key": "test-header-value"}
	c.AuthenticatorPublicKey = nil
	_, err = beholder.NewAuthenticator(c)
	assert.Error(t, err)

	// Configuring neither auth signer, nor auth heade, nor pub key should not error
	c = defaultTestingConfig()
	c.AuthenticatorSigner = nil
	c.AuthenticatorPublicKey = nil
	c.AuthenticatorHeaders = nil
	_, err = beholder.NewAuthenticator(c)
	assert.NoError(t, err)
}
