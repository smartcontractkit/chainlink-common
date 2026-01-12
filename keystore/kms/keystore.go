package kms

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	kmslib "github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/smartcontractkit/chainlink-common/keystore"
	kms "github.com/smartcontractkit/chainlink-common/keystore/kms/internal"
	kmsinternal "github.com/smartcontractkit/chainlink-common/keystore/kms/internal"
)

type KeystoreConfig struct {
	AWSProfile string
	KeyIDs     []string
	KeyRegion  string
}

type kmsKeystoreSignerReader struct {
	client kmsinternal.Client
	config KeystoreConfig
}

func NewKMSKeystore(config KeystoreConfig) (keystore.KeystoreSignerReader, error) {
	client, err := kmsinternal.NewClient(kmsinternal.ClientConfig{
		KeyRegion:  config.KeyRegion,
		AWSProfile: config.AWSProfile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}
	return &kmsKeystoreSignerReader{
		client: client,
		config: config,
	}, nil
}

// keySpecToKeyType converts an AWS KMS KeySpec to a keystore KeyType.
// AWS KMS supports:
//   - ECC_SECG_P256K1 (secp256k1) -> ECDSA_S256
//   - ECC_NIST_P256, ECC_NIST_P384, ECC_NIST_P521 -> not supported (different curves)
//   - RSA_* -> not supported
//   - SYMMETRIC_DEFAULT -> not supported
//
// Note: AWS KMS does not support Ed25519 keys.
func keySpecToKeyType(keySpec string) (keystore.KeyType, error) {
	switch keySpec {
	case "ECC_SECG_P256K1":
		return keystore.ECDSA_S256, nil
	default:
		return "", fmt.Errorf("unsupported KMS key spec: %s (only ECC_SECG_P256K1 is supported)", keySpec)
	}
}

func (k *kmsKeystoreSignerReader) GetKeys(ctx context.Context, req keystore.GetKeysRequest) (keystore.GetKeysResponse, error) {
	keys := make([]keystore.GetKeyResponse, 0, len(k.config.KeyIDs))
	for _, keyID := range k.config.KeyIDs {
		// Get public key
		key, err := k.client.GetPublicKey(&kmslib.GetPublicKeyInput{
			KeyId: aws.String(keyID),
		})
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("failed to get public key for key %s: %w", keyID, err)
		}

		// Get key metadata to determine key type and creation date
		describeKey, err := k.client.DescribeKey(&kmslib.DescribeKeyInput{
			KeyId: aws.String(keyID),
		})
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("failed to describe key %s: %w", keyID, err)
		}

		// Convert KMS KeySpec to keystore KeyType
		keySpec := aws.StringValue(describeKey.KeyMetadata.KeySpec)
		keyType, err := keySpecToKeyType(keySpec)
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("key %s: %w", keyID, err)
		}

		// Get creation date from metadata
		createdAt := time.Now() // fallback
		if describeKey.KeyMetadata.CreationDate != nil {
			createdAt = *describeKey.KeyMetadata.CreationDate
		}

		keys = append(keys, keystore.GetKeyResponse{
			KeyInfo: keystore.KeyInfo{
				Name:      keyID,
				KeyType:   keyType,
				PublicKey: key.PublicKey,
				CreatedAt: createdAt,
			},
		})
	}
	return keystore.GetKeysResponse{Keys: keys}, nil
}

var (
	// secp256k1N is the N value of the secp256k1 curve, used to adjust the S value in signatures.
	secp256k1N = crypto.S256().Params().N
	// secp256k1HalfN is half of the secp256k1 N value, used to adjust the S value in signatures.
	secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

// Sign signs data using the KMS key specified by the key name.
func (k *kmsKeystoreSignerReader) Sign(ctx context.Context, req keystore.SignRequest) (keystore.SignResponse, error) {
	// TODO: Handle other key types
	if len(req.Data) != 32 {
		return keystore.SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), keystore.ErrInvalidSignRequest)
	}
	key, err := k.client.GetPublicKey(&kmslib.GetPublicKeyInput{
		KeyId: aws.String(req.KeyName),
	})
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to get public key for key %s: %w", req.KeyName, err)
	}
	var spki kms.SPKI
	if _, err = asn1.Unmarshal(key.PublicKey, &spki); err != nil {
		return keystore.SignResponse{}, fmt.Errorf("cannot parse asn1 public key for KeyId=%s: %w", req.KeyName, err)
	}
	pubKey, err := crypto.UnmarshalPubkey(spki.SubjectPublicKey.Bytes)
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to unmarshal public key: %w", err)
	}
	pubKeyBytes := secp256k1.S256().Marshal(pubKey.X, pubKey.Y)

	// MessageType is digest because its prehashed.
	sig, err := k.client.Sign(&kmslib.SignInput{
		KeyId:            aws.String(req.KeyName),
		Message:          req.Data,
		SigningAlgorithm: aws.String(string(kmslib.SigningAlgorithmSpecEcdsaSha256)),
		MessageType:      aws.String(string(kmslib.MessageTypeDigest)),
	})
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to sign data: %w", err)
	}
	signature, err := kmsToEVMSig(sig.Signature, pubKeyBytes, req.Data)
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to convert KMS signature to EVM signature: %w", err)
	}
	return keystore.SignResponse{
		Signature: signature,
	}, nil
}

// kmsToEVMSig converts a KMS signature to an Ethereum-compatible signature. This follows this
// example provided by AWS Guides.
//
// [AWS Guides]: https://aws.amazon.com/blogs/database/part2-use-aws-kms-to-securely-manage-ethereum-accounts/
func kmsToEVMSig(kmsSig, ecdsaPubKeyBytes, hash []byte) ([]byte, error) {
	var ecdsaSig kms.ECDSASig
	if _, err := asn1.Unmarshal(kmsSig, &ecdsaSig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal KMS signature: %w", err)
	}

	rBytes := ecdsaSig.R.Bytes
	sBytes := ecdsaSig.S.Bytes

	// Adjust S value from signature to match EVM standard.
	//
	// After we extract r and s successfully, we have to test if the value of s is greater than
	// secp256k1n/2 as specified in EIP-2 and flip it if required.
	sBigInt := new(big.Int).SetBytes(sBytes)
	if sBigInt.Cmp(secp256k1HalfN) > 0 {
		sBytes = new(big.Int).Sub(secp256k1N, sBigInt).Bytes()
	}

	return recoverEVMSignature(ecdsaPubKeyBytes, hash, rBytes, sBytes)
}

// recoverEVMSignature attempts to reconstruct the EVM signature by trying both possible recovery
// IDs (v = 0 and v = 1). It compares the recovered public key with the expected public key bytes
// to determine the correct signature.
//
// Returns the valid EVM signature if successful, or an error if neither recovery ID matches.
func recoverEVMSignature(expectedPublicKey, txHash, r, s []byte) ([]byte, error) {
	// Ethereum signatures require r and s to be exactly 32 bytes each.
	rsSig := append(padTo32Bytes(r), padTo32Bytes(s)...)
	// Ethereum signatures have a 65th byte called the recovery ID (v), which can be 0 or 1.
	// Here we append 0 to the signature to start with for the first recovery attempt.
	evmSig := append(rsSig, []byte{0}...)

	recoveredPublicKey, err := crypto.Ecrecover(txHash, evmSig)
	if err != nil {
		return nil, fmt.Errorf("failed to recover signature with v=0: %w", err)
	}

	if hex.EncodeToString(recoveredPublicKey) != hex.EncodeToString(expectedPublicKey) {
		// If the first recovery attempt failed, we try with v=1.
		evmSig = append(rsSig, []byte{1}...)
		recoveredPublicKey, err = crypto.Ecrecover(txHash, evmSig)
		if err != nil {
			return nil, fmt.Errorf("failed to recover signature with v=1: %w", err)
		}

		if hex.EncodeToString(recoveredPublicKey) != hex.EncodeToString(expectedPublicKey) {
			return nil, errors.New("cannot reconstruct public key from sig")
		}
	}

	return evmSig, nil
}

// padTo32Bytes pads the given byte slice to 32 bytes by trimming leading zeros and prepending
// zeros.
func padTo32Bytes(buffer []byte) []byte {
	buffer = bytes.TrimLeft(buffer, "\x00")
	for len(buffer) < 32 {
		zeroBuf := []byte{0}
		buffer = append(zeroBuf, buffer...)
	}

	return buffer
}

func (k *kmsKeystoreSignerReader) Verify(ctx context.Context, req keystore.VerifyRequest) (keystore.VerifyResponse, error) {
	if req.KeyType != keystore.ECDSA_S256 {
		return keystore.VerifyResponse{}, fmt.Errorf("KMS keystore only supports ECDSA_S256, got %s: %w", req.KeyType, keystore.ErrInvalidVerifyRequest)
	}

	if len(req.Data) != 32 {
		return keystore.VerifyResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), keystore.ErrInvalidVerifyRequest)
	}

	// ECDSA_S256 public keys are in SEC1 (uncompressed) format
	if len(req.PublicKey) != 65 {
		return keystore.VerifyResponse{}, fmt.Errorf("public key must be 65 bytes for ECDSA_S256, got %d: %w", len(req.PublicKey), keystore.ErrInvalidVerifyRequest)
	}

	if len(req.Signature) != 65 {
		return keystore.VerifyResponse{}, fmt.Errorf("signature must be 65 bytes for ECDSA_S256, got %d: %w", len(req.Signature), keystore.ErrInvalidVerifyRequest)
	}

	// VerifySignature expects 64 bytes [R || S] without the V byte
	// Strip the V byte (last byte) from the 65-byte signature
	valid := crypto.VerifySignature(req.PublicKey, req.Data, req.Signature[:64])
	return keystore.VerifyResponse{
		Valid: valid,
	}, nil
}
