package kms

import (
	"context"
	"crypto/ecdsa"
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
	PrivateKey *ecdsa.PrivateKey
	KeyID      string
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
			asn1PubKey, err := SEC1ToASN1PublicKey(crypto.FromECDSAPub(&key.PrivateKey.PublicKey))
			if err != nil {
				return nil, err
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
			sig, err := crypto.Sign(input.Message, key.PrivateKey)
			if err != nil {
				return nil, err
			}
			asn1Sig, err := SEC1ToASN1Sig(sig)
			if err != nil {
				return nil, err
			}
			return &kms.SignOutput{
				KeyId:     &key.KeyID,
				Signature: asn1Sig,
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
			keySpec := kmstypes.KeySpecEccSecgP256k1
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
