package store

import (
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeystore(t *testing.T) {
	gormDB, err := test.NewGormDB()
	require.NoError(t, err)

	// create new tables
	require.NoError(t, test.CreateTable(gormDB, &models.EncryptedKeyRings{}))
	require.NoError(t, test.CreateTable(gormDB, &models.EthKeyStates{}))

	keyring, pubKeys, err := KeystoreInit(gormDB, "test-password")
	require.NoError(t, err)

	// check if mapped public keys exist
	for _, v := range []string{"OCRKeyID", "OCRConfigPublicKey", "OCROffchainPublicKey", "P2PID", "P2PPublicKey"} {
		val, ok := pubKeys[v]
		assert.True(t, ok)          // value exists
		assert.NotEqual(t, "", val) // not empty
	}

	keys := []struct {
		name string
		f    func() (int, error)
	}{
		{"ocr-key", func() (int, error) {
			ocrKeys, err := keyring.OCR().GetAll()
			return len(ocrKeys), err
		}},
		{"p2p-key", func() (int, error) {
			p2pKeys, err := keyring.P2P().GetAll()
			return len(p2pKeys), err
		}},
	}

	// test presence of keys
	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			res, err := k.f()
			assert.NoError(t, err)
			assert.Equal(t, 1, res)
		})
	}

	t.Run("ocr2-key-wrapper", func(t *testing.T) {
		ocrKeys, _ := keyring.OCR().GetAll()
		ocrKey := ocrKeys[0]

		ocr2 := NewOCR2KeyWrapper(ocrKey)
		assert.NotEmpty(t, ocr2.OffchainPublicKey())

		signed, err := ocr2.OffchainSign([]byte{})
		assert.NotEmpty(t, signed)
		assert.NoError(t, err)

		dh, err := ocr2.ConfigDiffieHellman([32]byte{10})
		assert.NotEmpty(t, dh)
		assert.NoError(t, err)

		assert.NotEmpty(t, ocr2.ConfigEncryptionPublicKey())
	})
}
