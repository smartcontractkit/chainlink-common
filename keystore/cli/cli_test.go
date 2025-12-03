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
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	keystoreFile := filepath.Join(tempDir, "keystore.json")
	f, err := os.Create(keystoreFile)
	require.NoError(t, err)
	defer f.Close()
	os.Setenv("KEYSTORE_FILE_PATH", keystoreFile)
	os.Setenv("KEYSTORE_PASSWORD", "testpassword")

	// No error just listing help.
	_, err = runCommand(t, nil, "")
	require.NoError(t, err)

	// Create a key.
	_, err = runCommand(t, nil, "create", "-d", `{"Keys": [{"KeyName": "testkey", "KeyType": "X25519"}]}`)
	require.NoError(t, err)

	// List keys.
	out, err := runCommand(t, nil, "get", "-d", `{"KeyNames": ["testkey"]}`)
	require.NoError(t, err)
	resp := ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)

	// Create a second key we export.
	_, err = runCommand(t, nil, "create", "-d", `{"Keys": [{"KeyName": "testkey2", "KeyType": "ECDSA_S256"}]}`)
	require.NoError(t, err)

	// Export the second key.
	out, err = runCommand(t, nil, "export", "-d", `{"Keys": [{"KeyName": "testkey2", "Enc": {"Password": "testpassword2", "ScryptParams": {"N": 1024, "P": 1, "R": 8}}}]}`)
	require.NoError(t, err)
	exportResp := ks.ExportKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &exportResp)
	require.NoError(t, err)
	exportedKey2Data := base64.StdEncoding.EncodeToString(exportResp.Keys[0].Data)

	// Delete the second key.
	_, err = runCommand(t, nil, "delete", "-d", `{"KeyNames": ["testkey2"]}`, "--yes")
	require.NoError(t, err)

	// List key should only see first.
	out, err = runCommand(t, nil, "list")
	require.NoError(t, err)
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)

	// Import the exported key.
	_, err = runCommand(t, nil, "import", "-d", `{"Keys": [{"KeyName": "testkey2", "Data": "`+exportedKey2Data+`", "Password": "testpassword2"}]}`)
	require.NoError(t, err)

	// List keys.
	out, err = runCommand(t, nil, "list")
	require.NoError(t, err)
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 2)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)
	require.Equal(t, "testkey2", resp.Keys[1].KeyInfo.Name)
	require.Equal(t, ks.ECDSA_S256, resp.Keys[1].KeyInfo.KeyType)

	// Set metadata on testkey.
	metadata := base64.StdEncoding.EncodeToString([]byte("my-custom-metadata"))
	_, err = runCommand(t, nil, "set-metadata", "-d", `{"Updates": [{"KeyName": "testkey", "Metadata": "`+metadata+`"}]}`)
	require.NoError(t, err)

	// Verify metadata was set.
	out, err = runCommand(t, nil, "get", "-d", `{"KeyNames": ["testkey"]}`)
	require.NoError(t, err)
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "testkey", resp.Keys[0].KeyInfo.Name)
	// Metadata is []byte, Go's JSON unmarshaler automatically decodes base64 strings
	require.Equal(t, "my-custom-metadata", string(resp.Keys[0].KeyInfo.Metadata))

	// Delete the keys with confirmation.
	out, err = runCommand(t, bytes.NewBufferString("yes\n"), "delete", "-d", `{"KeyNames": ["testkey", "testkey2"]}`)
	require.NoError(t, err)
	t.Log("out", out.String())

	// List keys should be empty.
	out, err = runCommand(t, nil, "list")
	require.NoError(t, err)
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &resp)
	require.NoError(t, err)
	require.Empty(t, resp.Keys)
}

func runCommand(t *testing.T, in *bytes.Buffer, args ...string) (bytes.Buffer, error) {
	// Cobra commands are stateful which can cause subtle bugs if not reset.
	// For simplicity just create a fresh object.
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	cmd.SetArgs(args)
	cmd.SetIn(in)
	err := cmd.ExecuteContext(t.Context())
	if err != nil {
		return bytes.Buffer{}, err
	}
	return *buf, nil
}
