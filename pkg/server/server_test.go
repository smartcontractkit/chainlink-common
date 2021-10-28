package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/pkg/server/types"
	"github.com/smartcontractkit/chainlink-relay/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://github.com/gin-gonic/gin#testing

func TestNewHTTPService_Routes(t *testing.T) {
	t.Parallel()

	srv := NewHTTPService("", "", test.MockStore{}, test.MockPipeline{}, &map[string]string{})

	routes := []struct {
		name   string
		route  string
		method string
		body   interface{}
		code   int
	}{
		// requests that pass
		{"health-get", "/health", http.MethodGet, struct{}{}, 200},
		{"keys-get", "/health", http.MethodGet, struct{}{}, 200},
		{"runs-post", "/runs", http.MethodPost, types.JobRunData{}, 201},
		{"jobs-post", "/jobs", http.MethodPost, types.CreateJobReq{}, 307},
		{"jobs-delete", "/jobs/test-job-id", http.MethodDelete, struct{}{}, 200},
		// requests that fail
		{"fail-method", "/jobs", http.MethodGet, types.CreateJobReq{}, 404},
		{"fail-endpoint", "/random", http.MethodPost, types.JobRunData{}, 404},
		{"fail-payload", "/runs", http.MethodPost, false, 400},
	}

	for _, r := range routes {
		t.Run(r.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body, err := test.RequestBody(r.body)
			require.NoError(t, err)
			req, err := http.NewRequest(r.method, r.route, body)
			assert.NoError(t, err)
			srv.Router.ServeHTTP(w, req)
			assert.Equal(t, r.code, w.Code)
		})
	}
}
