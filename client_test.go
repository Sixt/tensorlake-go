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
	"testing"
)

func TestNewClient(t *testing.T) {

	t.Run("with default options", func(t *testing.T) {
		client := NewClient(WithAPIKey("test-key"))

		if client == nil {
			t.Fatal("NewClient returned nil")
		}

		if client.apiKey != "test-key" {
			t.Errorf("expected API key 'test-key', got '%s'", client.apiKey)
		}

		if client.baseURL != "https://api.eu.tensorlake.ai/documents/v2" {
			t.Errorf("expected default base URL 'https://api.eu.tensorlake.ai/documents/v2', got '%s'", client.baseURL)
		}

		if client.httpClient == nil {
			t.Error("expected httpClient to be set, got nil")
		}
	})

	t.Run("with custom base URL", func(t *testing.T) {
		client := NewClient(WithBaseURL("https://api.custom.com"))

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		client := NewClient(WithHTTPClient(&http.Client{}))

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
	})

	t.Run("with custom region", func(t *testing.T) {
		client := NewClient(WithRegion(RegionUS))

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
	})

	t.Run("with custom API key", func(t *testing.T) {
		client := NewClient(WithAPIKey("test-key"))

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
	})

	t.Run("with custom base URL and region", func(t *testing.T) {
		client := NewClient(WithBaseURL("https://api.custom.com"), WithRegion(RegionUS))

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
	})
}
