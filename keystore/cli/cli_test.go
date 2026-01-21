package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	ks "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/cli"
)

func setupKeystore(t *testing.T) func(t *testing.T) {
	tempDir := t.TempDir()
	keystoreFile := filepath.Join(tempDir, "keystore.json")
	f, err := os.Create(keystoreFile)
	require.NoError(t, err)
	t.Setenv("KEYSTORE_FILE_PATH", keystoreFile)
	t.Setenv("KEYSTORE_PASSWORD", "testpassword")
	// Set to empty string to test regular keystore mode.
	t.Setenv("KEYSTORE_KMS_PROFILE", "")
	return func(t *testing.T) {
		f.Close()
		os.RemoveAll(tempDir)
	}
}

func TestAdminCLI(t *testing.T) {
	teardown := setupKeystore(t)
	defer teardown(t)

	// No error just listing help.
	_, err := runCommand(t, nil, "")
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
	originalPublicKey := resp.Keys[0].KeyInfo.PublicKey

	// Rename testkey to renamedkey.
	_, err = runCommand(t, nil, "rename", "-d", `{"OldName": "testkey", "NewName": "renamedkey"}`)
	require.NoError(t, err)

	// Verify the old name doesn't exist.
	out, err = runCommand(t, nil, "get", "-d", `{"KeyNames": ["testkey"]}`)
	require.Error(t, err)

	// Verify the new name exists with the same key material.
	out, err = runCommand(t, nil, "get", "-d", `{"KeyNames": ["renamedkey"]}`)
	require.NoError(t, err)
	resp = ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Keys, 1)
	require.Equal(t, "renamedkey", resp.Keys[0].KeyInfo.Name)
	require.Equal(t, ks.X25519, resp.Keys[0].KeyInfo.KeyType)
	require.Equal(t, originalPublicKey, resp.Keys[0].KeyInfo.PublicKey)
	// Metadata should be preserved
	require.Equal(t, "my-custom-metadata", string(resp.Keys[0].KeyInfo.Metadata))

	// Rename it back to testkey for cleanup.
	_, err = runCommand(t, nil, "rename", "-d", `{"OldName": "renamedkey", "NewName": "testkey"}`)
	require.NoError(t, err)

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

func TestSignerCLI(t *testing.T) {
	teardown := setupKeystore(t)
	defer teardown(t)

	// Create an ECDSA key for signing.
	_, err := runCommand(t, nil, "create", "-d", `{"Keys": [{"KeyName": "ecdsakey", "KeyType": "ECDSA_S256"}]}`)
	require.NoError(t, err)

	// Get the key to retrieve the public key.
	out, err := runCommand(t, nil, "get", "-d", `{"KeyNames": ["ecdsakey"]}`)
	require.NoError(t, err)
	getResp := ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &getResp)
	require.NoError(t, err)
	require.Len(t, getResp.Keys, 1)
	publicKey := getResp.Keys[0].KeyInfo.PublicKey

	// ECDSA_S256 requires a 32-byte hash to sign.
	dataToSign := sha256.Sum256([]byte("hello world"))
	dataB64 := base64.StdEncoding.EncodeToString(dataToSign[:])
	out, err = runCommand(t, nil, "sign", "-d", `{"KeyName": "ecdsakey", "Data": "`+dataB64+`"}`)
	require.NoError(t, err)
	signResp := ks.SignResponse{}
	err = json.Unmarshal(out.Bytes(), &signResp)
	require.NoError(t, err)
	require.NotEmpty(t, signResp.Signature)

	// Verify the signature.
	sigB64 := base64.StdEncoding.EncodeToString(signResp.Signature)
	pubKeyB64 := base64.StdEncoding.EncodeToString(publicKey)
	out, err = runCommand(t, nil, "verify", "-d", `{"KeyType": "ECDSA_S256", "PublicKey": "`+pubKeyB64+`", "Data": "`+dataB64+`", "Signature": "`+sigB64+`"}`)
	require.NoError(t, err)
	verifyResp := ks.VerifyResponse{}
	err = json.Unmarshal(out.Bytes(), &verifyResp)
	require.NoError(t, err)
	require.True(t, verifyResp.Valid)
}

func TestEncryptDecryptCLI(t *testing.T) {
	teardown := setupKeystore(t)
	defer teardown(t)

	// Create an X25519 key for encryption.
	_, err := runCommand(t, nil, "create", "-d", `{"Keys": [{"KeyName": "x25519key", "KeyType": "X25519"}]}`)
	require.NoError(t, err)

	// Get the key to retrieve the public key.
	out, err := runCommand(t, nil, "get", "-d", `{"KeyNames": ["x25519key"]}`)
	require.NoError(t, err)
	getResp := ks.GetKeysResponse{}
	err = json.Unmarshal(out.Bytes(), &getResp)
	require.NoError(t, err)
	require.Len(t, getResp.Keys, 1)
	publicKey := getResp.Keys[0].KeyInfo.PublicKey

	// Encrypt some data to the key's public key.
	plaintext := []byte("secret message")
	pubKeyB64 := base64.StdEncoding.EncodeToString(publicKey)
	plaintextB64 := base64.StdEncoding.EncodeToString(plaintext)
	out, err = runCommand(t, nil, "encrypt", "-d", `{"RemoteKeyType": "X25519", "RemotePubKey": "`+pubKeyB64+`", "Data": "`+plaintextB64+`"}`)
	require.NoError(t, err)
	encryptResp := ks.EncryptResponse{}
	err = json.Unmarshal(out.Bytes(), &encryptResp)
	require.NoError(t, err)
	require.NotEmpty(t, encryptResp.EncryptedData)

	// Decrypt the data.
	encryptedB64 := base64.StdEncoding.EncodeToString(encryptResp.EncryptedData)
	out, err = runCommand(t, nil, "decrypt", "-d", `{"KeyName": "x25519key", "EncryptedData": "`+encryptedB64+`"}`)
	require.NoError(t, err)
	decryptResp := ks.DecryptResponse{}
	err = json.Unmarshal(out.Bytes(), &decryptResp)
	require.NoError(t, err)
	require.Equal(t, plaintext, decryptResp.Data)
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
