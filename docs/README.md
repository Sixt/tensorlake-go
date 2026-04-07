# Tensorlake Go SDK Documentation

Welcome to the Tensorlake Go SDK documentation! This guide covers all APIs for intelligent document processing.

## Documentation Index

### Core API Guides

1. **[File Management APIs](./file-apis.md)**
   - Upload files to Tensorlake
   - List and search files
   - Retrieve file metadata
   - Delete files
   - Supported file types and limits

2. **[Parse APIs](./parse-apis.md)**
   - Parse documents to markdown
   - Extract structured data with JSON schemas
   - Classify pages by content
   - Get parse results with SSE streaming
   - List and manage parse jobs
   - Advanced parsing configurations

3. **[Dataset APIs](./dataset-apis.md)**
   - Create reusable parsing configurations
   - Update dataset settings
   - List and manage datasets
   - Batch process documents with datasets
   - Use cases and best practices

4. **[Sandbox APIs](./sandbox-apis.md)**
   - Create, list, get, update, and delete sandboxes
   - Snapshot, suspend, and resume sandboxes
   - File operations (read, write, delete, list)
   - PTY sessions (create, connect via WebSocket, resize, kill)
   - Process management (start, signal, kill, stdin pipe)
   - Output capture and SSE streaming (stdout, stderr, merged)

## Quick Navigation

### Common Tasks

- **Getting Started**: See [Main README](../README.md#quick-start)
- **Upload a file**: [File APIs - Upload](./file-apis.md#upload-file)
- **Parse a document**: [Parse APIs - Parse Document](./parse-apis.md#parse-document)
- **Extract data**: [Parse APIs - Extract Document](./parse-apis.md#extract-document)
- **Create a dataset**: [Dataset APIs - Create Dataset](./dataset-apis.md#create-dataset)
- **Stream results with SSE**: [Parse APIs - Get Parse Result](./parse-apis.md#get-parse-result)

- **Create a sandbox**: [Sandbox APIs - Create](./sandbox-apis.md#create-a-sandbox)
- **Run a process**: [Sandbox APIs - Start Process](./sandbox-apis.md#start-a-process)
- **Interactive terminal**: [Sandbox APIs - PTY](./sandbox-apis.md#connect-via-websocket)
- **Follow output**: [Sandbox APIs - Follow Output](./sandbox-apis.md#follow-output-via-sse)

### By Use Case

#### Simple Document Parsing
1. [Upload file](./file-apis.md#upload-file)
2. [Parse document](./parse-apis.md#parse-document)
3. [Get results](./parse-apis.md#get-parse-result)

#### Structured Data Extraction
1. [Define JSON schema](./parse-apis.md#extract-document)
2. [Extract with schema](./parse-apis.md#extract-document)
3. [Access extracted data](./parse-apis.md#example-access-structured-data)

#### Page Classification
1. [Define page classes](./parse-apis.md#classify-document)
2. [Classify document](./parse-apis.md#classify-document)
3. [Access classifications](./parse-apis.md#classify-document)

#### Batch Processing
1. [Create dataset](./dataset-apis.md#create-dataset)
2. [Parse with dataset](./dataset-apis.md#parse-with-dataset)
3. [Process multiple files](./dataset-apis.md#example-parse-multiple-files-with-same-configuration)

## Key Concepts

### File Sources
Documents can be provided in three ways:
- **File ID**: Upload first, then reference by ID (recommended)
- **File URL**: Publicly accessible internet URL
- **Raw Text**: Plain text content

See [Parse APIs](./parse-apis.md#parse-document) for details.

### Pagination
List operations support cursor-based pagination:
- Use `Cursor`, `Limit`, and `Direction` parameters
- Or use convenient iterator methods (`IterFiles`, `IterParseJobs`, `IterDatasets`)

See [File APIs - List Files](./file-apis.md#iterate-all-files) for examples.

### Server-Sent Events (SSE)
Get real-time updates for long-running parse jobs:
- Enable with `WithSSE(true)`
- Receive callbacks with `WithOnUpdate()`

See [Parse APIs - SSE Streaming](./parse-apis.md#example-sse-streaming) for examples.

## API Reference by Category

### File Management
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Upload file | `UploadFile()` | [Link](./file-apis.md#upload-file) |
| List files | `ListFiles()` / `IterFiles()` | [Link](./file-apis.md#list-files) |
| Get metadata | `GetFileMetadata()` | [Link](./file-apis.md#get-file-metadata) |
| Delete file | `DeleteFile()` | [Link](./file-apis.md#delete-file) |

### Parse Operations
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Parse document | `ParseDocument()` | [Link](./parse-apis.md#parse-document) |
| Read document | `ReadDocument()` | [Link](./parse-apis.md#read-document) |
| Extract data | `ExtractDocument()` | [Link](./parse-apis.md#extract-document) |
| Classify pages | `ClassifyDocument()` | [Link](./parse-apis.md#classify-document) |
| Get results | `GetParseResult()` | [Link](./parse-apis.md#get-parse-result) |
| List jobs | `ListParseJobs()` / `IterParseJobs()` | [Link](./parse-apis.md#list-parse-jobs) |
| Delete job | `DeleteParseJob()` | [Link](./parse-apis.md#delete-parse-job) |

### Dataset Management
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Create dataset | `CreateDataset()` | [Link](./dataset-apis.md#create-dataset) |
| Get dataset | `GetDataset()` | [Link](./dataset-apis.md#get-dataset) |
| Update dataset | `UpdateDataset()` | [Link](./dataset-apis.md#update-dataset) |
| List datasets | `ListDatasets()` / `IterDatasets()` | [Link](./dataset-apis.md#list-datasets) |
| Delete dataset | `DeleteDataset()` | [Link](./dataset-apis.md#delete-dataset) |
| Parse with dataset | `ParseDataset()` | [Link](./dataset-apis.md#parse-with-dataset) |

### Sandbox Management
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Create sandbox | `CreateSandbox()` | [Link](./sandbox-apis.md#create-a-sandbox) |
| List sandboxes | `ListSandboxes()` | [Link](./sandbox-apis.md#list-sandboxes) |
| Get sandbox | `GetSandbox()` | [Link](./sandbox-apis.md#get-sandbox-details) |
| Update sandbox | `UpdateSandbox()` | [Link](./sandbox-apis.md#update-sandbox) |
| Delete sandbox | `DeleteSandbox()` | [Link](./sandbox-apis.md#delete-terminate-sandbox) |
| Snapshot | `SnapshotSandbox()` | [Link](./sandbox-apis.md#snapshot-and-restore) |
| Suspend/Resume | `SuspendSandbox()` / `ResumeSandbox()` | [Link](./sandbox-apis.md#suspend-and-resume) |

### Sandbox Files
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Read file | `ReadSandboxFile()` | [Link](./sandbox-apis.md#read-a-file) |
| Write file | `WriteSandboxFile()` | [Link](./sandbox-apis.md#write-a-file) |
| Delete file | `DeleteSandboxFile()` | [Link](./sandbox-apis.md#delete-a-file) |
| List directory | `ListSandboxDirectory()` | [Link](./sandbox-apis.md#list-a-directory) |

### PTY Sessions
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Create PTY | `CreatePTY()` | [Link](./sandbox-apis.md#create-a-pty-session) |
| Connect WebSocket | `ConnectPTY()` | [Link](./sandbox-apis.md#connect-via-websocket) |
| List/Get/Resize/Kill | `ListPTY()` / `GetPTY()` / `ResizePTY()` / `KillPTY()` | [Link](./sandbox-apis.md#list--get--resize--kill-pty-sessions) |

### Process Management
| Operation | Method | Documentation |
|-----------|--------|---------------|
| Start process | `StartProcess()` | [Link](./sandbox-apis.md#start-a-process) |
| List/Get | `ListProcesses()` / `GetProcess()` | [Link](./sandbox-apis.md#list--get-processes) |
| Signal/Kill | `SignalProcess()` / `KillProcess()` | [Link](./sandbox-apis.md#send-signal--kill) |
| Stdin | `WriteProcessStdin()` / `CloseProcessStdin()` | [Link](./sandbox-apis.md#stdin-pipe) |
| Read output | `GetProcessStdout()` / `GetProcessStderr()` / `GetProcessOutput()` | [Link](./sandbox-apis.md#read-captured-output) |
| Follow output | `FollowProcessStdout()` / `FollowProcessStderr()` / `FollowProcessOutput()` | [Link](./sandbox-apis.md#follow-output-via-sse) |

## Configuration Reference

### Parsing Options
Customize document parsing behavior:
- Chunking strategies
- OCR models
- Table output formats
- Image inclusion
- Detection features (signatures, barcodes, etc.)

See [Parse APIs - Parsing Options](./parse-apis.md#parsingoptions)

### Enrichment Options
Add AI-powered enhancements:
- Table summarization
- Figure summarization
- Custom prompts

See [Parse APIs - Enrichment Options](./parse-apis.md#enrichmentoptions)

### Extraction Options
Configure structured data extraction:
- JSON schema definition
- Partition strategies
- Model providers
- Page class filters
- Citation support

See [Parse APIs - Structured Extraction](./parse-apis.md#extract-document)

## Examples

Each API guide includes comprehensive examples:
- Basic usage
- Advanced configurations
- Error handling
- Real-world use cases

Start with the [Main README](../README.md) for an overview, then dive into specific API guides as needed.

## External Resources

- [Tensorlake API Documentation](https://docs.tensorlake.ai/)
- [API Reference v2](https://docs.tensorlake.ai/api-reference/v2/introduction)
- [Go Package Documentation](https://pkg.go.dev/github.com/sixt/tensorlake-go)

## Need Help?

- Check the relevant API guide for detailed examples
- Review [best practices](./parse-apis.md#best-practices) in each guide
- Consult the [error handling](./file-apis.md#error-handling) section

---

[← Back to Main README](../README.md)

