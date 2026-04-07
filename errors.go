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

import "fmt"

// ErrorResponse represents an error returned by the Tensorlake API.
type ErrorResponse struct {
	// Message is a human-readable error message.
	Message string `json:"message"`
	// Code is the error code for programmatic handling.
	Code ErrorCode `json:"code"`
	// Timestamp is the Unix epoch timestamp in milliseconds when the error occurred.
	Timestamp int64 `json:"timestamp,omitempty"`
	// TraceId is the trace ID of the error.
	TraceId string `json:"trace_id,omitempty"`
	// Details is the details of the error.
	Details any `json:"details,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("code: %s, message: %s, trace_id: %s, details: %v", e.Code, e.Message, e.TraceId, e.Details)
}

// ErrorCode represents error codes for Document AI API.
//
// These codes are used to identify specific error conditions in the API.
// They can be used for programmatic handling of errors.
type ErrorCode string

const (
	ErrorCodeQuotaExceeded        ErrorCode = "QUOTA_EXCEEDED"
	ErrorCodeInvalidJSONSchema    ErrorCode = "INVALID_JSON_SCHEMA"
	ErrorCodeInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
	ErrorCodeInvalidPageClass     ErrorCode = "INVALID_PAGE_CLASSIFICATION"
	ErrorCodeEntityNotFound       ErrorCode = "ENTITY_NOT_FOUND"
	ErrorCodeEntityAlreadyExists  ErrorCode = "ENTITY_ALREADY_EXISTS"
	ErrorCodeInvalidFile          ErrorCode = "INVALID_FILE"
	ErrorCodeInvalidPageRange     ErrorCode = "INVALID_PAGE_RANGE"
	ErrorCodeInvalidMimeType      ErrorCode = "INVALID_MIME_TYPE"
	ErrorCodeInvalidDatasetName   ErrorCode = "INVALID_DATASET_NAME"
	ErrorCodeInternalError        ErrorCode = "INTERNAL_ERROR"
	ErrorCodeInvalidMultipart     ErrorCode = "INVALID_MULTIPART"
	ErrorCodeMultipartStreamEnd   ErrorCode = "MULTIPART_STREAM_END"
	ErrorCodeInvalidQueryParams   ErrorCode = "INVALID_QUERY_PARAMS"
	ErrorCodeInvalidJobState      ErrorCode = "INVALID_JOB_STATE"
	ErrorCodeClientDisconnect     ErrorCode = "CLIENT_DISCONNECT"
	ErrorCodeInvalidID            ErrorCode = "INVALID_ID"
)
