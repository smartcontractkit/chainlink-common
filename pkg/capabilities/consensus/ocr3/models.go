package ocr3

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type config struct {
	AggregationMethod string      `mapstructure:"aggregation_method" json:"aggregation_method" jsonschema:"enum=data_feeds,enum=identical,enum=reduce"`
	AggregationConfig *values.Map `mapstructure:"aggregation_config" json:"aggregation_config"`
	Encoder           string      `mapstructure:"encoder" json:"encoder"`
	EncoderConfig     *values.Map `mapstructure:"encoder_config" json:"encoder_config"`
	ReportID          string      `mapstructure:"report_id" json:"report_id" jsonschema:"required,pattern=^[a-f0-9]{4}$"`
	RequestTimeoutMS  int64       `mapstructure:"request_timeout_ms" json:"request_timeout_ms"`

	KeyID string `mapstructure:"key_id" json:"key_id,omitempty" jsonschema:"required"`
}

type inputs struct {
	Observations  *values.List `json:"observations" jsonschema:""`
	EncoderName   string       `mapstructure:"encoder" json:"encoder,omitempty"`
	EncoderConfig *values.Map  `mapstructure:"encoder_config" json:"encoder_config,omitempty"`
}
