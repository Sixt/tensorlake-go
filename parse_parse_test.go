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
	"os"
	"path/filepath"
	"slices"
	"testing"

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

				result, err := c.GetParseResult(t.Context(), r.ParseId, WithSSE(true), WithOnUpdate(func(name ParseEventName, _ *ParseResult) {
					t.Logf("parse status: %s", name)
				}))
				if err != nil {
					t.Fatalf("failed to get parse result: %v", err)
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
				jobs := []string{}
				for j, err := range c.IterParseJobs(t.Context(), 1) {
					if err != nil {
						t.Fatalf("failed to list parse jobs: %v", err)
					}
					jobs = append(jobs, j.ParseId)
				}
				t.Logf("listed %d parse jobs: %v", len(jobs), jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}

				testCleanupFileAndParseJob(t, c, "", r.ParseId)
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
				result, err := c.GetParseResult(t.Context(), r.ParseId, WithSSE(true), WithOnUpdate(func(name ParseEventName, _ *ParseResult) {
					t.Logf("parse status: %s", name)
				}))
				if err != nil {
					t.Fatalf("failed to get parse result: %v", err)
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
				jobs := []string{}
				for j, err := range c.IterParseJobs(t.Context(), 1) {
					if err != nil {
						t.Fatalf("failed to list parse jobs: %v", err)
					}
					jobs = append(jobs, j.ParseId)
				}
				t.Logf("listed %d parse jobs: %v", len(jobs), jobs)
				if !slices.Contains(jobs, r.ParseId) {
					t.Fatalf("parse job is not found in list: %v", jobs)
				}

				testCleanupFileAndParseJob(t, c, resp.FileId, r.ParseId)
			}()
		})
	}
}
