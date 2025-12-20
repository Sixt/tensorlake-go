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
	"net/http"
	"os"
)

// Client is a Tensorlake API client.
type Client struct {
	httpClient *http.Client

	baseURL string
	apiKey  string
	region  Region
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

// WithRegion sets the region to use for the client.
func WithRegion(region Region) Option {
	return func(c *Client) {
		c.region = region
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
		region:     RegionEU,
		baseURL:    "https://api.eu.tensorlake.ai/documents/v2",
		apiKey:     os.Getenv("TENSORLAKE_API_KEY"),
	}

	for _, opt := range opts {
		opt(client)
	}

	// For non on-premise regions, use the default base URL.
	switch client.region {
	case RegionEU:
		client.baseURL = "https://api.eu.tensorlake.ai/documents/v2"
	case RegionUS:
		client.baseURL = "https://api.tensorlake.ai/documents/v2"
	}
	return client
}
