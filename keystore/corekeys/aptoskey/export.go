package aptoskey

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	commonkeystore "github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

const keyTypeIdentifier = "Aptos"

// aptosEncryptedKeyExport extends the base export with an AccountAddress field.
type aptosEncryptedKeyExport struct {
	KeyType        string              `json:"keyType"`
	PublicKey      string              `json:"publicKey"`
	AccountAddress string              `json:"accountAddress,omitempty"`
	Crypto         keystore.CryptoJSON `json:"crypto"`
}

func (x aptosEncryptedKeyExport) GetCrypto() keystore.CryptoJSON {
	return x.Crypto
}

// FromEncryptedJSON gets key from json and password.
// If the JSON was created before accountAddress was added (migration case),
// the address is derived from the public key and stored.
func FromEncryptedJSON(keyJSON []byte, password string) (Key, error) {
	return internal.FromEncryptedJSON(
		keyTypeIdentifier,
		keyJSON,
		password,
		adulteratedPassword,
		func(export aptosEncryptedKeyExport, rawPrivKey internal.Raw) (Key, error) {
			key := KeyFor(rawPrivKey)
			if export.AccountAddress != "" {
				// Stored address present — use it (supports key rotation)
				key = key.WithAccountAddress(export.AccountAddress)
			}
			// else: KeyFor already derived address from current pubkey (migration path)
			return key, nil
		},
	)
}

// ToEncryptedJSON returns encrypted JSON representing key.
// The accountAddress is stored in plaintext in the outer envelope so it can be
// read without decrypting the key.
func (key Key) ToEncryptedJSON(password string, scryptParams commonkeystore.ScryptParams) (export []byte, err error) {
	return internal.ToEncryptedJSON(
		keyTypeIdentifier,
		key,
		password,
		scryptParams.N,
		scryptParams.P,
		adulteratedPassword,
		func(id string, key Key, cryptoJSON keystore.CryptoJSON) aptosEncryptedKeyExport {
			return aptosEncryptedKeyExport{
				KeyType:        id,
				PublicKey:      hex.EncodeToString(key.pubKey),
				AccountAddress: key.accountAddress,
				Crypto:         cryptoJSON,
			}
		},
	)
}

func adulteratedPassword(password string) string {
	return "aptoskey" + password
}
