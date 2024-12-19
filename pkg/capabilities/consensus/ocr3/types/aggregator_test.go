package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata_padWorkflowName(t *testing.T) {
	type fields struct {
		WorkflowName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "padWorkflowName hex with 9 bytes",
			fields: fields{
				WorkflowName: "ABCD1234EF567890AB",
			},
			want: "abcd1234ef567890ab00",
		},
		{
			name: "padWorkflowName hex with 5 bytes",
			fields: fields{
				WorkflowName: "1234ABCD56",
			},
			want: "1234abcd560000000000",
		},
		{
			name: "padWorkflowName empty",
			fields: fields{
				WorkflowName: "",
			},
			want: "00000000000000000000",
		},
		{
			name: "padWorkflowName non-hex string",
			fields: fields{
				WorkflowName: "not-hex",
			},
			want: "not-hex   ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metadata{
				WorkflowName: tt.fields.WorkflowName,
			}
			m.padWorkflowName()
			assert.Equal(t, tt.want, m.WorkflowName, tt.name)
		})
	}
}
