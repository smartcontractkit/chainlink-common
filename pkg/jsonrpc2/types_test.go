package jsonrpc2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypes(t *testing.T) {
	t.Run("Request.Digest", func(t *testing.T) {
		type TestType struct {
			ID        string `json:"id"`
			Namespace string `json:"namespace"`
			Owner     string `json:"owner"`
		}

		req := Request[TestType]{
			Version: JsonRpcVersion,
			ID:      "1",
			Method:  "service.method",
			Params: &TestType{
				ID:        "1",
				Namespace: "test",
				Owner:     "test",
			},
		}
		digest, err := req.Digest()
		require.NoError(t, err)
		require.Equal(t, "0d390d82191fcc4fe7b321124f55e3880bf3f754c725d10bcc0668cebf6a6ded", digest)
	})

	t.Run("Request.Digest - JSON marshal error", func(t *testing.T) {
		type UnmarshalableType struct {
			ID      string      `json:"id"`
			Channel chan string `json:"channel"` // channels can't be marshaled to JSON
		}

		req := Request[UnmarshalableType]{
			Version: JsonRpcVersion,
			ID:      "1",
			Method:  "service.method",
			Params: &UnmarshalableType{
				ID:      "1",
				Channel: make(chan string),
			},
		}

		digest, err := req.Digest()
		require.Error(t, err)
		require.Equal(t, "error marshaling JSON: canonicaljson: unsupported type: chan string", err.Error())
		require.Empty(t, digest)
	})

	t.Run("WireError.Error", func(t *testing.T) {
		msg := "Invalid request format"
		wireErr := WireError{
			Code:    ErrInvalidRequest,
			Message: msg,
		}

		result := wireErr.Error()
		require.Equal(t, msg, result)
	})
}

func TestResponseDigest(t *testing.T) {
	t.Run("Response.Digest - with result", func(t *testing.T) {
		type ResultType struct {
			Value string `json:"value"`
		}
		resp := Response[ResultType]{
			Version: JsonRpcVersion,
			ID:      "42",
			Method:  "service.method",
			Result: &ResultType{
				Value: "success",
			},
		}
		digest, err := resp.Digest()
		require.NoError(t, err)
		require.NotEmpty(t, digest)
		require.Equal(t, "b2915f4ca34315c3cade73ce8e1d69225d3512550a12819374e287ed8327ba25", digest)
	})

	t.Run("Response.Digest - with error", func(t *testing.T) {
		resp := Response[any]{
			Version: JsonRpcVersion,
			ID:      "err1",
			Method:  "service.method",
			Error: &WireError{
				Code:    ErrInvalidRequest,
				Message: "bad request",
			},
		}
		digest, err := resp.Digest()
		require.NoError(t, err)
		require.NotEmpty(t, digest)
		require.Equal(t, "a8e6e0e131f05fc98fedc4826550a13bbdeb93b01bd9b97aa462a2411860e7d9", digest)
	})

	t.Run("Response.Digest - JSON marshal error", func(t *testing.T) {
		type UnmarshalableResult struct {
			Ch chan int `json:"ch"`
		}
		resp := Response[UnmarshalableResult]{
			Version: JsonRpcVersion,
			ID:      "bad",
			Method:  "service.method",
			Result: &UnmarshalableResult{
				Ch: make(chan int),
			},
		}
		digest, err := resp.Digest()
		require.Error(t, err)
		require.Contains(t, err.Error(), "error marshaling JSON")
		require.Empty(t, digest)
	})
}
