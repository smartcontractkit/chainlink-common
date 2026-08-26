package gcpkms

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/kms"
)

// cryptoKeyVersionsSegment separates a CryptoKey resource name from its version number in a
// CryptoKeyVersion resource name.
const cryptoKeyVersionsSegment = "/cryptoKeyVersions/"

// errNoEnabledVersion and errUnsupportedAlgorithm mark the two ways a CryptoKey can turn out to be
// unusable by this keystore. Both are matchable so that a key ring listing can skip such keys
// instead of failing outright, while an explicitly requested key still surfaces the error.
var (
	errNoEnabledVersion     = errors.New("has no enabled version")
	errUnsupportedAlgorithm = errors.New("unsupported Cloud KMS key algorithm")
)

// castagnoliTable is the CRC32C (Castagnoli) table used by Google Cloud KMS for integrity checks.
var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

// crc32c returns the CRC32C checksum of data, matching the value Google Cloud KMS uses in its
// *_crc32c integrity fields.
func crc32c(data []byte) int64 {
	return int64(crc32.Checksum(data, castagnoliTable))
}

// checkCrc32c verifies that a received CRC32C checksum matches the computed value of the data.
//
// Cloud KMS documents this as the required client-side step for detecting corruption in transit:
// "you should verify the integrity of the response" by recomputing the CRC32C over the returned
// bytes and comparing it against the response's *_crc32c field.
// https://cloud.google.com/kms/docs/data-integrity-guidelines
//
// A missing checksum is treated as a failure: every response we check carries one, so its absence
// means the response was truncated or tampered with in transit.
func checkCrc32c(data []byte, received *wrapperspb.Int64Value) error {
	if received == nil {
		return errors.New("CRC32C integrity check failed: response is missing its checksum")
	}
	if want := crc32c(data); want != received.Value {
		return fmt.Errorf("CRC32C integrity check failed: computed %d, received %d", want, received.Value)
	}
	return nil
}

type keystoreSignerReader struct {
	client      Client
	keyRingName string

	mu sync.RWMutex
	// versions pins each key name to the CryptoKeyVersion it first resolved to, so that the public
	// key reported by GetKeys and the version used by Sign can never diverge. See resolveKeyVersion.
	versions map[string]resolvedKey
	// publicKeys caches public keys by CryptoKeyVersion resource name. A version's public key is
	// immutable, so this is safe to cache indefinitely and saves a round trip per signature.
	publicKeys map[string][]byte
}

type KeystoreOptions struct {
	// KeyRingName is required when GetKeys is called without an explicit key allowlist.
	// It must be a resource name in the format projects/<p>/locations/<l>/keyRings/<r>.
	KeyRingName string
}

func NewKeystore(client Client, opts KeystoreOptions) (interface {
	keystore.Reader
	keystore.Signer
}, error) {
	if client == nil {
		return nil, errors.New("GCP KMS client is required")
	}
	return &keystoreSignerReader{
		client:      client,
		keyRingName: opts.KeyRingName,
		versions:    make(map[string]resolvedKey),
		publicKeys:  make(map[string][]byte),
	}, nil
}

// cryptoKeyVersionAlgorithmToKeyType converts a Cloud KMS CryptoKeyVersionAlgorithm to a keystore
// KeyType. Google Cloud KMS supports:
//   - EC_SIGN_SECP256K1_SHA256 (secp256k1) -> ECDSA_S256
//   - EC_SIGN_ED25519 (Ed25519) -> Ed25519
func cryptoKeyVersionAlgorithmToKeyType(algo kmspb.CryptoKeyVersion_CryptoKeyVersionAlgorithm) (keystore.KeyType, error) {
	switch algo {
	case kmspb.CryptoKeyVersion_EC_SIGN_SECP256K1_SHA256:
		return keystore.ECDSA_S256, nil
	case kmspb.CryptoKeyVersion_EC_SIGN_ED25519:
		return keystore.Ed25519, nil
	default:
		return "", fmt.Errorf("%w: %s (supported: EC_SIGN_SECP256K1_SHA256, EC_SIGN_ED25519)", errUnsupportedAlgorithm, algo)
	}
}

type resolvedKey struct {
	version *kmspb.CryptoKeyVersion
	keyType keystore.KeyType
}

// resolveKeyVersion resolves a key name to a specific, enabled CryptoKeyVersion and its keystore
// KeyType.
//
// Cloud KMS only populates CryptoKey.Primary for ENCRYPT_DECRYPT keys — asymmetric signing keys
// never have a primary version — so a concrete version has to be selected here. keyName may be
// either:
//   - a CryptoKeyVersion resource name (.../cryptoKeys/<k>/cryptoKeyVersions/<n>), which names one
//     version outright, or
//   - a CryptoKey resource name, in which case the highest-numbered enabled version is selected.
//
// The result is pinned for the lifetime of the keystore: the first resolution of a given key name
// wins, and every later GetKeys and Sign for that name reuses it. Re-resolving per call would let a
// rotation land between the two, so a caller could hold the public key of version N while its
// signatures came from version N+1 — a silent verification failure with nothing in SignResponse to
// identify which version signed. Pinning makes the public key reported by GetKeys authoritative for
// every signature this keystore will produce.
//
// The cost is that a rotation is picked up on restart rather than immediately. That is the safer
// default for signing keys, whose public keys are typically registered with peers or on-chain and
// cannot change under a running system. Deployments that want rotation on an explicit schedule
// should configure CryptoKeyVersion names directly.
func (k *keystoreSignerReader) resolveKeyVersion(ctx context.Context, keyName string) (resolvedKey, error) {
	k.mu.RLock()
	pinned, ok := k.versions[keyName]
	k.mu.RUnlock()
	if ok {
		return pinned, nil
	}

	var version *kmspb.CryptoKeyVersion
	if strings.Contains(keyName, cryptoKeyVersionsSegment) {
		got, err := k.client.GetCryptoKeyVersion(ctx, &kmspb.GetCryptoKeyVersionRequest{Name: keyName})
		if err != nil {
			return resolvedKey{}, fmt.Errorf("failed to get crypto key version %s: %w", keyName, err)
		}
		if got == nil {
			return resolvedKey{}, fmt.Errorf("empty crypto key version response from Cloud KMS for %s", keyName)
		}
		if got.Name != keyName {
			return resolvedKey{}, fmt.Errorf("crypto key version response has name %q, expected %q", got.Name, keyName)
		}
		version = got
	} else {
		latest, err := k.latestEnabledVersion(ctx, keyName)
		if err != nil {
			return resolvedKey{}, err
		}
		version = latest
	}

	if version.State != kmspb.CryptoKeyVersion_ENABLED {
		return resolvedKey{}, fmt.Errorf("crypto key version %s is not enabled (state=%s)", version.Name, version.State)
	}
	if version.CreateTime == nil {
		return resolvedKey{}, fmt.Errorf("crypto key version %s has no creation time", version.Name)
	}
	keyType, err := cryptoKeyVersionAlgorithmToKeyType(version.Algorithm)
	if err != nil {
		return resolvedKey{}, fmt.Errorf("crypto key %s: %w", keyName, err)
	}

	resolved := resolvedKey{version: version, keyType: keyType}
	k.mu.Lock()
	defer k.mu.Unlock()
	// First writer wins, so concurrent resolutions racing a rotation still converge on one version.
	if existing, ok := k.versions[keyName]; ok {
		return existing, nil
	}
	k.versions[keyName] = resolved
	return resolved, nil
}

// latestEnabledVersion returns the highest-numbered enabled version of a CryptoKey.
func (k *keystoreSignerReader) latestEnabledVersion(ctx context.Context, cryptoKeyName string) (*kmspb.CryptoKeyVersion, error) {
	versions, err := k.client.ListCryptoKeyVersions(ctx, cryptoKeyName)
	if err != nil {
		return nil, err
	}
	var latest *kmspb.CryptoKeyVersion
	var latestNumber uint64
	for _, version := range versions {
		if version == nil || version.State != kmspb.CryptoKeyVersion_ENABLED {
			continue
		}
		number, err := cryptoKeyVersionNumber(version.Name)
		if err != nil {
			return nil, err
		}
		if latest == nil || number > latestNumber {
			latest, latestNumber = version, number
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("crypto key %s %w", cryptoKeyName, errNoEnabledVersion)
	}
	return latest, nil
}

// cryptoKeyVersionNumber extracts the trailing version number from a CryptoKeyVersion resource
// name. Cloud KMS assigns these sequentially, so a higher number means a newer version.
func cryptoKeyVersionNumber(versionName string) (uint64, error) {
	index := strings.LastIndex(versionName, cryptoKeyVersionsSegment)
	if index < 0 {
		return 0, fmt.Errorf("unexpected crypto key version resource name %q", versionName)
	}
	number, err := strconv.ParseUint(versionName[index+len(cryptoKeyVersionsSegment):], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unexpected crypto key version resource name %q: %w", versionName, err)
	}
	return number, nil
}

// getPublicKeyBytes fetches the public key for a crypto key version and converts it to the
// keystore's native format for the given key type. Results are cached per version name.
func (k *keystoreSignerReader) getPublicKeyBytes(ctx context.Context, versionName string, keyType keystore.KeyType) ([]byte, error) {
	k.mu.RLock()
	cached, ok := k.publicKeys[versionName]
	k.mu.RUnlock()
	if ok {
		return cached, nil
	}

	pk, err := k.client.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: versionName})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key for %s: %w", versionName, err)
	}
	if pk == nil {
		return nil, fmt.Errorf("empty public key response from Cloud KMS for %s", versionName)
	}
	if pk.Name != versionName {
		return nil, fmt.Errorf("public key response has name %q, expected %q", pk.Name, versionName)
	}
	if err = checkCrc32c([]byte(pk.Pem), pk.PemCrc32C); err != nil {
		return nil, fmt.Errorf("public key for %s: %w", versionName, err)
	}

	block, _ := pem.Decode([]byte(pk.Pem))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM public key for %s", versionName)
	}

	var publicKeyBytes []byte
	switch keyType {
	case keystore.ECDSA_S256:
		// GCP returns the public key in ASN.1 DER-encoded SubjectPublicKeyInfo (SPKI) format,
		// identical to AWS. Reuse the shared conversion.
		publicKeyBytes, err = kms.ASN1ToSEC1PublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	case keystore.Ed25519:
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Ed25519 public key for %s: %w", versionName, err)
		}
		ed25519PubKey, ok := pubKey.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to convert Ed25519 public key for %s to ed25519.PublicKey", versionName)
		}
		publicKeyBytes = ed25519PubKey
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}

	k.mu.Lock()
	k.publicKeys[versionName] = publicKeyBytes
	k.mu.Unlock()
	return publicKeyBytes, nil
}

// signingKeyNames filters a key ring listing down to the asymmetric signing keys this keystore can
// use. Key rings are commonly shared, so anything unusable is skipped rather than failing the whole
// listing.
//
// The version-template check is an optimisation: it rejects keys whose configured algorithm is
// unsupported without spending an RPC on them. VersionTemplate is optional in the API, so a key that
// survives this filter is not necessarily usable; resolveKeyVersion checks the algorithm of the
// version it actually resolves, and GetKeys skips anything that fails there for the same reasons.
func signingKeyNames(listedKeys []*kmspb.CryptoKey) []string {
	keyNames := make([]string, 0, len(listedKeys))
	for _, key := range listedKeys {
		if key == nil || key.Name == "" {
			continue
		}
		if key.Purpose != kmspb.CryptoKey_ASYMMETRIC_SIGN {
			continue
		}
		if template := key.VersionTemplate; template != nil {
			if _, err := cryptoKeyVersionAlgorithmToKeyType(template.Algorithm); err != nil {
				continue
			}
		}
		keyNames = append(keyNames, key.Name)
	}
	return keyNames
}

// GetKeys lists keys in the Cloud KMS keystore.
//
// Key names are either CryptoKey resource names
// (projects/<p>/locations/<l>/keyRings/<r>/cryptoKeys/<k>), for which the highest-numbered enabled
// version is resolved and pinned, or CryptoKeyVersion resource names, which name one version
// outright. Keys are returned sorted by name, per [keystore.Reader]; correlate them by
// KeyInfo.Name rather than by position.
//
// When no key names are given, the configured key ring is listed and keys that this keystore
// cannot use — another purpose, an unsupported algorithm, or no enabled version — are skipped,
// since a key ring may hold keys that have nothing to do with this keystore. Explicitly requested
// keys always surface their errors.
//
// Listing requires a client that implements [KeyRingLister] and credentials holding a ring-wide
// cloudkms.cryptoKeys.list. Deployments that follow least privilege grant neither and should pass an
// explicit key allowlist instead.
func (k *keystoreSignerReader) GetKeys(ctx context.Context, req keystore.GetKeysRequest) (keystore.GetKeysResponse, error) {
	keyNames := append([]string(nil), req.KeyNames...)
	listed := len(keyNames) == 0
	if listed {
		if k.keyRingName == "" {
			return keystore.GetKeysResponse{}, errors.New("key ring name is required to list Cloud KMS keys")
		}
		lister, ok := k.client.(KeyRingLister)
		if !ok {
			return keystore.GetKeysResponse{}, fmt.Errorf("cannot list key ring %s: this Cloud KMS client does not implement KeyRingLister; request keys explicitly by name instead", k.keyRingName)
		}
		listedKeys, err := lister.ListCryptoKeys(ctx, k.keyRingName)
		if err != nil {
			return keystore.GetKeysResponse{}, err
		}
		keyNames = signingKeyNames(listedKeys)
	}
	// keystore.Reader specifies keys sorted by name. Sorting the names up front also gives the
	// listing path a stable order, which Cloud KMS does not otherwise guarantee.
	sort.Strings(keyNames)

	keys := make([]keystore.GetKeyResponse, 0, len(keyNames))
	seen := make(map[string]struct{}, len(keyNames))
	for _, keyName := range keyNames {
		if _, ok := seen[keyName]; ok {
			return keystore.GetKeysResponse{}, fmt.Errorf("key %s provided multiple times", keyName)
		}
		seen[keyName] = struct{}{}
		resolved, err := k.resolveKeyVersion(ctx, keyName)
		if err != nil {
			if listed && (errors.Is(err, errNoEnabledVersion) || errors.Is(err, errUnsupportedAlgorithm)) {
				continue
			}
			return keystore.GetKeysResponse{}, err
		}
		publicKeyBytes, err := k.getPublicKeyBytes(ctx, resolved.version.Name, resolved.keyType)
		if err != nil {
			return keystore.GetKeysResponse{}, err
		}
		createdAt := resolved.version.CreateTime.AsTime()
		keys = append(keys, keystore.GetKeyResponse{
			KeyInfo: keystore.NewKeyInfo(keyName, resolved.keyType, createdAt, publicKeyBytes, []byte{}),
		})
	}
	return keystore.GetKeysResponse{Keys: keys}, nil
}

// Sign signs data using the Cloud KMS crypto key specified by the key name.
func (k *keystoreSignerReader) Sign(ctx context.Context, req keystore.SignRequest) (keystore.SignResponse, error) {
	resolved, err := k.resolveKeyVersion(ctx, req.KeyName)
	if err != nil {
		return keystore.SignResponse{}, err
	}
	versionName := resolved.version.Name

	switch resolved.keyType {
	case keystore.ECDSA_S256:
		if len(req.Data) != 32 {
			return keystore.SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), keystore.ErrInvalidSignRequest)
		}
		// Needed only to recover the SEC1 `v` byte below. Cached per version, so this is at most
		// one extra round trip per key version for the lifetime of the keystore.
		pubKeyBytes, err := k.getPublicKeyBytes(ctx, versionName, resolved.keyType)
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to get public key for key %s: %w", req.KeyName, err)
		}

		// The data is a pre-hashed 32-byte digest. For EC_SIGN_SECP256K1_SHA256 the digest field is
		// the exact bytes signed; Cloud KMS does not re-hash.
		sig, err := k.client.AsymmetricSign(ctx, &kmspb.AsymmetricSignRequest{
			Name:         versionName,
			Digest:       &kmspb.Digest{Digest: &kmspb.Digest_Sha256{Sha256: req.Data}},
			DigestCrc32C: wrapperspb.Int64(crc32c(req.Data)),
		})
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to sign data: %w", err)
		}
		if sig == nil {
			return keystore.SignResponse{}, errors.New("empty signing response from Cloud KMS")
		}
		if sig.Name != versionName {
			return keystore.SignResponse{}, fmt.Errorf("signing response has name %q, expected %q", sig.Name, versionName)
		}
		if !sig.VerifiedDigestCrc32C {
			return keystore.SignResponse{}, errors.New("digest CRC32C checksum was not verified by Cloud KMS")
		}
		if err = checkCrc32c(sig.Signature, sig.SignatureCrc32C); err != nil {
			return keystore.SignResponse{}, fmt.Errorf("signature for key %s: %w", req.KeyName, err)
		}
		// Cloud KMS returns the ECDSA signature in ASN.1 DER format, identical to AWS. Reuse the
		// shared conversion to SEC1 (R || S || V).
		signature, err := kms.ASN1ToSEC1Sig(sig.Signature, pubKeyBytes, req.Data)
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to convert Cloud KMS signature to SEC1 signature: %w", err)
		}
		return keystore.SignResponse{Signature: signature}, nil
	case keystore.Ed25519:
		// Ed25519 signs arbitrary length messages. For EC_SIGN_ED25519 the raw data field is the
		// exact bytes signed; Cloud KMS does not hash.
		sig, err := k.client.AsymmetricSign(ctx, &kmspb.AsymmetricSignRequest{
			Name:       versionName,
			Data:       req.Data,
			DataCrc32C: wrapperspb.Int64(crc32c(req.Data)),
		})
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to sign data: %w", err)
		}
		if sig == nil {
			return keystore.SignResponse{}, errors.New("empty signing response from Cloud KMS")
		}
		if sig.Name != versionName {
			return keystore.SignResponse{}, fmt.Errorf("signing response has name %q, expected %q", sig.Name, versionName)
		}
		if !sig.VerifiedDataCrc32C {
			return keystore.SignResponse{}, errors.New("data CRC32C checksum was not verified by Cloud KMS")
		}
		if err = checkCrc32c(sig.Signature, sig.SignatureCrc32C); err != nil {
			return keystore.SignResponse{}, fmt.Errorf("signature for key %s: %w", req.KeyName, err)
		}
		// Ed25519 signatures from Cloud KMS are already in the correct format.
		if len(sig.Signature) != ed25519.SignatureSize {
			return keystore.SignResponse{}, fmt.Errorf("invalid Ed25519 signature length: expected %d bytes, got %d", ed25519.SignatureSize, len(sig.Signature))
		}
		return keystore.SignResponse{Signature: sig.Signature}, nil
	default:
		return keystore.SignResponse{}, fmt.Errorf("key %s: %w", req.KeyName, keystore.ErrInvalidSignRequest)
	}
}

func (k *keystoreSignerReader) Verify(ctx context.Context, req keystore.VerifyRequest) (keystore.VerifyResponse, error) {
	return keystore.Verify(ctx, req)
}
