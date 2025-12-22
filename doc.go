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

// Package tensorlake provides a Go SDK for the Tensorlake API.
//
// Tensorlake enables document parsing, structured data extraction, and page
// classification for various document formats including PDF, DOCX, PPTX, images,
// and more.
//
// # Getting Started
//
// Create a client with your API key:
//
//	c := tensorlake.NewClient(
//		tensorlake.WithBaseURL("https://api.your-domain.com"),
//		tensorlake.WithAPIKey("your-api-key"),
//	)
//
// # Uploading a File
//
// Upload a file to the project:
//
//	file, err := os.Open("path/to/your/file.pdf")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer file.Close()
//
//	r, err := c.UploadFile(context.Background(), &tensorlake.UploadFileRequest{
//		FileBytes: file,
//		FileName:  "your-file.pdf",
//		Labels:    map[string]string{"category": "label-1", "subcategory": "label-2"},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Parsing a Document
//
// Parse an uploaded file and retrieve the results:
//
//	// Start parsing using the file ID from upload
//	parseJob, err := c.ParseDocument(context.Background(), &tensorlake.ParseDocumentRequest{
//		FileSource: tensorlake.FileSource{
//			FileId: r.FileId,
//		},
//		Labels: map[string]string{"type": "invoice"},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Retrieve parse results with streaming updates
//	result, err := c.GetParseResult(context.Background(), parseJob.ParseId,
//		tensorlake.WithSSE(true),
//		tensorlake.WithOnUpdate(func(eventName string, r *tensorlake.ParseResult) {
//			log.Printf("Parse status: %s", eventName)
//		}),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access the parsed content
//	for _, page := range result.Pages {
//		log.Printf("Page %d: %s", page.PageNumber, page.Markdown)
//	}
package tensorlake
