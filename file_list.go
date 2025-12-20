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

// ListFilesRequest holds options for listing files.
type ListFilesRequest struct {
	// Cursor is the cursor to use for pagination.
	// This is a base64-encoded string representing a timestamp.
	// It is used to paginate through the results.
	//
	// Optional.
	Cursor string `json:"cursor,omitempty"`

	// Direction of pagination.
	//
	// This can be either next or prev.
	// next means to get the next page of results,
	// while prev means to get the previous page of results.
	//
	// Optional.
	Direction PaginationDirection `json:"direction,omitempty"`

	// Limit is the limits for the number of results to return.
	//
	// This is a positive integer that specifies the maximum number of results
	// to return. If not provided, a default value will be used.
	//
	// Required range: x >= 0.
	Limit int `json:"limit,omitempty"`

	// FileName is the name to filter results by.
	// This is a case-sensitive substring that will be matched against the file names.
	// If provided, only files with names containing this substring will be returned.
	FileName string `json:"file_name,omitempty"`

	// CreatedAfter is the date and time to filter results by.
	// The date should be in RFC 3339 format.
	CreatedAfter string `json:"created_after,omitempty"`

	// CreatedBefore is the date and time to filter results by.
	// The date should be in RFC 3339 format.
	CreatedBefore string `json:"created_before,omitempty"`
}

// ListFiles lists files in the Tensorlake project.
//
// This operation allows to list every file that has been uploaded to the Project specified by the API key used in the request.
// The response will include metadata about each file, such as the file ID, name, size, and type.
// We use cursor-based pagination to return the files in pages. A page has the following fields:
//
//   - Items: An array of file metadata, each containing the fields described below.
//   - HasMore: A boolean indicating whether there are more files available beyond the current page.
//   - NextCursor: A base64-encoded cursor for the next page of results. If HasMore is false, this field will be null.
//   - PrevCursor: A base64-encoded cursor for the previous page of results. If this is the first page, this field will be null.
func (c *Client) ListFiles(ctx context.Context, in *ListFilesRequest) (*PaginationResult[FileInfo], error) {
	reqURL := fmt.Sprintf("%s/files", c.baseURL)
	params := url.Values{}
	if in.Cursor != "" {
		params.Add("cursor", in.Cursor)
	}
	if in.Direction != "" {
		params.Add("direction", string(in.Direction))
	}
	if in.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", in.Limit))
	}
	if in.FileName != "" {
		params.Add("file_name", in.FileName)
	}
	if in.CreatedAfter != "" {
		params.Add("created_after", in.CreatedAfter)
	}
	if in.CreatedBefore != "" {
		params.Add("created_before", in.CreatedBefore)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return do(c, req, func(r io.Reader) (*PaginationResult[FileInfo], error) {
		var result PaginationResult[FileInfo]
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode list files response: %w", err)
		}
		return &result, nil
	})
}
