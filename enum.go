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

const (
	// EndpointEU is the European endpoint.
	EndpointEU string = "https://api.eu.tensorlake.ai/documents/v2"
	// EndpointUS is the United States endpoint.
	EndpointUS string = "https://api.tensorlake.ai/documents/v2"
)

// ChunkingStrategy determines how the document is chunked into smaller pieces.
//
// Every text block, image, table, etc. is considered a fragment.
type ChunkingStrategy string

const (
	// ChunkingStrategyNone: No chunking is applied.
	ChunkingStrategyNone ChunkingStrategy = "none"
	// ChunkingStrategyPage: The document is chunked by page.
	ChunkingStrategyPage ChunkingStrategy = "page"
	// ChunkingStrategySection: The document is chunked into sections.
	// Title and section headers are used as chunking markers.
	ChunkingStrategySection ChunkingStrategy = "section"
	// ChunkingStrategyFragment: Each page element is converted into markdown form.
	ChunkingStrategyFragment ChunkingStrategy = "fragment"
)

// MimeType represents supported MIME types for document parsing.
type MimeType string

const (
	// MimeTypeTXT represents plain text files.
	MimeTypeTXT MimeType = "text/plain"
	// MimeTypeCSV represents a comma-separated values files.
	MimeTypeCSV MimeType = "text/csv"
	// MimeTypeHTML represents HTML files.
	MimeTypeHTML MimeType = "text/html"
	// MimeTypeJPEG represents JPEG image files.
	MimeTypeJPEG MimeType = "image/jpeg"
	// MimeTypePNG represents PNG image files.
	MimeTypePNG MimeType = "image/png"
	// MimeTypePDF represents Portable Document Format files.
	MimeTypePDF MimeType = "application/pdf"
	// MimeTypeDOCX represents Microsoft Word documents.
	MimeTypeDOCX MimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	// MimeTypePPTX represents Microsoft PowerPoint presentations.
	MimeTypePPTX MimeType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	// MimeTypeKEYNOTE represents Apple Keynote presentations.
	MimeTypeKEYNOTE MimeType = "application/vnd.apple.keynote"
	// MimeTypeXLS represents Microsoft Excel spreadsheets (legacy format).
	MimeTypeXLS MimeType = "application/vnd.ms-excel"
	// MimeTypeXLSX represents Microsoft Excel spreadsheets.
	MimeTypeXLSX MimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	// MimeTypeXLSM represents Microsoft Excel spreadsheets (macros enabled).
	MimeTypeXLSM MimeType = "application/vnd.ms-excel.sheet.macroenabled.12"

	// Note: The following formats are not explicitly supported in Tensorlake.
	//
	// MimeTypeDOC   MimeType = "application/msword"
	// MimeTypeTIFF  MimeType = "image/tiff"
	// MimeTypeMD    MimeType = "text/markdown"
	// MimeTypeXMD   MimeType = "text/x-markdown"
	// MimeTypeXML   MimeType = "text/xml"
	// MimeTypeOCTET MimeType = "application/octet-stream"
)

// ModelProvider represents the LLM provider to use for structured data extraction.
type ModelProvider string

const (
	// ModelProviderTensorlake represents private models, running on Tensorlake infrastructure.
	ModelProviderTensorlake ModelProvider = "tensorlake"

	// ModelProviderGemini3 represents Google Gemini 3 models.
	ModelProviderGemini3 ModelProvider = "gemini-3"

	// ModelProviderSonnet represents Anthropic Sonnet models.
	ModelProviderSonnet ModelProvider = "sonnet"

	// ModelProviderGPT4oMini represents OpenAI GPT-4o-mini model.
	ModelProviderGPT4oMini ModelProvider = "gpt_4o_mini"
)

// OCRPipelineProvider represents the different models for OCR (Optical Character Recognition).
type OCRPipelineProvider string

const (
	// OCRPipelineProviderDefault is the default OCR model (same as model01).
	OCRPipelineProviderDefault OCRPipelineProvider = ""

	// OCRPipelineProviderTensorlake01 is fast but could have lower accuracy on complex tables.
	// It's good for legal documents with footnotes.
	OCRPipelineProviderTensorlake01 OCRPipelineProvider = "model01"

	// OCRPipelineProviderTensorlake02 is slower but could have higher accuracy on complex tables.
	// It's good for financial documents with merged cells.
	OCRPipelineProviderTensorlake02 OCRPipelineProvider = "model02"

	// OCRPipelineProviderTensorlake03 is a compact model delivered to on-premise users.
	// It takes about 2 minutes to startup on Tensorlake's Cloud because it's meant
	// for testing for users who are eventually going to deploy this model on
	// dedicated hardware in their own datacenter.
	OCRPipelineProviderTensorlake03 OCRPipelineProvider = "model03"

	// OCRPipelineProviderGemini3 calls Google Gemini 3 API for OCR processing.
	OCRPipelineProviderGemini3 OCRPipelineProvider = "gemini3"
)

// ParseStatus indicates the status of the parse job.
type ParseStatus string

const (
	// ParseStatusFailure means the job has failed.
	ParseStatusFailure ParseStatus = "failure"

	// ParseStatusPending means the job is waiting to be processed.
	ParseStatusPending ParseStatus = "pending"

	// ParseStatusProcessing means the job is currently being processed.
	ParseStatusProcessing ParseStatus = "processing"

	// ParseStatusSuccessful means the job has been successfully completed and the results are available.
	ParseStatusSuccessful ParseStatus = "successful"

	// ParseStatusDetectingLayout means the job is detecting the layout of the document.
	ParseStatusDetectingLayout ParseStatus = "detecting_layout"

	// ParseStatusLayoutDetected means the layout of the document has been detected.
	ParseStatusLayoutDetected ParseStatus = "detected_layout"

	// ParseStatusExtractingData means the job is extracting the data from the document.
	ParseStatusExtractingData ParseStatus = "extracting_data"

	// ParseStatusExtractedData means the data has been extracted from the document.
	ParseStatusExtractedData ParseStatus = "extracted_data"

	// ParseStatusFormattingOutput means the output is being formatted.
	ParseStatusFormattingOutput ParseStatus = "formatting_output"

	// ParseStatusFormattedOutput means the output has been formatted.
	ParseStatusFormattedOutput ParseStatus = "formatted_output"
)

// PartitionStrategy determines how documents are partitioned before structured data extraction.
//
// The API will return one structured data object per partition.
type PartitionStrategy string

const (
	// PartitionStrategyNone: No partitioning is applied.
	// The entire document is treated as a single unit for extraction.
	PartitionStrategyNone PartitionStrategy = "none"

	// PartitionStrategyPage: The document is partitioned by individual pages.
	// Each page is treated as a separate unit for extraction.
	PartitionStrategyPage PartitionStrategy = "page"

	// PartitionStrategySection: The document is partitioned into sections based on
	// detected section headers. Each section is treated as a separate unit for extraction.
	PartitionStrategySection PartitionStrategy = "section"

	// PartitionStrategyFragment: The document is partitioned by individual page elements.
	// Each fragment is treated as a separate unit for extraction.
	PartitionStrategyFragment PartitionStrategy = "fragment"

	// PartitionStrategyPatterns: The document is partitioned based on user-defined
	// start and end patterns.
	PartitionStrategyPatterns PartitionStrategy = "patterns"
)

// TableOutputMode is the format for tables extracted from the document.
type TableOutputMode string

const (
	// TableOutputModeHTML outputs tables as HTML strings.
	TableOutputModeHTML TableOutputMode = "html"
	// TableOutputModeMarkdown outputs tables as Markdown strings.
	TableOutputModeMarkdown TableOutputMode = "markdown"
)

// TableParsingFormat determines which model the system uses to identify
// and extract tables from the document.
type TableParsingFormat string

const (
	// TableParsingFormatTSR identifies the structure of the table first,
	// then the cells of the tables. Better suited for clean, grid-like tables.
	TableParsingFormatTSR TableParsingFormat = "tsr"
	// TableParsingFormatVLM uses a vision language model to identify
	// and extract the cells of the tables. Better suited for tables
	// with merged cells or irregular structures.
	TableParsingFormatVLM TableParsingFormat = "vlm"
)

// PageFragmentType represents the type of a page fragment.
type PageFragmentType string

const (
	PageFragmentTypeSectionHeader  PageFragmentType = "section_header"
	PageFragmentTypeTitle          PageFragmentType = "title"
	PageFragmentTypeText           PageFragmentType = "text"
	PageFragmentTypeTable          PageFragmentType = "table"
	PageFragmentTypeFigure         PageFragmentType = "figure"
	PageFragmentTypeFormula        PageFragmentType = "formula"
	PageFragmentTypeForm           PageFragmentType = "form"
	PageFragmentTypeKeyValueRegion PageFragmentType = "key_value_region"
	PageFragmentTypeDocumentIndex  PageFragmentType = "document_index"
	PageFragmentTypeListItem       PageFragmentType = "list_item"
	PageFragmentTypeTableCaption   PageFragmentType = "table_caption"
	PageFragmentTypeFigureCaption  PageFragmentType = "figure_caption"
	PageFragmentTypeFormulaCaption PageFragmentType = "formula_caption"
	PageFragmentTypePageFooter     PageFragmentType = "page_footer"
	PageFragmentTypePageHeader     PageFragmentType = "page_header"
	PageFragmentTypePageNumber     PageFragmentType = "page_number"
	PageFragmentTypeSignature      PageFragmentType = "signature"
	PageFragmentTypeStrikethrough  PageFragmentType = "strikethrough"
	PageFragmentTypeBarcode        PageFragmentType = "barcode"
)

type PaginationDirection string

const (
	PaginationDirectionNext PaginationDirection = "next"
	PaginationDirectionPrev PaginationDirection = "prev"
)

type DatasetStatus string

const (
	DatasetStatusIdle       DatasetStatus = "idle"
	DatasetStatusProcessing DatasetStatus = "processing"
)

type JobType string

const (
	JobTypeParse    JobType = "parse"
	JobTypeRead     JobType = "read"
	JobTypeExtract  JobType = "extract"
	JobTypeClassify JobType = "classify"
	JobTypeLegacy   JobType = "legacy"
	JobTypeDataset  JobType = "dataset"
)
