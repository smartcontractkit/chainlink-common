package ocr2key

import (
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// OCR3ReportContext expresses an OCR3 round as the ReportContext an OCR2 key
// bundle signs: the sequence number as the epoch, no round, and no extra hash.
//
// A key bundle signs a ReportContext and knows nothing about sequence numbers, so
// every OCR3 caller has to make this translation. It is one function rather than
// each caller's own literal because the result is a signing preimage: two callers
// that disagree about it produce signatures the other rejects, and the failure
// surfaces as a DON that will not come to consensus rather than as a mismatch
// anyone can see.
func OCR3ReportContext(digest ocrtypes.ConfigDigest, seqNr uint64) ocrtypes.ReportContext {
	return ocrtypes.ReportContext{
		ReportTimestamp: ocrtypes.ReportTimestamp{
			ConfigDigest: digest,
			//nolint:gosec // G115: truncation is part of the established preimage.
			Epoch: uint32(seqNr),
			Round: 0,
		},
		ExtraHash: [32]byte{},
	}
}
