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
	"iter"
	"net/http"
	"net/url"
)

// IterParseJobs iterates over all parse jobs in the project.
func (c *Client) IterParseJobs(ctx context.Context, limit int, direction PaginationDirection) iter.Seq2[ParseResult, error] {
	return func(yield func(ParseResult, error) bool) {
		cursor := ""
		for {
			listResp, err := c.ListParseJobs(ctx, &ListParseJobsRequest{
				Cursor:    cursor,
				Limit:     limit,
				Direction: direction,
			})
			if err != nil {
				yield(ParseResult{}, err)
				return
			}
			for _, parseJob := range listResp.Items {
				if !yield(parseJob, nil) {
					return
				}
			}
			cursor = listResp.NextCursor
			if !listResp.HasMore {
				return
			}
		}
	}
}

type ListParseJobsRequest struct {
	Cursor         string              `json:"cursor,omitempty"`
	Direction      PaginationDirection `json:"direction,omitempty"`
	DatasetName    string              `json:"dataset_name,omitempty"`
	Limit          int                 `json:"limit,omitempty"`
	FileName       string              `json:"file_name,omitempty"`
	Status         ParseStatus         `json:"status,omitempty"`
	CreatedAfter   string              `json:"created_after,omitempty"`
	CreatedBefore  string              `json:"created_before,omitempty"`
	FinishedAfter  string              `json:"finished_after,omitempty"`
	FinishedBefore string              `json:"finished_before,omitempty"`
}

// ListParseJobs lists parse jobs in the Tensorlake project.
//
// See also: [List Parse Jobs API Reference]
//
// [List Parse Jobs API Reference]: https://docs.tensorlake.ai/api-reference/v2/parse/list
func (c *Client) ListParseJobs(ctx context.Context, in *ListParseJobsRequest) (*PaginationResult[ParseResult], error) {
	reqURL := c.baseURL + "/parse"
	params := url.Values{}
	if in.Cursor != "" {
		params.Add("cursor", in.Cursor)
	}
	if in.Direction != "" {
		params.Add("direction", string(in.Direction))
	}
	if in.DatasetName != "" {
		params.Add("dataset_name", in.DatasetName)
	}
	if in.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", in.Limit))
	}
	if in.FileName != "" {
		params.Add("file_name", in.FileName)
	}
	if in.Status != "" {
		params.Add("status", string(in.Status))
	}
	if in.CreatedAfter != "" {
		params.Add("created_after", in.CreatedAfter)
	}
	if in.CreatedBefore != "" {
		params.Add("created_before", in.CreatedBefore)
	}
	if in.FinishedAfter != "" {
		params.Add("finished_after", in.FinishedAfter)
	}
	if in.FinishedBefore != "" {
		params.Add("finished_before", in.FinishedBefore)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return do(c, req, func(r io.Reader) (*PaginationResult[ParseResult], error) {
		var result PaginationResult[ParseResult]
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &result, nil
	})
}
