package kms

import (
	"crypto/ecdsa"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	kmslib "github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/crypto"
)

type MockKMSClient struct {
	keys      []Key
	createdAt time.Time
}

type Key struct {
	PrivateKey *ecdsa.PrivateKey
	KeyID      string
}

func NewMockKMSClient(keys []Key) (*MockKMSClient, error) {
	return &MockKMSClient{
		keys:      keys,
		createdAt: time.Now(),
	}, nil
}

func (m *MockKMSClient) GetPublicKey(input *kmslib.GetPublicKeyInput) (*kmslib.GetPublicKeyOutput, error) {
	for _, key := range m.keys {
		if aws.StringValue(input.KeyId) == key.KeyID {
			asn1PubKey, err := SEC1ToASN1PublicKey(crypto.FromECDSAPub(&key.PrivateKey.PublicKey))
			if err != nil {
				return nil, err
			}
			return &kmslib.GetPublicKeyOutput{
				KeyId:     aws.String(key.KeyID),
				PublicKey: asn1PubKey,
			}, nil
		}
	}
	return nil, awserr.New(kmslib.ErrCodeNotFoundException, "key not found", errors.New("key not found"))
}

func (m *MockKMSClient) Sign(input *kmslib.SignInput) (*kmslib.SignOutput, error) {
	for _, key := range m.keys {
		if aws.StringValue(input.KeyId) == key.KeyID {
			sig, err := crypto.Sign(input.Message, key.PrivateKey)
			if err != nil {
				return nil, err
			}
			asn1Sig, err := SEC1ToASN1Sig(sig)
			if err != nil {
				return nil, err
			}
			return &kmslib.SignOutput{
				KeyId:     aws.String(key.KeyID),
				Signature: asn1Sig,
			}, nil
		}
	}
	return nil, awserr.New(kmslib.ErrCodeNotFoundException, "key not found", errors.New("key not found"))
}

// DescribeKey returns metadata about the key.
func (m *MockKMSClient) DescribeKey(input *kmslib.DescribeKeyInput) (*kmslib.DescribeKeyOutput, error) {
	for _, key := range m.keys {
		if aws.StringValue(input.KeyId) == key.KeyID {
			return &kmslib.DescribeKeyOutput{
				KeyMetadata: &kmslib.KeyMetadata{
					KeyId:        aws.String(key.KeyID),
					KeySpec:      aws.String(kmslib.KeySpecEccSecgP256k1),
					CreationDate: aws.Time(m.createdAt),
				},
			}, nil
		}
	}
	return nil, awserr.New(kmslib.ErrCodeNotFoundException, "key not found", errors.New("key not found"))
}

// ListKeys returns a list of key IDs.
func (m *MockKMSClient) ListKeys(_ *kmslib.ListKeysInput) (*kmslib.ListKeysOutput, error) {
	keys := make([]*kmslib.KeyListEntry, 0, len(m.keys))
	for _, key := range m.keys {
		keys = append(keys, &kmslib.KeyListEntry{
			KeyId: aws.String(key.KeyID),
		})
	}
	return &kmslib.ListKeysOutput{
		Keys: keys,
	}, nil
}
