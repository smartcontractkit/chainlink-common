package jsonrpc

import (
	"encoding/json"
	"errors"
)

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}
func (c *Codec) DecodeRequestWithAuthToken(requestBytes []byte, authTokenFromHeader string) (*Request, error) {
	req, err := c.DecodeRequest(requestBytes)
	if err != nil {
		return nil, err
	}
	if authTokenFromHeader == "" {
		return nil, errors.New("missing auth token")
	}
	req.Auth = authTokenFromHeader
	return req, nil
}

func (*Codec) DecodeRequest(requestBytes []byte) (*Request, error) {
	var request Request
	err := json.Unmarshal(requestBytes, &request)
	if err != nil {
		return nil, err
	}
	if request.Version != "2.0" {
		return nil, errors.New("incorrect jsonrpc version")
	}
	if request.Method == "" {
		return nil, errors.New("empty method field")
	}
	if request.Params == nil {
		return nil, errors.New("missing params attribute")
	}
	return &request, nil
}

func (*Codec) EncodeRequest(request *Request) ([]byte, error) {
	return json.Marshal(request)
}

func (*Codec) DecodeResponse(responseBytes []byte) (*Response, error) {
	var response Response
	err := json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (*Codec) EncodeResponse(response *Response) ([]byte, error) {
	return json.Marshal(response)
}
