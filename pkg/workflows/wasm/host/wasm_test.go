package host

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
)

const (
	oomBinaryLocation    = "test/oom/cmd/testmodule.wasm"
	oomBinaryCmd         = "test/oom/cmd"
	sleepBinaryLocation  = "test/sleep/cmd/testmodule.wasm"
	sleepBinaryLocation2 = "test/sleep/cmd/testmodule_2.wasm" // used to avoid a build race between tests
	sleepBinaryCmd       = "test/sleep/cmd"
	// distinct output paths so the parallel size tests don't race on the
	// same .wasm file while building the same package.
	stdioCompressedBinaryLocation   = "test/stdio/cmd/testmodule_size1.wasm"
	stdioDecompressedBinaryLocation = "test/stdio/cmd/testmodule_size2.wasm"
)

func createTestBinary(outputPath, path string, uncompressed bool, t testing.TB) []byte {
	cmd := exec.Command("go", "build", "-o", path, "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/"+outputPath) // #nosec
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	binary, err := os.ReadFile(path)
	require.NoError(t, err)

	if uncompressed {
		return binary
	}

	var b bytes.Buffer
	bwr := brotli.NewWriter(&b)
	_, err = bwr.Write(binary)
	require.NoError(t, err)
	require.NoError(t, bwr.Close())

	cb, err := io.ReadAll(&b)
	require.NoError(t, err)
	return cb
}

func TestModule_CompressedBinarySize(t *testing.T) {
	t.Parallel()

	t.Run("compressed binary size is smaller than the default 10mb limit", func(t *testing.T) {
		binary := createTestBinary(stdioBinaryCmd, stdioCompressedBinaryLocation, false, t)

		_, err := NewModule(t.Context(), &ModuleConfig{IsUncompressed: false, Logger: logger.Test(t)}, binary)
		require.NoError(t, err)
	})

	t.Run("compressed binary size is bigger than the default 10mb limit", func(t *testing.T) {
		binary := make([]byte, defaultMaxCompressedBinarySize+1)

		var b bytes.Buffer
		bwr := brotli.NewWriter(&b)
		_, err := bwr.Write(binary)
		require.NoError(t, err)
		require.NoError(t, bwr.Close())

		_, err = NewModule(t.Context(), &ModuleConfig{IsUncompressed: false, Logger: logger.Test(t)}, binary)
		require.ErrorContains(t, err, "binary size exceeds the maximum allowed size")
		var limitErr limits.ErrorBoundLimited[config.Size]
		require.ErrorAs(t, err, &limitErr)
		assert.Equal(t, defaultMaxCompressedBinarySize, int(limitErr.Limit))
	})

	t.Run("compressed binary size is bigger than the custom limit", func(t *testing.T) {
		customMaxCompressedBinarySize := uint64(1 * 1024 * 1024)
		binary := make([]byte, customMaxCompressedBinarySize+1)

		var b bytes.Buffer
		bwr := brotli.NewWriter(&b)
		_, err := bwr.Write(binary)
		require.NoError(t, err)
		require.NoError(t, bwr.Close())

		_, err = NewModule(t.Context(), &ModuleConfig{IsUncompressed: false, MaxCompressedBinarySize: customMaxCompressedBinarySize, Logger: logger.Test(t)}, binary)
		require.ErrorContains(t, err, "binary size exceeds the maximum allowed size")
		var limitErr limits.ErrorBoundLimited[config.Size]
		require.ErrorAs(t, err, &limitErr)
		assert.Equal(t, customMaxCompressedBinarySize, uint64(limitErr.Limit))
	})
}

func TestModule_DecompressedBinarySize(t *testing.T) {
	t.Parallel()

	binary := createTestBinary(stdioBinaryCmd, stdioDecompressedBinaryLocation, false, t)
	rdr := brotli.NewReader(bytes.NewBuffer(binary))
	decompedBinary, err := io.ReadAll(rdr)
	require.NoError(t, err)
	t.Run("decompressed binary size is within the limit", func(t *testing.T) {
		customDecompressedBinarySize := uint64(len(decompedBinary))
		_, err := NewModule(t.Context(), &ModuleConfig{IsUncompressed: false, MaxDecompressedBinarySize: customDecompressedBinarySize, Logger: logger.Test(t)}, binary)
		require.NoError(t, err)
	})

	t.Run("decompressed binary size is bigger than the limit", func(t *testing.T) {
		customDecompressedBinarySize := uint64(len(decompedBinary) - 1)
		_, err := NewModule(t.Context(), &ModuleConfig{IsUncompressed: false, MaxDecompressedBinarySize: customDecompressedBinarySize, Logger: logger.Test(t)}, binary)
		require.ErrorContains(t, err, "decompressed binary size reached the maximum allowed size")
		var limitErr limits.ErrorBoundLimited[config.Size]
		require.ErrorAs(t, err, &limitErr)
		assert.Equal(t, customDecompressedBinarySize, uint64(limitErr.Limit))
	})
}
