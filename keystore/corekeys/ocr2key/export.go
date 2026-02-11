package ocr2key

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys/starkkey"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/types"
)

const keyTypeIdentifier = "OCR2"

// EncryptedOCRKeyExport represents encrypted OCR key export
type EncryptedOCRKeyExport struct {
	KeyType           string              `json:"keyType"`
	ChainType         types.ChainType     `json:"chainType"`
	ID                string              `json:"id"`
	OnchainPublicKey  string              `json:"onchainPublicKey"`
	OffChainPublicKey string              `json:"offchainPublicKey"`
	ConfigPublicKey   string              `json:"configPublicKey"`
	Crypto            keystore.CryptoJSON `json:"crypto"`
}

func (x EncryptedOCRKeyExport) GetCrypto() keystore.CryptoJSON {
	return x.Crypto
}

// FromEncryptedJSON returns key from encrypted json
func FromEncryptedJSON(keyJSON []byte, password string) (KeyBundle, error) {
	return internal.FromEncryptedJSON(
		keyTypeIdentifier,
		keyJSON,
		password,
		adulteratedPassword,
		func(export EncryptedOCRKeyExport, rawPrivKey internal.Raw) (KeyBundle, error) {
			var kb KeyBundle
			switch export.ChainType {
			case types.EVM:
				kb = newKeyBundle(new(evmKeyring))
			case types.Cosmos:
				kb = newKeyBundle(new(cosmosKeyring))
			case types.Solana:
				kb = newKeyBundle(new(solanaKeyring))
			case types.StarkNet:
				kb = newKeyBundle(new(starkkey.OCR2Key))
			case types.Aptos:
				kb = newKeyBundle(new(ed25519Keyring))
			case types.Tron:
				kb = newKeyBundle(new(evmKeyring))
			case types.TON:
				kb = newKeyBundle(new(tonKeyring))
			case types.Sui:
				kb = newKeyBundle(new(ed25519Keyring))
			default:
				return nil, types.NewErrInvalidChainType(export.ChainType)
			}
			if err := kb.Unmarshal(internal.Bytes(rawPrivKey)); err != nil {
				return nil, err
			}
			return kb, nil
		},
	)
}

// ToEncryptedJSON returns encrypted JSON representing key
func ToEncryptedJSON(key KeyBundle, password string, scryptParams scrypt.ScryptParams) (export []byte, err error) {
	return internal.ToEncryptedJSON(
		keyTypeIdentifier,
		key,
		password,
		scryptParams,
		adulteratedPassword,
		func(id string, key KeyBundle, cryptoJSON keystore.CryptoJSON) EncryptedOCRKeyExport {
			pubKeyConfig := key.ConfigEncryptionPublicKey()
			pubKey := key.OffchainPublicKey()
			return EncryptedOCRKeyExport{
				KeyType:           id,
				ChainType:         key.ChainType(),
				ID:                key.ID(),
				OnchainPublicKey:  key.OnChainPublicKey(),
				OffChainPublicKey: hex.EncodeToString(pubKey[:]),
				ConfigPublicKey:   hex.EncodeToString(pubKeyConfig[:]),
				Crypto:            cryptoJSON,
			}
		},
	)
}
