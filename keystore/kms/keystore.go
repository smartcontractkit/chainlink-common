package kms

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	kmslib "github.com/aws/aws-sdk-go/service/kms"
	"github.com/smartcontractkit/chainlink-common/keystore"
	kms "github.com/smartcontractkit/chainlink-common/keystore/kms/internal"
	kmsinternal "github.com/smartcontractkit/chainlink-common/keystore/kms/internal"
)

// KeystoreConfig is the configuration for the KMS keystore.
// Struct for extensibility in the future.
// Uses the default region from the AWS profile.
type KeystoreConfig struct {
	AWSProfile string
}

type kmsKeystoreSignerReader struct {
	client kmsinternal.Client
	config KeystoreConfig
}

func NewKMSKeystore(config KeystoreConfig) (keystore.KeystoreSignerReader, error) {
	client, err := kmsinternal.NewClient(config.AWSProfile)
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
func keySpecToKeyType(keySpec string) (keystore.KeyType, error) {
	switch keySpec {
	case kmslib.KeySpecEccSecgP256k1:
		return keystore.ECDSA_S256, nil
	default:
		return "", fmt.Errorf("unsupported KMS key spec: %s (only ECC_SECG_P256K1 is supported)", keySpec)
	}
}

// GetKeys lists all keys in the KMS keystore.
// Note: possible that ListKeys is not supported for a given AWS role
// but specific GetPublicKey/DescribeKey are supported. We transparently
// let the user use whatever they are permitted to given their AWS perms.
func (k *kmsKeystoreSignerReader) GetKeys(ctx context.Context, req keystore.GetKeysRequest) (keystore.GetKeysResponse, error) {
	if len(req.KeyNames) == 0 {
		listResp, err := k.client.ListKeys(&kmslib.ListKeysInput{})
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("failed to list KMS keys: %w", err)
		}
		req.KeyNames = make([]string, 0, len(listResp.Keys))
		for _, key := range listResp.Keys {
			req.KeyNames = append(req.KeyNames, aws.StringValue(key.KeyId))
		}
	}

	keys := make([]keystore.GetKeyResponse, 0, len(req.KeyNames))
	for _, keyID := range req.KeyNames {
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
		createdAt := time.Unix(describeKey.KeyMetadata.CreationDate.Unix(), 0)

		// Convert KMS KeySpec to keystore KeyType
		keySpec := aws.StringValue(describeKey.KeyMetadata.KeySpec)
		keyType, err := keySpecToKeyType(keySpec)
		if err != nil {
			return keystore.GetKeysResponse{}, fmt.Errorf("key %s: %w", keyID, err)
		}
		var publicKeyBytes []byte
		switch keyType {
		case keystore.ECDSA_S256:
			publicKeyBytes, err = kms.ASN1ToSEC1PublicKey(key.PublicKey)
			if err != nil {
				return keystore.GetKeysResponse{}, fmt.Errorf("failed to convert public key for key %s: %w", keyID, err)
			}
		default:
			return keystore.GetKeysResponse{}, fmt.Errorf("unsupported key type: %s", keyType)
		}

		keys = append(keys, keystore.GetKeyResponse{
			KeyInfo: keystore.KeyInfo{
				Name:      keyID,
				KeyType:   keyType,
				PublicKey: publicKeyBytes,
				CreatedAt: createdAt,
			},
		})
	}
	return keystore.GetKeysResponse{Keys: keys}, nil
}

// Sign signs data using the KMS key specified by the key name.
func (k *kmsKeystoreSignerReader) Sign(ctx context.Context, req keystore.SignRequest) (keystore.SignResponse, error) {
	key, err := k.client.GetPublicKey(&kmslib.GetPublicKeyInput{
		KeyId: aws.String(req.KeyName),
	})
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to get public key for key %s: %w", req.KeyName, err)
	}
	describeKey, err := k.client.DescribeKey(&kmslib.DescribeKeyInput{
		KeyId: aws.String(req.KeyName),
	})
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("failed to describe key %s: %w", req.KeyName, err)
	}
	keySpec := aws.StringValue(describeKey.KeyMetadata.KeySpec)
	keyType, err := keySpecToKeyType(keySpec)
	if err != nil {
		return keystore.SignResponse{}, fmt.Errorf("key %s: %w", req.KeyName, err)
	}

	switch keyType {
	case keystore.ECDSA_S256:
		if len(req.Data) != 32 {
			return keystore.SignResponse{}, fmt.Errorf("data must be 32 bytes for ECDSA_S256, got %d: %w", len(req.Data), keystore.ErrInvalidSignRequest)
		}
		pubKeyBytes, err := kms.ASN1ToSEC1PublicKey(key.PublicKey)
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to convert public key for KeyId=%s: %w", req.KeyName, err)
		}

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
		signature, err := kms.KMSToSEC1Sig(sig.Signature, pubKeyBytes, req.Data)
		if err != nil {
			return keystore.SignResponse{}, fmt.Errorf("failed to convert KMS signature to SEC1 signature: %w", err)
		}
		return keystore.SignResponse{
			Signature: signature,
		}, nil
	default:
		return keystore.SignResponse{}, fmt.Errorf("key %s: %w", req.KeyName, keystore.ErrInvalidSignRequest)
	}

}

func (k *kmsKeystoreSignerReader) Verify(ctx context.Context, req keystore.VerifyRequest) (keystore.VerifyResponse, error) {
	return keystore.Verify(ctx, req)
}
