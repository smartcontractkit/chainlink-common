package corekeys

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/ocr2offchain"
)

type ChainType string

const (
	// Must match ChainType in core.
	chainTypeEVM ChainType = "evm"
)

const (
	TypeOCR           = "ocr"
	PrefixOCR2Onchain = "ocr2_onchain"
)

type OCRKeyBundle struct {
	ChainType             ChainType
	OffchainSigningKey    []byte
	OffchainEncryptionKey []byte
	OnchainSigningKey     []byte
}

func (ks *Store) GenerateEncryptedOCRKeyBundle(ctx context.Context, chainType ChainType, password string) ([]byte, error) {
	_, err := ocr2offchain.CreateOCR2OffchainKeyring(ctx, ks.Keystore, nameDefault)
	if err != nil {
		return nil, err
	}

	var onchainKeyPath keystore.KeyPath
	switch chainType {
	case chainTypeEVM:
		path := keystore.NewKeyPath(PrefixOCR2Onchain, nameDefault, string(chainType))
		_, ierr := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
			Keys: []keystore.CreateKeyRequest{
				{
					KeyName: path.String(),
					KeyType: keystore.ECDSA_S256,
				},
			},
		})
		if ierr != nil {
			return nil, fmt.Errorf("failed to generate exportable key: %w", ierr)
		}

		onchainKeyPath = path
	default:
		return nil, fmt.Errorf("unsupported chain type: %s", chainType)
	}

	er, err := ks.ExportKeys(ctx, keystore.ExportKeysRequest{
		Keys: []keystore.ExportKeyParam{
			{
				KeyName: keystore.NewKeyPath(ocr2offchain.PrefixOCR2Offchain, nameDefault, ocr2offchain.OCR2OffchainSigning).String(),
				Enc: keystore.EncryptionParams{
					Password:     password,
					ScryptParams: keystore.DefaultScryptParams,
				},
			},
			{
				KeyName: keystore.NewKeyPath(ocr2offchain.PrefixOCR2Offchain, nameDefault, ocr2offchain.OCR2OffchainEncryption).String(),
				Enc: keystore.EncryptionParams{
					Password:     password,
					ScryptParams: keystore.DefaultScryptParams,
				},
			},
			{
				KeyName: onchainKeyPath.String(),
				Enc: keystore.EncryptionParams{
					Password:     password,
					ScryptParams: keystore.DefaultScryptParams,
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to export OCR key bundle: %w", err)
	}

	envelope := Envelope{
		Type:         TypeOCR,
		Keys:         er.Keys,
		ExportFormat: exportFormat,
	}

	data, err := json.Marshal(&envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OCR key bundle envelope: %w", err)
	}

	return data, nil
}

func FromEncryptedOCRKeyBundle(data []byte, password string) (*OCRKeyBundle, error) {
	envelope := Envelope{}
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal import data into envelope: %w", err)
	}

	if envelope.ExportFormat != exportFormat {
		return nil, fmt.Errorf("invalid export format: %w", ErrInvalidExportFormat)
	}

	if envelope.Type != TypeOCR {
		return nil, fmt.Errorf("invalid key type: expected %s, got %s", TypeOCR, envelope.Type)
	}

	if len(envelope.Keys) != 3 {
		return nil, fmt.Errorf("expected exactly three keys in envelope, got %d", len(envelope.Keys))
	}

	bundle := &OCRKeyBundle{}

	for _, key := range envelope.Keys {
		keypb, err := decryptKey(key.Data, password)
		if err != nil {
			return nil, err
		}

		if strings.Contains(key.KeyName, ocr2offchain.OCR2OffchainSigning) {
			bundle.OffchainSigningKey = keypb.PrivateKey
		} else if strings.Contains(key.KeyName, ocr2offchain.OCR2OffchainEncryption) {
			bundle.OffchainEncryptionKey = keypb.PrivateKey
		} else if strings.Contains(key.KeyName, PrefixOCR2Onchain) {
			bundle.OnchainSigningKey = keypb.PrivateKey
			// Extract chain type from the key path
			keyPath := keystore.NewKeyPathFromString(key.KeyName)
			bundle.ChainType = ChainType(strings.ToLower(keyPath.Base()))
		}
	}

	return bundle, nil
}
