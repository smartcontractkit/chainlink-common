package jsonrpc

import (
	"encoding/json"
	"errors"
)

type Codec struct {
}

func (*Codec) DecodeRequest(requestBytes []byte) (Request, error) {
	var request Request
	err := json.Unmarshal(requestBytes, &request)
	if err != nil {
		return Request{}, err
	}
	if request.Version != "2.0" {
		return Request{}, errors.New("incorrect jsonrpc version")
	}
	if request.Method == "" {
		return Request{}, errors.New("empty method field")
	}
	if request.Params == nil {
		return Request{}, errors.New("missing params attribute")
	}
	if request.Auth != "" {
		return request, nil
	}

	return request, nil
}

func (*Codec) EncodeRequest(request *Request) ([]byte, error) {
	return json.Marshal(request)
}

func (*Codec) DecodeResponse(responseBytes []byte) (Response, error) {
	var response Response
	err := json.Unmarshal(responseBytes, &response)
	if err != nil {
		return Response{}, err
	}
	return response, nil
}

func (*Codec) EncodeResponse(response *Response) ([]byte, error) {
	return json.Marshal(response)
}
