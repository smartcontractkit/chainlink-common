package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_tokenStore(t *testing.T) {
	s := NewTokenStore("./cfg.yml")
	token := "authtoken"

	// Ensure we remove the cookie even if the test fails
	t.Cleanup(func() { os.Remove("./authToken.cache") })

	// Save the cookie to disk
	err := s.Save(token)
	require.NoError(t, err)

	_, err = os.Stat("./authToken.cache")
	assert.NoError(t, err)

	// Retrieve the cookie from disk
	cs, err := s.Get()
	require.NoError(t, err)
	assert.Equal(t, token, cs)

	// Delete the cookie from disk
	err = s.Delete()
	require.NoError(t, err)

	// Should not return an error when the tokenStore is empty
	_, err = s.Get()
	assert.NoError(t, err)
}
