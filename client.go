// Copyright 2025 SIXT SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tensorlake

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Client is a Tensorlake API client.
type Client struct {
	httpClient *http.Client

	baseURL string
	apiKey  string
}

// Option defines a configuration option for the Client.
type Option func(*Client)

// WithBaseURL sets the base URL to use for the client.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithAPIKey sets the API key to use for the client.
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithHTTPClient sets the HTTP client to use for the client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new Tensorlake API client.
func NewClient(opts ...Option) *Client {
	client := &Client{
		httpClient: http.DefaultClient,
		baseURL:    EndpointEU,
		apiKey:     os.Getenv("TENSORLAKE_API_KEY"),
	}

	for _, opt := range opts {
		opt(client)
	}
	return client
}

func do[T any](c *Client, req *http.Request, successHandler func(io.Reader) (T, error)) (T, error) {
	var zero T

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if successHandler != nil {
			return successHandler(resp.Body)
		}
		return zero, nil

	case resp.StatusCode >= 400:
		var errRes ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errRes); err != nil {
			return zero, fmt.Errorf("failed to decode error response (%d): %w", resp.StatusCode, err)
		}
		return zero, &errRes

	default:
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Limit to 1MB
		return zero, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

}
