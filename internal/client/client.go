package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "http://localhost:3678"
)

type Client struct {
	BaseURL    string
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		BaseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) WaitForServer(maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		resp, err := c.doRequest("GET", "/", nil)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}

		}
		time.Sleep(retryInterval)

	}
	return fmt.Errorf("server not available after %d retries", maxRetries)
}
