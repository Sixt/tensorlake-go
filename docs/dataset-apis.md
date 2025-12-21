# Dataset APIs

Datasets in Tensorlake allow you to define reusable parsing configurations that can be applied to multiple documents. This is useful when you have a specific document type (e.g., invoices, contracts) that requires consistent processing settings.

## Table of Contents

- [Overview](#overview)
- [Create Dataset](#create-dataset)
- [Get Dataset](#get-dataset)
- [Update Dataset](#update-dataset)
- [List Datasets](#list-datasets)
- [Delete Dataset](#delete-dataset)
- [Parse with Dataset](#parse-with-dataset)

## Overview

A dataset encapsulates:
- **Parsing options**: How to parse documents
- **Structured extraction schemas**: What data to extract
- **Page classifications**: How to classify pages
- **Enrichment settings**: Additional AI-powered enhancements

Once created, you can parse multiple documents using the same dataset configuration, ensuring consistency across similar document types.

---

## Create Dataset

Create a new dataset with specific parsing configuration.

### Method

```go
func (c *Client) CreateDataset(ctx context.Context, in *CreateDatasetRequest) (*CreateDatasetResponse, error)
```

### Request Parameters

```go
type CreateDatasetRequest struct {
    // Name of the dataset (required, must be unique)
    // Can only contain alphanumeric characters, hyphens, and underscores
    Name string

    // Optional description
    Description string

    // Parsing configuration (optional)
    ParsingOptions              *ParsingOptions
    StructuredExtractionOptions []StructuredExtractionOptions
    PageClassifications         []PageClassConfig
    EnrichmentOptions           *EnrichmentOptions
}
```

### Response

```go
type CreateDatasetResponse struct {
    Name      string  // Name of the dataset
    DatasetId string  // Unique dataset identifier
    CreatedAt string  // Creation timestamp (RFC 3339)
}
```

### Example

```go
// Create a dataset for invoice processing
dataset, err := client.CreateDataset(context.Background(), &tensorlake.CreateDatasetRequest{
    Name:        "invoice-processing",
    Description: "Standard invoice parsing with data extraction",
    ParsingOptions: &tensorlake.ParsingOptions{
        IncludeImages:       false,
        TableOutputMode:     tensorlake.TableOutputModeMarkdown,
        SignatureDetection:  true,
    },
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName: "invoice_data",
            JSONSchema: &jsonschema.Schema{
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "invoice_number": {Type: "string"},
                    "vendor_name":    {Type: "string"},
                    "total_amount":   {Type: "number"},
                    "invoice_date":   {Type: "string"},
                    "due_date":       {Type: "string"},
                    "line_items": {
                        Type: "array",
                        Items: &jsonschema.Schema{
                            Type: "object",
                            Properties: map[string]*jsonschema.Schema{
                                "description": {Type: "string"},
                                "quantity":    {Type: "number"},
                                "unit_price":  {Type: "number"},
                                "total":       {Type: "number"},
                            },
                        },
                    },
                },
                Required: []string{"invoice_number", "vendor_name", "total_amount"},
            },
            PartitionStrategy: tensorlake.PartitionStrategyNone,
            ProvideCitations:  true,
        },
    },
    PageClassifications: []tensorlake.PageClassConfig{
        {
            Name:        "invoice_page",
            Description: "Pages containing invoice information with line items",
        },
        {
            Name:        "payment_terms",
            Description: "Pages with payment terms and conditions",
        },
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dataset created: %s (ID: %s)\n", dataset.Name, dataset.DatasetId)
```

---

## Get Dataset

Retrieve details for a specific dataset.

### Method

```go
func (c *Client) GetDataset(ctx context.Context, datasetId string) (*Dataset, error)
```

### Parameters

- `datasetId`: The unique identifier of the dataset

### Response

```go
type Dataset struct {
    Name        string        // Dataset name
    DatasetId   string        // Unique identifier
    Description string        // Dataset description
    Status      DatasetStatus // Current status
    CreatedAt   string        // Creation timestamp
    UpdatedAt   string        // Last update timestamp
}
```

### Example

```go
dataset, err := client.GetDataset(context.Background(), "dataset_abc123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dataset: %s\n", dataset.Name)
fmt.Printf("Status: %s\n", dataset.Status)
fmt.Printf("Description: %s\n", dataset.Description)
fmt.Printf("Created: %s\n", dataset.CreatedAt)
fmt.Printf("Updated: %s\n", dataset.UpdatedAt)
```

---

## Update Dataset

Update an existing dataset's configuration.

### Method

```go
func (c *Client) UpdateDataset(ctx context.Context, in *UpdateDatasetRequest) (*Dataset, error)
```

### Request Parameters

```go
type UpdateDatasetRequest struct {
    DatasetId string  // Required: which dataset to update

    // All fields are optional - only provide what you want to update
    Description                 string
    ParsingOptions              *ParsingOptions
    StructuredExtractionOptions []StructuredExtractionOptions
    PageClassifications         []PageClassConfig
    EnrichmentOptions           *EnrichmentOptions
}
```

### Response

Returns the updated `Dataset` object.

### Example

```go
// Update dataset to include table summarization
updatedDataset, err := client.UpdateDataset(context.Background(), &tensorlake.UpdateDatasetRequest{
    DatasetId:   "dataset_abc123",
    Description: "Invoice processing with enhanced table summaries",
    EnrichmentOptions: &tensorlake.EnrichmentOptions{
        TableSummarization:       true,
        TableSummarizationPrompt: "Summarize line items with totals",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dataset updated: %s\n", updatedDataset.Name)
```

---

## List Datasets

List all datasets in your project with cursor-based pagination.

### Method

```go
func (c *Client) ListDatasets(ctx context.Context, in *ListDatasetsRequest) (*PaginationResult[Dataset], error)
```

### Request Parameters

```go
type ListDatasetsRequest struct {
    Cursor    string              // Pagination cursor
    Direction PaginationDirection // "next" or "prev"
    Limit     int                 // Maximum results per page
    Name      string              // Filter by name (substring match)
}
```

### Response

```go
type PaginationResult[Dataset] struct {
    Items      []Dataset  // Array of datasets
    HasMore    bool       // More results available
    NextCursor string     // Cursor for next page
    PrevCursor string     // Cursor for previous page
}
```

### Example

```go
// List first 10 datasets
response, err := client.ListDatasets(context.Background(), &tensorlake.ListDatasetsRequest{
    Limit:     10,
    Direction: tensorlake.PaginationDirectionNext,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d datasets:\n", len(response.Items))
for _, dataset := range response.Items {
    fmt.Printf("  - %s (ID: %s) - %s\n",
        dataset.Name, dataset.DatasetId, dataset.Status)
}

// Paginate through results
if response.HasMore {
    nextPage, err := client.ListDatasets(context.Background(), &tensorlake.ListDatasetsRequest{
        Cursor:    response.NextCursor,
        Limit:     10,
        Direction: tensorlake.PaginationDirectionNext,
    })
    // ... process next page
}
```

### Iterate All Datasets

For convenience, use the `IterDatasets` method:

```go
func (c *Client) IterDatasets(ctx context.Context, batchSize int) iter.Seq2[Dataset, error]
```

**Example:**

```go
for dataset, err := range client.IterDatasets(context.Background(), 50) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Dataset: %s - %s\n", dataset.Name, dataset.Description)
}
```

---

## Delete Dataset

Delete a dataset from Tensorlake.

### Method

```go
func (c *Client) DeleteDataset(ctx context.Context, datasetId string) error
```

### Parameters

- `datasetId`: The unique identifier of the dataset to delete

### Example

```go
err := client.DeleteDataset(context.Background(), "dataset_abc123")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Dataset deleted successfully")
```

### Notes

- Deleting a dataset is permanent
- Does not delete files or parse jobs created using the dataset
- You cannot undo this operation

---

## Parse with Dataset

Parse a document using a dataset's predefined configuration.

### Method

```go
func (c *Client) ParseDataset(ctx context.Context, in *ParseDatasetRequest) (*ParseJob, error)
```

### Request Parameters

```go
type ParseDatasetRequest struct {
    DatasetId string      // Required: which dataset to use
    FileSource            // One of: FileId, FileURL, or RawText (required)

    // Optional overrides
    PageRange string
    FileName  string  // Only used with FileId
    MimeType  MimeType
    Labels    map[string]string
}
```

### Response

```go
type ParseJob struct {
    ParseId   string  // Unique identifier for tracking
    CreatedAt string  // RFC 3339 timestamp
}
```

### Example: Parse Multiple Files with Same Configuration

```go
// Create a dataset for contracts
dataset, err := client.CreateDataset(context.Background(), &tensorlake.CreateDatasetRequest{
    Name:        "legal-contracts",
    Description: "Parse legal contracts and extract key terms",
    ParsingOptions: &tensorlake.ParsingOptions{
        SignatureDetection:       true,
        CrossPageHeaderDetection: true,
    },
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName: "contract_terms",
            JSONSchema: &jsonschema.Schema{
                Type: "object",
                Properties: map[string]*jsonschema.Schema{
                    "parties":        {Type: "array", Items: &jsonschema.Schema{Type: "string"}},
                    "effective_date": {Type: "string"},
                    "term_length":    {Type: "string"},
                    "payment_terms":  {Type: "string"},
                },
            },
        },
    },
    PageClassifications: []tensorlake.PageClassConfig{
        {
            Name:        "signature_page",
            Description: "Pages with signature blocks",
        },
    },
})
if err != nil {
    log.Fatal(err)
}

// Parse multiple contracts using the same dataset
fileIds := []string{"file_contract1", "file_contract2", "file_contract3"}

for _, fileId := range fileIds {
    parseJob, err := client.ParseDataset(context.Background(), &tensorlake.ParseDatasetRequest{
        DatasetId: dataset.DatasetId,
        FileSource: tensorlake.FileSource{
            FileId: fileId,
        },
        Labels: map[string]string{
            "document_type": "contract",
            "batch":         "2024-Q1",
        },
    })
    if err != nil {
        log.Printf("Error parsing %s: %v", fileId, err)
        continue
    }

    fmt.Printf("Started parse job: %s for file %s\n", parseJob.ParseId, fileId)
}
```

### Example: Parse from URL with Dataset

```go
parseJob, err := client.ParseDataset(context.Background(), &tensorlake.ParseDatasetRequest{
    DatasetId: "dataset_abc123",
    FileSource: tensorlake.FileSource{
        FileURL: "https://example.com/invoice.pdf",
    },
    Labels: map[string]string{
        "source": "vendor_portal",
    },
})
if err != nil {
    log.Fatal(err)
}

// Retrieve results
result, err := client.GetParseResult(context.Background(), parseJob.ParseId)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Parse completed with dataset: %s\n", result.DatasetId)
```

---

## Common Types

### DatasetStatus

```go
type DatasetStatus string

const (
    DatasetStatusActive   DatasetStatus = "active"
    DatasetStatusInactive DatasetStatus = "inactive"
)
```

### Dataset Naming Conventions

When creating datasets, follow these guidelines:

- Use descriptive, lowercase names with hyphens: `invoice-processing`, `legal-contracts`
- Keep names concise but meaningful: `quarterly-reports` not `q-rpts`
- Group related datasets with prefixes: `hr-employee-docs`, `hr-payroll-docs`

---

## Use Cases

### 1. Batch Processing Similar Documents

```go
// Create dataset for expense reports
dataset, _ := client.CreateDataset(ctx, &tensorlake.CreateDatasetRequest{
    Name: "expense-reports",
    StructuredExtractionOptions: []tensorlake.StructuredExtractionOptions{
        {
            SchemaName: "expenses",
            JSONSchema: expenseSchema,
        },
    },
})

// Process all expense PDFs
for _, fileId := range expenseFileIds {
    client.ParseDataset(ctx, &tensorlake.ParseDatasetRequest{
        DatasetId:  dataset.DatasetId,
        FileSource: tensorlake.FileSource{FileId: fileId},
    })
}
```

### 2. Multi-Tenant Processing

```go
// Create dataset per customer/tenant
customerDataset, _ := client.CreateDataset(ctx, &tensorlake.CreateDatasetRequest{
    Name:        fmt.Sprintf("customer-%s-invoices", customerId),
    Description: fmt.Sprintf("Invoice processing for customer %s", customerId),
    // ... customer-specific configuration
})

// Parse customer documents
client.ParseDataset(ctx, &tensorlake.ParseDatasetRequest{
    DatasetId:  customerDataset.DatasetId,
    FileSource: tensorlake.FileSource{FileId: fileId},
    Labels:     map[string]string{"customer_id": customerId},
})
```

### 3. Document Type Classification Pipeline

```go
// Create datasets for different document types
contractDataset, _ := client.CreateDataset(ctx, &tensorlake.CreateDatasetRequest{
    Name: "contracts",
    // Contract-specific schemas
})

invoiceDataset, _ := client.CreateDataset(ctx, &tensorlake.CreateDatasetRequest{
    Name: "invoices",
    // Invoice-specific schemas
})

// Route documents to appropriate dataset
func processDocument(fileId, docType string) {
    var datasetId string
    switch docType {
    case "contract":
        datasetId = contractDataset.DatasetId
    case "invoice":
        datasetId = invoiceDataset.DatasetId
    }

    client.ParseDataset(ctx, &tensorlake.ParseDatasetRequest{
        DatasetId:  datasetId,
        FileSource: tensorlake.FileSource{FileId: fileId},
    })
}
```

---

## Best Practices

1. **Reusability**: Create datasets for document types you process frequently
2. **Naming**: Use clear, descriptive names that indicate the document type and purpose
3. **Versioning**: Create new datasets instead of updating existing ones for major config changes
4. **Testing**: Test dataset configurations with sample documents before batch processing
5. **Organization**: Use consistent naming patterns for related datasets
6. **Documentation**: Keep dataset descriptions up-to-date with their purpose and configuration
7. **Monitoring**: Track parse job results by dataset to identify configuration issues

---

## Error Handling

Common error scenarios:

```go
dataset, err := client.CreateDataset(ctx, request)
if err != nil {
    var apiErr *tensorlake.ErrorResponse
    if errors.As(err, &apiErr) {
        switch apiErr.ErrorCode {
        case "dataset_already_exists":
            fmt.Println("Dataset name must be unique")
        case "invalid_schema":
            fmt.Println("JSON schema validation failed")
        case "quota_exceeded":
            fmt.Println("Dataset limit reached")
        default:
            fmt.Printf("API Error: %s\n", apiErr.Message)
        }
    }
    return
}
```

---

For more information, see:
- [File Management APIs](./file-apis.md)
- [Parse APIs](./parse-apis.md)
- [Main Documentation](../README.md)

