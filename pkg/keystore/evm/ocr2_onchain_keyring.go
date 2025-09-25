package evm

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	evmutil "github.com/smartcontractkit/libocr/offchainreporting2plus/chains/evmutil"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type OCR2OnchainKeyringCreateRequest struct {
	Name string
}

type OCR2OnchainKeyringCreateResponse struct {
	Keyring ocrtypes.OnchainKeyring
}

type OCR2OnchainKeyringGetRequest struct {
	Name string
}

type OCR2OnchainKeyringGetResponse struct {
	Keyring ocrtypes.OnchainKeyring
}

type OCR2OnchainKeyringDeleteRequest struct {
	Name string
}

type OCR2OnchainKeyringDeleteResponse struct{}

type OCR2OnchainKeyringListRequest struct{}

type OCR2OnchainKeyringListResponse struct {
	Keyrings []ocrtypes.OnchainKeyring
}

type OCR2OnchainKeyringImportRequest struct {
	Name string
	Data []byte
}

type OCR2OnchainKeyringImportResponse struct {
	Keyring ocrtypes.OnchainKeyring
}

type OCR2OnchainKeyringExportRequest struct {
	Name string
}

type OCR2OnchainKeyringExportResponse struct {
	Data []byte
}

type OCR2OnchainKeyringStore interface {
	CreateKeyring(ctx context.Context, req OCR2OnchainKeyringCreateRequest) (OCR2OnchainKeyringCreateResponse, error)
	GetKeyring(ctx context.Context, req OCR2OnchainKeyringGetRequest) (OCR2OnchainKeyringGetResponse, error)
	DeleteKeyring(ctx context.Context, req OCR2OnchainKeyringDeleteRequest) (OCR2OnchainKeyringDeleteResponse, error)
	ListKeyrings(ctx context.Context, req OCR2OnchainKeyringListRequest) (OCR2OnchainKeyringListResponse, error)
	ImportKeyring(ctx context.Context, req OCR2OnchainKeyringImportRequest) (OCR2OnchainKeyringImportResponse, error)
	ExportKeyring(ctx context.Context, req OCR2OnchainKeyringExportRequest) (OCR2OnchainKeyringExportResponse, error)
}

type ocr2OnchainKeyringStore struct {
	ks keystore.Keystore
}

func NewOCR2OnchainKeyringStore(ks keystore.Keystore) OCR2OnchainKeyringStore {
	return &ocr2OnchainKeyringStore{ks: ks}
}

func (s *ocr2OnchainKeyringStore) buildKeyName(name string) string {
	return fmt.Sprintf("%s_%s", "evm_ocr2_onchain", name)
}

func (s *ocr2OnchainKeyringStore) CreateKeyring(ctx context.Context, req OCR2OnchainKeyringCreateRequest) (OCR2OnchainKeyringCreateResponse, error) {
	createReq := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				Name:    s.buildKeyName(req.Name),
				KeyType: keystore.Secp256k1,
			},
		},
	}
	resp, err := s.ks.CreateKeys(ctx, createReq)
	if err != nil {
		return OCR2OnchainKeyringCreateResponse{}, err
	}
	if len(resp.Keys) != 1 {
		return OCR2OnchainKeyringCreateResponse{}, fmt.Errorf("expected 1 key, got %d", len(resp.Keys))
	}

	return OCR2OnchainKeyringCreateResponse{
		Keyring: &evmOnchainKeyring{
			ks:         s.ks,
			OnchainKey: resp.Keys[0].KeyInfo,
		},
	}, nil
}

func (s *ocr2OnchainKeyringStore) GetKeyring(ctx context.Context, req OCR2OnchainKeyringGetRequest) (OCR2OnchainKeyringGetResponse, error) {
	return OCR2OnchainKeyringGetResponse{}, nil
}

func (s *ocr2OnchainKeyringStore) DeleteKeyring(ctx context.Context, req OCR2OnchainKeyringDeleteRequest) (OCR2OnchainKeyringDeleteResponse, error) {
	return OCR2OnchainKeyringDeleteResponse{}, nil
}

func (s *ocr2OnchainKeyringStore) ListKeyrings(ctx context.Context, req OCR2OnchainKeyringListRequest) (OCR2OnchainKeyringListResponse, error) {
	return OCR2OnchainKeyringListResponse{}, nil
}

func (s *ocr2OnchainKeyringStore) ImportKeyring(ctx context.Context, req OCR2OnchainKeyringImportRequest) (OCR2OnchainKeyringImportResponse, error) {
	return OCR2OnchainKeyringImportResponse{}, nil
}

func (s *ocr2OnchainKeyringStore) ExportKeyring(ctx context.Context, req OCR2OnchainKeyringExportRequest) (OCR2OnchainKeyringExportResponse, error) {
	return OCR2OnchainKeyringExportResponse{}, nil
}

var _ ocrtypes.OnchainKeyring = &evmOnchainKeyring{}

type evmOnchainKeyring struct {
	ks         keystore.Keystore
	OnchainKey keystore.KeyInfo
}

func (k *evmOnchainKeyring) PublicKey() ocrtypes.OnchainPublicKey {
	// XXX: PublicKey returns the address of the public key not the public key itself
	return nil
}

func ReportToSigData(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) []byte {
	rawReportContext := evmutil.RawReportContext(reportCtx)
	sigData := crypto.Keccak256(report)
	sigData = append(sigData, rawReportContext[0][:]...)
	sigData = append(sigData, rawReportContext[1][:]...)
	sigData = append(sigData, rawReportContext[2][:]...)
	return crypto.Keccak256(sigData)
}

func (k *evmOnchainKeyring) Sign(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) ([]byte, error) {
	signResp, err := k.ks.Sign(context.Background(), keystore.SignRequest{
		Name: k.OnchainKey.Name,
		Data: ReportToSigData(reportCtx, report),
	})
	return signResp.Signature, err
}

func (k *evmOnchainKeyring) Verify(publicKey ocrtypes.OnchainPublicKey, reportCtx ocrtypes.ReportContext, report ocrtypes.Report, signature []byte) bool {
	verifyResp, err := k.ks.Verify(context.Background(), keystore.VerifyRequest{
		Name:      k.OnchainKey.Name,
		Data:      ReportToSigData(reportCtx, report),
		Signature: signature,
	})
	if err != nil {
		// Log?
		return false
	}
	return verifyResp.Valid
}

func (k *evmOnchainKeyring) MaxSignatureLength() int {
	return 65
}
