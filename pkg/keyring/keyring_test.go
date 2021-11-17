package keyring

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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

func TestNewServiceMemKeystore(t *testing.T) {
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
