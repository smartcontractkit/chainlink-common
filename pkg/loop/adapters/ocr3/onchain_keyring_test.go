package ocr3

import (
	"fmt"
	"reflect"
	"testing"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/require"
)

var _ ocrtypes.OnchainKeyring = (*fakeOnchainKeyring)(nil)

var (
	pubKey             = ocrtypes.OnchainPublicKey("pub-key")
	maxSignatureLength = 12
	sigs               = []byte("some-signatures")
)

type fakeOnchainKeyring struct {
}

func (f fakeOnchainKeyring) PublicKey() ocrtypes.OnchainPublicKey {
	return pubKey
}

func (f fakeOnchainKeyring) Sign(rc ocrtypes.ReportContext, r ocrtypes.Report) (signature []byte, err error) {
	if !reflect.DeepEqual(rc.ConfigDigest, configDigest) {
		return nil, fmt.Errorf("expected configDigest %v but got %v", configDigest, rc.ReportTimestamp.ConfigDigest)
	}

	if rc.Epoch != uint32(seqNr) {
		return nil, fmt.Errorf("expected Epoch %v but got %v", seqNr, rc.Epoch)
	}

	if rc.Round != 0 {
		return nil, fmt.Errorf("expected Round %v but got %v", 0, rc.Round)
	}

	if !reflect.DeepEqual(r, rwi.Report) {
		return nil, fmt.Errorf("expected Report %v but got %v", rwi.Report, r)
	}
	return nil, nil
}

func (f fakeOnchainKeyring) Verify(pk ocrtypes.OnchainPublicKey, rc ocrtypes.ReportContext, r ocrtypes.Report, signature []byte) bool {
	if !reflect.DeepEqual(pk, pubKey) {
		return false
	}

	if !reflect.DeepEqual(rc.ConfigDigest, configDigest) {
		return false
	}

	if rc.Epoch != uint32(seqNr) {
		return false
	}

	if rc.Round != 0 {
		return false
	}

	if !reflect.DeepEqual(r, rwi.Report) {
		return false
	}

	if !reflect.DeepEqual(signature, sigs) {
		return false
	}

	return true
}

func (f fakeOnchainKeyring) MaxSignatureLength() int {
	return maxSignatureLength
}

func TestOnchainKeyring(t *testing.T) {
	kr := NewOnchainKeyring(fakeOnchainKeyring{})

	_, err := kr.Sign(configDigest, seqNr, rwi)
	require.NoError(t, err)
	require.True(t, kr.Verify(pubKey, configDigest, seqNr, rwi, sigs))

	require.Equal(t, pubKey, kr.PublicKey())
	require.Equal(t, maxSignatureLength, kr.MaxSignatureLength())
}
