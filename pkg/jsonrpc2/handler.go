package jsonrpc2

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Handler struct {
}

func (*Handler) DecodeRequest(requestBytes []byte, jwtTokenFromHeader string) (Request, error) {
	var request Request
	err := json.Unmarshal(requestBytes, &request)
	if err != nil {
		return Request{}, err
	}
	if request.Version != JsonRpcVersion {
		return Request{}, errors.New("incorrect jsonrpc version")
	}
	if request.Method == "" {
		return Request{}, errors.New("empty method field")
	}
	if request.Params == nil {
		return Request{}, ErrInvalidParams
	}
	if request.Auth != "" {
		return request, nil
	}

	if jwtTokenFromHeader == "" {
		return request, errors.New("missing auth token")
	}

	request.Auth = jwtTokenFromHeader

	return request, nil
}

func (*Handler) EncodeRequest(request *Request) ([]byte, error) {
	return json.Marshal(request)
}

func (*Handler) DecodeResponse(responseBytes []byte) (Response, error) {
	var response Response
	err := json.Unmarshal(responseBytes, &response)
	if err != nil {
		return Response{}, err
	}
	if response.Error != nil {
		return Response{}, fmt.Errorf("received non-empty error field: %v", response.Error)
	}
	return response, nil
}

func (*Handler) EncodeResponse(response *Response) ([]byte, error) {
	return json.Marshal(response)
}

func (r *Request) EncodeErrorReponse(err *WireError) ([]byte, error) {
	return json.Marshal(Response{
		Version: JsonRpcVersion,
		ID:      r.ID,
		Error:   err,
	})
}

func (r *Request) ServiceName() string {
	return strings.Split(r.Method, ".")[0]
}
