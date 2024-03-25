package mercury

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

var (
	feedOne   = Must(FromFeedIDString("0x1111111111111111111100000000000000000000000000000000000000000000"))
	feedTwo   = Must(FromFeedIDString("0x2222222222222222222200000000000000000000000000000000000000000000"))
	feedThree = Must(FromFeedIDString("0x3333333333333333333300000000000000000000000000000000000000000000"))
)

// hex-encoded 32-byte value, prefixed with "0x", all lowercase
type FeedID [32]byte

const FeedIDBytesLen = 32

var ErrInvalidFeedID = errors.New("invalid feed ID")

func (id FeedID) String() string {
	return "0x" + hex.EncodeToString(id[:])
}

func FromFeedIDString(s string) (FeedID, error) {
	if !strings.HasPrefix(s, "0x") {
		return FeedID{}, ErrInvalidFeedID
	}

	if len(s) != 2*FeedIDBytesLen+2 {
		return FeedID{}, ErrInvalidFeedID
	}

	b, err := hex.DecodeString(s[2:])
	return FeedID(b), err
}

func Must[O any](o O, err error) O {
	if err != nil {
		panic(err)
	}

	return o
}

type ReportSet struct {
	// feedID -> report
	Reports map[FeedID]Report
}

type Report struct {
	Info       ReportInfo // minimal data extracted from the report for convenience
	FullReport []byte     // full report, acceptable by the verifier contract
}

type ReportInfo struct {
	Timestamp uint32
	Price     float64
}

// TODO: fix this by adding support for uint64 in value.go
type FeedReport struct {
	FeedID               FeedID `json:"feedID"`
	FullReport           []byte `json:"fullreport"`
	BenchmarkPrice       int64  `json:"benchmarkPrice"`
	ObservationTimestamp int64  `json:"observationTimestamp"`
}

type TriggerEvent struct {
	TriggerType string       `json:"triggerType"`
	ID          string       `json:"id"`
	Timestamp   string       `json:"timestamp"`
	Payload     []FeedReport `json:"payload"`
}

// TODO implement an actual codec
type Codec struct {
}

// What do with this?
func (m Codec) Unwrap(raw values.Value) (ReportSet, error) {
	now := uint32(time.Now().Unix())
	return ReportSet{
		Reports: map[FeedID]Report{
			feedOne: {
				Info: ReportInfo{
					Timestamp: now,
					Price:     100.00,
				},
			},
			feedTwo: {
				Info: ReportInfo{
					Timestamp: now,
					Price:     100.00,
				},
			},
			feedThree: {
				Info: ReportInfo{
					Timestamp: now,
					Price:     100.00,
				},
			},
		},
	}, nil
}

func (m Codec) Wrap(reportSet ReportSet) (values.Value, error) {
	return values.NewMap(
		map[string]any{
			feedOne.String(): map[string]any{
				"timestamp": 42,
				"price":     decimal.NewFromFloat(100.00),
			},
		},
	)
}

type feedReport struct {
	FeedID               string `json:"feedId"`
	FullReport           []byte `json:"fullreport"`
	BenchmarkPrice       int64  `json:"benchmarkPrice"`
	ObservationTimestamp int64  `json:"observationTimestamp"`
}

// triggerEvent is a values-compatible representation of `TriggerEvent`
// which stores the `feedID` as a string rather than a [32]byte.
type triggerEvent struct {
	TriggerType string       `json:"triggerType"`
	ID          string       `json:"id"`
	Timestamp   string       `json:"timestamp"`
	Payload     []feedReport `json:"payload"`
}

func (m Codec) WrapMercuryTriggerEvent(event TriggerEvent) (values.Value, error) {
	te := &triggerEvent{
		TriggerType: event.TriggerType,
		ID:          event.ID,
		Timestamp:   event.Timestamp,
		Payload:     []feedReport{},
	}

	for _, p := range event.Payload {
		p := feedReport{
			FeedID:               p.FeedID.String(),
			FullReport:           p.FullReport,
			BenchmarkPrice:       p.BenchmarkPrice,
			ObservationTimestamp: p.ObservationTimestamp,
		}

		te.Payload = append(te.Payload, p)
	}

	return values.Wrap(te)
}

func (m Codec) UnwrapMercuryTriggerEvent(raw values.Value) (TriggerEvent, error) {
	mercuryTriggerEvent := TriggerEvent{}
	val, err := raw.Unwrap()
	if err != nil {
		return mercuryTriggerEvent, err
	}
	event := val.(map[string]any)
	mercuryTriggerEvent.TriggerType = event["TriggerType"].(string)
	mercuryTriggerEvent.ID = event["ID"].(string)
	mercuryTriggerEvent.Timestamp = event["Timestamp"].(string)
	mercuryTriggerEvent.Payload = make([]FeedReport, 0)
	for _, report := range event["Payload"].([]any) {
		reportMap := report.(map[string]any)
		var mercuryReport feedReport
		err = mapstructure.Decode(reportMap, &mercuryReport)
		if err != nil {
			return mercuryTriggerEvent, err
		}

		fid, err := FromFeedIDString(mercuryReport.FeedID)
		if err != nil {
			return mercuryTriggerEvent, err
		}

		mr := FeedReport{
			FeedID:               fid,
			FullReport:           mercuryReport.FullReport,
			ObservationTimestamp: mercuryReport.ObservationTimestamp,
			BenchmarkPrice:       mercuryReport.BenchmarkPrice,
		}

		mercuryTriggerEvent.Payload = append(mercuryTriggerEvent.Payload, mr)
	}
	return mercuryTriggerEvent, nil
}

func NewCodec() Codec {
	return Codec{}
}
