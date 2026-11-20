package host

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v47"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func writeTestArtifact(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "module")
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

func serializeTestModule(t *testing.T, body string) string {
	t.Helper()
	raw, err := wasmtime.Wat2Wasm(`(module
  (import "env" "version_v2" (func))
  (memory (export "memory") 1)
  (func (export "_start") ` + body + `))`)
	require.NoError(t, err)
	compiled, err := NewModule(t.Context(), &ModuleConfig{Logger: logger.Test(t), IsUncompressed: true}, raw)
	require.NoError(t, err)
	defer compiled.Close()
	artifact, err := compiled.Serialize()
	require.NoError(t, err)
	return writeTestArtifact(t, artifact)
}

func TestSerializedModuleRoundTrip(t *testing.T) {
	path := serializeTestModule(t, "")
	var sdk string
	cfg := &ModuleConfig{Logger: logger.Test(t), SdkLabeler: func(name string) { sdk = name }}
	loaded, err := NewModuleFromSerializedFile(t.Context(), cfg, path)
	require.NoError(t, err)
	defer loaded.Close()
	require.Equal(t, "version_v2", sdk)
	require.NoError(t, os.Remove(path))

	helper := mocks.NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("id")
	result, err := loaded.Execute(t.Context(), &sdkpb.ExecuteRequest{}, helper)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestSerializedModuleRejectsInvalidArtifacts(t *testing.T) {
	cfg := &ModuleConfig{Logger: logger.Test(t)}
	_, err := NewModuleFromSerializedFile(t.Context(), cfg, filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)

	raw, err := wasmtime.Wat2Wasm("(module)")
	require.NoError(t, err)
	native, err := wasmtime.NewModule(GetEngine(cfg.Logger).Engine, raw)
	require.NoError(t, err)
	defer native.Close()
	artifact, err := native.Serialize()
	require.NoError(t, err)
	for _, tc := range []struct {
		name string
		data []byte
		want string
	}{
		{"source wasm", raw, "error deserializing wasmtime module"},
		{"missing memory", artifact, `module does not export "memory"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewModuleFromSerializedFile(t.Context(), cfg, writeTestArtifact(t, tc.data))
			require.ErrorContains(t, err, tc.want)
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = NewModuleFromSerializedFile(ctx, cfg, "unused")
	require.ErrorIs(t, err, context.Canceled)
}

func TestSerializedModuleEpochTimeout(t *testing.T) {
	path := serializeTestModule(t, "(loop $spin (br $spin))")
	timeout := 200 * time.Millisecond
	cfg := &ModuleConfig{Logger: logger.Test(t), Timeout: &timeout}
	loaded, err := NewModuleFromSerializedFile(t.Context(), cfg, path)
	require.NoError(t, err)
	defer loaded.Close()
	helper := mocks.NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("id")
	_, err = loaded.Execute(t.Context(), &sdkpb.ExecuteRequest{}, helper)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
