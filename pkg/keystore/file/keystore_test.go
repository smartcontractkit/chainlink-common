package file

import (
	"context"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	"github.com/stretchr/testify/require"
)

func TestFileKeystore(t *testing.T) {
	tmpFile := "/tmp/test_keystore"
	// defer os.Remove(tmpFile)

	ks, err := NewFileKeystore("test_password", tmpFile)
	require.NoError(t, err)

	ctx := context.Background()
	keyInfo, err := ks.CreateKey(ctx, "test_key", keystore.Ed25519)
	require.NoError(t, err)
	require.Equal(t, "test_key", keyInfo.Name)
	require.Equal(t, keystore.Ed25519, keyInfo.KeyType)

	keys, err := ks.ListKeys(ctx)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Equal(t, "test_key", keys[0].Name)
	require.Equal(t, keystore.Ed25519, keys[0].KeyType)

	keyInfo, err = ks.GetKey(ctx, "test_key")
	require.NoError(t, err)
	require.Equal(t, "test_key", keyInfo.Name)
	require.Equal(t, keystore.Ed25519, keyInfo.KeyType)
	require.NotEmpty(t, keyInfo.PublicKey)

	// Restrict access to just signing/verifying
	var signer keystore.Signer = ks
	testData := []byte("hello world")
	signature, err := signer.Sign(ctx, "test_key", testData)
	require.NoError(t, err)
	require.NotEmpty(t, signature)
	valid, err := signer.Verify(ctx, "test_key", testData, signature)
	require.NoError(t, err)
	require.True(t, valid)

	// err = ks.DeleteKey(ctx, "test_key")
	// require.NoError(t, err)

	// keys, err = ks.ListKeys(ctx)
	// require.NoError(t, err)
	// require.Empty(t, keys)
}
