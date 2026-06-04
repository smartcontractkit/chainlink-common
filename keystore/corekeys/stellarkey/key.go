package stellarkey

import (
	"crypto"
	"crypto/ed25519"
	crypto_rand "crypto/rand"
	"io"

	"github.com/stellar/go-stellar-sdk/strkey"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

// Key represents a Stellar (Soroban) key.
//
// Stellar accounts use ed25519 (same primitive as Solana/Aptos), but the
// account identifier is a StrKey-encoded "G..." address: base32 of
// (versionByte || ed25519-pubkey || crc16). Using the canonical StrKey
// encoding is load-bearing: the Stellar relayer/TXM looks signing keys up by
// that exact "G..." address (loop.Keystore.Sign(account, data)), so ID() must
// match it byte-for-byte.
type Key struct {
	raw    internal.Raw
	signFn func(io.Reader, []byte, crypto.SignerOpts) ([]byte, error)
	pubKey ed25519.PublicKey
}

func KeyFor(raw internal.Raw) Key {
	privKey := ed25519.NewKeyFromSeed(internal.Bytes(raw))
	pubKey := privKey.Public().(ed25519.PublicKey)
	return Key{
		raw:    raw,
		signFn: privKey.Sign,
		pubKey: pubKey,
	}
}

// New creates a new Key
func New() (Key, error) {
	return newFrom(crypto_rand.Reader)
}

// MustNewInsecure returns a Key if no error
func MustNewInsecure(reader io.Reader) Key {
	key, err := newFrom(reader)
	if err != nil {
		panic(err)
	}
	return key
}

// newFrom creates a new Key from a provided random reader
func newFrom(reader io.Reader) (Key, error) {
	pub, priv, err := ed25519.GenerateKey(reader)
	if err != nil {
		return Key{}, err
	}
	return Key{
		raw:    internal.NewRaw(priv.Seed()),
		signFn: priv.Sign,
		pubKey: pub,
	}, nil
}

// ID gets the Key ID, which is the StrKey "G..." account address.
func (key Key) ID() string {
	return key.Account()
}

// Account returns the StrKey-encoded ("G...") Stellar account address.
func (key Key) Account() string {
	// strkey.Encode only returns an error for an invalid version byte, which is
	// a compile-time constant here, so this cannot fail in practice.
	addr, err := strkey.Encode(strkey.VersionByteAccountID, key.pubKey)
	if err != nil {
		panic(err)
	}
	return addr
}

// GetPublic gets the Key's ed25519 public key
func (key Key) GetPublic() ed25519.PublicKey {
	return key.pubKey
}

// PublicKeyStr returns the StrKey "G..." account address (canonical string form).
func (key Key) PublicKeyStr() string {
	return key.Account()
}

// Raw returns the seed of the private key
func (key Key) Raw() internal.Raw { return key.raw }

// Sign signs a message with the ed25519 private key. For Stellar the caller
// passes the transaction hash; the returned 64-byte signature is wrapped into
// an xdr.DecoratedSignature by the relayer/TXM.
func (key Key) Sign(msg []byte) ([]byte, error) {
	return key.signFn(crypto_rand.Reader, msg, crypto.Hash(0)) // no specific hash function used
}
