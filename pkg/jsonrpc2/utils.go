package jsonrpc2

import (
	"encoding/json"
	"errors"
)

func DecodeRequest[Params any](requestBytes []byte, jwtTokenFromHeader string) (Request[Params], error) {
	var request Request[Params]
	err := json.Unmarshal(requestBytes, &request)
	if err != nil {
		return Request[Params]{}, err
	}
	if request.Version != JsonRpcVersion {
		return Request[Params]{}, errors.New("incorrect jsonrpc version")
	}
	if request.Method == "" {
		return Request[Params]{}, errors.New("empty method field")
	}
	if request.Params == nil {
		return Request[Params]{}, errors.New("invalid params")
	}
	if request.Auth != "" {
		return request, nil
	}
	request.Auth = jwtTokenFromHeader
	return request, nil
}

func EncodeRequest[Params any](request *Request[Params]) ([]byte, error) {
	return json.Marshal(request)
}

func DecodeResponse[Result any](responseBytes []byte) (Response[Result], error) {
	var response Response[Result]
	err := json.Unmarshal(responseBytes, &response)
	if err != nil {
		return Response[Result]{}, err
	}
	return response, nil
}

func EncodeResponse[Result any](response *Response[Result]) ([]byte, error) {
	return json.Marshal(response)
}

func EncodeErrorReponse[ErrorData any](id string, err *WireError) ([]byte, error) {
	return json.Marshal(Response[any]{
		Version: JsonRpcVersion,
		ID:      id,
		Error:   err,
	})
}
