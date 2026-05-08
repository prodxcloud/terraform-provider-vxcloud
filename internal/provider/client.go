package provider

import (
	"net/http"
	"time"
)

type Client struct {
	Endpoint string
	Email    string
	APIToken string
	HTTP     *http.Client
}

func NewClient(endpoint, email, apiToken string) *Client {
	return &Client{
		Endpoint: endpoint,
		Email:    email,
		APIToken: apiToken,
		HTTP:     &http.Client{Timeout: 30 * time.Second},
	}
}
