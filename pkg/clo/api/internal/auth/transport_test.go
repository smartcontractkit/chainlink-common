package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Transport(t *testing.T) {
	token := tokenStore{rootPath: "./"}
	err := token.Save("abc")
	require.NoError(t, err)

	// Ensure we remove the token even if the test fails
	t.Cleanup(func() { os.Remove("./authToken.cache") })

	server := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("X-Session-Token"), "abc")
	})
	defer server.Close()

	client := newHttpClient(token)
	res, err := client.Get(server.URL)
	assert.NoError(t, err)

	gqlClient := NewGqlClient(server.URL, token)
	assert.NotNil(t, gqlClient)

	err = token.Delete()
	require.NoError(t, err)

	res.Body.Close()
}

func newMockServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
