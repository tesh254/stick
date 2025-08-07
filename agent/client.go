package agent

import (
	"net/http"
	"time"

	"github.com/tesh254/stick/internal/config"
)

type Client struct {
	Endpoint   string
	APIKey     string
	HTTPClient *http.Client
}

func Init() *Client {
	apiKey := config.GetAPIKey()
	return &Client{
		Endpoint:   "https://api.together.xyz/v1/chat/completions",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}
