package ocr2key

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
	"github.com/smartcontractkit/chainlink-common/keystore/corekeys/starkkey"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

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
	ChainType() corekeys.ChainType
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
var _ KeyBundle = &keyBundle[*keccakEd25519Keyring]{}
var _ KeyBundle = &keyBundle[*tonKeyring]{}
var _ KeyBundle = &keyBundle[*ed25519Keyring]{}

var curve = secp256k1.S256()

type keyBundleFactory struct {
	new func(io.Reader, io.Reader, io.Reader) (KeyBundle, error)
	// insecure constructs a key bundle using caller-provided key material.
	insecure func(io.Reader) KeyBundle
	// empty constructs an uninitialized key bundle for unmarshalling persisted key material.
	empty func() KeyBundle
}

func newKeyBundleFactory[K keyring](chain corekeys.ChainType, newKeyring func(io.Reader) (K, error), empty func() K) keyBundleFactory {
	return keyBundleFactory{
		new: func(onchainSigningKeyMaterial, onchainEncryptionKeyMaterial, offchainKeyMaterial io.Reader) (KeyBundle, error) {
			return newKeyBundleFrom(chain, newKeyring, onchainSigningKeyMaterial, onchainEncryptionKeyMaterial, offchainKeyMaterial)
		},
		insecure: func(reader io.Reader) KeyBundle {
			return mustNewKeyBundleInsecure(chain, newKeyring, reader)
		},
		empty: func() KeyBundle {
			return newKeyBundle(empty())
		},
	}
}

var keyBundleFactories = map[corekeys.ChainType]keyBundleFactory{
	corekeys.EVM:      newKeyBundleFactory(corekeys.EVM, newEVMKeyring, func() *evmKeyring { return new(evmKeyring) }),
	corekeys.Cosmos:   newKeyBundleFactory(corekeys.Cosmos, newCosmosKeyring, func() *cosmosKeyring { return new(cosmosKeyring) }),
	corekeys.Solana:   newKeyBundleFactory(corekeys.Solana, newSolanaKeyring, func() *solanaKeyring { return new(solanaKeyring) }),
	corekeys.StarkNet: newKeyBundleFactory(corekeys.StarkNet, starkkey.NewOCR2Key, func() *starkkey.OCR2Key { return new(starkkey.OCR2Key) }),
	corekeys.Aptos:    newKeyBundleFactory(corekeys.Aptos, newEd25519Keyring, func() *ed25519Keyring { return new(ed25519Keyring) }),
	corekeys.Tron:     newKeyBundleFactory(corekeys.Tron, newEVMKeyring, func() *evmKeyring { return new(evmKeyring) }),
	corekeys.TON:      newKeyBundleFactory(corekeys.TON, newTONKeyring, func() *tonKeyring { return new(tonKeyring) }),
	corekeys.Sui:      newKeyBundleFactory(corekeys.Sui, newEd25519Keyring, func() *ed25519Keyring { return new(ed25519Keyring) }),
	corekeys.Stellar: newKeyBundleFactory(corekeys.Stellar, newKeccakEd25519Keyring, func() *keccakEd25519Keyring {
		return &keccakEd25519Keyring{ed25519Keyring: new(ed25519Keyring)}
	}),
}

// New returns key bundle based on the chain type
func New(chainType corekeys.ChainType) (KeyBundle, error) {
	if factory, ok := keyBundleFactories[chainType]; ok {
		return factory.new(cryptorand.Reader, cryptorand.Reader, cryptorand.Reader)
	}
	return nil, corekeys.NewErrInvalidChainType(chainType)
}

// MustNewInsecure returns key bundle based on the chain type or panics
func MustNewInsecure(reader io.Reader, chainType corekeys.ChainType) KeyBundle {
	if factory, ok := keyBundleFactories[chainType]; ok {
		return factory.insecure(reader)
	}
	panic(corekeys.NewErrInvalidChainType(chainType))
}

type keyBundleBase struct {
	offchainKeyring
	id        corekeys.Sha256Hash
	chainType corekeys.ChainType
}

func (kb keyBundleBase) ID() string {
	return hex.EncodeToString(kb.id[:])
}

// ChainType gets the chain type from the key bundle
func (kb keyBundleBase) ChainType() corekeys.ChainType {
	return kb.chainType
}

func KeyFor(raw internal.Raw) (kb KeyBundle) {
	var temp struct{ ChainType corekeys.ChainType }
	err := json.Unmarshal(internal.Bytes(raw), &temp)
	if err != nil {
		panic(err)
	}
	factory, ok := keyBundleFactories[temp.ChainType]
	if !ok {
		return nil
	}
	kb = factory.empty()
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
