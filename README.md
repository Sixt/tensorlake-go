# Tensorlake Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/sixt/tensorlake-go.svg)](https://pkg.go.dev/github.com/sixt/tensorlake-go)

A comprehensive Go SDK for the [Tensorlake API](https://docs.tensorlake.ai/api-reference/v2/introduction), enabling intelligent document processing with parsing, structured data extraction, and page classification capabilities.

## Features

- **Document Parsing**: Convert PDFs, DOCX, images, and more to structured markdown
- **Data Extraction**:  Extract structured data using JSON schemas
- **Page Classification**: Classify pages by content type
- **File Management**: Upload and manage documents
- **Datasets**: Reusable parsing configurations for consistent processing
- **SSE Support**: Real-time progress updates via Server-Sent Events
- **Iterator Pattern**: Easy pagination through results

## Installation

```bash
go get github.com/sixt/tensorlake-go
```

**Requirements:** Go 1.25 or later

## Quick Start

### 1. Initialize the Client

```go
import "github.com/sixt/tensorlake-go"

c := tensorlake.NewClient(
    tensorlake.WithRegion(tensorlake.RegionOnPrem),
    tensorlake.WithBaseURL("https://api.your-domain.com"),
    tensorlake.WithAPIKey("your-api-key"),
)
```

### 2. Upload a File

```go
file, _ := os.Open("document.pdf")
defer file.Close()

uploadResp, _ := c.UploadFile(context.Background(), &tensorlake.UploadFileRequest{
    FileBytes: file,
    FileName:  "document.pdf",
    Labels:    map[string]string{"category": "invoice"},
})

fmt.Printf("File uploaded: %s\n", uploadResp.FileId)
```

### 3. Parse the Document

```go
parseJob, _ := c.ParseDocument(context.Background(), &tensorlake.ParseDocumentRequest{
    FileSource: tensorlake.FileSource{
        FileId: uploadResp.FileId,
    },
})

// Get results with real-time updates
result, _ := c.GetParseResult(
    context.Background(),
    parseJob.ParseId,
    tensorlake.WithSSE(true),
    tensorlake.WithOnUpdate(func(name tensorlake.ParseEventName, r *tensorlake.ParseResult) {
        fmt.Printf("Status: %s - %d/%d pages\n", name, r.ParsedPagesCount, r.TotalPages)
    }),
)

// Access parsed content
for _, page := range result.Pages {
    fmt.Printf("Page %d:\n", page.PageNumber)
    // Process page content...
}
```

## Documentation

### Core APIs

- **[File Management APIs](./docs/file-apis.md)** - Upload, list, retrieve metadata, and delete files
- **[Parse APIs](./docs/parse-apis.md)** - Parse documents, extract data, and classify pages
- **[Dataset APIs](./docs/dataset-apis.md)** - Create reusable parsing configurations

### Comprehensive Examples

#### Extract Structured Data

```go
import "github.com/google/jsonschema-go/jsonschema"

// Define extraction schema
type InvoiceData struct {
    InvoiceNumber string     `json:"invoice_number"`
    VendorName    string     `json:"vendor_name"`
    TotalAmount   float64    `json:"total_amount"`
    LineItems     []LineItem `json:"line_items"`
}

type LineItem struct {
    Description string  `json:"description"`
    Amount      float64 `json:"amount"`
}

schema, _ := jsonschema.For[InvoiceData](nil)

// Parse with extraction
parseJob, _ := c.ParseDocument(context.Background(), &tensorlake.ParseDocumentRequest{
    FileSource: tensorlake.FileSource{FileId: fileId},
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName:        "invoice_data",
            JSONSchema:        schema,
            PartitionStrategy: tensorlake.PartitionStrategyNone,
            ProvideCitations:  true,
        },
    },
})

// Retrieve and unmarshal extracted data
result, _ := c.GetParseResult(context.Background(), parseJob.ParseId)
for _, data := range result.StructuredData {
    var extracted map[string]interface{}
    json.Unmarshal(data.Data, &extracted)
    fmt.Printf("Extracted: %+v\n", extracted)
}
```

#### Classify Pages

```go
parseJob, err := c.ClassifyDocument(context.Background(), &tensorlake.ClassifyDocumentRequest{
    FileSource: tensorlake.FileSource{FileId: fileId},
    PageClassifications: []tensorlake.PageClassConfig{
        {
            Name:        "signature_page",
            Description: "Pages containing signatures or signature blocks",
        },
        {
            Name:        "terms_and_conditions",
            Description: "Pages with legal terms and conditions",
        },
    },
})

result, _ := c.GetParseResult(context.Background(), parseJob.ParseId)
for _, pageClass := range result.PageClasses {
    fmt.Printf("Class '%s' found on pages: %v\n", pageClass.PageClass, pageClass.PageNumbers)
}
```

#### Use Datasets for Batch Processing

```go
// Create a reusable dataset
dataset, err := c.CreateDataset(context.Background(), &tensorlake.CreateDatasetRequest{
    Name:        "invoice-processing",
    Description: "Standard invoice parsing configuration",
    ParsingOptions: &tensorlake.ParsingOptions{
        TableOutputMode: tensorlake.TableOutputModeMarkdown,
    },
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName: "invoice",
            JSONSchema: schema,
        },
    },
})

// Process multiple files with the same configuration
fileIds := []string{"file_001", "file_002", "file_003"}
for _, fileId := range fileIds {
    parseJob, err := c.ParseDataset(context.Background(), &tensorlake.ParseDatasetRequest{
        DatasetId:  dataset.DatasetId,
        FileSource: tensorlake.FileSource{FileId: fileId},
    })
    // Process results...
}
```

## Advanced Features

### Server-Sent Events (SSE)

Get real-time progress updates for long-running parse jobs:

```go
result, err := c.GetParseResult(
    ctx,
    parseId,
    tensorlake.WithSSE(true),
    tensorlake.WithOnUpdate(func(name tensorlake.ParseEventName, r *tensorlake.ParseResult) {
        switch eventName {
        case tensorlake.sseEventParseQueued:
            fmt.Println("Job queued")
        case tensorlake.sseEventParseUpdate:
            fmt.Printf("Progress: %d/%d pages\n", r.ParsedPagesCount, r.TotalPages)
        case tensorlake.sseEventParseDone:
            fmt.Println("Complete!")
        }
    }),
)
```

### Iterator Pattern

Easily iterate through paginated results:

```go
// Iterate all files
for file, err := range c.IterFiles(ctx, 50, tensorlake.PaginationDirectionNext) {
    if err != nil {
        panic(err)
    }
    fmt.Printf("File: %s\n", file.FileName)
}

// Iterate all parse jobs
for job, err := range c.IterParseJobs(ctx, 50, tensorlake.PaginationDirectionNext) {
    if err != nil {
        panic(err)
    }
    fmt.Printf("Job %s: Status: %s\n", job.ParseId, job.Status)
}

// Iterate all datasets
for dataset, err := range c.IterDatasets(ctx, 50, tensorlake.PaginationDirectionNext) {
    if err != nil {
        panic(err)
    }
    fmt.Printf("Dataset %s: Name: %s, Status: %s\n", dataset.DatasetId, dataset.Name, dataset.Status)
}
```

## Supported File Types

- **Documents**: PDF, DOCX
- **Spreadsheets**: XLS, XLSX, XLSM, CSV
- **Presentations**: PPTX, Apple Keynote
- **Images**: PNG, JPG, JPEG
- **Text**: Plain text, HTML

Maximum file size: 1 GB

## Error Handling

All API methods return structured errors:

```go
result, err := c.ParseDocument(ctx, request)
if err != nil {
    var apiErr *tensorlake.ErrorResponse
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s (Code: %s)\n", apiErr.Message, apiErr.ErrorCode)
        // Handle specific error codes
    } else {
        fmt.Printf("Network/Client Error: %v\n", err)
    }
}
```

## Best Practices

1. **Reuse Datasets** - Create datasets for frequently processed document types
2. **Use SSE** - Enable SSE for large documents to track progress
3. **Batch Processing** - Process similar documents with the same dataset configuration
4. **Error Handling** - Always check error responses and handle retries appropriately
5. **Labels** - Use labels to organize and filter files and parse jobs
6. **Iterators** - Use iterator methods for efficient pagination through large result sets

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Related Resources

- [Tensorlake API Documentation](https://docs.tensorlake.ai/)
- [API Reference](https://docs.tensorlake.ai/api-reference/v2/introduction)
- [Go Package Documentation](https://pkg.go.dev/github.com/sixt/tensorlake-go)

## License

Copyright 2025 SIXT SE. Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

<a href="https://www.sixt.com">
    <picture>
        <source media="(prefers-color-scheme: dark)" srcset=".github/sixt_dark.png">
        <source media="(prefers-color-scheme: light)" srcset=".github/sixt_light.png">
        <img width="100px" alt="Sixt logo" src=".github/sixt_dark.png">
    </picture>
</a>

