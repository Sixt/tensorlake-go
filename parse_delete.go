package tensorlake

import (
	"context"
	"fmt"
	"net/http"
)

// DeleteParseJob deletes a previously submitted parse job. This will
// remove the parse job and its associated settings from the system.
// Deleting a parse job does not delete the original file used for parsing,
// nor does it affect any other parse jobs that may have been created from the same file.
func (c *Client) DeleteParseJob(ctx context.Context, parseId string) error {
	reqURL := fmt.Sprintf("%s/parse/%s", c.baseURL, parseId)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	_, err = do[struct{}](c, req, nil)
	return err
}
