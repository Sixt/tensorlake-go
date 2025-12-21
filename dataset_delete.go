// Copyright 2025 SIXT SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tensorlake

import (
	"context"
	"fmt"
	"net/http"
)

// DeleteDataset deletes a dataset from Tensorlake.
//
// See also: [Delete Dataset API Reference]
//
// [Delete Dataset API Reference]: https://docs.tensorlake.ai/api-reference/v2/datasets/delete
func (c *Client) DeleteDataset(ctx context.Context, datasetId string) error {
	reqURL := fmt.Sprintf("%s/datasets/%s", c.baseURL, datasetId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	_, err = do[struct{}](c, req, nil)
	return err
}
