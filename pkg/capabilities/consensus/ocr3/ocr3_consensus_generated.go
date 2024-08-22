// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package ocr3

import "encoding/json"
import "fmt"
import streams "github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
import "reflect"
import "regexp"

// OCR3 consensus exposed as a capability.
type Consensus struct {
	// Config corresponds to the JSON schema field "config".
	Config ConsensusConfig `json:"config" yaml:"config" mapstructure:"config"`

	// Inputs corresponds to the JSON schema field "inputs".
	Inputs ConsensusInputs `json:"inputs" yaml:"inputs" mapstructure:"inputs"`

	// Outputs corresponds to the JSON schema field "outputs".
	Outputs SignedReport `json:"outputs" yaml:"outputs" mapstructure:"outputs"`
}

type ConsensusConfig struct {
	// AggregationConfig corresponds to the JSON schema field "aggregation_config".
	AggregationConfig ConsensusConfigAggregationConfig `json:"aggregation_config" yaml:"aggregation_config" mapstructure:"aggregation_config"`

	// AggregationMethod corresponds to the JSON schema field "aggregation_method".
	AggregationMethod ConsensusConfigAggregationMethod `json:"aggregation_method" yaml:"aggregation_method" mapstructure:"aggregation_method"`

	// Encoder corresponds to the JSON schema field "encoder".
	Encoder ConsensusConfigEncoder `json:"encoder" yaml:"encoder" mapstructure:"encoder"`

	// EncoderConfig corresponds to the JSON schema field "encoder_config".
	EncoderConfig ConsensusConfigEncoderConfig `json:"encoder_config" yaml:"encoder_config" mapstructure:"encoder_config"`

	// ReportId corresponds to the JSON schema field "report_id".
	ReportId string `json:"report_id" yaml:"report_id" mapstructure:"report_id"`
}

type ConsensusConfigAggregationConfig struct {
	// Allowed partial staleness as a number between 0 and 1.
	AllowedPartialStaleness string `json:"allowedPartialStaleness" yaml:"allowedPartialStaleness" mapstructure:"allowedPartialStaleness"`

	// Feeds corresponds to the JSON schema field "feeds".
	Feeds ConsensusConfigAggregationConfigFeeds `json:"feeds" yaml:"feeds" mapstructure:"feeds"`
}

type ConsensusConfigAggregationConfigFeeds map[string]FeedValue

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConsensusConfigAggregationConfig) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["allowedPartialStaleness"]; raw != nil && !ok {
		return fmt.Errorf("field allowedPartialStaleness in ConsensusConfigAggregationConfig: required")
	}
	if _, ok := raw["feeds"]; raw != nil && !ok {
		return fmt.Errorf("field feeds in ConsensusConfigAggregationConfig: required")
	}
	type Plain ConsensusConfigAggregationConfig
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ConsensusConfigAggregationConfig(plain)
	return nil
}

type ConsensusConfigAggregationMethod string

const ConsensusConfigAggregationMethodDataFeeds ConsensusConfigAggregationMethod = "data_feeds"

var enumValues_ConsensusConfigAggregationMethod = []interface{}{
	"data_feeds",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConsensusConfigAggregationMethod) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConsensusConfigAggregationMethod {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConsensusConfigAggregationMethod, v)
	}
	*j = ConsensusConfigAggregationMethod(v)
	return nil
}

type ConsensusConfigEncoder string

type ConsensusConfigEncoderConfig struct {
	// The ABI for report encoding.
	Abi string `json:"abi" yaml:"abi" mapstructure:"abi"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConsensusConfigEncoderConfig) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["abi"]; raw != nil && !ok {
		return fmt.Errorf("field abi in ConsensusConfigEncoderConfig: required")
	}
	type Plain ConsensusConfigEncoderConfig
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ConsensusConfigEncoderConfig(plain)
	return nil
}

const ConsensusConfigEncoderEVM ConsensusConfigEncoder = "EVM"

var enumValues_ConsensusConfigEncoder = []interface{}{
	"EVM",
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConsensusConfigEncoder) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	var ok bool
	for _, expected := range enumValues_ConsensusConfigEncoder {
		if reflect.DeepEqual(v, expected) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid value (expected one of %#v): %#v", enumValues_ConsensusConfigEncoder, v)
	}
	*j = ConsensusConfigEncoder(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConsensusConfig) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["aggregation_config"]; raw != nil && !ok {
		return fmt.Errorf("field aggregation_config in ConsensusConfig: required")
	}
	if _, ok := raw["aggregation_method"]; raw != nil && !ok {
		return fmt.Errorf("field aggregation_method in ConsensusConfig: required")
	}
	if _, ok := raw["encoder"]; raw != nil && !ok {
		return fmt.Errorf("field encoder in ConsensusConfig: required")
	}
	if _, ok := raw["encoder_config"]; raw != nil && !ok {
		return fmt.Errorf("field encoder_config in ConsensusConfig: required")
	}
	if _, ok := raw["report_id"]; raw != nil && !ok {
		return fmt.Errorf("field report_id in ConsensusConfig: required")
	}
	type Plain ConsensusConfig
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if matched, _ := regexp.MatchString("^[a-f0-9]{4}$", string(plain.ReportId)); !matched {
		return fmt.Errorf("field %s pattern match: must match %s", "^[a-f0-9]{4}$", "ReportId")
	}
	*j = ConsensusConfig(plain)
	return nil
}

type ConsensusInputs struct {
	// Observations corresponds to the JSON schema field "observations".
	Observations [][]streams.Feed `json:"observations" yaml:"observations" mapstructure:"observations"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *ConsensusInputs) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["observations"]; raw != nil && !ok {
		return fmt.Errorf("field observations in ConsensusInputs: required")
	}
	type Plain ConsensusInputs
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = ConsensusInputs(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Consensus) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["config"]; raw != nil && !ok {
		return fmt.Errorf("field config in Consensus: required")
	}
	if _, ok := raw["inputs"]; raw != nil && !ok {
		return fmt.Errorf("field inputs in Consensus: required")
	}
	if _, ok := raw["outputs"]; raw != nil && !ok {
		return fmt.Errorf("field outputs in Consensus: required")
	}
	type Plain Consensus
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Consensus(plain)
	return nil
}

type FeedValue struct {
	// The deviation that is required to generate a new report. Expressed as a
	// percentage. For example, 0.01 is 1% deviation.
	Deviation string `json:"deviation" yaml:"deviation" mapstructure:"deviation"`

	// The interval in seconds after which a new report is generated, regardless of
	// whether any deviations have occurred. New reports reset the timer.
	Heartbeat int `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`

	// An optional remapped ID for the feed.
	RemappedID *string `json:"remappedID,omitempty" yaml:"remappedID,omitempty" mapstructure:"remappedID,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *FeedValue) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["deviation"]; raw != nil && !ok {
		return fmt.Errorf("field deviation in FeedValue: required")
	}
	if _, ok := raw["heartbeat"]; raw != nil && !ok {
		return fmt.Errorf("field heartbeat in FeedValue: required")
	}
	type Plain FeedValue
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if 1 > plain.Heartbeat {
		return fmt.Errorf("field %s: must be >= %v", "heartbeat", 1)
	}
	*j = FeedValue(plain)
	return nil
}

type SignedReport struct {
	// Context corresponds to the JSON schema field "Context".
	Context string `json:"Context" yaml:"Context" mapstructure:"Context"`

	// ID corresponds to the JSON schema field "ID".
	ID string `json:"ID" yaml:"ID" mapstructure:"ID"`

	// Report corresponds to the JSON schema field "Report".
	Report string `json:"Report" yaml:"Report" mapstructure:"Report"`

	// Signatures corresponds to the JSON schema field "Signatures".
	Signatures []string `json:"Signatures,omitempty" yaml:"Signatures,omitempty" mapstructure:"Signatures,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *SignedReport) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["Context"]; raw != nil && !ok {
		return fmt.Errorf("field Context in SignedReport: required")
	}
	if _, ok := raw["ID"]; raw != nil && !ok {
		return fmt.Errorf("field ID in SignedReport: required")
	}
	if _, ok := raw["Report"]; raw != nil && !ok {
		return fmt.Errorf("field Report in SignedReport: required")
	}
	type Plain SignedReport
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = SignedReport(plain)
	return nil
}
