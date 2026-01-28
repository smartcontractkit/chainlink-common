// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/serialization"
)

var (
	ErrInvalidExportFormat = errors.New("invalid export format")
)

const (
	TypeCSA      = "csa"
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

func (ks *Store) GenerateEncryptedCSAKey(ctx context.Context, password string) ([]byte, error) {
	path := keystore.NewKeyPath(TypeCSA, nameDefault)
	_, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				KeyName: path.String(),
				KeyType: keystore.Ed25519,
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
		Type:         TypeCSA,
		Keys:         er.Keys,
		ExportFormat: exportFormat,
	}

	data, err := json.Marshal(&envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envelope: %w", err)
	}

	return data, nil
}

func FromEncryptedCSAKey(data []byte, password string) ([]byte, error) {
	envelope := Envelope{}
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal import data into envelope: %w", err)
	}

	if envelope.ExportFormat != exportFormat {
		return nil, fmt.Errorf("invalid export format: %w", ErrInvalidExportFormat)
	}

	if envelope.Type != TypeCSA {
		return nil, fmt.Errorf("invalid key type: expected %s, got %s", TypeCSA, envelope.Type)
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
