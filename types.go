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
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
)

// PaginationResult represents the result of a pagination operation.
type PaginationResult[T any] struct {
	Items      []T    `json:"items"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// FileSource represents the source of a document (FileId, FileURL, or RawText).
type FileSource struct {
	// ID of the file previously uploaded to Tensorlake.
	// Has tensorlake- (V1) or file_ (V2) prefix.
	// Example: "file_abc123xyz"
	FileId string `json:"file_id,omitempty"`
	// External URL of the file to parse. Must be publicly accessible.
	// Examples: "https://pub-226479de18b2493f96b64c6674705dd8.r2.dev/real-estate-purchase-all-signed.pdf"
	FileURL string `json:"file_url,omitempty"`
	// The raw text content to parse.
	// Examples: "This is the document content..."
	RawText string `json:"raw_text,omitempty"`
}

// SourceProvided checks exactly one source is provided.
func (fs *FileSource) SourceProvided() bool {
	count := 0
	if fs.FileId != "" {
		count++
	}
	if fs.FileURL != "" {
		count++
	}
	if fs.RawText != "" {
		count++
	}
	return count == 1
}

// ParsingOptions holds configuration for document parsing.
type ParsingOptions struct {
	// Chunking strategy determines how the document is chunked into smaller pieces.
	// Different strategies can be used to optimize the parsing process.
	// Choose the one that best fits your use case. The default is `None`,
	// which means no chunking is applied.
	ChunkingStrategy ChunkingStrategy `json:"chunking_strategy,omitempty"`

	// CrossPageHeaderDetection enables header-hierarchy detection across pages.
	// When set to `true`, the parser will consider headers from different pages
	// when determining the hierarchy of headers within a single page.
	CrossPageHeaderDetection bool `json:"cross_page_header_detection,omitempty"`

	// DisableLayoutDetection disables bounding box detection for the document.
	// Leads to faster document parsing.
	DisableLayoutDetection bool `json:"disable_layout_detection,omitempty"`

	// OCRModel indicates the model to use for OCR (Optical Character Recognition).
	//
	//   - model01: It's fast but could have lower accuracy on complex tables.
	//              It's good for legal documents with footnotes.
	//   - model02: It's slower but could have higher accuracy on complex tables.
	//              It's good for financial documents with merged cells.
	//   - model03: A compact model that we deliver to on-premise users.
	//              It takes about 2 minutes to startup on Tensorlake's Cloud
	//              because it's meant for testing for users who are eventually
	//              going to deploy this model on dedicated hardware in their
	//              own datacenter.
	OCRModel OCRPipelineProvider `json:"ocr_model,omitempty"`

	// RemoveStrikethroughLines enables the detection, and removal, of
	// strikethrough text in the document. This flag incurs additional billing costs.
	RemoveStrikethroughLines bool `json:"remove_strikethrough_lines,omitempty"`

	// SignatureDetection enables the detection of signatures in the document.
	// This flag incurs additional billing costs.
	// The default is false.
	SignatureDetection bool `json:"signature_detection,omitempty"`

	// SkewDetection enables detect and correct skewed or rotated pages in the
	// document. Setting this to true will increase the processing time of the
	// document. The default is false.
	SkewDetection bool `json:"skew_detection,omitempty"`

	// TableOutputMode is the format for the tables extracted from the document.
	// The default is HTML.
	TableOutputMode TableOutputMode `json:"table_output_mode,omitempty"`

	// TableParsingFormat determines which model the system uses to identify
	// and extract tables from the document. The default is tsr.
	TableParsingFormat TableParsingFormat `json:"table_parsing_format,omitempty"`

	// IgnoreSections contain a set of page fragment types to ignore during parsing.
	//
	// This can be used to skip certain types of content that are not relevant
	// for the parsing process, such as headers, footers, or other
	// non-essential elements.
	//
	// The default is an empty set.
	IgnoreSections []PageFragmentType `json:"ignore_sections,omitempty"`

	// IncludeImages embeded images from document in the markdown.
	// The default is false.
	IncludeImages bool `json:"include_images,omitempty"`

	// BarcodeDetection enable barcode detection in the document.
	// Setting this to true will increase the processing time of the document.
	// The default is false.
	BarcodeDetection bool `json:"barcode_detection,omitempty"`

	// MergeTables enables merging of tables that span across multiple pages.
	// The default is false.
	MergeTables bool `json:"merge_tables,omitempty"`
}

// EnrichmentOptions holds configuration for document enrichment.
type EnrichmentOptions struct {
	// FigureSummarization enables summary generation for parsed figures.
	// The default is false.
	FigureSummarization bool `json:"figure_summarization,omitempty"`

	// FigureSummarizationPrompt is the prompt to guide the figure summarization.
	// If not provided, a default prompt will be used. It is not required to provide a prompt.
	// The prompt only has effect if [FigureSummarization] is set to `true`.
	FigureSummarizationPrompt string `json:"figure_summarization_prompt,omitempty"`

	// TableSummarization enables summary generation for parsed tables.
	// The default is false.
	TableSummarization bool `json:"table_summarization,omitempty"`

	// TableSummarizationPrompt is the prompt to guide the table summarization.
	// If not provided, a default prompt will be used. It is not required to provide a prompt.
	// The prompt only has effect if [TableSummarization] is set to `true`.
	TableSummarizationPrompt string `json:"table_summarization_prompt,omitempty"`

	// IncludeFullPageImage includes the full page image in addition to the cropped table and figure images.
	// This provides Language Models context about the table and figure they are summarizing in addition to the cropped images, and could improve the summarization quality.
	// The default is false.
	IncludeFullPageImage bool `json:"include_full_page_image,omitempty"`

	// TableCellGrounding enables grounding of table cells with bounding boxes.
	// The default is false.
	TableCellGrounding bool `json:"table_cell_grounding,omitempty"`

	// ChartExtraction enables extraction of data from charts.
	// The default is false.
	ChartExtraction bool `json:"chart_extraction,omitempty"`

	// KeyValueExtraction enables extraction of key-value pairs.
	// The default is false.
	KeyValueExtraction bool `json:"key_value_extraction,omitempty"`
}

// ParseJob represents a parse job.
type ParseJob struct {
	// ParseId is the unique identifier for the parse job.
	// This is the ID that can be used to track the status of the parse job.
	// Used in the GET /documents/v2/parse/{parse_id} endpoint to retrieve
	// the status and results of the parse job.
	ParseId string `json:"parse_id"`
	// CreatedAt is the creation date and time of the parse job.
	CreatedAt string `json:"created_at"`
}

// ParseResult represents the result of a parse job.
type ParseResult struct {
	//
	// ParseResult specific fields.
	//

	// The unique identifier for the parse job. This is the same value
	// returned from ReadDocument or ParseDocument.
	// Example: "parse_abcd1234"
	ParseId string `json:"parse_id"`

	// The number of pages that were parsed successfully.
	// This is the total number of pages that were successfully parsed
	// in the document. Required range: x >= 0. Example: 5
	ParsedPagesCount int `json:"parsed_pages_count"`

	// The current status of the parse job. This indicates whether the
	// job is pending, in progress, completed, or failed.
	// This can be used to track the progress of the parse operation.
	Status ParseStatus `json:"status"`

	// The date and time when the parse job was created.
	// The date is in RFC 3339 format. This can be used to track when
	// the parse job was initiated. Example: "2023-10-01T12:00:00Z"
	CreatedAt string `json:"created_at"`

	//
	// Optional fields.
	//

	// Error occurred during any part of the parse execution.
	// This is only populated if the parse operation failed.
	Error string `json:"error,omitempty"`

	// The date and time when the parse job was finished.
	// The date is in RFC 3339 format.
	// This can be undefined if the parse job is still in progress or pending.
	FinishedAt string `json:"finished_at,omitempty"`

	// Labels associated with the parse job.
	//
	// These are the key-value, or json, pairs submitted with the parse
	// request.
	//
	// This can be used to categorize or tag the parse job for easier
	// identification and filtering.
	//
	// It can be undefined if no labels were provided in the request.
	Labels map[string]string `json:"labels,omitempty"`

	// TotalPages is the total number of pages in the document that was parsed.
	TotalPages int `json:"total_pages,omitempty"`

	// MessageUpdate is the message update for the parse job.
	MessageUpdate string `json:"message_update,omitempty"`

	// PdfBase64 is the base64-encoded PDF content of the parsed document.
	PdfBase64 string `json:"pdf_base64,omitempty"`

	// TasksCompletedCount is the number of tasks completed for the parse job.
	TasksCompletedCount *int `json:"tasks_completed_count,omitempty"`

	// TasksTotalCount is the total number of tasks for the parse job.
	TasksTotalCount *int `json:"tasks_total_count,omitempty"`

	// If the parse job was scheduled from a dataset, this field contains
	// the dataset id. This is the identifier used in URLs and API endpoints
	// to refer to the dataset.
	DatasetId string `json:"dataset_id,omitempty"`

	//
	// Parsed document specific fields.
	//

	// Chunks of the document.
	//
	// This is a vector of Chunk objects, each containing a chunk of the
	// document.
	//
	// The number of chunks depend on the chunking strategy used during
	// parsing.
	Chunks []Chunk `json:"chunks,omitempty"`

	// List of pages parsed from the document.
	//
	// Each page has a list of fragments, which are detected objects such as
	// tables, text, figures, section headers, etc.
	//
	// We also return the detected text, structure of the table(if its a
	// table), and the bounding box of the object.
	Pages []Page `json:"pages"`

	// Page classes extracted from the document.
	//
	// This is a map where the keys are page class names provided in the parse
	// request under the page_classification_options field,
	// and the values are vectors of page numbers (1-indexed) where each page
	// class appears.
	//
	// This is used to categorize pages in the document based on the
	// classify options provided.
	PageClasses []PageClass `json:"page_classes,omitempty"`

	// Structured data extracted from the document.
	//
	// The structured data is a map where the keys are the schema names
	// provided in the parse request, and the values are
	// StructuredData objects containing the structured data extracted from
	// the document.
	//
	// The number of structured data objects depends on the partition strategy
	// None - one structured data object for the entire document.
	// Page - one structured data object for each page.
	StructuredData []StructuredData `json:"structured_data,omitempty"`

	// MergedTables contains tables that were merged across multiple pages.
	MergedTables []MergedTable `json:"merged_tables,omitempty"`

	// Options contains the options used for the parse job.
	Options *ParseResultOptions `json:"options,omitempty"`

	// Resource usage associated with the parse job.
	//
	// This includes details such as number of pages parsed, tokens used for
	// OCR and extraction, etc.
	//
	// Usage is only populated for successful jobs.
	//
	// Billing is based on the resource usage.
	Usage Usage `json:"usage"`
}

// MergeTableActions describes the merge operations performed on the table.
type MergeTableActions struct {
	// Pages is the list of page numbers that were merged.
	Pages []int `json:"pages,omitempty"`
	// TargetColumns is the target column count for the merged table.
	TargetColumns *int `json:"target_columns,omitempty"`
}

// MergedTable represents a table that was merged across multiple pages.
type MergedTable struct {
	// MergedTableId is the unique identifier for the merged table.
	MergedTableId string `json:"merged_table_id"`
	// MergedTableHTML is the HTML representation of the merged table.
	MergedTableHTML string `json:"merged_table_html"`
	// StartPage is the first page of the merged table.
	StartPage int `json:"start_page"`
	// EndPage is the last page of the merged table.
	EndPage int `json:"end_page"`
	// PagesMerged is the number of pages that were merged.
	PagesMerged int `json:"pages_merged"`
	// Summary is an optional summary of the merged table.
	Summary string `json:"summary,omitempty"`
	// MergeActions describes the merge operations performed.
	MergeActions *MergeTableActions `json:"merge_actions,omitempty"`
}

// ParseConfiguration contains the full configuration used for a parse job.
type ParseConfiguration struct {
	ParsingOptions              *ParsingOptions               `json:"parsing_options,omitempty"`
	StructuredExtractionOptions []StructuredExtractionOptions `json:"structured_extraction_options,omitempty"`
	PageClassifications         []PageClassConfig             `json:"page_classifications,omitempty"`
	EnrichmentOptions           *EnrichmentOptions            `json:"enrichment_options,omitempty"`
}

// ParseResultOptions contains the options used for the parse job.
// It includes the configuration options used for the parse job,
// including the file ID, file URL, raw text, mime type,
// and structured extraction options, etc.
type ParseResultOptions struct {
	FileSource
	FileName      string            `json:"file_name"`
	FileLabels    map[string]string `json:"file_labels"`
	MimeType      MimeType          `json:"mime_type"`
	TraceId       string            `json:"trace_id"`
	PageRange     string            `json:"page_range"`
	JobType       JobType           `json:"job_type"`
	Configuration *ParseConfiguration `json:"configuration"`
	Usage         *Usage            `json:"usage,omitempty"`
	MessageUpdate string            `json:"message_update,omitempty"`
}

// Usage contains resource usage associated with the parse job.
// This includes details such as number of pages parsed, tokens used for
// OCR and extraction, etc.
// Usage is only populated for successful jobs.
// Billing is based on the resource usage.
type Usage struct {
	PagesParsed                   int `json:"pages_parsed"`
	SignatureDetectedPages        int `json:"signature_detected_pages"`
	StrikethroughDetectedPages    int `json:"strikethrough_detected_pages"`
	OCRInputTokensUsed            int `json:"ocr_input_tokens_used"`
	OCROutputTokensUsed           int `json:"ocr_output_tokens_used"`
	ExtractionInputTokensUsed     int `json:"extraction_input_tokens_used"`
	ExtractionOutputTokensUsed    int `json:"extraction_output_tokens_used"`
	SummarizationInputTokensUsed  int `json:"summarization_input_tokens_used"`
	SummarizationOutputTokensUsed int `json:"summarization_output_tokens_used"`
}

// StructuredExtractionOptions holds configuration for structured data extraction.
type StructuredExtractionOptions struct {

	//
	// Required fields
	//

	// The name of the schema. This is used to tag the structured data output
	// with a name in the response.
	SchemaName string `json:"schema_name"`

	// 	The JSON schema to guide structured data extraction from the file.
	//
	// This schema should be a valid JSON schema that defines the structure of
	// the data to be extracted.
	//
	// The API supports a subset of the JSON schema specification.
	//
	// This value must be provided if structured_extraction is present in the
	// request.
	JSONSchema *jsonschema.Schema `json:"json_schema"` // Can be any JSON schema structure

	//
	// Optional fields
	//

	// Strategy to partition the document before structured data extraction.
	// The API will return one structured data object per partition. This is
	// useful when you want to extract certain fields from every page.
	PartitionStrategy PartitionStrategy `json:"partition_strategy,omitempty"`

	// 	The model provider to use for structured data extraction.
	//
	// The default is tensorlake, which uses our private model, and runs on
	// our servers.
	ModelProvider ModelProvider `json:"model_provider,omitempty"`

	// Filter the pages of the document to be used for structured data
	// extraction by providing a list of page classes.
	PageClasses []string `json:"page_classes,omitempty"`

	// The prompt to use for structured data extraction.
	//
	// If not provided, the default prompt will be used.
	Prompt string `json:"prompt,omitempty"`

	// Flag to enable visual citations in the structured data output.
	// It returns the bounding boxes of the coordinates of the document
	// where the structured data was extracted from.
	ProvideCitations bool `json:"provide_citations,omitempty"`

	// Boolean flag to skip converting the document blob to OCR text before
	// structured data extraction.
	//
	// If set to true, the API will skip the OCR step and directly extract
	// structured data from the document.
	SkipOCR bool `json:"skip_ocr,omitempty"`
}

// Chunk represents a chunk of the document.
type Chunk struct {
	Content    string `json:"content"`
	PageNumber int    `json:"page_number"` // >= 0
}

// Page represents a page in the parsed document.
type Page struct {
	// Dimensions is a 2-element vector representing the width and height of
	// the page in points.
	Dimensions []int `json:"dimensions,omitempty"`

	// PageDimensions is a 2-element vector representing the width and height of
	// the page in points.
	PageDimensions PageDimensions `json:"page_dimensions,omitempty"`

	// Vector of text fragments extracted from the page.
	// Each fragment represents a distinct section of text, such as titles,
	// paragraphs, tables, figures, etc.
	PageFragments []PageFragment `json:"page_fragments,omitempty"`

	// 1-indexed page number in the document.
	PageNumber int `json:"page_number"`

	// If the page was classified into a specific class, this field contains
	// the reason for the classification.
	ClassificationReason string `json:"classification_reason,omitempty"`
}

// PageFragment represents a fragment of a page in the parsed document.
type PageFragment struct {
	FragmentType PageFragmentType    `json:"fragment_type"`
	Content      PageFragmentContent `json:"content"`
	ReadingOrder int64               `json:"reading_order,omitempty"`
	BoundingBox  map[string]float64  `json:"bbox,omitempty"`
}

// PageDimensions represents the dimensions of a page.
type PageDimensions struct {
	// Width is the width of the page in points.
	Width int `json:"width"`
	// Height is the height of the page in points.
	Height int `json:"height"`
}

// PageClass extracted from the document.
type PageClass struct {
	// PageClass is the name of the page class given in the parse request.
	// This value should match one of the class names provided in the
	// page_classification_options field of the parse request.
	//
	// Required.
	PageClass string `json:"page_class"`

	// PageNumbers is a list of page numbers (1-indexed) where
	// the page class was detected. Required.
	PageNumbers []int `json:"page_numbers"`

	// ClassificationReasons is a map of classification reasons per page number
	// The key is the page number, and the value is the reason for the classification.
	ClassificationReasons map[int]string `json:"classification_reasons,omitempty"`
}

type PageClassConfig struct {
	// Name is the name of the page class.
	Name string `json:"name"`

	// Description is the description of the page class to guide the model
	// to classify the pages. Describe what the model should look for in
	// the page to classify it.
	Description string `json:"description,omitempty"`
}

// StructuredData extracted from the document.
// The structured data is a map where the keys are the schema names
// provided in the parse request, and the values are
// StructuredData objects containing the structured data extracted from
// the document.
type StructuredData struct {
	// Data is a JSON object containing the structured data extracted from the document.
	// The schema is specified in the StructuredExtractionOptions.JSONSchema field.
	Data json.RawMessage `json:"data"`
	// PageNumber contains either an integer or an array of integers regarding page numbers.
	// Example: [1, 2, 3] or 1
	PageNumbers UnionValues[int] `json:"page_numbers"`
	// SchemaName is the name of the schema used to extract the structured data.
	// It is specified in the StructuredExtractionOptions.SchemaName field.
	SchemaName string `json:"schema_name,omitempty"`
}

type PageFragmentContent struct {
	// One of these will be set depending on the JSON input:
	Text      *PageFragmentText      `json:"text,omitempty"`
	Header    *PageFragmentHeader    `json:"header,omitempty"`
	Table     *PageFragmentTable     `json:"table,omitempty"`
	Figure    *PageFragmentFigure    `json:"figure,omitempty"`
	Signature *PageFragmentSignature `json:"signature,omitempty"`
}

type PageFragmentText struct {
	Content string `json:"content"`
}

type PageFragmentHeader struct {
	Content string `json:"content"`
	Level   int    `json:"level"`
}
type PageFragmentTableCell struct {
	Text        string             `json:"text"`
	BoundingBox map[string]float64 `json:"bounding_box"`
}

type PageFragmentTable struct {
	Content  string                  `json:"content"`
	Cells    []PageFragmentTableCell `json:"cells"`
	HTML     string                  `json:"html,omitempty"`
	Markdown string                  `json:"markdown,omitempty"`
	Summary  string                  `json:"summary,omitempty"`
}

type PageFragmentFigure struct {
	Content string `json:"content"`
	Summary string `json:"summary,omitempty"`
}

type PageFragmentSignature struct {
	Content string `json:"content"`
}

// DatasetParseJobAnalytics contains analytics about parse jobs in a dataset.
type DatasetParseJobAnalytics struct {
	TotalProcessingParseJobs int `json:"total_processing_parse_jobs"`
	TotalPendingParseJobs    int `json:"total_pending_parse_jobs"`
	TotalErrorParseJobs      int `json:"total_error_parse_jobs"`
	TotalSuccessfulParseJobs int `json:"total_successful_parse_jobs"`
	TotalJobs                int `json:"total_jobs"`
}

// Dataset represents a dataset.
type Dataset struct {
	Name        string                    `json:"name"`
	DatasetId   string                    `json:"dataset_id"`
	Description string                    `json:"description,omitempty"`
	Status      DatasetStatus             `json:"status"`
	CreatedAt   string                    `json:"created_at"`
	UpdatedAt   string                    `json:"updated_at"`
	Analytics   *DatasetParseJobAnalytics `json:"analytics,omitempty"`
}
