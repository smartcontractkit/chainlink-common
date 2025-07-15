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
		require.Equal(t, "cfa8212d72dd8414d120b8e49d0173fe37ac7fb815fc9d582724ea865645aec5", digest)
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
		require.Equal(t, "597ac3b0a7659eabd02527779218b69fec5434afbc1c077d1bcecf8c1ae014f6", digest)
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

func TestNormalize(t *testing.T) {
	t.Run("Normalize - primitive types", func(t *testing.T) {
		str := "hello"
		normalizedStr, err := Normalize(str)
		require.NoError(t, err)
		require.Equal(t, str, normalizedStr)

		num := 42
		normalizedNum, err := Normalize(num)
		require.NoError(t, err)
		require.Equal(t, num, normalizedNum)

		b := true
		normalizedBool, err := Normalize(b)
		require.NoError(t, err)
		require.Equal(t, b, normalizedBool)
	})

	t.Run("Normalize - map with sorted keys", func(t *testing.T) {
		input := map[string]interface{}{
			"zebra": "animal",
			"apple": "fruit",
			"car":   "vehicle",
		}

		normalized, err := Normalize(input)
		require.NoError(t, err)

		expected := map[string]interface{}{
			"apple": "fruit",
			"car":   "vehicle",
			"zebra": "animal",
		}
		require.Equal(t, expected, normalized)
	})

	t.Run("Normalize - nested map", func(t *testing.T) {
		input := map[string]interface{}{
			"outer": map[string]interface{}{
				"z": "last",
				"a": "first",
			},
		}

		normalized, err := Normalize(input)
		require.NoError(t, err)

		expected := map[string]interface{}{
			"outer": map[string]interface{}{
				"a": "first",
				"z": "last",
			},
		}
		require.Equal(t, expected, normalized)
	})

	t.Run("Normalize - slice", func(t *testing.T) {
		input := []interface{}{"hello", 42, true}

		normalized, err := Normalize(input)
		require.NoError(t, err)

		expected := []interface{}{"hello", 42, true}
		require.Equal(t, expected, normalized)
	})

	t.Run("Normalize - slice with nested maps", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{
				"z": "last",
				"a": "first",
			},
			"simple string",
		}

		normalized, err := Normalize(input)
		require.NoError(t, err)

		expected := []interface{}{
			map[string]interface{}{
				"a": "first",
				"z": "last",
			},
			"simple string",
		}
		require.Equal(t, expected, normalized)
	})

	t.Run("Normalize - complex nested structure", func(t *testing.T) {
		input := map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{
					"name": "John",
					"age":  30,
					"address": map[string]interface{}{
						"zip":    "12345",
						"city":   "NYC",
						"street": "Main St",
					},
				},
			},
			"config": map[string]interface{}{
				"timeout": 5000,
				"debug":   true,
			},
		}

		normalized, err := Normalize(input)
		require.NoError(t, err)

		expected := map[string]interface{}{
			"config": map[string]interface{}{
				"debug":   true,
				"timeout": 5000,
			},
			"users": []interface{}{
				map[string]interface{}{
					"address": map[string]interface{}{
						"city":   "NYC",
						"street": "Main St",
						"zip":    "12345",
					},
					"age":  30,
					"name": "John",
				},
			},
		}
		require.Equal(t, expected, normalized)
	})

	t.Run("Normalize - nil value", func(t *testing.T) {
		var input interface{}
		normalized, err := Normalize(input)
		require.NoError(t, err)
		require.Nil(t, normalized)
	})

	t.Run("Normalize - empty map", func(t *testing.T) {
		input := map[string]interface{}{}
		normalized, err := Normalize(input)
		require.NoError(t, err)
		require.Equal(t, input, normalized)
	})

	t.Run("Normalize - empty slice", func(t *testing.T) {
		input := []interface{}{}
		normalized, err := Normalize(input)
		require.NoError(t, err)
		require.Equal(t, input, normalized)
	})
}

func TestDigestConsistency(t *testing.T) {
	t.Run("Request digest consistency with different param order", func(t *testing.T) {
		req1 := Request[map[string]interface{}]{
			Version: JsonRpcVersion,
			ID:      "test",
			Method:  "test.method",
			Params: &map[string]interface{}{
				"z": "last",
				"a": "first",
				"m": "middle",
			},
		}

		req2 := Request[map[string]interface{}]{
			Version: JsonRpcVersion,
			ID:      "test",
			Method:  "test.method",
			Params: &map[string]interface{}{
				"a": "first",
				"m": "middle",
				"z": "last",
			},
		}

		digest1, err1 := req1.Digest()
		digest2, err2 := req2.Digest()

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, digest1, digest2, "Digests should be equal regardless of param order")
	})

	t.Run("Response digest consistency with different result order", func(t *testing.T) {
		resp1 := Response[map[string]interface{}]{
			Version: JsonRpcVersion,
			ID:      "test",
			Result: &map[string]interface{}{
				"z": "last",
				"a": "first",
			},
		}

		resp2 := Response[map[string]interface{}]{
			Version: JsonRpcVersion,
			ID:      "test",
			Result: &map[string]interface{}{
				"a": "first",
				"z": "last",
			},
		}

		digest1, err1 := resp1.Digest()
		digest2, err2 := resp2.Digest()

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, digest1, digest2, "Digests should be equal regardless of result order")
	})

	t.Run("Auth field excluded from digest", func(t *testing.T) {
		type ParamsType struct {
			Value string `json:"value"`
		}

		req1 := Request[ParamsType]{
			Version: JsonRpcVersion,
			ID:      "test",
			Method:  "test.method",
			Params:  &ParamsType{Value: "test"},
			Auth:    "token123",
		}

		req2 := Request[ParamsType]{
			Version: JsonRpcVersion,
			ID:      "test",
			Method:  "test.method",
			Params:  &ParamsType{Value: "test"},
			Auth:    "differentToken456",
		}

		req3 := Request[ParamsType]{
			Version: JsonRpcVersion,
			ID:      "test",
			Method:  "test.method",
			Params:  &ParamsType{Value: "test"},
			// No Auth field
		}

		digest1, err1 := req1.Digest()
		digest2, err2 := req2.Digest()
		digest3, err3 := req3.Digest()

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		require.Equal(t, digest1, digest2, "Digests should be equal regardless of Auth token")
		require.Equal(t, digest1, digest3, "Digests should be equal whether Auth is present or not")
	})
}
