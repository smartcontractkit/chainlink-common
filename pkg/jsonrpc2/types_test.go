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
		require.Equal(t, "01025753dbedc82b0771489d77f87b4fe8bbe7255be218caafce6704cc4b45a6", digest)
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
		require.Equal(t, "error marshaling JSON: json: unsupported type: chan string", err.Error())
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
			Result: &ResultType{
				Value: "success",
			},
		}
		digest, err := resp.Digest()
		require.NoError(t, err)
		require.NotEmpty(t, digest)
		require.Equal(t, "1dcc380a15b81b33bda5b69b8ea1e01671a7633cdc989df65e68c53aba587087", digest)
	})

	t.Run("Response.Digest - with error", func(t *testing.T) {
		resp := Response[any]{
			Version: JsonRpcVersion,
			ID:      "err1",
			Error: &WireError{
				Code:    ErrInvalidRequest,
				Message: "bad request",
			},
		}
		digest, err := resp.Digest()
		require.NoError(t, err)
		require.NotEmpty(t, digest)
		require.Equal(t, "df7b0b7347b46915adab6e50b4c9475fc42106dd61c8d2340b7dec260248204c", digest)
	})

	t.Run("Response.Digest - JSON marshal error", func(t *testing.T) {
		type UnmarshalableResult struct {
			Ch chan int `json:"ch"`
		}
		resp := Response[UnmarshalableResult]{
			Version: JsonRpcVersion,
			ID:      "bad",
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
