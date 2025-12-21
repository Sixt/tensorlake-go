# Parse APIs

The Parse APIs provide comprehensive document processing capabilities including parsing, structured data extraction, and page classification.

## Table of Contents

- [Overview](#overview)
- [Parse Document](#parse-document)
- [Read Document](#read-document)
- [Extract Document](#extract-document)
- [Classify Document](#classify-document)
- [Get Parse Result](#get-parse-result)
- [List Parse Jobs](#list-parse-jobs)
- [Delete Parse Job](#delete-parse-job)
- [Configuration Options](#configuration-options)

## Overview

Tensorlake offers multiple parsing operations:

- **Parse Document**: Comprehensive parsing with optional extraction and classification
- **Read Document**: Basic document parsing to markdown
- **Extract Document**: Structured data extraction using JSON schemas
- **Classify Document**: Page classification based on content

All parsing operations follow the same pattern:
1. Submit a parse request → receive a `ParseJob` with `parse_id`
2. Use the `parse_id` to query results via `GetParseResult`
3. Optionally use SSE (Server-Sent Events) for real-time progress updates

---

## Parse Document

Submit a document for comprehensive parsing, including reading, extraction, and classification in a single operation.

### Method

```go
func (c *Client) ParseDocument(ctx context.Context, in *ParseDocumentRequest) (*ParseJob, error)
```

### Request Parameters

```go
type ParseDocumentRequest struct {
    FileSource  // One of: FileId, FileURL, or RawText (required)

    // Optional configuration
    ParsingOptions              *ParsingOptions
    EnrichmentOptions           *EnrichmentOptions
    StructuredExtractionOptions []StructuredExtractionOptions
    PageClassificationOptions   []PageClassConfig

    // Page range to parse (e.g., "1-5,8,10")
    PageRange string

    // Additional metadata
    Labels   map[string]string
    MimeType MimeType
}

// FileSource - exactly one must be provided
type FileSource struct {
    FileId  string  // File ID from UploadFile
    FileURL string  // Internet-reachable URL
    RawText string  // Plain text content
}
```

### Response

```go
type ParseJob struct {
    ParseId   string  // Unique identifier for tracking
    CreatedAt string  // RFC 3339 timestamp
}
```

### Example

```go
// Parse an uploaded file with extraction
parseJob, err := client.ParseDocument(context.Background(), &tensorlake.ParseDocumentRequest{
    FileSource: tensorlake.FileSource{
        FileId: "file_abc123",
    },
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName: "invoice_data",
            JSONSchema: &jsonschema.Schema{
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "invoice_number": {Type: "string"},
                    "total_amount":   {Type: "number"},
                    "date":          {Type: "string"},
                },
            },
        },
    },
    Labels: map[string]string{"type": "invoice"},
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Parse job created: %s\n", parseJob.ParseId)
```

---

## Read Document

Submit a document for basic parsing to markdown format.

### Method

```go
func (c *Client) ReadDocument(ctx context.Context, in *ReadDocumentRequest) (*ParseJob, error)
```

### Request Parameters

```go
type ReadDocumentRequest struct {
    FileSource  // One of: FileId, FileURL, or RawText

    ParsingOptions    *ParsingOptions
    EnrichmentOptions *EnrichmentOptions
    PageRange         string
    Labels            map[string]string
    FileName          string  // Only used with FileId
    MimeType          MimeType
}
```

### Example

```go
// Read a PDF from URL
parseJob, err := client.ReadDocument(context.Background(), &tensorlake.ReadDocumentRequest{
    FileSource: tensorlake.FileSource{
        FileURL: "https://example.com/document.pdf",
    },
    ParsingOptions: &tensorlake.ParsingOptions{
        IncludeImages: true,
        TableOutputMode: tensorlake.TableOutputModeMarkdown,
    },
})
if err != nil {
    log.Fatal(err)
}
```

---

## Extract Document

Submit a document for structured data extraction using JSON schemas.

### Method

```go
func (c *Client) ExtractDocument(ctx context.Context, in *ExtractDocumentRequest) (*ParseJob, error)
```

### Request Parameters

```go
type ExtractDocumentRequest struct {
    FileSource  // One of: FileId, FileURL, or RawText

    // At least one extraction schema required
    StructuredExtractionOptions []StructuredExtractionOptions

    PageRange string
    MimeType  string
    Labels    map[string]string
}

type StructuredExtractionOptions struct {
    SchemaName        string              // Name for this schema
    JSONSchema        *jsonschema.Schema  // JSON schema definition
    PartitionStrategy PartitionStrategy   // How to partition document
    ModelProvider     ModelProvider       // LLM provider to use
    PageClasses       []string            // Filter by page classes
    Prompt            string              // Custom extraction prompt
    ProvideCitations  bool                // Include bounding boxes
    SkipOCR           bool                // Skip OCR processing
}
```

### Example

```go
// Extract invoice data
parseJob, err := client.ExtractDocument(context.Background(), &tensorlake.ExtractDocumentRequest{
    FileSource: tensorlake.FileSource{
        FileId: "file_abc123",
    },
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName: "invoice",
            JSONSchema: &jsonschema.Schema{
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "vendor_name":    {Type: "string"},
                    "invoice_number": {Type: "string"},
                    "total_amount":   {Type: "number"},
                    "line_items": {
                        Type: "array",
                        Items: &jsonschema.Schema{
                            Type: "object",
                            Properties: map[string]*jsonschema.Schema{
                                "description": {Type: "string"},
                                "amount":      {Type: "number"},
                            },
                        },
                    },
                },
            },
            PartitionStrategy: tensorlake.PartitionStrategyNone,
            ProvideCitations:  true,
        },
    },
})
if err != nil {
    log.Fatal(err)
}
```

---

## Classify Document

Submit a document for page classification.

### Method

```go
func (c *Client) ClassifyDocument(ctx context.Context, in *ClassifyDocumentRequest) (*ParseJob, error)
```

### Request Parameters

```go
type ClassifyDocumentRequest struct {
    FileSource  // One of: FileId, FileURL, or RawText

    // At least one classification config required
    PageClassifications []PageClassConfig

    PageRange string
    MimeType  string
    Labels    map[string]string
}

type PageClassConfig struct {
    Name        string  // Class name
    Description string  // What to look for in pages
}
```

### Example

```go
// Classify pages in a legal document
parseJob, err := client.ClassifyDocument(context.Background(), &tensorlake.ClassifyDocumentRequest{
    FileSource: tensorlake.FileSource{
        FileId: "file_abc123",
    },
    PageClassifications: []tensorlake.PageClassConfig{
        {
            Name:        "signature_page",
            Description: "Pages containing signatures or signature blocks",
        },
        {
            Name:        "terms_and_conditions",
            Description: "Pages with legal terms and conditions text",
        },
        {
            Name:        "exhibits",
            Description: "Appendix or exhibit pages",
        },
    },
})
if err != nil {
    log.Fatal(err)
}
```

---

## Get Parse Result

Retrieve the result of a parse job, with optional SSE streaming for real-time updates.

### Method

```go
func (c *Client) GetParseResult(ctx context.Context, parseId string, opts ...GetParseResultOption) (*ParseResult, error)
```

### Options

```go
// Enable Server-Sent Events for streaming updates
func WithSSE(enable bool) GetParseResultOption

// Callback for intermediate updates during SSE
func WithOnUpdate(onUpdate ParseResultUpdateFunc) GetParseResultOption

type ParseResultUpdateFunc func(name ParseEventName, result *ParseResult)
```

### Response

```go
type ParseResult struct {
    // Job metadata
    ParseId          string
    Status           ParseStatus
    ParsedPagesCount int
    TotalPages       int
    CreatedAt        string
    FinishedAt       string
    Error            string
    Labels           map[string]string
    DatasetId        string

    // Parsed content
    Pages          []Page            // Page-by-page content
    Chunks         []Chunk           // Text chunks (if chunking enabled)
    PageClasses    []PageClass       // Classification results
    StructuredData []StructuredData  // Extracted structured data
}
```

### Example: Basic Retrieval

```go
result, err := client.GetParseResult(context.Background(), parseJob.ParseId)
if err != nil {
    log.Fatal(err)
}

if result.Status == tensorlake.ParseStatusCompleted {
    fmt.Printf("Parsed %d pages\n", result.ParsedPagesCount)

    // Access page content
    for _, page := range result.Pages {
        fmt.Printf("Page %d:\n", page.PageNumber)
        for _, fragment := range page.PageFragments {
            fmt.Printf("  Type: %s\n", fragment.FragmentType)
        }
    }
}
```

### Example: SSE Streaming

```go
result, err := client.GetParseResult(
    context.Background(),
    parseJob.ParseId,
    tensorlake.WithSSE(true),
    tensorlake.WithOnUpdate(func(name tensorlake.ParseEventName, r *tensorlake.ParseResult) {
        switch eventName {
        case tensorlake.sseEventParseQueued:
            fmt.Println("Parse job queued")
        case tensorlake.sseEventParseUpdate:
            fmt.Printf("Processing... %d/%d pages\n", r.ParsedPagesCount, r.TotalPages)
        case tensorlake.sseEventParseDone:
            fmt.Println("Parse complete!")
        case tensorlake.sseEventParseFailed:
            fmt.Printf("Parse failed: %s\n", r.Error)
        }
    }),
)
if err != nil {
    log.Fatal(err)
}

// Process final result
fmt.Printf("Final status: %s\n", result.Status)
```

### Example: Access Structured Data

```go
result, err := client.GetParseResult(context.Background(), parseJob.ParseId)
if err != nil {
    log.Fatal(err)
}

for _, data := range result.StructuredData {
    fmt.Printf("Schema: %s\n", data.SchemaName)
    fmt.Printf("Pages: %v\n", data.PageNumbers)

    // Unmarshal the extracted data
    var extracted map[string]interface{}
    if err := json.Unmarshal(data.Data, &extracted); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Extracted data: %+v\n", extracted)
}
```

---

## List Parse Jobs

List all parse jobs in your project with pagination.

### Method

```go
func (c *Client) ListParseJobs(ctx context.Context, in *ListParseJobsRequest) (*PaginationResult[ParseResult], error)
```

### Request Parameters

```go
type ListParseJobsRequest struct {
    Cursor        string
    Direction     PaginationDirection
    Limit         int
    FileId        string  // Filter by file ID
    CreatedAfter  string  // RFC 3339 timestamp
    CreatedBefore string  // RFC 3339 timestamp
}
```

### Example

```go
// List recent parse jobs
response, err := client.ListParseJobs(context.Background(), &tensorlake.ListParseJobsRequest{
    Limit:     20,
    Direction: tensorlake.PaginationDirectionNext,
})
if err != nil {
    log.Fatal(err)
}

for _, job := range response.Items {
    fmt.Printf("Parse ID: %s, Status: %s, Pages: %d\n",
        job.ParseId, job.Status, job.ParsedPagesCount)
}
```

### Iterate All Parse Jobs

```go
func (c *Client) IterParseJobs(ctx context.Context, batchSize int) iter.Seq2[ParseResult, error]
```

**Example:**

```go
for job, err := range client.IterParseJobs(context.Background(), 50) {
    if err != nil {
        log.Fatal(err)
    }
    if job.Status == tensorlake.ParseStatusFailed {
        fmt.Printf("Failed job: %s - %s\n", job.ParseId, job.Error)
    }
}
```

---

## Delete Parse Job

Delete a previously submitted parse job.

### Method

```go
func (c *Client) DeleteParseJob(ctx context.Context, parseId string) error
```

### Example

```go
err := client.DeleteParseJob(context.Background(), "parse_abc123")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Parse job deleted successfully")
```

### Notes

- Deleting a parse job removes the job and its settings
- Does not delete the original file used for parsing
- Does not affect other parse jobs from the same file

---

## Configuration Options

### ParsingOptions

Configure how documents are parsed:

```go
type ParsingOptions struct {
    ChunkingStrategy         ChunkingStrategy     // How to chunk document
    CrossPageHeaderDetection bool                 // Detect headers across pages
    DisableLayoutDetection   bool                 // Skip layout detection for speed
    OCRModel                 OCRPipelineProvider  // OCR model to use
    RemoveStrikethroughLines bool                 // Remove strikethrough text
    SignatureDetection       bool                 // Detect signatures
    SkewDetection            bool                 // Correct skewed/rotated pages
    TableOutputMode          TableOutputMode      // HTML or Markdown
    TableParsingFormat       TableParsingFormat   // Table extraction method
    IgnoreSections           []PageFragmentType   // Skip certain content types
    IncludeImages            bool                 // Include images in markdown
    BarcodeDetection         bool                 // Detect barcodes
}
```

**Example:**

```go
parsingOpts := &tensorlake.ParsingOptions{
    IncludeImages:            true,
    TableOutputMode:          tensorlake.TableOutputModeMarkdown,
    SignatureDetection:       true,
    CrossPageHeaderDetection: true,
}
```

### EnrichmentOptions

Enhance parsed content with AI-generated summaries:

```go
type EnrichmentOptions struct {
    FigureSummarization       bool   // Summarize figures
    FigureSummarizationPrompt string // Custom prompt for figures
    TableSummarization        bool   // Summarize tables
    TableSummarizationPrompt  string // Custom prompt for tables
    IncludeFullPageImage      bool   // Include full page for context
}
```

**Example:**

```go
enrichmentOpts := &tensorlake.EnrichmentOptions{
    TableSummarization:        true,
    TableSummarizationPrompt:  "Summarize this financial table in 2-3 sentences",
    IncludeFullPageImage:      true,
}
```

### Common Enums

#### ParseStatus

```go
const (
    ParseStatusPending    ParseStatus = "pending"
    ParseStatusQueued     ParseStatus = "queued"
    ParseStatusProcessing ParseStatus = "processing"
    ParseStatusCompleted  ParseStatus = "completed"
    ParseStatusFailed     ParseStatus = "failed"
)
```

#### ChunkingStrategy

```go
const (
    ChunkingStrategyNone     ChunkingStrategy = "none"
    ChunkingStrategyPage     ChunkingStrategy = "page"
    ChunkingStrategySemantic ChunkingStrategy = "semantic"
)
```

#### TableOutputMode

```go
const (
    TableOutputModeHTML     TableOutputMode = "html"
    TableOutputModeMarkdown TableOutputMode = "markdown"
)
```

#### PartitionStrategy

```go
const (
    PartitionStrategyNone PartitionStrategy = "none"  // One result for entire doc
    PartitionStrategyPage PartitionStrategy = "page"  // One result per page
)
```

---

## Best Practices

1. **Use SSE for Long Documents**: Enable SSE streaming to get progress updates for large documents
2. **Optimize with Parsing Options**: Disable unnecessary features (e.g., `DisableLayoutDetection`) for faster processing
3. **Structured Extraction**: Use specific, well-defined JSON schemas for better extraction accuracy
4. **Page Classification**: Provide detailed descriptions in `PageClassConfig` to improve classification accuracy
5. **Error Handling**: Always check the `Status` field and handle failed parse jobs appropriately
6. **Pagination**: Use `IterParseJobs` for easier iteration through all jobs

---

For more information, see:
- [File Management APIs](./file-apis.md)
- [Dataset APIs](./dataset-apis.md)
- [Main Documentation](../README.md)

