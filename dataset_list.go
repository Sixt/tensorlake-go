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

// IterDatasets iterates over all datasets in the organization.
func (c *Client) IterDatasets(ctx context.Context, limit int, direction PaginationDirection) iter.Seq2[Dataset, error] {
	return func(yield func(Dataset, error) bool) {
		cursor := ""
		for {
			listResp, err := c.ListDatasets(ctx, &ListDatasetsRequest{
				Cursor:    cursor,
				Limit:     limit,
				Direction: direction,
			})
			if err != nil {
				yield(Dataset{}, err)
				return
			}
			for _, dataset := range listResp.Items {
				if !yield(dataset, nil) {
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

// IterDatasetData iterates over all dataset data in the organization.
func (c *Client) IterDatasetData(ctx context.Context, limit int, direction PaginationDirection) iter.Seq2[ParseResult, error] {
	return func(yield func(ParseResult, error) bool) {
		cursor := ""
		for {
			listResp, err := c.ListDatasetData(ctx, &ListDatasetDataRequest{
				Cursor:    cursor,
				Limit:     limit,
				Direction: direction,
			})
			if err != nil {
				yield(ParseResult{}, err)
				return
			}
			for _, datasetData := range listResp.Items {
				if !yield(datasetData, nil) {
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

// ListDatasetsRequest holds options for listing datasets.
type ListDatasetsRequest struct {
	Cursor    string
	Direction PaginationDirection
	Limit     int
	Status    DatasetStatus
	Name      string
}

// ListDatasets lists all datasets in the organization.
func (c *Client) ListDatasets(ctx context.Context, in *ListDatasetsRequest) (*PaginationResult[Dataset], error) {
	reqURL := c.baseURL + "/datasets"

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
	if in.Status != "" {
		params.Add("status", string(in.Status))
	}
	if in.Name != "" {
		params.Add("name", in.Name)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	return do(c, req, func(r io.Reader) (*PaginationResult[Dataset], error) {
		var result PaginationResult[Dataset]
		if err := json.NewDecoder(r).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		return &result, nil
	})
}

// ListDatasetDataRequest holds options for listing dataset parse jobs.
type ListDatasetDataRequest struct {
	// Query parameter.

	DatasetId string `json:"-"`

	// Body parameters.

	Cursor         string              `json:"cursor,omitempty"`
	Direction      PaginationDirection `json:"direction,omitempty"`
	Limit          int                 `json:"limit,omitempty"`
	Status         ParseStatus         `json:"status,omitempty"`
	ParseId        string              `json:"parse_id,omitempty"`
	FileName       string              `json:"file_name,omitempty"`
	CreatedAfter   string              `json:"created_after,omitempty"`   // RFC3339
	CreatedBefore  string              `json:"created_before,omitempty"`  // RFC3339
	FinishedAfter  string              `json:"finished_after,omitempty"`  // RFC3339
	FinishedBefore string              `json:"finished_before,omitempty"` // RFC3339
}

// ListDatasetData lists all the parse jobs associated with a specific dataset.
// This endpoint allows you to retrieve the status and metadata of each parse job
// that has been submitted under the specified dataset.
func (c *Client) ListDatasetData(ctx context.Context, in *ListDatasetDataRequest) (*PaginationResult[ParseResult], error) {
	reqURL := fmt.Sprintf("%s/datasets/%s/data", c.baseURL, in.DatasetId)
	params := url.Values{}
	if in.Cursor != "" {
		params.Add("cursor", in.Cursor)
	}
	if in.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", in.Limit))
	}
	if in.Direction != "" {
		params.Add("direction", string(in.Direction))
	}
	if in.Status != "" {
		params.Add("status", string(in.Status))
	}
	if in.ParseId != "" {
		params.Add("parse_id", in.ParseId)
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
