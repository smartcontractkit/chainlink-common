package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	cmd := NewRootCmd()
	tempDir := t.TempDir()
	keystoreFile := filepath.Join(tempDir, "keystore.json")
	os.Create(keystoreFile)
	cmd.Flags().Set("file-path", keystoreFile)
	// cmd.Flags().Set("password", "testpassword")
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	cmd.SetErr(buf)
	require.NoError(t, cmd.Execute())
	t.Log(buf.String())
}
