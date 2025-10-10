package keystore_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/stretchr/testify/require"
)

func TestSigner(t *testing.T) {
	ks := NewKeystoreTH(t)
	ks.CreateTestKeys(t)
	ctx := t.Context()

	var tt = []struct {
		name          string
		keyName       string
		data          []byte
		signature     []byte
		expectedError error
	}{
		{
			name:    "ECDSA_S256 sign/verify",
			keyName: ks.KeyName(keystore.ECDSA_S256, 0),
			data:    make([]byte, 32), // 32 byte digest
		},
		{
			name:          "ECDSA_S256 sign/verify no such key",
			keyName:       "no-such-key",
			data:          make([]byte, 32), // 32 byte digest
			expectedError: keystore.ErrKeyNotFound,
		},
		{
			name:          "ECDSA_S256 sign/verify wrong data length",
			keyName:       ks.KeyName(keystore.ECDSA_S256, 0),
			data:          make([]byte, 31),
			expectedError: keystore.ErrInvalidSignRequest,
		},
		{
			name:    "Ed25519 sign/verify",
			keyName: ks.KeyName(keystore.Ed25519, 0),
			data:    []byte("test_data"),
		},
		{
			name:          "Ed25519 sign/verify no such key",
			keyName:       "no-such-key",
			data:          make([]byte, 2),
			expectedError: keystore.ErrKeyNotFound,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			signature, err := ks.Keystore.Sign(ctx, keystore.SignRequest{KeyName: tc.keyName, Data: tc.data})
			if tc.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			valid, err := ks.Keystore.Verify(ctx, keystore.VerifyRequest{
				KeyType:   ks.KeysByName()[tc.keyName].KeyType,
				PublicKey: ks.KeysByName()[tc.keyName].PublicKey,
				Data:      tc.data,
				Signature: signature.Signature,
			})
			require.NoError(t, err)
			require.True(t, valid.Valid)
		})
	}
}
