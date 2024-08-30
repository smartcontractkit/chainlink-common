package api

import (
	"os"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	resty *resty.Client
}

func NewClient(url string, apiKey string) *Client {
	isDebug := os.Getenv("DEBUG_RESTY") == "true"

	return &Client{
		resty: resty.New().
			SetDebug(isDebug).
			SetBaseURL(url).
			SetHeader("Authorization", "Bearer "+apiKey),
	}
}
