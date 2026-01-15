package kms

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/x509"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type FakeKMSClient struct {
	keys      []Key
	createdAt time.Time
}

type Key struct {
	ECDSAKey   *ecdsa.PrivateKey
	Ed25519Key ed25519.PrivateKey
	KeyID      string
	KeySpec    kmstypes.KeySpec
}

// IsECDSA returns true if this key is an ECDSA key.
func (k Key) IsECDSA() bool {
	return k.ECDSAKey != nil
}

// IsEd25519 returns true if this key is an Ed25519 key.
func (k Key) IsEd25519() bool {
	return len(k.Ed25519Key) > 0
}

func NewFakeKMSClient(keys []Key) (*FakeKMSClient, error) {
	return &FakeKMSClient{
		keys:      keys,
		createdAt: time.Now(),
	}, nil
}

func (m *FakeKMSClient) GetPublicKey(ctx context.Context, input *kms.GetPublicKeyInput, opts ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	if input.KeyId == nil {
		return nil, errors.New("key ID is required")
	}
	for _, key := range m.keys {
		if *input.KeyId == key.KeyID {
			var asn1PubKey []byte
			var err error
			if key.IsECDSA() {
				asn1PubKey, err = SEC1ToASN1PublicKey(crypto.FromECDSAPub(&key.ECDSAKey.PublicKey))
				if err != nil {
					return nil, err
				}
			} else if key.IsEd25519() {
				// Ed25519 public key is the last 32 bytes of the private key
				pubKey := ed25519.PublicKey(key.Ed25519Key[32:])
				asn1PubKey, err = x509.MarshalPKIXPublicKey(pubKey)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("key has no valid private key")
			}
			return &kms.GetPublicKeyOutput{
				KeyId:     &key.KeyID,
				PublicKey: asn1PubKey,
			}, nil
		}
	}
	return nil, errors.New("key not found")
}

func (m *FakeKMSClient) Sign(ctx context.Context, input *kms.SignInput, opts ...func(*kms.Options)) (*kms.SignOutput, error) {
	if input.KeyId == nil {
		return nil, errors.New("key ID is required")
	}
	for _, key := range m.keys {
		if *input.KeyId == key.KeyID {
			var signature []byte
			if key.IsECDSA() {
				sig, err := crypto.Sign(input.Message, key.ECDSAKey)
				if err != nil {
					return nil, err
				}
				signature, err = SEC1ToASN1Sig(sig)
				if err != nil {
					return nil, err
				}
			} else if key.IsEd25519() {
				// Ed25519 signatures are 64 bytes and don't need ASN.1 encoding
				signature = ed25519.Sign(key.Ed25519Key, input.Message)
			} else {
				return nil, errors.New("key has no valid private key")
			}
			return &kms.SignOutput{
				KeyId:     &key.KeyID,
				Signature: signature,
			}, nil
		}
	}
	return nil, errors.New("key not found")
}

// DescribeKey returns metadata about the key.
func (m *FakeKMSClient) DescribeKey(ctx context.Context, input *kms.DescribeKeyInput, opts ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	if input.KeyId == nil {
		return nil, errors.New("key ID is required")
	}
	for _, key := range m.keys {
		if *input.KeyId == key.KeyID {
			keySpec := key.KeySpec
			// Default based on key type if not specified
			if keySpec == "" {
				if key.IsECDSA() {
					keySpec = kmstypes.KeySpecEccSecgP256k1
				} else if key.IsEd25519() {
					keySpec = kmstypes.KeySpecEccNistEdwards25519
				} else {
					return nil, errors.New("key has no valid private key")
				}
			}
			return &kms.DescribeKeyOutput{
				KeyMetadata: &kmstypes.KeyMetadata{
					KeyId:        &key.KeyID,
					KeySpec:      keySpec,
					CreationDate: &m.createdAt,
				},
			}, nil
		}
	}
	return nil, errors.New("key not found")
}

// ListKeys returns a list of key IDs.
func (m *FakeKMSClient) ListKeys(ctx context.Context, input *kms.ListKeysInput, opts ...func(*kms.Options)) (*kms.ListKeysOutput, error) {
	keys := make([]kmstypes.KeyListEntry, 0, len(m.keys))
	for _, key := range m.keys {
		keyID := key.KeyID
		keys = append(keys, kmstypes.KeyListEntry{
			KeyId: &keyID,
		})
	}
	return &kms.ListKeysOutput{
		Keys: keys,
	}, nil
}
