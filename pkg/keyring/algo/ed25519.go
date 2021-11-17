package algo

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
	bip39 "github.com/cosmos/go-bip39"

	ed25519 "crypto/ed25519"

	cosmosEd25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

// inspired by https://github.com/cosmos/cosmos-sdk/blob/master/crypto/hd/algo.go, as Ed25519 is not supported natively
var (
	Ed25519 = ed25519Algo{}
)

type ed25519Algo struct{}

func (s ed25519Algo) Name() hd.PubKeyType {
	return hd.Ed25519Type
}

// Derive derives and returns the ed25519 private key for the given seed and HD path.
// TODO: There might be issues with HD derivation of ed25519, currently not working properly. Potentially heck relevant issues on Cosmos SDK
func (s ed25519Algo) Derive() hd.DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := hd.ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return masterPriv[:], nil
		}
		derivedKey, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath)
		publicKey := ed25519.PublicKey(derivedKey)

		// cosmos-sdk expecting ed25519 PrivKey to include concatenated pubkey bytes
		return append(derivedKey[:], publicKey[:]...), err
	}
}

// Generate generates a ed25519 private key from the given bytes.
func (s ed25519Algo) Generate() hd.GenerateFn {
	return func(bz []byte) types.PrivKey {
		var bzArr = make([]byte, cosmosEd25519.PrivKeySize)
		copy(bzArr, bz)

		return &cosmosEd25519.PrivKey{Key: bzArr}
	}
}
