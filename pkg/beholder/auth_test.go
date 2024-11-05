package beholder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAuthHeaders(t *testing.T) {
	mockPubKey := []byte("test-public-key")
	mockSigner := func(data []byte) []byte {
		return append(data, []byte("__test-signature")...)
	}

	expectedHeaders := map[string]string{
		"X-Beholder-Node-Auth-Token": "1:746573742d7075626c69632d6b6579:746573742d7075626c69632d6b65795f5f746573742d7369676e6174757265",
	}

	assert.Equal(t, expectedHeaders, BuildAuthHeaders(mockSigner, mockPubKey))
}
