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
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// UploadFileRequest holds options for uploading a file.
type UploadFileRequest struct {
	// FileBytes is the reader for the file to upload.
	//
	// Required.
	FileBytes io.Reader `json:"file_bytes"`

	// FileName is the name of the file to upload.
	//
	// Optional.
	FileName string `json:"file_name"`

	// Labels are the labels to add to the file.
	//
	// Optional.
	Labels map[string]string `json:"labels,omitempty"`
}

// FileUploadResponse represents the response from uploading a file.
type FileUploadResponse struct {
	// FileId is the ID of the created file.
	// Use this ID to reference the file in parse, datasets, and other operations.
	FileId string `json:"file_id"`

	// CreatedAt is the creation date and time of the file.
	// This is in RFC 3339 format.
	CreatedAt time.Time `json:"created_at"`
}

// UploadFile uploads a file to Tensorlake Cloud.
//
// The file will be associated with the project specified by the API key
// used in the request.
//
// The file can be of any of the following types:
// - PDF
// - Word (DOCX)
// - Spreadsheets (XLS, XLSX, XSLM, CSV)
// - Presentations (PPTX, Apple Keynote)
// - Images (PNG, JPG, JPEG)
// - Raw text (plain text, HTML)
//
// The file type is automatically detected based on Content-Type header.
// In case the Content-Type header is not provided, the file extension
// will be used to infer the type. If the file type cannot be determined,
// it will default to application/octet-stream.
//
// We only keep one copy of the file, so uploading the same file multiple
// times will return the same file_id.
//
// # Labels
//
// Labels can be added to the file to help categorize the parse jobs associated with it.
// Labels are key-value pairs that can be used to filter and organize files.
// These should be provided in the a labels text field in the multipart form data.
// Labels are optional, but they can be very useful for organizing and managing parse jobs.
//
// # Limits
//
// There is an upload limit of 1 GB per file.
func (c *Client) UploadFile(ctx context.Context, in *UploadFileRequest) (*FileUploadResponse, error) {
	if in.FileName == "" {
		return nil, errors.New("file name is empty")
	}

	if in.FileBytes == nil {
		return nil, errors.New("file bytes is nil")
	}

	// Use pipe to stream upload without buffering entire file in memory
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	contentType := writer.FormDataContentType()

	// Create request before spawning goroutine to avoid goroutine leak
	// if request creation fails (e.g., cancelled context or invalid URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/files", pipeReader)
	if err != nil {
		pipeReader.Close()
		pipeWriter.Close()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", contentType)

	// Write multipart data in goroutine
	errChan := make(chan error, 1)
	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		// Add file_bytes field
		fileWriter, err := writer.CreateFormFile("file_bytes", in.FileName)
		if err != nil {
			errChan <- fmt.Errorf("failed to create form file: %w", err)
			return
		}

		if _, err := io.Copy(fileWriter, in.FileBytes); err != nil {
			errChan <- fmt.Errorf("failed to copy file data: %w", err)
			return
		}

		// Add labels if provided
		if len(in.Labels) > 0 {
			labelsJSON, err := json.Marshal(in.Labels)
			if err != nil {
				errChan <- fmt.Errorf("failed to marshal labels: %w", err)
				return
			}
			if err := writer.WriteField("labels", string(labelsJSON)); err != nil {
				errChan <- fmt.Errorf("failed to write labels field: %w", err)
				return
			}
		}

		errChan <- nil
	}()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Close pipeReader to unblock the goroutine writing to pipeWriter
		pipeReader.Close()
		<-errChan // Wait for goroutine to finish
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the goroutine encountered an error
	if writeErr := <-errChan; writeErr != nil {
		return nil, writeErr
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Limit to 1MB
		return nil, fmt.Errorf("failed to upload file: unexpected status code (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode upload file response (%d): %w", resp.StatusCode, err)
	}
	return &result, nil
}
