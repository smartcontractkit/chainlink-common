package ocr3

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

func Test_config_validate(t *testing.T) {
	type fields struct {
		AggregationMethod string
		AggregationConfig *values.Map
		Encoder           string
		EncoderConfig     *values.Map
		ReportID          string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				ReportID: "1234",
			},
		},
		{
			name: "not hex report id",
			fields: fields{
				ReportID: "123z",
			},
			wantErr: true,
		},
		{
			name: "report id not len 4",
			fields: fields{
				ReportID: "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				AggregationMethod: tt.fields.AggregationMethod,
				AggregationConfig: tt.fields.AggregationConfig,
				Encoder:           tt.fields.Encoder,
				EncoderConfig:     tt.fields.EncoderConfig,
				ReportID:          tt.fields.ReportID,
			}
			if err := c.validate(); (err != nil) != tt.wantErr {
				t.Errorf("config.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
