package keyring

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/smartcontractkit/chainlink-relay/pkg/keyring/algo"
	"github.com/smartcontractkit/chainlink-relay/pkg/keyring/backend/sql"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

func getCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

var algoOptionFn = func(options *keyring.Options) {
	options.SupportedAlgos = keyring.SigningAlgoList{
		algo.Secp256k1,
		algo.Ed25519,
	}

	options.SupportedAlgosLedger = keyring.SigningAlgoList{
		algo.Secp256k1,
		algo.Ed25519,
	}
}

func newKeyringInMem() keyring.Keyring {
	return keyring.NewInMemory(getCodec(), algoOptionFn)
}

func newKeyringSql() keyring.Keyring {
	backend := sql.NewKeyring(nil)
	return keyring.NewKeystore(backend, getCodec(), algoOptionFn)
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
	mnemonic := "health amateur need boy enough bless april march dove rabbit satoshi purse"
	bip39Passphrase := ""
	hdPath := "m/44'/60'/0'/0/0"

	testCases := []struct {
		name string
		uid  string
		algo keyring.SignatureAlgo
	}{
		{
			name: "can sign/verify secp256k1",
			uid:  "uuid-secp256k1-1",
			algo: algo.Secp256k1,
		},
		{
			name: "can sign/verify ed25519",
			uid:  "uuid-ed25519-1",
			algo: algo.Ed25519,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := keyrings[keyringEthTest].NewAccount(tc.uid, mnemonic, bip39Passphrase, hdPath, tc.algo)
			assert.NoError(t, err)

			// add keyrings to service
			s := NewService(nil, keyrings)
			assert.NotEmpty(t, s)

			// sign
			msg := []byte{0}
			sig, pk, err := s.Signers[keyringEthTest].Sign(tc.uid, msg)
			assert.NoError(t, err)

			// debug
			fmt.Println(pk.Address())
			fmt.Println(pk.String())

			// verify
			ok := pk.VerifySignature(msg, sig)
			assert.True(t, ok)
		})
	}
}

// TODO: test keyring.Keyring.ImportPrivKey
