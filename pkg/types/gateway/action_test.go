package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutboundHTTPRequest_Hash_Deterministic_HeadersOnly(t *testing.T) {
	// Same logical request built in different map insertion orders must produce the same hash.
	req1 := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/api",
		WorkflowOwner: "owner-1",
		Body:          []byte(`{"a":1}`),
		Headers: map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
			"X-Request-Id": "req-123",
		},
	}
	req2 := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/api",
		WorkflowOwner: "owner-1",
		Body:          []byte(`{"a":1}`),
		Headers: map[string]string{
			"X-Request-Id": "req-123",
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
	}

	hash1 := req1.Hash()
	hash2 := req2.Hash()

	require.Equal(t, hash1, hash2, "Hash() must be same for same content regardless of map iteration order - Headers only")
}

func TestOutboundHTTPRequest_Hash_Deterministic_MultiHeadersOnly(t *testing.T) {
	// Same logical request with MultiHeaders: keys and values per key may be in different order; hash must be identical.
	req1 := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/api",
		WorkflowOwner: "owner-1",
		Body:          []byte(`{"a":1}`),
		MultiHeaders: map[string][]string{
			"Accept":       {"application/json"},
			"Content-Type": {"application/json"},
			"Set-Cookie":   {"s1=abc", "s2=def"},
		},
	}
	req2 := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/api",
		WorkflowOwner: "owner-1",
		Body:          []byte(`{"a":1}`),
		MultiHeaders: map[string][]string{
			"Set-Cookie":   {"s2=def", "s1=abc"}, // different key and value order
			"Content-Type": {"application/json"},
			"Accept":       {"application/json"},
		},
	}

	hash1 := req1.Hash()
	hash2 := req2.Hash()

	require.Equal(t, hash1, hash2, "Hash() must be same for same content regardless of map/slice order - MultiHeaders only (keys and values are sorted)")
}

func TestOutboundHTTPRequest_Hash_DifferentContent_DifferentHash(t *testing.T) {
	base := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/",
		WorkflowOwner: "owner",
		Headers:       map[string]string{"A": "1"},
	}
	hBase := base.Hash()
	require.NotEmpty(t, hBase)

	// Different Headers value
	rH := base
	rH.Headers = map[string]string{"A": "2"}
	require.NotEqual(t, hBase, rH.Hash(), "different Headers value must yield different hash")

	// MultiHeaders-only: different value order for same key should still be sorted, so same hash for same set of values
	baseMulti := OutboundHTTPRequest{
		Method:        "GET",
		URL:           "https://example.com/",
		WorkflowOwner: "owner",
		MultiHeaders:  map[string][]string{"A": {"1", "2"}},
	}
	hMultiA := baseMulti.Hash()
	hMultiB := baseMulti.Hash()
	require.Equal(t, hMultiA, hMultiB)

	// Different MultiHeaders content
	otherMulti := baseMulti
	otherMulti.MultiHeaders = map[string][]string{"A": {"1", "3"}}
	require.NotEqual(t, hMultiA, otherMulti.Hash(), "different MultiHeaders content must yield different hash")
}

func TestOutboundHTTPRequest_HashValidated_Success(t *testing.T) {
	req := OutboundHTTPRequest{
		Method:       "GET",
		URL:          "https://example.com/",
		WorkflowOwner: "owner",
		Headers:       map[string]string{"A": "1"},
	}
	hash, err := req.HashValidated()
	require.NoError(t, err)
	require.Equal(t, req.Hash(), hash, "HashValidated() must return same hash as Hash() for valid request")
}

func TestOutboundHTTPRequest_HashValidated_ReturnsErrorWhenBothHeadersAndMultiHeadersSet(t *testing.T) {
	req := OutboundHTTPRequest{
		Method:       "GET",
		URL:          "https://example.com/",
		WorkflowOwner: "owner",
		Headers:       map[string]string{"A": "1"},
		MultiHeaders:  map[string][]string{"B": {"2"}},
	}
	hash, err := req.HashValidated()
	require.Error(t, err)
	require.Empty(t, hash)
	require.ErrorIs(t, err, ErrBothHeadersAndMultiHeaders)
}
