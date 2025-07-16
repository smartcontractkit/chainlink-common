package report

import (
	"encoding/binary"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// GenerateReportContext generate the report context that is the config digest + the sequence number padded with zeros
func GenerateReportContext(seqNr uint64, configDigest types.ConfigDigest) []byte {
	seqToEpoch := make([]byte, 32)
	binary.BigEndian.PutUint32(seqToEpoch[32-5:32-1], uint32(seqNr)) //nolint:gosec
	zeros := make([]byte, 32)
	repContext := append(append(configDigest[:], seqToEpoch[:]...), zeros...)
	return repContext
}
