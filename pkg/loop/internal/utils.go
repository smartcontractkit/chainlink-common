package internal

import (
	"math"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func PbReportTimestamp(ts libocr.ReportTimestamp) *pb.ReportTimestamp {
	return &pb.ReportTimestamp{
		ConfigDigest: ts.ConfigDigest[:],
		Epoch:        ts.Epoch,
		Round:        uint32(ts.Round),
	}
}

func ReportTimestamp(ts *pb.ReportTimestamp) (r libocr.ReportTimestamp, err error) {
	if l := len(ts.ConfigDigest); l != 32 {
		err = ErrConfigDigestLen(l)
		return
	}
	copy(r.ConfigDigest[:], ts.ConfigDigest)
	r.Epoch = ts.Epoch
	if ts.Round > math.MaxUint8 {
		err = ErrUint8Bounds{Name: "Round", U: ts.Round}
		return
	}
	r.Round = uint8(ts.Round)
	return
}
