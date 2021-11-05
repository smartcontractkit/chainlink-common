package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/config"
	"github.com/smartcontractkit/chainlink-relay/core/server/types"
	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://github.com/gin-gonic/gin#testing

func TestNewHTTPService_Routes(t *testing.T) {
	t.Parallel()

	srv := NewHTTPService(&test.MockWebhookConfig{}, test.MockStore{}, test.MockPipeline{}, &map[string]string{})

	routes := []struct {
		name   string
		route  string
		method string
		body   interface{}
		code   int
	}{
		// requests that pass
		{"health-get", "/health", http.MethodGet, struct{}{}, 200},
		{"keys-get", "/keys", http.MethodGet, struct{}{}, 200},
		{"keys-set", "/keys", http.MethodPost, types.SetKeyData{"test", "test", "test", "test"}, 201},
		{"runs-post", "/runs", http.MethodPost, types.JobRunData{}, 201},
		{"jobs-post", "/jobs", http.MethodPost, types.CreateJobReq{}, 307},
		{"jobs-delete", "/jobs/test-job-id", http.MethodDelete, struct{}{}, 200},
		// requests that fail
		{"fail-method", "/jobs", http.MethodGet, types.CreateJobReq{}, 404},
		{"fail-endpoint", "/random", http.MethodPost, types.JobRunData{}, 404},
		{"fail-payload", "/runs", http.MethodPost, false, 400},
		{"fail-missingKeys", "/keys", http.MethodPost, types.SetKeyData{}, 400},
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

	// clear secrest set during POST /keys
	assert.NoError(t, test.UnsetEIKeysSecrets())
}

func TestPostNewKeys_IntegrationTest(t *testing.T) {
	// create environment variables
	test.MockSetRequiredConfigs(t, []string{"IC_ACCESSKEY", "IC_SECRET", "CI_SECRET", "CI_ACCESSKEY"})
	cfg, err := config.NewWebhookConfig()
	require.NoError(t, err)

	// start up relay server
	srv := NewHTTPService(cfg, test.MockStore{}, test.MockPipeline{}, &map[string]string{})

	// create webhook outgoing trigger + mock CL node endpoint
	serverCL := test.MockServer(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "test_ICKey", req.Header.Get("X-Chainlink-EA-AccessKey"))
		assert.Equal(t, "test_ICSecret", req.Header.Get("X-Chainlink-EA-Secret"))
		require.NoError(t, test.WriteResponse(rw, http.StatusOK, nil))
	})
	defer serverCL.Close()
	outgoing := webhook.NewTrigger(serverCL.URL, cfg)

	// track starting env vars
	ICKey0 := cfg.ICKey()
	ICSecret0 := cfg.ICSecret()
	CIKey0 := cfg.CIKey()
	CISecret0 := cfg.CISecret()

	// update env vars by using POST /keys
	w := httptest.NewRecorder()
	body, err := test.RequestBody(types.SetKeyData{"test_ICKey", "test_ICSecret", "test_CIKey", "test_CISecret"})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/keys", body)
	assert.NoError(t, err)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	// verify that keys and secrets changed
	assert.NotEqual(t, ICKey0, cfg.ICKey())
	assert.NotEqual(t, ICSecret0, cfg.ICSecret())
	assert.NotEqual(t, CIKey0, cfg.CIKey())
	assert.NotEqual(t, CISecret0, cfg.CISecret())

	// verify webhook incoming uses new keys
	w = httptest.NewRecorder()
	body, err = test.RequestBody(struct{}{})
	require.NoError(t, err)
	req, err = http.NewRequest(http.MethodDelete, "/jobs/test-job-id", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Chainlink-EA-AccessKey", "test_CIKey")
	req.Header.Add("X-Chainlink-EA-Secret", "test_CISecret")
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// verify webhook outgoing uses new keys (triggers the mock CL node server above)
	outgoing.TriggerJob("test-job-id")

	// clear secrest set during POST /keys
	assert.NoError(t, test.UnsetEIKeysSecrets())
}
