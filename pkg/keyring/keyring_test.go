package keyring

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/smartcontractkit/chainlink-relay/pkg/keyring/backend/sql"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

func getCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func newKeyringInMem() keyring.Keyring {
	return keyring.NewInMemory(getCodec(), func(options *keyring.Options) {
		// TODO: edit keyring options with custom supported signing algorithms
	})
}

func newKeyringSql() keyring.Keyring {
	backend := sql.NewKeyring(nil)
	return keyring.NewKeystore(backend, getCodec(), func(options *keyring.Options) {
		// TODO: edit keyring options with custom supported signing algorithms
	})
}

func TestNewService(t *testing.T) {
	keyrings := Keyrings{
		"ethereum.tx-ocr_report.test/mem": newKeyringInMem(),
		"ethereum.tx-ocr_report.test/sql": newKeyringSql(),
		"solana.tx.test/mem":              newKeyringInMem(),
		"solana.ocr-report.test/mem":      newKeyringInMem(),
		"solana.tx.test/sql":              newKeyringSql(),
		"solana.ocr-report.test/sql":      newKeyringSql(),
	}

	s := NewService(nil, keyrings)
	assert.NotEmpty(t, s)
	assert.NotEmpty(t, s.Signers["ethereum.tx-ocr_report.test/mem"])
	assert.NotEmpty(t, s.Signers["ethereum.tx-ocr_report.test/sql"])
	assert.NotEmpty(t, s.Signers["solana.tx.test/mem"])
	assert.NotEmpty(t, s.Signers["solana.ocr-report.test/mem"])
}

func TestNewServicePortsSignVerify(t *testing.T) {
	keyringEthTest := "ethereum.tx.test/mem"
	keyrings := Keyrings{
		"ethereum.tx.test/mem": newKeyringInMem(),
	}

	// generate a key to test
	uid := "uuid-1"
	mnemonic := "health amateur need boy enough bless april march dove rabbit satoshi purse"
	bip39Passphrase := ""
	hdPath := "m/44'/60'/0'/0"
	algo := hd.Secp256k1

	_, err := keyrings[keyringEthTest].NewAccount(uid, mnemonic, bip39Passphrase, hdPath, algo)
	assert.NoError(t, err)

	// add keyrings to service
	s := NewService(nil, keyrings)
	assert.NotEmpty(t, s)

	// sign
	msg := []byte{0}
	sig, pk, err := s.Signers[keyringEthTest].Sign(uid, msg)
	assert.NoError(t, err)

	// verify
	ok := pk.VerifySignature(msg, sig)
	assert.True(t, ok)
}

// TODO: test keyring.Keyring.ImportPrivKey
