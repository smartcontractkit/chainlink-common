// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

const (
	TypeWorkflowKey = "workflow"
)

func (ks *Store) GenerateEncryptedWorkflowKey(ctx context.Context, password string) ([]byte, error) {
	path := keystore.NewKeyPath(TypeWorkflowKey, nameDefault)
	_, err := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				KeyName: path.String(),
				KeyType: keystore.X25519,
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
		Type:         TypeWorkflowKey,
		Keys:         er.Keys,
		ExportFormat: exportFormat,
	}

	data, err := json.Marshal(&envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envelope: %w", err)
	}

	return data, nil
}

func FromEncryptedWorkflowKey(data []byte, password string) ([]byte, error) {
	envelope := Envelope{}
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal import data into envelope: %w", err)
	}

	if envelope.ExportFormat != exportFormat {
		return nil, fmt.Errorf("invalid export format: %w", ErrInvalidExportFormat)
	}

	if envelope.Type != TypeWorkflowKey {
		return nil, fmt.Errorf("invalid key type: expected %s, got %s", TypeWorkflowKey, envelope.Type)
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
