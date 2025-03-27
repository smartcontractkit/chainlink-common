package datastreams

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// hex-encoded 32-byte value, prefixed with "0x", all lowercase
type FeedID string

const FeedIDBytesLen = 32

var ErrInvalidFeedID = errors.New("invalid feed ID")

func (id FeedID) String() string {
	return string(id)
}

// Bytes() converts the FeedID string into a [32]byte
// value.
// Note: this function panics if the underlying
// string isn't of the right length. For production (i.e.)
// non-test uses, please create the FeedID via the NewFeedID
// constructor, which will validate the string.
func (id FeedID) Bytes() [FeedIDBytesLen]byte {
	b, _ := hex.DecodeString(string(id)[2:])
	return [FeedIDBytesLen]byte(b)
}

func (id FeedID) validate() error {
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

func NewFeedID(s string) (FeedID, error) {
	id := FeedID(s)
	return id, id.validate()
}

func FeedIDFromBytes(b [FeedIDBytesLen]byte) FeedID {
	return FeedID("0x" + hex.EncodeToString(b[:]))
}

type FeedReport struct {
	FeedID        string
	FullReport    []byte
	ReportContext []byte
	Signatures    [][]byte

	// Fields below are derived from FullReport
	// NOTE: BenchmarkPrice is a byte representation of big.Int. We can't use big.Int
	// directly due to Value serialization problems using mapstructure.
	BenchmarkPrice       []byte
	ObservationTimestamp int64
}

// passed alongside Streams trigger events
type Metadata struct {
	Signers               [][]byte
	MinRequiredSignatures int
}

// StreamsTriggerEvent is the underlying type passed to the dataFeedsAggregator.Aggregate
// function via the untyped observation, which originates in the asset don.
type StreamsTriggerEvent struct {
	Payload   []FeedReport
	Metadata  Metadata
	Timestamp int64
}

// LLOStreamsTriggerEvent is the underlying type passed to the LLOAggregator.Aggregate
// function via the untyped observation, which originates on the asset don via the LLO OCR3 plugin.
type LLOStreamsTriggerEvent struct {
	Payload                         []*LLOStreamDecimal
	ObservationTimestampNanoseconds uint64
}

type LLOStreamDecimal struct {
	StreamID uint32
	Decimal  []byte // binary representation of [llo.Decimal]: https://github.com/smartcontractkit/chainlink-data-streams/blob/d33e95631485bbcfdc22d209875035e3c73199d0/llo/stream_value.go#L147
	// future: may add aggregation type {MODE, MEDIAN, etc...}
}

type ReportCodec interface {
	// unwrap StreamsTriggerEvent and convert to a list of FeedReport
	Unwrap(wrapped values.Value) ([]FeedReport, error)

	// wrap a list of FeedReport to a wrapped StreamsTriggerEvent Value
	Wrap(reports []FeedReport) (values.Value, error)

	// validate signatures on a single FeedReport
	Validate(feedReport FeedReport, allowedSigners [][]byte, minRequiredSignatures int) error
}

// Helpers for unwrapping a StreamsTriggerPayload into a []FeedReport - more efficient than using mapstructure/reflection
func UnwrapStreamsTriggerEventToFeedReportList(wrapped values.Value) ([]FeedReport, error) {
	result := []FeedReport{}
	triggerEvent, ok := wrapped.(*values.Map)
	if !ok {
		return nil, fmt.Errorf("unexpected value %+v for trigger payload: expected map, got %T", wrapped, wrapped)
	}

	p, ok := triggerEvent.Underlying["Payload"]
	if !ok {
		return nil, errors.New("expected map to have Payload field")
	}

	plst, ok := p.(*values.List)
	if !ok {
		return nil, errors.New("expected Payload to be a list")
	}
	for _, v := range plst.Underlying {
		report := FeedReport{}
		mp, ok := v.(*values.Map)
		if !ok {
			return nil, fmt.Errorf("unexpected value %+v for feed report: expected map, got %T", v, v)
		}
		var err error
		report.FeedID, err = getStringField(mp, "FeedID")
		if err != nil {
			return nil, err
		}
		report.FullReport, err = getBytesField(mp, "FullReport")
		if err != nil {
			return nil, err
		}
		report.ReportContext, err = getBytesField(mp, "ReportContext")
		if err != nil {
			return nil, err
		}
		sigListVal, ok := mp.Underlying["Signatures"]
		if !ok {
			return nil, errors.New("missing Signatures key")
		}
		sigList, ok := sigListVal.(*values.List)
		if !ok {
			return nil, errors.New("expected list type for Signatures")
		}
		for idx, sig := range sigList.Underlying {
			sigVal, ok := sig.(*values.Bytes)
			if !ok {
				return nil, fmt.Errorf("expected bytes type for signature %d", idx)
			}
			report.Signatures = append(report.Signatures, sigVal.Underlying)
		}
		result = append(result, report)
	}

	return result, nil
}

func getStringField(mp *values.Map, key string) (string, error) {
	val, ok := mp.Underlying[key]
	if !ok {
		return "", fmt.Errorf("missing key %s", key)
	}
	strVal, ok := val.(*values.String)
	if !ok {
		return "", fmt.Errorf("expected string type for key %s", key)
	}
	return strVal.Underlying, nil
}

func getBytesField(mp *values.Map, key string) ([]byte, error) {
	val, ok := mp.Underlying[key]
	if !ok {
		return nil, fmt.Errorf("missing key %s", key)
	}
	byleVal, ok := val.(*values.Bytes)
	if !ok {
		return nil, fmt.Errorf("expected bytes type for key %s", key)
	}
	return byleVal.Underlying, nil
}
