package resourcemanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOwnerLabel(t *testing.T) {
	tests := []struct {
		name  string
		owner string
		want  string
	}{
		{name: "lowercase 0x prefix stripped", owner: "0xabc123def", want: "abc123def"},
		{name: "uppercase 0X prefix stripped", owner: "0Xabc123def", want: "abc123def"},
		{name: "no prefix unchanged", owner: "abc123def", want: "abc123def"},
		{name: "mixed case lowercased", owner: "0xAbC123dEF", want: "abc123def"},
		{name: "mixed case without prefix lowercased", owner: "AbC123dEF", want: "abc123def"},
		{name: "empty string", owner: "", want: ""},
		{name: "prefix only", owner: "0x", want: ""},
		{name: "only one prefix stripped", owner: "0x0Xabc", want: "0xabc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, OwnerLabel(tt.owner))
		})
	}
}
