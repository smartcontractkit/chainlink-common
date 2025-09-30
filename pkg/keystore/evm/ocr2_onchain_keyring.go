package evm

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/chainlink-common/pkg/keystore"
	evmutil "github.com/smartcontractkit/libocr/offchainreporting2plus/chains/evmutil"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const (
	OCR2OnchainPrefix = "ocr2_onchain"
)

// GetOCR2OnchainKeystoreName builds EVM/OCR2_ONCHAIN local name into a fully-qualified keystore name.
func GetOCR2OnchainKeystoreName(localName string) string {
	return keystore.JoinKeySegments(EVM_PREFIX, OCR2OnchainPrefix, localName)
}

// IsOCR2OnchainKey checks if a keystore name belongs to the OCR2 onchain namespace.
func IsOCR2OnchainKey(name string) bool {
	return strings.HasPrefix(name, keystore.JoinKeySegments(EVM_PREFIX, OCR2OnchainPrefix, ""))
}

type OCR2OnchainKeyringCreateRequest struct {
	LocalName string
}

type OCR2OnchainKeyringCreateResponse struct {
	Keyring ocrtypes.OnchainKeyring
}

type OCR2OnchainKeyringGetKeyringsRequest struct {
	Names []string // Empty slice means get all OCR2 onchain keyrings
}

type OCR2OnchainKeyringGetKeyringsResponse struct {
	Keyrings []ocrtypes.OnchainKeyring
}

// CreateOCR2OnchainKeyring creates an OCR2 onchain keyring using the base keystore and returns the handle.
func CreateOCR2OnchainKeyring(ctx context.Context, ks keystore.Keystore, localName string) (ocrtypes.OnchainKeyring, error) {
	createReq := keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{
				Name:    GetOCR2OnchainKeystoreName(localName),
				KeyType: keystore.Secp256k1,
			},
		},
	}
	resp, err := ks.CreateKeys(ctx, createReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Keys) != 1 {
		return nil, fmt.Errorf("expected 1 key, got %d", len(resp.Keys))
	}
	return &evmOnchainKeyring{ks: ks, OnchainKey: resp.Keys[0].KeyInfo}, nil
}

// ListOCR2OnchainKeyrings lists OCR2 onchain keyrings. If no local names provided, returns all OCR2 onchain keyrings.
func ListOCR2OnchainKeyrings(ctx context.Context, ks keystore.Keystore, localNames ...string) ([]ocrtypes.OnchainKeyring, error) {
	// Build names if explicitly provided
	var names []string
	if len(localNames) > 0 {
		for _, ln := range localNames {
			names = append(names, GetOCR2OnchainKeystoreName(ln))
		}
	}

	getReq := keystore.GetKeysRequest{Names: names}
	resp, err := ks.GetKeys(ctx, getReq)
	if err != nil {
		return nil, err
	}

	var keyrings []ocrtypes.OnchainKeyring
	for _, key := range resp.Keys {
		if IsOCR2OnchainKey(key.Name) {
			keyrings = append(keyrings, &evmOnchainKeyring{ks: ks, OnchainKey: key})
		}
	}
	return keyrings, nil
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
