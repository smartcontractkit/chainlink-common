package cosmoskey

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"
)

const keyTypeIdentifier = "Cosmos"

// FromEncryptedJSON gets key from json and password
func FromEncryptedJSON(keyJSON []byte, password string) (Key, error) {
	return internal.FromEncryptedJSON(
		keyTypeIdentifier,
		keyJSON,
		password,
		adulteratedPassword,
		func(_ internal.EncryptedKeyExport, rawPrivKey internal.Raw) (Key, error) {
			return KeyFor(rawPrivKey), nil
		},
	)
}

// ToEncryptedJSON returns encrypted JSON representing key
func (key Key) ToEncryptedJSON(password string, scryptParams scrypt.ScryptParams) (export []byte, err error) {
	return internal.ToEncryptedJSON(
		keyTypeIdentifier,
		key,
		password,
		scryptParams,
		adulteratedPassword,
		func(id string, key Key, cryptoJSON keystore.CryptoJSON) internal.EncryptedKeyExport {
			return internal.EncryptedKeyExport{
				KeyType:   id,
				PublicKey: hex.EncodeToString(key.PublicKey().Bytes()),
				Crypto:    cryptoJSON,
			}
		},
	)
}

func adulteratedPassword(password string) string {
	return "cosmoskey" + password
}
