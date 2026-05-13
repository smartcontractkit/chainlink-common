package stellar_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	stellarcap "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/stellar"
	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestConvertGetLatestLedgerResponseRoundtrip(t *testing.T) {
	t.Parallel()

	rawHash := []byte{0xde, 0xad, 0xbe, 0xef, 0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13,
		0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b}
	rawHeader := []byte{0xaa, 0xbb, 0xcc}
	rawMeta := []byte{0x11, 0x22, 0x33, 0x44}

	protoResp := &stellarcap.GetLatestLedgerResponse{
		Hash:              rawHash,
		ProtocolVersion:   22,
		Sequence:          987654,
		LedgerCloseTime:   1_700_000_000,
		LedgerHeaderXdr:   rawHeader,
		LedgerMetadataXdr: rawMeta,
	}

	domain, err := stellarcap.ConvertGetLatestLedgerResponseFromProto(protoResp)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(rawHash), domain.Hash)
	require.Equal(t, uint32(22), domain.ProtocolVersion)
	require.Equal(t, uint32(987654), domain.Sequence)
	require.Equal(t, int64(1_700_000_000), domain.LedgerCloseTime)
	require.Equal(t, base64.StdEncoding.EncodeToString(rawHeader), domain.LedgerHeaderXDR)
	require.Equal(t, base64.StdEncoding.EncodeToString(rawMeta), domain.LedgerMetadataXDR)

	// Round-trip back to proto.
	roundTripped, err := stellarcap.ConvertGetLatestLedgerResponseToProto(domain)
	require.NoError(t, err)
	require.Equal(t, rawHash, roundTripped.Hash)
	require.Equal(t, uint32(22), roundTripped.ProtocolVersion)
	require.Equal(t, uint32(987654), roundTripped.Sequence)
	require.Equal(t, int64(1_700_000_000), roundTripped.LedgerCloseTime)
	require.Equal(t, rawHeader, roundTripped.LedgerHeaderXdr)
	require.Equal(t, rawMeta, roundTripped.LedgerMetadataXdr)
}

func TestConvertGetLatestLedgerResponseFromProto_RejectsNil(t *testing.T) {
	t.Parallel()

	_, err := stellarcap.ConvertGetLatestLedgerResponseFromProto(nil)
	require.ErrorContains(t, err, "getLatestLedgerResponse is nil")
}

func TestConvertGetLatestLedgerResponseFromProto_EmptyFieldsReturnZeroValues(t *testing.T) {
	t.Parallel()

	domain, err := stellarcap.ConvertGetLatestLedgerResponseFromProto(&stellarcap.GetLatestLedgerResponse{})
	require.NoError(t, err)
	require.Equal(t, "", domain.Hash) // hex.EncodeToString(nil) == ""
	require.Equal(t, base64.StdEncoding.EncodeToString(nil), domain.LedgerHeaderXDR)
}

func TestConvertGetLatestLedgerResponseToProto_RejectsInvalidHexHash(t *testing.T) {
	t.Parallel()

	_, err := stellarcap.ConvertGetLatestLedgerResponseToProto(stellartypes.GetLatestLedgerResponse{
		Hash: "not-hex!",
	})
	require.ErrorContains(t, err, "invalid hex hash")
}

func TestConvertGetLatestLedgerResponseToProto_RejectsInvalidBase64Header(t *testing.T) {
	t.Parallel()

	_, err := stellarcap.ConvertGetLatestLedgerResponseToProto(stellartypes.GetLatestLedgerResponse{
		Hash:            "",
		LedgerHeaderXDR: "not-valid-base64!!!",
	})
	require.ErrorContains(t, err, "invalid base64 ledger_header_xdr")
}

func TestConvertGetLatestLedgerResponseToProto_RejectsInvalidBase64Metadata(t *testing.T) {
	t.Parallel()

	_, err := stellarcap.ConvertGetLatestLedgerResponseToProto(stellartypes.GetLatestLedgerResponse{
		Hash:              "",
		LedgerHeaderXDR:   "",
		LedgerMetadataXDR: "not-valid-base64!!!",
	})
	require.ErrorContains(t, err, "invalid base64 ledger_metadata_xdr")
}

func TestValidateReadContractRequest_AcceptsValidRequest(t *testing.T) {
	t.Parallel()

	err := stellarcap.ValidateReadContractRequest(&stellarcap.ReadContractRequest{
		ContractId: "CAHJJJKK7777AAAA1111BBBB2222CCCC3333DDDD4444EEEE5555FFFF6666",
		Function:   "balance",
	})
	require.NoError(t, err)
}

func TestValidateReadContractRequest_RejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		req     *stellarcap.ReadContractRequest
		wantErr string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: "readContractRequest is nil",
		},
		{
			name:    "empty contract_id",
			req:     &stellarcap.ReadContractRequest{Function: "balance"},
			wantErr: "contract_id is required",
		},
		{
			name:    "empty function",
			req:     &stellarcap.ReadContractRequest{ContractId: "CABC"},
			wantErr: "function is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := stellarcap.ValidateReadContractRequest(tc.req)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestValidateReadContractRequest_ZeroLedgerSequenceIsAllowed(t *testing.T) {
	t.Parallel()

	// ledger_sequence == 0 means "use latest" and is explicitly allowed.
	err := stellarcap.ValidateReadContractRequest(&stellarcap.ReadContractRequest{
		ContractId:     "CABC",
		Function:       "my_fn",
		LedgerSequence: 0,
	})
	require.NoError(t, err)
}

func TestValidateWriteReportRequest_AcceptsValidRequest(t *testing.T) {
	t.Parallel()

	err := stellarcap.ValidateWriteReportRequest(&stellarcap.WriteReportRequest{
		ContractId: "CAHJJJKK7777AAAA1111BBBB2222CCCC3333DDDD4444EEEE5555FFFF6666",
		Report:     &sdkpb.ReportResponse{},
	})
	require.NoError(t, err)
}

func TestValidateWriteReportRequest_RejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		req     *stellarcap.WriteReportRequest
		wantErr string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: "writeReportRequest is nil",
		},
		{
			name:    "empty contract_id",
			req:     &stellarcap.WriteReportRequest{Report: &sdkpb.ReportResponse{}},
			wantErr: "contract_id is required",
		},
		{
			name:    "nil report",
			req:     &stellarcap.WriteReportRequest{ContractId: "CABC"},
			wantErr: "report is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := stellarcap.ValidateWriteReportRequest(tc.req)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
