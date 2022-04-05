package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func MockServer(handlerFunc http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handlerFunc)
}

func WriteResponse(rw http.ResponseWriter, statusCode int, obj interface{}) error {
	rw.WriteHeader(statusCode)
	if obj == nil {
		return nil
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = rw.Write(b)
	return err
}

func RequestBody(r interface{}) (io.Reader, error) {
	data, err := json.Marshal(r)
	return bytes.NewBuffer(data), err
}

func MockGinContext(reqBody interface{}) (*httptest.ResponseRecorder, *gin.Context, error) {
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	body, err := RequestBody(reqBody)
	if err != nil {
		return res, ctx, err
	}

	req, err := http.NewRequest(http.MethodPost, "", body)
	if err != nil {
		return res, ctx, err
	}

	ctx.Request = req
	return res, ctx, err
}
