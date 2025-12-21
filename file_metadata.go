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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// FileInfo represents metadata about a file.
type FileInfo struct {
	FileId         string            `json:"file_id"`
	FileName       string            `json:"file_name"`
	MimeType       MimeType          `json:"mime_type"`
	FileSize       int64             `json:"file_size"`
	ChecksumSHA256 string            `json:"checksum_sha256,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// GetFileMetadata retrieves metadata for a specific file.
//
// See also: [Get File Metadata API Reference]
//
// [Get File Metadata API Reference]: https://docs.tensorlake.ai/api-reference/v2/files/get-metadata
func (c *Client) GetFileMetadata(ctx context.Context, fileId string) (*FileInfo, error) {
	reqURL := fmt.Sprintf("%s/files/%s/metadata", c.baseURL, url.PathEscape(fileId))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	return do(c, req, func(r io.Reader) (*FileInfo, error) {
		var result FileInfo
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode get file metadata response: %w", err)
		}

		return &result, nil
	})
}
