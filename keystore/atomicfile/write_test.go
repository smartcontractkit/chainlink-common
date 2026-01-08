package atomicfile

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFile_WriteAndRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	data := []byte("test")
	err := WriteFile(path, bytes.NewReader(data), 0600)
	require.NoError(t, err)
	readData, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, readData, data)
}
