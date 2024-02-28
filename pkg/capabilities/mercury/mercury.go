package mercury

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// hex-encoded 32-byte value, prefixed with "0x", all lowercase
type FeedID string

const FeedIDBytesLen = 32

var ErrInvalidFeedID = errors.New("invalid feed ID")

func (id FeedID) String() string {
	return string(id)
}

func (id FeedID) Validate() error {
	if len(id) != 2*FeedIDBytesLen+2 {
		return ErrInvalidFeedID
	}
	if !strings.HasPrefix(string(id), "0x") {
		return ErrInvalidFeedID
	}
	if strings.ToLower(string(id)) != string(id) {
		return ErrInvalidFeedID
	}
	_, err := hex.DecodeString(string(id)[2:])
	return err
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
	FeedID               int64  `json:"feedId"`
	Fullreport           []byte `json:"fullreport"`
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

func (m Codec) Unwrap(raw values.Value) (ReportSet, error) {
	now := uint32(time.Now().Unix())
	return ReportSet{
		Reports: map[FeedID]Report{
			FeedID("0x1111111111111111111100000000000000000000000000000000000000000000"): {
				Info: ReportInfo{
					Timestamp: now,
					Price:     100.00,
				},
			},
			FeedID("0x2222222222222222222200000000000000000000000000000000000000000000"): {
				Info: ReportInfo{
					Timestamp: now,
					Price:     100.00,
				},
			},
			FeedID("0x3333333333333333333300000000000000000000000000000000000000000000"): {
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
			"0x1111111111111111111100000000000000000000000000000000000000000000": map[string]any{
				"timestamp": 42,
				"price":     decimal.NewFromFloat(100.00),
			},
		},
	)
}

func (m Codec) WrapMercuryTriggerEvent(event TriggerEvent) (values.Value, error) {
	return values.Wrap(event)
}

func (m Codec) UnwrapMercuryTriggerEvent(raw values.Value) (TriggerEvent, error) {
	mercuryTriggerEvent := TriggerEvent{}
	val, err := raw.Unwrap()
	if err != nil {
		return mercuryTriggerEvent, err
	}
	event := val.(map[string]any)
	mercuryTriggerEvent.TriggerType = event["triggerType"].(string)
	mercuryTriggerEvent.ID = event["id"].(string)
	mercuryTriggerEvent.Timestamp = event["timestamp"].(string)
	mercuryTriggerEvent.Payload = make([]FeedReport, 0)
	for _, report := range event["payload"].([]any) {
		reportMap := report.(map[string]any)
		decodedData, err := base64.StdEncoding.DecodeString(reportMap["fullreport"].(string))
		if err != nil {
			return TriggerEvent{}, err
		}
		mercuryReport := FeedReport{
			FeedID:               reportMap["feedId"].(int64),
			Fullreport:           decodedData,
			BenchmarkPrice:       reportMap["benchmarkPrice"].(int64),
			ObservationTimestamp: reportMap["observationTimestamp"].(int64),
		}
		mercuryTriggerEvent.Payload = append(mercuryTriggerEvent.Payload, mercuryReport)
	}
	return mercuryTriggerEvent, nil
}

func NewCodec() Codec {
	return Codec{}
}
