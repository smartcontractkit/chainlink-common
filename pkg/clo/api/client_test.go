package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/feeds-manager/api/operations"
)

func Test_Client(t *testing.T) {
	testFile := "./authToken.cache"

	_, err := os.Create(testFile)
	require.NoError(t, err)

	// Ensure we remove the token even if the test fails
	t.Cleanup(func() { os.Remove("./authToken.cache") })

	server := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/query")
		assert.Equal(t, r.Method, "POST")
		_, werr := w.Write([]byte(`{
			"login": {
				"session": {
					"id": "123",
					"token": "abc"
				}
			}
		}`))
		require.NoError(t, werr)
	})
	defer server.Close()

	cfg := &Config{
		BaseUrl:  server.URL,
		Email:    "123@abc.xyz",
		Password: "password123",
		FilePath: testFile,
	}

	client, err := NewClient(context.Background(), cfg)
	assert.NoError(t, err)

	ctx := client.Ctx()
	assert.Equal(t, ctx, client.context)

	gql := client.Gql()
	assert.Equal(t, gql, client.graphQL)

	resp, err := client.Login()
	assert.NoError(t, err)
	assert.Equal(t, *resp, operations.LoginResponse{})

	err = client.Logout()
	assert.NoError(t, err)
}

func newMockServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
