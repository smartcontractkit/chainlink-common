package tonkey

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/keystore/scrypt"
)

const keyTypeIdentifier = "TON"

// FromEncryptedJSON gets key from json and password
func FromEncryptedJSON(keyJSON []byte, password string) (Key, error) {
	builder := func(_ internal.EncryptedKeyExport, rawPrivKey internal.Raw) (Key, error) {
		return KeyFor(rawPrivKey), nil
	}
	return internal.FromEncryptedJSON(
		keyTypeIdentifier,
		keyJSON,
		password,
		adulteratedPassword,
		builder,
	)
}

// ToEncryptedJSON returns encrypted JSON representing key
func (key Key) ToEncryptedJSON(password string, scryptParams scrypt.ScryptParams) ([]byte, error) {
	exporter := func(id string, key Key, cryptoJSON keystore.CryptoJSON) internal.EncryptedKeyExport {
		return internal.EncryptedKeyExport{
			KeyType:   id,
			PublicKey: hex.EncodeToString(key.pubKey),
			Crypto:    cryptoJSON,
		}
	}
	return internal.ToEncryptedJSON(
		keyTypeIdentifier,
		key,
		password,
		scryptParams,
		adulteratedPassword,
		exporter,
	)
}

func adulteratedPassword(password string) string {
	return "tonkey" + password
}
