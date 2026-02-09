package ocr2key

import (
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys/starkkey"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/types"
)

type Sha256Hash [32]byte

type OCR3SignerVerifier interface {
	SignBlob(b []byte) (sig []byte, err error)
	VerifyBlob(publicKey ocrtypes.OnchainPublicKey, b []byte, sig []byte) bool
	Sign3(digest ocrtypes.ConfigDigest, seqNr uint64, r ocrtypes.Report) (signature []byte, err error)
	Verify3(publicKey ocrtypes.OnchainPublicKey, cd ocrtypes.ConfigDigest, seqNr uint64, r ocrtypes.Report, signature []byte) bool
}

type KeyBundle interface {
	// OnchainKeyring is used for signing reports (groups of observations, verified onchain)
	ocrtypes.OnchainKeyring
	// offchainKeyring is used for signing observations
	ocrtypes.OffchainKeyring

	OCR3SignerVerifier

	ID() string
	ChainType() types.ChainType
	Marshal() ([]byte, error)
	Unmarshal(b []byte) (err error)
	Raw() internal.Raw
	OnChainPublicKey() string
	// Decrypts ciphertext using the encryptionKey from an OCR2 offchainKeyring
	NaclBoxOpenAnonymous(ciphertext []byte) (plaintext []byte, err error)
}

// check generic keybundle for each chain conforms to KeyBundle interface
var _ KeyBundle = &keyBundle[*evmKeyring]{}
var _ KeyBundle = &keyBundle[*cosmosKeyring]{}
var _ KeyBundle = &keyBundle[*solanaKeyring]{}
var _ KeyBundle = &keyBundle[*starkkey.OCR2Key]{}
var _ KeyBundle = &keyBundle[*ed25519Keyring]{}
var _ KeyBundle = &keyBundle[*tonKeyring]{}
var _ KeyBundle = &keyBundle[*ed25519Keyring]{}

var curve = secp256k1.S256()

// New returns key bundle based on the chain type
func New(chainType types.ChainType) (KeyBundle, error) {
	switch chainType {
	case types.EVM:
		return newKeyBundleRand(types.EVM, newEVMKeyring)
	case types.Cosmos:
		return newKeyBundleRand(types.Cosmos, newCosmosKeyring)
	case types.Solana:
		return newKeyBundleRand(types.Solana, newSolanaKeyring)
	case types.StarkNet:
		return newKeyBundleRand(types.StarkNet, starkkey.NewOCR2Key)
	case types.Aptos:
		return newKeyBundleRand(types.Aptos, newEd25519Keyring)
	case types.Tron:
		return newKeyBundleRand(types.Tron, newEVMKeyring)
	case types.TON:
		return newKeyBundleRand(types.TON, newTONKeyring)
	case types.Sui:
		return newKeyBundleRand(types.Sui, newEd25519Keyring)
	}
	return nil, types.NewErrInvalidChainType(chainType)
}

// MustNewInsecure returns key bundle based on the chain type or panics
func MustNewInsecure(reader io.Reader, chainType types.ChainType) KeyBundle {
	switch chainType {
	case types.EVM:
		return mustNewKeyBundleInsecure(types.EVM, newEVMKeyring, reader)
	case types.Cosmos:
		return mustNewKeyBundleInsecure(types.Cosmos, newCosmosKeyring, reader)
	case types.Solana:
		return mustNewKeyBundleInsecure(types.Solana, newSolanaKeyring, reader)
	case types.StarkNet:
		return mustNewKeyBundleInsecure(types.StarkNet, starkkey.NewOCR2Key, reader)
	case types.Aptos:
		return mustNewKeyBundleInsecure(types.Aptos, newEd25519Keyring, reader)
	case types.Tron:
		return mustNewKeyBundleInsecure(types.Tron, newEVMKeyring, reader)
	case types.TON:
		return mustNewKeyBundleInsecure(types.TON, newTONKeyring, reader)
	case types.Sui:
		return mustNewKeyBundleInsecure(types.Sui, newEd25519Keyring, reader)
	}
	panic(types.NewErrInvalidChainType(chainType))
}

type keyBundleBase struct {
	offchainKeyring
	id        Sha256Hash
	chainType types.ChainType
}

func (kb keyBundleBase) ID() string {
	return hex.EncodeToString(kb.id[:])
}

// ChainType gets the chain type from the key bundle
func (kb keyBundleBase) ChainType() types.ChainType {
	return kb.chainType
}

func KeyFor(raw internal.Raw) (kb KeyBundle) {
	var temp struct{ ChainType types.ChainType }
	err := json.Unmarshal(internal.Bytes(raw), &temp)
	if err != nil {
		panic(err)
	}
	switch temp.ChainType {
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
		return nil
	}
	if err := kb.Unmarshal(internal.Bytes(raw)); err != nil {
		panic(err)
	}
	return
}

// type is added to the beginning of the passwords for OCR key bundles,
// so that the keys can't accidentally be mis-used in the wrong place
func adulteratedPassword(auth string) string {
	s := "ocr2key" + auth
	return s
}
