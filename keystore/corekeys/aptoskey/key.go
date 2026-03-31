package aptoskey

import (
	"crypto"
	"crypto/ed25519"
	crypto_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/crypto/sha3"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

// Key represents Aptos key
type Key struct {
	raw            internal.Raw
	signFn         func(io.Reader, []byte, crypto.SignerOpts) ([]byte, error)
	pubKey         ed25519.PublicKey
	accountAddress string // stored at creation; stable across key rotation
}

// deriveAccountAddress computes SHA3-256(pubKey + 0x00) per AIP-40.
// Used only at key creation and migration.
func deriveAccountAddress(pubKey ed25519.PublicKey) string {
	authKey := sha3.Sum256(append([]byte(pubKey), 0x00))
	return fmt.Sprintf("%064x", authKey)
}

func KeyFor(raw internal.Raw) Key {
	privKey := ed25519.NewKeyFromSeed(internal.Bytes(raw))
	pubKey := privKey.Public().(ed25519.PublicKey)
	return Key{
		raw:            raw,
		signFn:         privKey.Sign,
		pubKey:         pubKey,
		accountAddress: deriveAccountAddress(pubKey),
	}
}

// New creates new Key
func New() (Key, error) {
	return newFrom(crypto_rand.Reader)
}

// MustNewInsecure return Key if no error
func MustNewInsecure(reader io.Reader) Key {
	key, err := newFrom(reader)
	if err != nil {
		panic(err)
	}
	return key
}

// newFrom creates new Key from a provided random reader
func newFrom(reader io.Reader) (Key, error) {
	pub, priv, err := ed25519.GenerateKey(reader)
	if err != nil {
		return Key{}, err
	}
	return Key{
		raw:            internal.NewRaw(priv.Seed()),
		signFn:         priv.Sign,
		pubKey:         pub,
		accountAddress: deriveAccountAddress(pub),
	}, nil
}

// ID gets Key ID
func (key Key) ID() string {
	return key.PublicKeyStr()
}

// https://github.com/aptos-foundation/AIPs/blob/main/aips/aip-40.md#long
// Account returns the stored Aptos account address. This is set at key creation
// and is stable across authentication key rotation.
func (key Key) Account() string {
	return key.accountAddress
}

// WithAccountAddress returns a copy of the key with the account address overridden.
// Use after on-chain key rotation where the account address differs from what
// would be derived from the current public key.
func (key Key) WithAccountAddress(addr string) Key {
	key.accountAddress = addr
	return key
}

// GetPublic get Key's public key
func (key Key) GetPublic() ed25519.PublicKey {
	return key.pubKey
}

// PublicKeyStr returns hex encoded public key
func (key Key) PublicKeyStr() string {
	return fmt.Sprintf("%064x", key.pubKey)
}

// Raw returns the seed from private key
func (key Key) Raw() internal.Raw { return key.raw }

// serializedKey is the JSON structure used by Marshal/Unmarshal for KeyRing
// serialization. Extensible — add fields here as needed.
type serializedKey struct {
	Seed           []byte `json:"seed"`
	AccountAddress string `json:"accountAddress"`
}

// Marshal serializes the key for KeyRing storage.
func (key Key) Marshal() []byte {
	b, _ := json.Marshal(serializedKey{
		Seed:           internal.Bytes(key.raw),
		AccountAddress: key.accountAddress,
	})
	return b
}

// Unmarshal deserializes a key from KeyRing storage. Handles both the current
// JSON format and legacy raw seed bytes (pre-accountAddress).
func Unmarshal(data []byte) (Key, error) {
	var sk serializedKey
	if err := json.Unmarshal(data, &sk); err != nil {
		// Legacy format: raw seed bytes
		return KeyFor(internal.NewRaw(data)), nil
	}
	key := KeyFor(internal.NewRaw(sk.Seed))
	if sk.AccountAddress != "" {
		key = key.WithAccountAddress(sk.AccountAddress)
	}
	return key, nil
}

// Sign is used to sign a message
func (key Key) Sign(msg []byte) ([]byte, error) {
	return key.signFn(crypto_rand.Reader, msg, crypto.Hash(0)) // no specific hash function used
}
