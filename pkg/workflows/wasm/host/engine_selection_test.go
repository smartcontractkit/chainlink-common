package host

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/engine"
	"github.com/stretchr/testify/require"
)

func TestEngineSelection_Default(t *testing.T) {
	binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}
	m, err := NewModule(t.Context(), mc, binary)
	require.NoError(t, err)
	defer m.Close()
	require.Equal(t, string(engine.EngineWasmtime), mc.Engine)
}

func TestEngineSelection_ExplicitWasmtime(t *testing.T) {
	binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		Engine:         string(engine.EngineWasmtime),
	}
	m, err := NewModule(t.Context(), mc, binary)
	require.NoError(t, err)
	defer m.Close()
}

func TestEngineSelection_Invalid(t *testing.T) {
	binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)

	mc := &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
		Engine:         "nonexistent",
	}
	_, err := NewModule(t.Context(), mc, binary)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonexistent")
}
