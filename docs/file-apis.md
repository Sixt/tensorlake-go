# File Management APIs

The File Management APIs allow you to upload, list, retrieve metadata, and delete files in your Tensorlake project.

## Table of Contents

- [Upload File](#upload-file)
- [List Files](#list-files)
- [Get File Metadata](#get-file-metadata)
- [Delete File](#delete-file)

## Upload File

Upload a file to Tensorlake Cloud. The file will be associated with the project specified by the API key used in the request.

### Supported File Types

- **PDF** documents
- **Word** (DOCX)
- **Spreadsheets** (XLS, XLSX, XSLM, CSV)
- **Presentations** (PPTX, Apple Keynote)
- **Images** (PNG, JPG, JPEG)
- **Raw text** (plain text, HTML)

### Method

```go
func (c *Client) UploadFile(ctx context.Context, in *UploadFileRequest) (*FileUploadResponse, error)
```

### Request Parameters

```go
type UploadFileRequest struct {
    // FileBytes is the reader for the file to upload (required)
    FileBytes io.Reader

    // FileName is the name of the file to upload (required)
    FileName string

    // Labels are optional key-value pairs to categorize the file
    Labels map[string]string
}
```

### Response

```go
type FileUploadResponse struct {
    // FileId is the ID of the created file
    // Use this ID to reference the file in parse and other operations
    FileId string

    // CreatedAt is the creation date and time in RFC 3339 format
    CreatedAt time.Time
}
```

### Example

```go
file, err := os.Open("path/to/document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

response, err := client.UploadFile(context.Background(), &tensorlake.UploadFileRequest{
    FileBytes: file,
    FileName:  "document.pdf",
    Labels: map[string]string{
        "category": "invoice",
        "year": "2024",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("File uploaded with ID: %s\n", response.FileId)
```

### Notes

- Upload limit: 1 GB per file
- Files are deduplicated - uploading the same file multiple times returns the same `file_id`
- File type is automatically detected based on Content-Type header or file extension
- Labels help categorize and filter files for better organization

---

## List Files

List all files in your Tensorlake project with cursor-based pagination.

### Method

```go
func (c *Client) ListFiles(ctx context.Context, in *ListFilesRequest) (*PaginationResult[FileInfo], error)
```

### Request Parameters

```go
type ListFilesRequest struct {
    // Cursor for pagination (base64-encoded timestamp)
    Cursor string

    // Direction of pagination: "next" or "prev"
    Direction PaginationDirection

    // Limit the number of results (x >= 0)
    Limit int

    // Filter by file name (case-sensitive substring match)
    FileName string

    // Filter by creation date (RFC 3339 format)
    CreatedAfter string
    CreatedBefore string
}
```

### Response

```go
type PaginationResult[FileInfo] struct {
    Items      []FileInfo  // Array of file metadata
    HasMore    bool        // Whether more results are available
    NextCursor string      // Cursor for next page
    PrevCursor string      // Cursor for previous page
}

type FileInfo struct {
    FileId         string
    FileName       string
    MimeType       MimeType
    FileSize       int64
    ChecksumSHA256 string
    CreatedAt      string
    Labels         map[string]string
}
```

### Example

```go
// List first 10 files
response, err := client.ListFiles(context.Background(), &tensorlake.ListFilesRequest{
    Limit: 10,
    Direction: tensorlake.PaginationDirectionNext,
})
if err != nil {
    log.Fatal(err)
}

for _, file := range response.Items {
    fmt.Printf("File: %s (ID: %s, Size: %d bytes)\n",
        file.FileName, file.FileId, file.FileSize)
}

// Get next page if available
if response.HasMore {
    nextPage, err := client.ListFiles(context.Background(), &tensorlake.ListFilesRequest{
        Cursor: response.NextCursor,
        Limit: 10,
        Direction: tensorlake.PaginationDirectionNext,
    })
    // ... process next page
}
```

### Iterate All Files

For convenience, use the `IterFiles` method to iterate through all files:

```go
func (c *Client) IterFiles(ctx context.Context, batchSize int) iter.Seq2[FileInfo, error]
```

**Example:**

```go
for file, err := range client.IterFiles(context.Background(), 50) {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Processing file: %s\n", file.FileName)
}
```

---

## Get File Metadata

Retrieve metadata for a specific file by its ID.

### Method

```go
func (c *Client) GetFileMetadata(ctx context.Context, fileId string) (*FileInfo, error)
```

### Parameters

- `fileId`: The unique identifier of the file

### Response

```go
type FileInfo struct {
    FileId         string            // Unique file identifier
    FileName       string            // Original file name
    MimeType       MimeType          // MIME type of the file
    FileSize       int64             // File size in bytes
    ChecksumSHA256 string            // SHA-256 checksum
    CreatedAt      string            // Creation timestamp (RFC 3339)
    Labels         map[string]string // Associated labels
}
```

### Example

```go
metadata, err := client.GetFileMetadata(context.Background(), "file_abc123xyz")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("File: %s\n", metadata.FileName)
fmt.Printf("Size: %d bytes\n", metadata.FileSize)
fmt.Printf("Type: %s\n", metadata.MimeType)
fmt.Printf("Uploaded: %s\n", metadata.CreatedAt)
```

---

## Delete File

Delete a file from Tensorlake Cloud.

### Method

```go
func (c *Client) DeleteFile(ctx context.Context, fileId string) error
```

### Parameters

- `fileId`: The unique identifier of the file to delete

### Example

```go
err := client.DeleteFile(context.Background(), "file_abc123xyz")
if err != nil {
    log.Fatal(err)
}

fmt.Println("File deleted successfully")
```

### Notes

- Deleting a file is permanent and cannot be undone
- Deleting a file does not automatically delete parse jobs created from it
- You must have appropriate permissions to delete files in the project

---

## Common Types

### PaginationDirection

```go
type PaginationDirection string

const (
    PaginationDirectionNext PaginationDirection = "next"
    PaginationDirectionPrev PaginationDirection = "prev"
)
```

### MimeType

Common MIME types for uploaded files:

```go
type MimeType string

const (
    MimeTypePDF              MimeType = "application/pdf"
    MimeTypeDOCX             MimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    MimeTypePNG              MimeType = "image/png"
    MimeTypeJPEG             MimeType = "image/jpeg"
    // ... and more
)
```

---

## Error Handling

All methods return an error as the last return value. Common error scenarios:

- **Network errors**: Connection issues, timeouts
- **Authentication errors**: Invalid or expired API key
- **Validation errors**: Invalid parameters, missing required fields
- **Not found errors**: File ID doesn't exist
- **Quota errors**: Upload limits exceeded

**Example error handling:**

```go
response, err := client.UploadFile(ctx, request)
if err != nil {
    var apiErr *tensorlake.ErrorResponse
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s (Code: %s)\n", apiErr.Message, apiErr.ErrorCode)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

---

For more information, see:
- [Parse APIs](./parse-apis.md)
- [Dataset APIs](./dataset-apis.md)
- [Main Documentation](../README.md)

