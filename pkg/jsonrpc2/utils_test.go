package jsonrpc2

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testJWT = "test.jwt.token"
)

func TestHandler_DecodeRequest(t *testing.T) {
	var paramsStr string = "params"
	rawParams, err := json.Marshal(paramsStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	validReq := Request{
		Version: JsonRpcVersion,
		Method:  "service.method",
		Params:  rawParams,
		Auth:    "",
	}
	reqBytes, _ := json.Marshal(validReq)

	t.Run("valid request with jwt", func(t *testing.T) {
		got, err := DecodeRequest(reqBytes, testJWT)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Auth != testJWT {
			t.Errorf("expected auth to be set from header")
		}
		if got.Version != JsonRpcVersion {
			t.Errorf("expected version to be 2.0")
		}
		if got.Method != "service.method" {
			t.Errorf("expected method to be 'service.method', got %s", got.Method)
		}
		var params string
		err = json.Unmarshal(got.Params, &params)
		if err != nil {
			t.Fatalf("failed to unmarshal params: %v", err)
		}
		if params != paramsStr {
			t.Errorf("expected params to be 'params', got %s", params)
		}
	})

	t.Run("valid request with auth in body", func(t *testing.T) {
		req := validReq
		req.Auth = "body.jwt"
		reqBytes, _ := json.Marshal(req)
		got, err := DecodeRequest(reqBytes, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Auth != "body.jwt" {
			t.Errorf("expected auth from body")
		}
	})

	t.Run("missing params", func(t *testing.T) {
		req := validReq
		req.Params = nil
		reqBytes, _ := json.Marshal(req)
		_, err := DecodeRequest(reqBytes, testJWT)
		if err == nil || err.Error() != "invalid params" {
			t.Errorf("expected missing params error, got %v", err)
		}
	})

	t.Run("empty method", func(t *testing.T) {
		req := validReq
		req.Method = ""
		reqBytes, _ := json.Marshal(req)
		_, err := DecodeRequest(reqBytes, testJWT)
		if err == nil || err.Error() != "empty method field" {
			t.Errorf("expected empty method error, got %v", err)
		}
	})

	t.Run("incorrect version", func(t *testing.T) {
		req := validReq
		req.Version = "1.0"
		reqBytes, _ := json.Marshal(req)
		_, err := DecodeRequest(reqBytes, testJWT)
		if err == nil || err.Error() != "incorrect jsonrpc version" {
			t.Errorf("expected version error, got %v", err)
		}
	})

	t.Run("missing auth token", func(t *testing.T) {
		req := validReq
		req.Auth = ""
		reqBytes, _ := json.Marshal(req)
		got, err := DecodeRequest(reqBytes, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Auth != "" {
			t.Errorf("unexpected auth token")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := DecodeRequest([]byte("{invalid"), testJWT)
		if err == nil {
			t.Errorf("expected json unmarshal error")
		}
	})
}

func TestHandler_EncodeRequest(t *testing.T) {
	var paramsStr string = "params"
	rawParams, err := json.Marshal(paramsStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := &Request{
		Version: JsonRpcVersion,
		Method:  "service.method",
		Params:  rawParams,
		Auth:    testJWT,
	}
	data, err := EncodeRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(got, *req) {
		t.Errorf("expected %v, got %v", *req, got)
	}
}

func TestHandler_DecodeResponse(t *testing.T) {
	result, err := json.Marshal(`"key": "value"`)
	require.NoError(t, err)
	resp := Response{
		Version: JsonRpcVersion,
		ID:      "1",
		Result:  result,
	}
	data, _ := json.Marshal(resp)

	t.Run("valid response", func(t *testing.T) {
		decodedResponse, err := DecodeResponse(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(decodedResponse, resp) {
			t.Errorf("expected %v, got %v", resp, decodedResponse)
		}
	})

	t.Run("response with error", func(t *testing.T) {
		resp = Response{
			Version: JsonRpcVersion,
			ID:      "1",
			Error:   &WireError{Code: 123, Message: "fail"},
		}
		respBytes, _ := json.Marshal(resp)
		decodedResponse, err := DecodeResponse(respBytes)
		require.NoError(t, err)
		if !reflect.DeepEqual(decodedResponse, resp) {
			t.Errorf("expected %v, got %v", resp, decodedResponse)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err = DecodeResponse([]byte("{invalid"))
		if err == nil {
			t.Errorf("expected json unmarshal error")
		}
	})
}

func TestHandler_EncodeResponse(t *testing.T) {
	rawResult, err := json.Marshal("result")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp := &Response{
		Version: JsonRpcVersion,
		ID:      "1",
		Result:  rawResult,
	}
	data, err := EncodeResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got Response
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(got, *resp) {
		t.Errorf("expected %v, got %v", *resp, got)
	}
}

func TestRequest_EncodeErrorReponse(t *testing.T) {
	wireErr := &WireError{Code: 1, Message: "fail"}
	data, err := EncodeErrorReponse("abc", wireErr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(resp.Error, wireErr) {
		t.Errorf("expected %v, got %v", wireErr, resp.Error)
	}
	if resp.ID != "abc" {
		t.Errorf("expected id 'abc', got %v", resp.ID)
	}
}

func TestRequest_ServiceName(t *testing.T) {
	req := &Request{Method: "foo.bar"}
	if got := req.ServiceName(); got != "foo" {
		t.Errorf("expected 'foo', got %v", got)
	}
	req.Method = "single"
	if got := req.ServiceName(); got != "single" {
		t.Errorf("expected 'single', got %v", got)
	}
}
