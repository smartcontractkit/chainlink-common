package evm

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"golang.org/x/crypto/curve25519"
)

type OCR2OffchainKeyringCreateRequest struct {
	Name string
}

type OCR2OffchainKeyringCreateResponse struct {
	Keyring ocrtypes.OffchainKeyring
}

type OCR2OffchainKeyringGetRequest struct {
	Name string
}

type OCR2OffchainKeyringGetResponse struct {
	Keyring ocrtypes.OffchainKeyring
}

type OCR2OffchainKeyringDeleteRequest struct {
	Name string
}

type OCR2OffchainKeyringDeleteResponse struct{}

type OCR2OffchainKeyringListRequest struct{}

type OCR2OffchainKeyringListResponse struct {
	Keyrings []ocrtypes.OffchainKeyring
}

type OCR2OffchainKeyringImportRequest struct {
	Name string
	Data []byte
}

type OCR2OffchainKeyringImportResponse struct {
	Keyring ocrtypes.OffchainKeyring
}

type OCR2OffchainKeyringExportRequest struct {
	Name string
}

type OCR2OffchainKeyringExportResponse struct {
	Data []byte
}

type OCR2OffchainKeyringStore interface {
	CreateKeyring(ctx context.Context, req OCR2OffchainKeyringCreateRequest) (OCR2OffchainKeyringCreateResponse, error)
	GetKeyring(ctx context.Context, req OCR2OffchainKeyringGetRequest) (OCR2OffchainKeyringGetResponse, error)
	DeleteKeyring(ctx context.Context, req OCR2OffchainKeyringDeleteRequest) (OCR2OffchainKeyringDeleteResponse, error)
	ListKeyrings(ctx context.Context, req OCR2OffchainKeyringListRequest) (OCR2OffchainKeyringListResponse, error)
	ImportKeyring(ctx context.Context, req OCR2OffchainKeyringImportRequest) (OCR2OffchainKeyringImportResponse, error)
	ExportKeyring(ctx context.Context, req OCR2OffchainKeyringExportRequest) (OCR2OffchainKeyringExportResponse, error)
}

type ocr2OffchainKeyringStore struct {
	ks keystore.Keystore
}

func NewOCR2OffchainKeyringStore(ks keystore.Keystore) OCR2OffchainKeyringStore {
	return &ocr2OffchainKeyringStore{ks: ks}
}

func (s *ocr2OffchainKeyringStore) buildOffchainKeyName(name string) string {
	return fmt.Sprintf("%s_%s", "evm_ocr2_offchain", name)
}

func (s *ocr2OffchainKeyringStore) buildOffchainEncryptionKeyName(name string) string {
	return fmt.Sprintf("%s_%s", "evm_ocr2_offchain_encryption", name)
}

func (s *ocr2OffchainKeyringStore) CreateKeyring(ctx context.Context, req OCR2OffchainKeyringCreateRequest) (OCR2OffchainKeyringCreateResponse, error) {
	// Create both Ed25519 (for signing) and X25519 (for encryption) keys
	createReq := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				Name:    s.buildOffchainKeyName(req.Name),
				KeyType: keystore.Ed25519,
			},
			{
				Name:    s.buildOffchainEncryptionKeyName(req.Name),
				KeyType: keystore.X25519,
			},
		},
	}
	resp, err := s.ks.CreateKeys(ctx, createReq)
	if err != nil {
		return OCR2OffchainKeyringCreateResponse{}, err
	}
	if len(resp.Keys) == 0 {
		return OCR2OffchainKeyringCreateResponse{}, fmt.Errorf("no keys created")
	}
	if len(resp.Keys) != 2 {
		return OCR2OffchainKeyringCreateResponse{}, fmt.Errorf("expected 2 keys, got %d", len(resp.Keys))
	}

	// Note today loops have a similar thing of returning
	// a connection to another service. Add complexity but manageable.
	return OCR2OffchainKeyringCreateResponse{
		Keyring: &evmOffchainKeyring{
			ks:                    s.ks,
			OffchainKey:           resp.Keys[0].KeyInfo,
			OffchainEncryptionKey: resp.Keys[1].KeyInfo,
		},
	}, nil
}

func (s *ocr2OffchainKeyringStore) GetKeyring(ctx context.Context, req OCR2OffchainKeyringGetRequest) (OCR2OffchainKeyringGetResponse, error) {
	return OCR2OffchainKeyringGetResponse{}, nil
}

func (s *ocr2OffchainKeyringStore) DeleteKeyring(ctx context.Context, req OCR2OffchainKeyringDeleteRequest) (OCR2OffchainKeyringDeleteResponse, error) {
	return OCR2OffchainKeyringDeleteResponse{}, nil
}

func (s *ocr2OffchainKeyringStore) ListKeyrings(ctx context.Context, req OCR2OffchainKeyringListRequest) (OCR2OffchainKeyringListResponse, error) {
	return OCR2OffchainKeyringListResponse{}, nil
}

func (s *ocr2OffchainKeyringStore) ImportKeyring(ctx context.Context, req OCR2OffchainKeyringImportRequest) (OCR2OffchainKeyringImportResponse, error) {
	return OCR2OffchainKeyringImportResponse{}, nil
}

func (s *ocr2OffchainKeyringStore) ExportKeyring(ctx context.Context, req OCR2OffchainKeyringExportRequest) (OCR2OffchainKeyringExportResponse, error) {
	return OCR2OffchainKeyringExportResponse{}, nil
}

var _ ocrtypes.OffchainKeyring = &evmOffchainKeyring{}

type evmOffchainKeyring struct {
	ks                    keystore.Keystore
	OffchainKey           keystore.KeyInfo
	OffchainEncryptionKey keystore.KeyInfo
}

func (k *evmOffchainKeyring) OffchainPublicKey() ocrtypes.OffchainPublicKey {
	var pubKey ocrtypes.OffchainPublicKey
	copy(pubKey[:], k.OffchainKey.PublicKey)
	return pubKey
}

func (k *evmOffchainKeyring) ConfigEncryptionPublicKey() ocrtypes.ConfigEncryptionPublicKey {
	var pubKey ocrtypes.ConfigEncryptionPublicKey
	copy(pubKey[:], k.OffchainEncryptionKey.PublicKey)
	return pubKey
}

func (k *evmOffchainKeyring) OffchainSign(msg []byte) ([]byte, error) {
	signResp, err := k.ks.Sign(context.Background(), keystore.SignRequest{
		Name: k.OffchainKey.Name,
		Data: msg,
	})
	return signResp.Signature, err
}

func (k *evmOffchainKeyring) ConfigDiffieHellman(point [curve25519.PointSize]byte) ([curve25519.PointSize]byte, error) {
	resp, err := k.ks.DeriveSharedSecret(context.Background(), keystore.DeriveSharedSecretRequest{
		LocalKeyName: k.OffchainEncryptionKey.Name,
		RemotePubKey: point[:],
	})
	if err != nil {
		return [curve25519.PointSize]byte{}, err
	}

	var sharedPoint [curve25519.PointSize]byte
	copy(sharedPoint[:], resp.SharedSecret)
	return sharedPoint, nil
}
