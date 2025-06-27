package jsonrpc2

import (
	"encoding/json"
	"errors"
)

func DecodeRequest(requestBytes []byte, jwtTokenFromHeader string) (Request, error) {
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
		return Request{}, errors.New("invalid params")
	}
	if request.Auth != "" {
		return request, nil
	}
	request.Auth = jwtTokenFromHeader
	return request, nil
}

func EncodeRequest(request *Request) ([]byte, error) {
	return json.Marshal(request)
}

func DecodeResponse(responseBytes []byte) (Response, error) {
	var response Response
	err := json.Unmarshal(responseBytes, &response)
	if err != nil {
		return Response{}, err
	}
	return response, nil
}

func EncodeResponse(response *Response) ([]byte, error) {
	return json.Marshal(response)
}

func EncodeErrorReponse(id string, err *WireError) ([]byte, error) {
	return json.Marshal(Response{
		Version: JsonRpcVersion,
		ID:      id,
		Error:   err,
	})
}
