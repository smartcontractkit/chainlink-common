package artifacts

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const testServerPort = "45001"

type UploadTestSuite struct {
	suite.Suite

	lggr    *slog.Logger
	server  *http.Server
	storage *testArtifactStorage
}

func (s *UploadTestSuite) SetupSuite() {
	s.lggr = slog.New(slog.NewTextHandler(os.Stdout, nil))
	s.storage = newTestArtifactStorage()
	s.server = newTestServer(testServerPort, s.storage)
	go func() {
		_ = s.server.ListenAndServe()
	}()
}

func (s *UploadTestSuite) TearDownSuite() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.server.Shutdown(ctx)
	}
}

func (s *UploadTestSuite) TestUpload() {
	artifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "./testdata/config.yaml",
		BinaryPath:    "testdata/binary",
	}, s.lggr)

	err := artifacts.Compile()
	s.NoError(err, "failed to compile workflow")

	err = artifacts.Prepare()
	s.NoError(err, "failed to prepare artifacts")

	uploadInput := &UploadInput{
		PresignedURL: fmt.Sprintf("http://localhost:%s/artifacts/%s/binary.wasm", testServerPort, artifacts.GetWorkflowID()),
		PresignedFields: []Field{
			{Key: "key1", Value: "value1"},
		},
		Filepath: artifacts.GetBinaryPath(),
		Timeout:  10 * time.Second,
	}
	err = artifacts.DurableUpload(uploadInput)
	s.NoError(err, "failed to upload artifact")
	expected := artifacts.GetBinaryData()
	actual := s.storage.getBinary(artifacts.GetWorkflowID())
	expLen, actLen := len(expected), len(actual)
	s.Equal(expLen, actLen, "binary data length do not match")
	n := 100
	if expLen > n && actLen > n {
		// Compare only the first 100 bytes and the last 100 bytes of the artifact binary
		// Otherwise, the entire binary is printed to console
		s.Equal(expected[:n], actual[:n], "first 100 bytes do not match")
		s.Equal(expected[expLen-n:], actual[actLen-n:], "last 100 bytes do not match")
	} else {
		// If the binary is smaller than 100 bytes, compare the whole thing
		s.Equal(expected, actual, "binary data do not match")
	}
}

func TestUploadTestSuite(t *testing.T) {
	suite.Run(t, new(UploadTestSuite))
}

// testArtifactStorage holds multipart form data received by the test server.
type testArtifactStorage struct {
	mu       sync.RWMutex
	binaries map[string][]byte // workflowID -> body
	configs  map[string][]byte // workflowID -> body
}

func newTestArtifactStorage() *testArtifactStorage {
	return &testArtifactStorage{
		binaries: make(map[string][]byte),
		configs:  make(map[string][]byte),
	}
}

func (t *testArtifactStorage) getBinary(workflowID string) []byte {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.binaries[workflowID]
}

func (t *testArtifactStorage) getConfig(workflowID string) []byte {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.configs[workflowID]
}

// testServer runs on the given port and accepts POST multipart to:
//   - /artifacts/<workflow-id>/binary.wasm
//   - /artifacts/<workflow-id>/configs
func newTestServer(port string, storage *testArtifactStorage) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/artifacts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, "/artifacts/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) != 2 {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		workflowID, suffix := parts[0], parts[1]
		if workflowID == "" {
			http.Error(w, "missing workflow id", http.StatusBadRequest)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		storage.mu.Lock()
		switch suffix {
		case "binary.wasm":
			storage.binaries[workflowID] = body
		case "configs":
			storage.configs[workflowID] = body
		default:
			storage.mu.Unlock()
			http.Error(w, "unknown artifact type: "+suffix, http.StatusBadRequest)
			return
		}
		storage.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	return &http.Server{
		Addr:    "localhost:" + port,
		Handler: mux,
	}
}
