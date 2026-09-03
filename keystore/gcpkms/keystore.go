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

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/kms"
)

// errUnsupportedAlgorithm marks a CryptoKeyVersion whose algorithm this keystore cannot use.
var errUnsupportedAlgorithm = errors.New("unsupported Cloud KMS key algorithm")

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
	client Client
}

func NewKeystore(client Client) (interface {
	keystore.Reader
	keystore.Signer
}, error) {
	if client == nil {
		return nil, errors.New("GCP KMS client is required")
	}
	return &keystoreSignerReader{client: client}, nil
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

// getKeyVersion fetches a CryptoKeyVersion from Cloud KMS and validates that it is enabled and
// uses a supported algorithm.
func (k *keystoreSignerReader) getKeyVersion(ctx context.Context, versionName string) (*kmspb.CryptoKeyVersion, keystore.KeyType, error) {
	version, err := k.client.GetCryptoKeyVersion(ctx, &kmspb.GetCryptoKeyVersionRequest{Name: versionName})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get crypto key version %s: %w", versionName, err)
	}
	if version == nil {
		return nil, "", fmt.Errorf("empty crypto key version response from Cloud KMS for %s", versionName)
	}
	if version.Name != versionName {
		return nil, "", fmt.Errorf("crypto key version response has name %q, expected %q", version.Name, versionName)
	}
	if version.State != kmspb.CryptoKeyVersion_ENABLED {
		return nil, "", fmt.Errorf("crypto key version %s is not enabled (state=%s)", version.Name, version.State)
	}
	if version.CreateTime == nil {
		return nil, "", fmt.Errorf("crypto key version %s has no creation time", version.Name)
	}
	keyType, err := cryptoKeyVersionAlgorithmToKeyType(version.Algorithm)
	if err != nil {
		return nil, "", fmt.Errorf("crypto key version %s: %w", version.Name, err)
	}
	return version, keyType, nil
}

// publicKeyBytes fetches the public key for a crypto key version and converts it to the
// keystore's native format for the given key type.
func (k *keystoreSignerReader) publicKeyBytes(ctx context.Context, versionName string, keyType keystore.KeyType) ([]byte, error) {
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

	switch keyType {
	case keystore.ECDSA_S256:
		// GCP returns the public key in ASN.1 DER-encoded SubjectPublicKeyInfo (SPKI) format,
		// identical to AWS. Reuse the shared conversion.
		return kms.ASN1ToSEC1PublicKey(block.Bytes)
	case keystore.Ed25519:
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Ed25519 public key for %s: %w", versionName, err)
		}
		ed25519PubKey, ok := pubKey.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to convert Ed25519 public key for %s to ed25519.PublicKey", versionName)
		}
		return ed25519PubKey, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// GetKeys returns the requested keys from the Cloud KMS keystore, sorted by name.
//
// Key names are CryptoKeyVersion resource names
// (projects/<p>/locations/<l>/keyRings/<r>/cryptoKeys/<k>/cryptoKeyVersions/<n>): a key name
// always names exactly one version, so rotating a key means configuring the new version's name.
func (k *keystoreSignerReader) GetKeys(ctx context.Context, req keystore.GetKeysRequest) (keystore.GetKeysResponse, error) {
	if len(req.KeyNames) == 0 {
		return keystore.GetKeysResponse{}, errors.New("key names are required: this keystore does not list key rings")
	}
	versionNames := append([]string(nil), req.KeyNames...)
	sort.Strings(versionNames)

	keys := make([]keystore.GetKeyResponse, 0, len(versionNames))
	seen := make(map[string]struct{}, len(versionNames))
	for _, versionName := range versionNames {
		if _, ok := seen[versionName]; ok {
			return keystore.GetKeysResponse{}, fmt.Errorf("key %s provided multiple times", versionName)
		}
		seen[versionName] = struct{}{}
		version, keyType, err := k.getKeyVersion(ctx, versionName)
		if err != nil {
			return keystore.GetKeysResponse{}, err
		}
		publicKey, err := k.publicKeyBytes(ctx, versionName, keyType)
		if err != nil {
			return keystore.GetKeysResponse{}, err
		}
		keys = append(keys, keystore.GetKeyResponse{
			KeyInfo: keystore.NewKeyInfo(versionName, keyType, version.CreateTime.AsTime(), publicKey, []byte{}),
		})
	}
	return keystore.GetKeysResponse{Keys: keys}, nil
}

// Sign signs data using the Cloud KMS crypto key version specified by the key name.
//
// The key name must be a CryptoKeyVersion resource name: Cloud KMS rejects bare CryptoKey names
// on AsymmetricSign, so a rotation can never change which version a configured name signs with.
func (k *keystoreSignerReader) Sign(ctx context.Context, req keystore.SignRequest) (keystore.SignResponse, error) {
	_, keyType, err := k.getKeyVersion(ctx, req.KeyName)
	if err != nil {
		return keystore.SignResponse{}, err
	}
	versionName := req.KeyName

	switch keyType {
	case keystore.ECDSA_S256:
		if len(req.Data) != 32 {
			return keystore.SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), keystore.ErrInvalidSignRequest)
		}
		// Needed to recover the SEC1 `v` byte from the ASN.1 signature below.
		pubKeyBytes, err := k.publicKeyBytes(ctx, versionName, keyType)
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
