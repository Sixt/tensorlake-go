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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestParseDocumentRemote(t *testing.T) {
	c := initializeTestClient(t)

	tests := []struct {
		req *ParseDocumentRequest
	}{
		{
			req: &ParseDocumentRequest{
				FileSource: FileSource{
					FileURL: "https://www.sixt.de/shared/t-c/sixt_DE_de.pdf",
				},
				ParsingOptions: &ParsingOptions{
					ChunkingStrategy: ChunkingStrategyNone,
				},
				EnrichmentOptions: &EnrichmentOptions{
					TableSummarization: true,
				},
				Labels: map[string]string{"category": "terms-and-conditions"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.req.FileSource.FileURL, func(t *testing.T) {
			func() {
				// Trigger read document operation.
				r, err := c.ParseDocument(t.Context(), tt.req)
				if err != nil {
					t.Fatalf("failed to parse document: %v", err)
				}
				if r == nil {
					t.Fatal("response is nil")
				}
				t.Logf("read document done, parse ID: %s", r.ParseId)

				// Read document status.

				var success bool
				var latestStatus ParseStatus
				var result *ParseResult
			pollLoop:
				for range 10 {
					result, err = c.GetParseResult(t.Context(), r.ParseId)
					if err != nil {
						t.Fatalf("failed to get parse result: %v", err)
					}
					latestStatus = result.Status

					switch latestStatus {
					case ParseStatusSuccessful:
						success = true
						break pollLoop
					case ParseStatusFailure:
						t.Fatalf("Parse failed")
					default:
						t.Logf("Parse status: %s, retrying...", latestStatus)
					}
					time.Sleep(2 * time.Second)
				}
				if !success {
					t.Fatalf("parse still on going... status (%s)", string(latestStatus))
				}
				if len(result.Chunks) == 0 {
					t.Fatalf("no chunks found")
				}
				peak := result.Chunks[0].Content
				if len(peak) > 100 {
					peak = peak[:100]
				}
				t.Logf("parse result: %+v", peak)

				// Validate parse results.
				jobs, err := fetchAllParseJobs(t, c)
				if err != nil {
					t.Fatalf("failed to list parse jobs: %v", err)
				}
				t.Logf("parse jobs: %v", jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}

				// Delete parse job.
				if err := c.DeleteParseJob(t.Context(), r.ParseId); err != nil {
					t.Fatalf("failed to delete parse job: %v", err)
				}

				// Check if parse job is deleted.
				jobs, err = fetchAllParseJobs(t, c)
				if err != nil {
					t.Fatalf("failed to list parse jobs: %v", err)
				}
				if slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not deleted and found in list: %v", jobs)
				}
			}()
		})
	}
}

type Address struct {
	Street  string `json:"street" jsonschema:"Street address"`
	City    string `json:"city" jsonschema:"City"`
	State   string `json:"state" jsonschema:"State/Province code or name"`
	ZipCode string `json:"zip_code" jsonschema:"Postal code"`
}

type BankTransaction struct {
	TransactionDeposit               float64 `json:"transaction_deposit" jsonschema:"Deposit amount"`
	TransactionDepositDate           string  `json:"transaction_deposit_date" jsonschema:"Date of the deposit"`
	TransactionDepositDescription    string  `json:"transaction_deposit_description" jsonschema:"Description of the deposit"`
	TransactionWithdrawal            float64 `json:"transaction_withdrawal" jsonschema:"Withdrawal amount"`
	TransactionWithdrawalDate        string  `json:"transaction_withdrawal_date" jsonschema:"Date of the withdrawal"`
	TransactionWithdrawalDescription string  `json:"transaction_withdrawal_description" jsonschema:"Description of the withdrawal"`
}

type BankStatement struct {
	AccountNumber      string            `json:"account_number" jsonschema:"Bank account number"`
	AccountType        string            `json:"account_type" jsonschema:"Type of the bank account (e.g. Checking/Savings)"`
	BankAddress        Address           `json:"bank_address" jsonschema:"Address of the bank"`
	BankName           string            `json:"bank_name" jsonschema:"Name of the bank"`
	ClientAddress      Address           `json:"client_address" jsonschema:"Address of the client"`
	ClientName         string            `json:"client_name" jsonschema:"Name of the client"`
	EndingBalance      float64           `json:"ending_balance" jsonschema:"Ending balance for the period"`
	StartingBalance    float64           `json:"starting_balance" jsonschema:"Starting balance for the period"`
	StatementDate      string            `json:"statement_date" jsonschema:"Overall statement date if applicable"`
	StatementStartDate string            `json:"statement_start_date" jsonschema:"Start date of the bank statement"`
	StatementEndDate   string            `json:"statement_end_date" jsonschema:"End date of the bank statement"`
	TableItem          []BankTransaction `json:"table_item" jsonschema:"List of transactions in the statement"`
	Others             map[string]any    `json:"others" jsonschema:"Any other additional data from the statement"`
}

func TestParseDocumentStructuredExtraction(t *testing.T) {
	c := initializeTestClient(t)

	tests := []struct {
		filepath string
		req      *ParseDocumentRequest
	}{
		{
			filepath: "testdata/bank_statement.pdf",
			req: &ParseDocumentRequest{
				FileSource: FileSource{
					FileId: "", // fill later.
				},
				ParsingOptions: &ParsingOptions{
					ChunkingStrategy: ChunkingStrategyNone,
				},
				StructuredExtractionOptions: []StructuredExtractionOptions{
					{
						SchemaName: "form125-basic",
						JSONSchema: func() *jsonschema.Schema {
							s, err := jsonschema.For[BankStatement](nil)
							if err != nil {
								t.Fatalf("failed to get JSON schema: %v", err)
							}
							return s
						}(),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			func() {
				// Upload file.
				file, err := os.Open(tt.filepath)
				if err != nil {
					t.Fatalf("failed to open file: %v", err)
				}
				defer file.Close()

				// Upload file.
				resp, err := c.UploadFile(t.Context(), &UploadFileRequest{
					FileBytes: file,
					FileName:  filepath.Base(tt.filepath),
				})
				if err != nil {
					t.Fatalf("failed to upload file: %v", err)
				}
				if resp == nil {
					t.Fatal("response is nil")
				}
				if resp.FileId == "" {
					t.Fatal("file ID is empty")
				}

				tt.req.FileSource.FileId = resp.FileId

				// Parse document.
				r, err := c.ParseDocument(t.Context(), tt.req)
				if err != nil {
					t.Fatalf("failed to parse document: %v", err)
				}
				if r == nil {
					t.Fatal("response is nil")
				}
				t.Logf("parse document done, parse ID: %s", r.ParseId)

				// Get parse result.
				var success bool
				var latestStatus ParseStatus
				var result *ParseResult
			pollLoop:
				for range 10 {
					result, err = c.GetParseResult(t.Context(), r.ParseId)
					if err != nil {
						t.Fatalf("failed to get parse result: %v", err)
					}
					latestStatus = result.Status

					switch latestStatus {
					case ParseStatusSuccessful:
						success = true
						break pollLoop
					case ParseStatusFailure:
						t.Fatalf("Parse failed: %s", result.Error)
					default:
						t.Logf("Parse status: %s, retrying...", latestStatus)
					}
					time.Sleep(2 * time.Second)
				}
				if !success {
					t.Fatalf("parse still on going... status (%s)", string(latestStatus))
				}
				if len(result.Chunks) == 0 {
					t.Fatalf("no chunks found")
				}
				peak := result.Chunks[0].Content
				if len(peak) > 100 {
					peak = peak[:100]
				}
				t.Logf("parse result: %+v", peak)

				// Validate parse results.
				jobs, err := fetchAllParseJobs(t, c)
				if err != nil {
					t.Fatalf("failed to list parse jobs: %v", err)
				}
				t.Logf("parse jobs: %v", jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}

				// Delete file.
				if err := c.DeleteFile(t.Context(), resp.FileId); err != nil {
					t.Fatalf("failed to delete file: %v", err)
				}

				// Delete parse job.
				if err := c.DeleteParseJob(t.Context(), r.ParseId); err != nil {
					t.Fatalf("failed to delete parse job: %v", err)
				}

				// Check if file is deleted.
				files, err := fetchAllFiles(t, c)
				if err != nil {
					t.Fatalf("failed to list files: %v", err)
				}
				if slices.Contains(files, resp.FileId) {
					t.Fatalf("file is not deleted and found in list: %v", files)
				}

				// Check if parse job is deleted.
				jobs, err = fetchAllParseJobs(t, c)
				if err != nil {
					t.Fatalf("failed to list parse jobs: %v", err)
				}
				if slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not deleted and found in list: %v", jobs)
				}
			}()
		})
	}
}

func fetchAllParseJobs(t *testing.T, c *Client) ([]string, error) {
	jobs, cursor := []string{}, ""
	var err error
	for {
		var listResp *PaginationResult[ParseResult]
		listResp, err = c.ListParseJobs(t.Context(), &ListParseJobsRequest{
			Cursor:    cursor,
			Limit:     1,
			Direction: PaginationDirectionNext,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list parse jobs: %v", err)
		}
		if len(listResp.Items) == 0 {
			break
		}
		jobs = append(jobs, listResp.Items[0].ParseId)
		cursor = listResp.NextCursor

		if !listResp.HasMore {
			break
		}
	}
	return jobs, err
}
