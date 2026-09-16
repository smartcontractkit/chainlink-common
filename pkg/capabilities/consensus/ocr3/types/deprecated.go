package types

import (
	"context"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// The OCR3 consensus capability that produced and consumed these types has been
// removed from chainlink-common. They are retained only so that existing
// encoder implementations keep compiling.

// MetadataFieldName was the key under which consensus metadata was stored in an
// aggregation outcome.
//
// Deprecated: the OCR3 consensus capability no longer produces these outcomes.
const MetadataFieldName = "INTERNAL_METADATA"

// Encoder encoded an aggregation outcome into the bytes to be signed.
//
// Deprecated: the OCR3 consensus capability no longer invokes encoders.
type Encoder interface {
	Encode(ctx context.Context, input values.Map) ([]byte, error)
}

// SignedReport was the consensus capability's response to a workflow.
//
// Deprecated: the OCR3 consensus capability no longer produces signed reports.
type SignedReport struct {
	Report []byte
	// Report context is appended to the report before signing by libOCR.
	// It contains config digest + round/epoch/sequence numbers (currently 96 bytes).
	// Has to be appended to the report before validating signatures.
	Context []byte
	// Always exactly F+1 signatures.
	Signatures [][]byte
	// Report ID defined in the workflow spec (2 bytes).
	ID []byte
}
