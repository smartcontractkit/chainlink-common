package corekeys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/serialization"
)

var keyTypeToAlgorithmMap = keyTypeToAlgorithm{
	TypeCSA:         keystore.Ed25519,
	TypeP2P:         keystore.Ed25519,
	TypeEVM:         keystore.ECDSA_S256,
	TypeSolana:      keystore.Ed25519,
	TypeDKG:         keystore.ECDH_P256,
	TypeWorkflowKey: keystore.X25519,
}

type keyTypeToAlgorithm map[string]keystore.KeyType

func (m keyTypeToAlgorithm) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

var (
	ErrInvalidExportFormat = errors.New("invalid export format")
)

const (
	nameDefault  = "default"
	exportFormat = "github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

type Store struct {
	keystore.Keystore
}

type Envelope struct {
	Type         string
	Keys         []keystore.ExportKeyResponse
	ExportFormat string
}

func NewStore(ks keystore.Keystore) *Store {
	return &Store{
		Keystore: ks,
	}
}

// decryptKey decrypts an encrypted key using the provided password and returns the deserialized key.
func decryptKey(encryptedData []byte, password string) (*serialization.Key, error) {
	encData := gethkeystore.CryptoJSON{}
	err := json.Unmarshal(encryptedData, &encData)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal key material into CryptoJSON: %w", err)
	}

	decData, err := gethkeystore.DecryptDataV3(encData, password)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt data: %w", err)
	}

	keypb := &serialization.Key{}
	err = proto.Unmarshal(decData, keypb)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal key into serialization.Key: %w", err)
	}

	return keypb, nil
}

func (ks *Store) generateEncryptedKey(ctx context.Context, keyType string, password string) ([]byte, error) {
	algo, ok := keyTypeToAlgorithmMap[keyType]
	if !ok {
		return nil, fmt.Errorf("unsupported key type: %s. Supported types are: %s", keyType, strings.Join(keyTypeToAlgorithmMap.Keys(), ", "))
	}

	path := keystore.NewKeyPath(keyType, nameDefault)
	_, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				KeyName: path.String(),
				KeyType: algo,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate exportable key: %w", err)
	}

	er, err := ks.ExportKeys(ctx, keystore.ExportKeysRequest{
		Keys: []keystore.ExportKeyParam{
			{
				KeyName: path.String(),
				Enc: keystore.EncryptionParams{
					Password:     password,
					ScryptParams: keystore.DefaultScryptParams,
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to export key: %w", err)
	}

	envelope := Envelope{
		Type:         keyType,
		Keys:         er.Keys,
		ExportFormat: exportFormat,
	}

	data, err := json.Marshal(&envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envelope: %w", err)
	}

	return data, nil
}

func fromEncryptedKey(data []byte, keyType string, password string) ([]byte, error) {
	_, ok := keyTypeToAlgorithmMap[keyType]
	if !ok {
		return nil, fmt.Errorf("unsupported key type: %s. Supported types are: %s", keyType, strings.Join(keyTypeToAlgorithmMap.Keys(), ", "))
	}

	envelope := Envelope{}
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal import data into envelope: %w", err)
	}

	if envelope.ExportFormat != exportFormat {
		return nil, fmt.Errorf("invalid export format: %w", ErrInvalidExportFormat)
	}

	if envelope.Type != keyType {
		return nil, fmt.Errorf("invalid key type: expected %s, got %s", keyType, envelope.Type)
	}

	if len(envelope.Keys) != 1 {
		return nil, fmt.Errorf("expected exactly one key in envelope, got %d", len(envelope.Keys))
	}

	keypb, err := decryptKey(envelope.Keys[0].Data, password)
	if err != nil {
		return nil, err
	}

	return keypb.PrivateKey, nil
}
