package jsonrpc2

import (
	"encoding/json"
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
		require.Contains(t, err.Error(), "error marshaling JSON: json:")
		require.Contains(t, err.Error(), "Go chan string within \"/params/channel\"")
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
		require.Equal(t, "6acc256f1ba5bf28f834040ecd80ff28d316aa1806ef21ace295f45226b0a66d", digest)
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
		require.Equal(t, "62c241cdefd9bfabcbae8a92803406608a5327b5c73085380d189472b5ff0fb5", digest)
	})

	t.Run("Response.Digest - excludes ID", func(t *testing.T) {
		type ResultType struct {
			Value string `json:"value"`
		}
		mkResp := func(id string) Response[ResultType] {
			return Response[ResultType]{
				Version: JsonRpcVersion,
				ID:      id,
				Method:  "service.method",
				Result:  &ResultType{Value: "success"},
			}
		}
		prefixedResp := mkResp("owner::42")
		prefixed, err := prefixedResp.Digest()
		require.NoError(t, err)
		strippedResp := mkResp("42")
		stripped, err := strippedResp.Digest()
		require.NoError(t, err)
		require.Equal(t, prefixed, stripped)
	})

	t.Run("Response.Digest - excludes NodeSignatures", func(t *testing.T) {
		type ResultType struct {
			Value string `json:"value"`
		}
		resp := Response[ResultType]{
			Version: JsonRpcVersion,
			ID:      "42",
			Method:  "service.method",
			Result:  &ResultType{Value: "success"},
		}
		before, err := resp.Digest()
		require.NoError(t, err)
		resp.NodeSignatures = [][]byte{{0x01}, {0x02, 0x03}}
		after, err := resp.Digest()
		require.NoError(t, err)
		require.Equal(t, before, after)
	})

	t.Run("Response.NodeSignatures - JSON roundtrip", func(t *testing.T) {
		type ResultType struct {
			Value string `json:"value"`
		}
		resp := Response[ResultType]{
			Version:        JsonRpcVersion,
			ID:             "42",
			Method:         "service.method",
			Result:         &ResultType{Value: "success"},
			NodeSignatures: [][]byte{{0x01, 0x02}, {0x03}},
		}
		b, err := json.Marshal(&resp)
		require.NoError(t, err)

		var got Response[ResultType]
		require.NoError(t, json.Unmarshal(b, &got))
		require.Equal(t, resp.NodeSignatures, got.NodeSignatures)
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
