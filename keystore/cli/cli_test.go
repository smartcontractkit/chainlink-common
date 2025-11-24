package cli_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	ks "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/cli"
)

func TestCLI(t *testing.T) {
	cmd := cli.NewRootCmd()
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	keystoreFile := filepath.Join(tempDir, "keystore.json")
	f, err := os.Create(keystoreFile)
	require.NoError(t, err)
	defer f.Close()
	os.Setenv("KEYSTORE_FILE_PATH", keystoreFile)
	os.Setenv("KEYSTORE_PASSWORD", "testpassword")

	// No error just listing help.
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.ExecuteContext(t.Context()))

	// Create a key.
	buf.Reset()
	cmd.SetArgs([]string{"create", "-d", `{"Keys": [{"KeyName": "testkey", "KeyType": "X25519"}]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))

	// List keys.
	buf.Reset()
	cmd.SetArgs([]string{"get", "-d", `{"KeyNames": ["testkey"]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	resp := ks.GetKeysResponse{}
	err = json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)

	// Create a second key we export.
	buf.Reset()
	cmd.SetArgs([]string{"create", "-d", `{"Keys": [{"KeyName": "testkey2", "KeyType": "ECDSA_S256"}]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))

	// Export the second key.
	buf.Reset()
	cmd.SetArgs([]string{"export", "-d", `{"Keys": [{"KeyName": "testkey2", "Enc": {"Password": "testpassword2", "ScryptParams": {"N": 1024, "P": 1, "R": 8}}}]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	exportResp := ks.ExportKeysResponse{}
	err = json.Unmarshal(buf.Bytes(), &exportResp)
	require.NoError(t, err)
	exportedKey2Data := base64.StdEncoding.EncodeToString(exportResp.Keys[0].Data)

	// Delete the second key.
	buf.Reset()
	// Force deletion without confirmation.
	cmd.SetArgs([]string{"delete", "-d", `{"KeyNames": ["testkey2"]}`, "--yes"})
	require.NoError(t, cmd.ExecuteContext(t.Context()))

	// List key should only see first.
	buf.Reset()
	cmd.SetArgs([]string{"list"})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)

	// Import the exported key.
	buf.Reset()
	cmd.SetArgs([]string{"import", "-d", `{"Keys": [{"KeyName": "testkey2", "Data": "` + exportedKey2Data + `", "Password": "testpassword2"}]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))

	// List keys.
	buf.Reset()
	cmd.SetArgs([]string{"list"})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 2)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)
	require.Equal(t, "testkey2", resp.Keys[1].KeyInfo.Name)
	require.Equal(t, ks.ECDSA_S256, resp.Keys[1].KeyInfo.KeyType)

	// Set metadata on testkey.
	buf.Reset()
	metadata := base64.StdEncoding.EncodeToString([]byte("my-custom-metadata"))
	cmd.SetArgs([]string{"set-metadata", "-d", `{"Updates": [{"KeyName": "testkey", "Metadata": "` + metadata + `"}]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))

	// Verify metadata was set.
	buf.Reset()
	cmd.SetArgs([]string{"get", "-d", `{"KeyNames": ["testkey"]}`})
	require.NoError(t, cmd.ExecuteContext(t.Context()))
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(buf.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	// Metadata is []byte, Go's JSON unmarshaler automatically decodes base64 strings
	require.Equal(t, "my-custom-metadata", string(resp.Keys[0].KeyInfo.Metadata))
}
