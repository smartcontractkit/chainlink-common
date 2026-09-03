package gcpkms

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/internal"
	"github.com/smartcontractkit/chainlink-common/keystore/kms"
)

// cryptoKeyVersionsSegment separates a CryptoKey resource name from its version number in a
// CryptoKeyVersion resource name.
const cryptoKeyVersionsSegment = "/cryptoKeyVersions/"

// Key identifies one in-memory CryptoKeyVersion held by FakeGCPKMSClient. KeyID is a CryptoKey
// resource name (projects/<p>/locations/<l>/keyRings/<r>/cryptoKeys/<k>); several Keys may share a
// KeyID to emulate a rotated key with multiple versions.
type Key struct {
	KeyType    keystore.KeyType
	KeyID      string
	PrivateKey internal.Raw

	// VersionNumber is the CryptoKeyVersion number. Defaults to 1.
	VersionNumber uint64
	// State is the version state. Defaults to ENABLED.
	State kmspb.CryptoKeyVersion_CryptoKeyVersionState
	// Algorithm is the algorithm reported for this version. Defaults to the algorithm matching
	// KeyType; set it to emulate an unsupported one.
	Algorithm kmspb.CryptoKeyVersion_CryptoKeyVersionAlgorithm
}

// FakeGCPKMSClient is an in-memory implementation of Client for tests. It emulates the parts of
// Google Cloud KMS that the keystore uses, producing the same wire formats (PEM SPKI public keys,
// DER ECDSA signatures, raw Ed25519 signatures) and the same resource-naming rules: asymmetric
// operations only accept CryptoKeyVersion names, and CryptoKeys never report a primary version.
type FakeGCPKMSClient struct {
	keys      []Key
	createdAt time.Time
}

func NewFakeGCPKMSClient(keys []Key) (*FakeGCPKMSClient, error) {
	keys = append([]Key(nil), keys...)
	for i := range keys {
		if err := normalizeKey(&keys[i]); err != nil {
			return nil, err
		}
	}
	return &FakeGCPKMSClient{
		keys:      keys,
		createdAt: time.Now(),
	}, nil
}

// AddVersion appends a CryptoKeyVersion after construction, emulating a rotation that lands while
// the keystore is live. Not safe for concurrent use with the client's read methods.
func (m *FakeGCPKMSClient) AddVersion(key Key) error {
	if err := normalizeKey(&key); err != nil {
		return err
	}
	if _, err := m.findVersion(key.versionName()); err == nil {
		return fmt.Errorf("version %s already exists", key.versionName())
	}
	m.keys = append(m.keys, key)
	return nil
}

// normalizeKey validates a Key and fills in the fields a test left at their zero value.
func normalizeKey(key *Key) error {
	if key.KeyID == "" {
		return errors.New("key ID is required")
	}
	if key.VersionNumber == 0 {
		key.VersionNumber = 1
	}
	if key.State == kmspb.CryptoKeyVersion_CRYPTO_KEY_VERSION_STATE_UNSPECIFIED {
		key.State = kmspb.CryptoKeyVersion_ENABLED
	}
	if key.Algorithm == kmspb.CryptoKeyVersion_CRYPTO_KEY_VERSION_ALGORITHM_UNSPECIFIED {
		algorithm, err := keyTypeToAlgorithm(key.KeyType)
		if err != nil {
			return err
		}
		key.Algorithm = algorithm
	}
	return nil
}

func keyTypeToAlgorithm(keyType keystore.KeyType) (kmspb.CryptoKeyVersion_CryptoKeyVersionAlgorithm, error) {
	switch keyType {
	case keystore.ECDSA_S256:
		return kmspb.CryptoKeyVersion_EC_SIGN_SECP256K1_SHA256, nil
	case keystore.Ed25519:
		return kmspb.CryptoKeyVersion_EC_SIGN_ED25519, nil
	default:
		return 0, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// versionName returns the CryptoKeyVersion resource name of a key.
func (k Key) versionName() string {
	return k.KeyID + cryptoKeyVersionsSegment + strconv.FormatUint(k.VersionNumber, 10)
}

func (m *FakeGCPKMSClient) toCryptoKeyVersion(key *Key) *kmspb.CryptoKeyVersion {
	return &kmspb.CryptoKeyVersion{
		Name:       key.versionName(),
		Algorithm:  key.Algorithm,
		State:      key.State,
		CreateTime: timestamppb.New(m.createdAt),
	}
}

// findVersion looks up a key by its CryptoKeyVersion resource name. Cloud KMS rejects a bare
// CryptoKey name on the asymmetric endpoints, so the fake does too.
func (m *FakeGCPKMSClient) findVersion(versionName string) (*Key, error) {
	if !strings.Contains(versionName, cryptoKeyVersionsSegment) {
		return nil, fmt.Errorf("%q is not a CryptoKeyVersion resource name", versionName)
	}
	for i := range m.keys {
		if m.keys[i].versionName() == versionName {
			return &m.keys[i], nil
		}
	}
	return nil, errors.New("key not found")
}

func (m *FakeGCPKMSClient) GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	if req.Name == "" {
		return nil, errors.New("key version name is required")
	}
	key, err := m.findVersion(req.Name)
	if err != nil {
		return nil, err
	}
	return m.toCryptoKeyVersion(key), nil
}

func (m *FakeGCPKMSClient) GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, opts ...gax.CallOption) (*kmspb.PublicKey, error) {
	if req.Name == "" {
		return nil, errors.New("key version name is required")
	}
	key, err := m.findVersion(req.Name)
	if err != nil {
		return nil, err
	}

	var derPubKey []byte
	switch key.KeyType {
	case keystore.ECDSA_S256:
		ecdsaKey, err := crypto.ToECDSA(internal.Bytes(key.PrivateKey))
		if err != nil {
			return nil, err
		}
		derPubKey, err = kms.SEC1ToASN1PublicKey(crypto.FromECDSAPub(&ecdsaKey.PublicKey))
		if err != nil {
			return nil, err
		}
	case keystore.Ed25519:
		ed25519PrivKey, err := ed25519PrivateKey(key)
		if err != nil {
			return nil, err
		}
		pubKey := ed25519PrivKey.Public().(ed25519.PublicKey)
		derPubKey, err = x509.MarshalPKIXPublicKey(pubKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.KeyType)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: derPubKey})
	return &kmspb.PublicKey{
		Name:            req.Name,
		Algorithm:       key.Algorithm,
		Pem:             string(pemBytes),
		PemCrc32C:       wrapperspb.Int64(crc32c(pemBytes)),
		PublicKeyFormat: kmspb.PublicKey_PEM,
	}, nil
}

func (m *FakeGCPKMSClient) AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, opts ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error) {
	if req.Name == "" {
		return nil, errors.New("key version name is required")
	}
	key, err := m.findVersion(req.Name)
	if err != nil {
		return nil, err
	}

	switch key.KeyType {
	case keystore.ECDSA_S256:
		if req.Digest == nil {
			return nil, errors.New("digest is required for ECDSA signing")
		}
		ecdsaKey, err := crypto.ToECDSA(internal.Bytes(key.PrivateKey))
		if err != nil {
			return nil, err
		}
		sec1Sig, err := crypto.Sign(req.Digest.GetSha256(), ecdsaKey)
		if err != nil {
			return nil, err
		}
		derSig, err := kms.SEC1ToASN1Sig(sec1Sig)
		if err != nil {
			return nil, err
		}
		return &kmspb.AsymmetricSignResponse{
			Name:                 req.Name,
			Signature:            derSig,
			SignatureCrc32C:      wrapperspb.Int64(crc32c(derSig)),
			VerifiedDigestCrc32C: true,
		}, nil
	case keystore.Ed25519:
		ed25519PrivKey, err := ed25519PrivateKey(key)
		if err != nil {
			return nil, err
		}
		signature := ed25519.Sign(ed25519PrivKey, req.Data)
		return &kmspb.AsymmetricSignResponse{
			Name:               req.Name,
			Signature:          signature,
			SignatureCrc32C:    wrapperspb.Int64(crc32c(signature)),
			VerifiedDataCrc32C: true,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.KeyType)
	}
}

// ed25519PrivateKey returns the key's Ed25519 private key, erroring rather than letting the
// crypto/ed25519 helpers panic on a wrong-sized key.
func ed25519PrivateKey(key *Key) (ed25519.PrivateKey, error) {
	privKey := ed25519.PrivateKey(internal.Bytes(key.PrivateKey))
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key length: expected %d bytes, got %d", ed25519.PrivateKeySize, len(privKey))
	}
	return privKey, nil
}

func (m *FakeGCPKMSClient) Close() error {
	return nil
}
