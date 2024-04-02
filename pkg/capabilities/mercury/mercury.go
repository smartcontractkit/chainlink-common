package mercury

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
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

// TODO: fix this by adding support for uint64 in value.go
type FeedReport struct {
	FeedID               int64  `json:"feedId"`
	FullReport           []byte `json:"fullreport"`
	BenchmarkPrice       int64  `json:"benchmarkPrice"`
	ObservationTimestamp int64  `json:"observationTimestamp"`
}

// TODO implement an actual codec
type Codec struct {
}

func (m Codec) WrapMercuryTriggerEvent(event capabilities.TriggerEvent) (values.Value, error) {
	return values.Wrap(event)
}

func (m Codec) UnwrapMercuryTriggerEvent(raw values.Value) (capabilities.TriggerEvent, error) {
	mercuryTriggerEvent := capabilities.TriggerEvent{}
	val, err := raw.Unwrap()
	if err != nil {
		return mercuryTriggerEvent, err
	}
	event := val.(map[string]any)
	mercuryTriggerEvent.TriggerType = event["TriggerType"].(string)
	mercuryTriggerEvent.ID = event["ID"].(string)
	mercuryTriggerEvent.Timestamp = event["Timestamp"].(string)
	mercuryTriggerEvent.BatchedPayload = make(map[string]any)
	for id, report := range event["BatchedPayload"].(map[string]any) {
		reportMap := report.(map[string]any)
		var mercuryReport FeedReport
		err = mapstructure.Decode(reportMap, &mercuryReport)
		if err != nil {
			return mercuryTriggerEvent, err
		}
		mercuryTriggerEvent.BatchedPayload[id] = mercuryReport
	}
	return mercuryTriggerEvent, nil
}

func NewCodec() Codec {
	return Codec{}
}
