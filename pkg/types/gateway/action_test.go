package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutboundHTTPRequest_Hash(t *testing.T) {
	baseWithHeaders := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/",
		WorkflowOwner: "owner",
		Headers:       map[string]string{"A": "1"},
	}
	baseWithMultiHeaders := OutboundHTTPRequest{
		Method:       "GET",
		URL:          "https://example.com/",
		WorkflowOwner: "owner",
		MultiHeaders:  map[string][]string{"A": {"1", "2"}},
	}

	tests := []struct {
		name     string
		reqA     OutboundHTTPRequest
		reqB     OutboundHTTPRequest
		sameHash bool
	}{
		{
			name: "Headers only same content different map order",
			reqA: OutboundHTTPRequest{
				Method:        "GET",
				URL:           "https://example.com/api",
				WorkflowOwner: "owner-1",
				Body:          []byte(`{"a":1}`),
				Headers:       map[string]string{"Accept": "application/json", "Content-Type": "application/json", "X-Request-Id": "req-123"},
			},
			reqB: OutboundHTTPRequest{
				Method:        "GET",
				URL:           "https://example.com/api",
				WorkflowOwner: "owner-1",
				Body:          []byte(`{"a":1}`),
				Headers:       map[string]string{"X-Request-Id": "req-123", "Content-Type": "application/json", "Accept": "application/json"},
			},
			sameHash: true,
		},
		{
			name: "MultiHeaders only same content different key and value order",
			reqA: OutboundHTTPRequest{
				Method:        "GET",
				URL:           "https://example.com/api",
				WorkflowOwner: "owner-1",
				Body:          []byte(`{"a":1}`),
				MultiHeaders:  map[string][]string{"Accept": {"application/json"}, "Content-Type": {"application/json"}, "Set-Cookie": {"s1=abc", "s2=def"}},
			},
			reqB: OutboundHTTPRequest{
				Method:        "GET",
				URL:           "https://example.com/api",
				WorkflowOwner: "owner-1",
				Body:          []byte(`{"a":1}`),
				MultiHeaders:  map[string][]string{"Set-Cookie": {"s2=def", "s1=abc"}, "Content-Type": {"application/json"}, "Accept": {"application/json"}},
			},
			sameHash: true,
		},
		{
			// Headers is the comma-joined version of MultiHeaders; same data, different value order in MultiHeaders.
			name: "Headers and MultiHeaders both present same content",
			reqA: OutboundHTTPRequest{
				Method:       "GET",
				URL:          "https://example.com/api",
				WorkflowOwner: "owner-1",
				Body:         []byte(`{"a":1}`),
				Headers:      map[string]string{"X": "1", "Y": "a,b"},
				MultiHeaders: map[string][]string{"X": {"1"}, "Y": {"a", "b"}},
			},
			reqB: OutboundHTTPRequest{
				Method:       "GET",
				URL:          "https://example.com/api",
				WorkflowOwner: "owner-1",
				Body:         []byte(`{"a":1}`),
				Headers:      map[string]string{"X": "1", "Y": "a,b"},
				MultiHeaders: map[string][]string{"X": {"1"}, "Y": {"b", "a"}},
			},
			sameHash: true,
		},
		{
			name:     "Same MultiHeaders values different slice order same hash",
			reqA:     baseWithMultiHeaders,
			reqB:     baseWithMultiHeaders,
			sameHash: true,
		},
		{
			name: "Different Headers value yields different hash",
			reqA: baseWithHeaders,
			reqB: OutboundHTTPRequest{
				Method:        "GET",
				URL:           "https://example.com/",
				WorkflowOwner: "owner",
				Headers:       map[string]string{"A": "2"},
			},
			sameHash: false,
		},
		{
			name: "Different MultiHeaders content yields different hash",
			reqA: baseWithMultiHeaders,
			reqB: OutboundHTTPRequest{
				Method:       "GET",
				URL:          "https://example.com/",
				WorkflowOwner: "owner",
				MultiHeaders: map[string][]string{"A": {"1", "3"}},
			},
			sameHash: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashA := tt.reqA.Hash()
			hashB := tt.reqB.Hash()
			require.NotEmpty(t, hashA)
			require.NotEmpty(t, hashB)
			if tt.sameHash {
				require.Equal(t, hashA, hashB)
			} else {
				require.NotEqual(t, hashA, hashB)
			}
		})
	}
}
