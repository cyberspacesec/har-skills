// Package har provides functionality for parsing and manipulating HAR (HTTP Archive) files.
// This is a compatibility wrapper that forwards to the implementation in pkg/har.
package har

import (
	"github.com/cyberspacesec/har-skills/pkg/har"
)

// Har represents a HAR file
type Har = har.Har

// Log represents the log section of a HAR file
type Log = har.Log

// Creator represents the creator information
type Creator = har.Creator

// Browser represents browser information
type Browser = har.Browser

// PageTimings represents page timing information
type PageTimings = har.PageTimings

// Pages represents a page in the HAR file
type Pages = har.Pages

// Headers represents HTTP headers
type Headers = har.Headers

// QueryString represents URL query parameters
type QueryString = har.QueryString

// PostData represents HTTP POST request data
type PostData = har.PostData

// Param represents a form parameter in POST data
type Param = har.Param

// Request represents an HTTP request
type Request = har.Request

// Cookie represents an HTTP cookie
type Cookie = har.Cookie

// Content represents response content
type Content = har.Content

// Response represents an HTTP response
type Response = har.Response

// BeforeRequest represents cache state before request
type BeforeRequest = har.BeforeRequest

// AfterRequest represents cache state after request
type AfterRequest = har.AfterRequest

// Cache represents cache information
type Cache = har.Cache

// Timings represents timing information for a request
type Timings = har.Timings

// Entries represents an entry in the HAR file
type Entries = har.Entries

// Initiator represents the initiator of a request
type Initiator = har.Initiator

// Stack represents a call stack
type Stack = har.Stack

// Parent represents parent information in a call stack
type Parent = har.Parent

// ParentID represents a parent ID in a call stack
type ParentID = har.ParentID

// CallFrame represents a frame in a call stack
type CallFrame = har.CallFrame

// CustomFields stores non-standard extension fields
type CustomFields = har.CustomFields

// HTTPMethod enum type for HTTP methods
type HTTPMethod = har.HTTPMethod

// HTTP Method constants
const (
	MethodUnknown = har.MethodUnknown
	MethodGET     = har.MethodGET
	MethodPOST    = har.MethodPOST
	MethodPUT     = har.MethodPUT
	MethodDELETE  = har.MethodDELETE
	MethodHEAD    = har.MethodHEAD
	MethodOPTIONS = har.MethodOPTIONS
	MethodPATCH   = har.MethodPATCH
	MethodCONNECT = har.MethodCONNECT
	MethodTRACE   = har.MethodTRACE
)

// ConvertFormat for conversion formats
type ConvertFormat = har.ConvertFormat

// Format constants
const (
	FormatCSV      = har.FormatCSV
	FormatMarkdown = har.FormatMarkdown
	FormatHTML     = har.FormatHTML
	FormatText     = har.FormatText
	FormatJSON     = har.FormatJSON
	FormatYAML     = har.FormatYAML
)

// HAR specification version constants
const (
	HarSpecVersion11 = har.HarSpecVersion11
	HarSpecVersion12 = har.HarSpecVersion12
	HarSpecVersion13 = har.HarSpecVersion13
)

// Error types
type (
	ErrorCode              = har.ErrorCode
	HarError               = har.HarError
	ParseOptions           = har.ParseOptions
	FilterOptions          = har.FilterOptions
	FilterResult           = har.FilterResult
	Result                 = har.Result
	ConvertOptions         = har.ConvertOptions
	OptimizedHar           = har.OptimizedHar
	OptimizedEntries       = har.OptimizedEntries
	OptimizedRequest       = har.OptimizedRequest
	OptimizedResponse      = har.OptimizedResponse
	OptimizedContent       = har.OptimizedContent
	OptimizedTimings       = har.OptimizedTimings
	StreamingHar           = har.StreamingHar
	EntryIterator          = har.EntryIterator
	StreamingEntryIterator = har.StreamingEntryIterator
	LazyHar                = har.LazyHar
	LazyContent            = har.LazyContent
	LazyResponse           = har.LazyResponse
	LazyEntries            = har.LazyEntries

	// Statistics types
	HarStatistics  = har.HarStatistics
	TimingsSummary = har.TimingsSummary
	DomainStats    = har.DomainStats

	// Diff types
	HarDiff       = har.HarDiff
	DiffEntry     = har.DiffEntry
	ModifiedEntry = har.ModifiedEntry
	FieldChange   = har.FieldChange
	DiffOptions   = har.DiffOptions

	// Replay types
	ReplayOptions = har.ReplayOptions
	ReplayResult  = har.ReplayResult

	// Merge types
	MergeOptions = har.MergeOptions

	// Builder types
	HarBuilder   = har.HarBuilder
	EntryBuilder = har.EntryBuilder
	Recorder     = har.Recorder

	// Interface types
	HARProvider         = har.HARProvider
	EntryProvider       = har.EntryProvider
	RequestProvider     = har.RequestProvider
	ResponseProvider    = har.ResponseProvider
	HeaderProvider      = har.HeaderProvider
	CookieProvider      = har.CookieProvider
	ContentProvider     = har.ContentProvider
	TimingsProvider     = har.TimingsProvider
	PageProvider        = har.PageProvider
	PageTimingsProvider = har.PageTimingsProvider

	// Option type
	Option = har.Option

	// Functional option types
	FilterOption     = har.FilterOption
	ReplayOption     = har.ReplayOption
	ConvertOption    = har.ConvertOption
	DiffOption       = har.DiffOption
	MergeOption      = har.MergeOption
	HarBuilderOption = har.HarBuilderOption

	// Validation extension types
	ValidationRule  = har.ValidationRule
	ValidationError = har.ValidationError

	// Postman export types
	PostmanCollection = har.PostmanCollection
	PostmanInfo       = har.PostmanInfo
	PostmanItem       = har.PostmanItem
	PostmanRequest    = har.PostmanRequest
	PostmanHeader     = har.PostmanHeader
	PostmanURL        = har.PostmanURL
	PostmanQuery      = har.PostmanQuery
	PostmanBody       = har.PostmanBody

	// XML export types
	XMLElement   = har.XMLElement
	HARXML       = har.HARXML
	LogXML       = har.LogXML
	CreatorXML   = har.CreatorXML
	EntryXML     = har.EntryXML
	RequestXML   = har.RequestXML
	ResponseXML  = har.ResponseXML
	HeaderXML    = har.HeaderXML
	PostDataXML  = har.PostDataXML
	ContentXML   = har.ContentXML

	// Redaction types
	RedactOptions = har.RedactOptions
	RedactURLRule = har.RedactURLRule

	// Content analysis types
	MIMECategory   = har.MIMECategory
	ContentSummary = har.ContentSummary

	// Timeline types
	WaterfallEntry    = har.WaterfallEntry
	TimingPhases      = har.TimingPhases
	SLARule           = har.SLARule
	SLAResult         = har.SLAResult
	ConcurrencyPoint  = har.ConcurrencyPoint
	PageTimingMetrics = har.PageTimingMetrics

	// Security types
	SecurityFinding      = har.SecurityFinding
	SecurityReport       = har.SecurityReport
	SecurityAuditOptions = har.SecurityAuditOptions
	SecurityCheckFunc    = har.SecurityCheckFunc

	// Transform types
	TransformRule = har.TransformRule
	TransformType = har.TransformType

	// Dedup types
	DedupStrategy      = har.DedupStrategy
	DeduplicateOptions = har.DeduplicateOptions
	DuplicateGroup     = har.DuplicateGroup

	// Index types
	HarIndex   = har.HarIndex
	IndexStats = har.IndexStats

	// Cookie analysis types
	CookieFinding        = har.CookieFinding
	CookieAuditReport    = har.CookieAuditReport
	CookieEvolutionEntry = har.CookieEvolutionEntry

	// Cache analysis types
	CacheControlDirectives = har.CacheControlDirectives
	CacheEntryAssessment   = har.CacheEntryAssessment
	CacheReport            = har.CacheReport

	// Performance types
	PerformanceCategory = har.PerformanceCategory
	PerformanceFinding  = har.PerformanceFinding
	PerformanceReport   = har.PerformanceReport
)

// Error code constants
const (
	ErrCodeUnknown       = har.ErrCodeUnknown
	ErrCodeFileSystem    = har.ErrCodeFileSystem
	ErrCodeJSONParse     = har.ErrCodeJSONParse
	ErrCodeInvalidFormat = har.ErrCodeInvalidFormat
	ErrCodeValidation    = har.ErrCodeValidation
	ErrCodeMissingField  = har.ErrCodeMissingField
	ErrCodeInvalidValue  = har.ErrCodeInvalidValue
	ErrCodeUnsupported   = har.ErrCodeUnsupported
)

// Forward all functions
var (
	// Basic operations
	ParseHarFile = har.ParseHarFile
	ParseHar     = har.ParseHar
	NewHar       = har.NewHar

	// Optimized parsing
	ParseHarFileOptimized = har.ParseHarFileOptimized
	ParseHarOptimized     = har.ParseHarOptimized
	ToOptimizedHar        = har.ToOptimizedHar

	// Lazy loading
	ParseHarWithLazyLoading     = har.ParseHarWithLazyLoading
	ParseHarFileWithLazyLoading = har.ParseHarFileWithLazyLoading

	// Streaming
	NewStreamingHarFromFile = har.NewStreamingHarFromFile

	// Enhanced parsing
	ParseHarWithOptions      = har.ParseHarWithOptions
	ParseHarFileWithOptions  = har.ParseHarFileWithOptions
	ParseHarEnhanced         = har.ParseHarEnhanced
	ParseHarFileEnhanced     = har.ParseHarFileEnhanced
	ParseHarLenient          = har.ParseHarLenient
	ParseHarFileLenient      = har.ParseHarFileLenient
	ParseHarWithWarnings     = har.ParseHarWithWarnings
	ParseHarFileWithWarnings = har.ParseHarFileWithWarnings
	DefaultParseOptions      = har.DefaultParseOptions

	// Reader-based parsing
	ParseHarFromReader            = har.ParseHarFromReader
	ParseHarFromReaderWithOptions = har.ParseHarFromReaderWithOptions
	ParseFromReader               = har.ParseFromReader
	ParseHarFileGzipped           = har.ParseHarFileGzipped
	ParseHarFileAuto              = har.ParseHarFileAuto
	NewStreamingParserFromReader  = har.NewStreamingParserFromReader
	SaveToFileGzipped             = har.SaveToFileGzipped

	// Error utilities
	NewHarError            = har.NewHarError
	NewFileSystemError     = har.NewFileSystemError
	NewJSONParseError      = har.NewJSONParseError
	WrapJSONUnmarshalError = har.WrapJSONUnmarshalError
	NewValidationError     = har.NewValidationError
	NewInvalidFormatError  = har.NewInvalidFormatError
	NewMissingFieldError   = har.NewMissingFieldError
	NewInvalidValueError   = har.NewInvalidValueError
	NewUnsupportedError    = har.NewUnsupportedError

	// Utilities
	ParseMethod           = har.ParseMethod
	DefaultConvertOptions = har.DefaultConvertOptions
	ExtractDomain         = har.ExtractDomain

	// Validation
	ValidateHarFile          = har.ValidateHarFile
	IsValidHarVersion        = har.IsValidHarVersion
	DetectHarVersion         = har.DetectHarVersion
	RegisterValidator        = har.RegisterValidator
	UnregisterValidator      = har.UnregisterValidator
	ListValidators           = har.ListValidators
	ValidateWithRules        = har.ValidateWithRules
	ValidateStrict           = har.ValidateStrict
	ValidateURL              = har.ValidateURL
	ValidateTimingsConsistency = har.ValidateTimingsConsistency

	// Diff
	Diff          = har.Diff
	DiffWith      = har.DiffWith
	DefaultDiffOptions = har.DefaultDiffOptions
	NewDiffOptions    = har.NewDiffOptions

	// Merge
	Merge            = har.Merge
	MergeWithOptions = har.MergeWithOptions
	MergeWith        = har.MergeWith
	DefaultMergeOptions = har.DefaultMergeOptions
	NewMergeOptions     = har.NewMergeOptions

	// Replay
	DefaultReplayOptions = har.DefaultReplayOptions
	NewReplayOptions     = har.NewReplayOptions
	HTTPResponseToEntries = har.HTTPResponseToEntries
	ReplayResultsToHar    = har.ReplayResultsToHar

	// Functional options parsing API
	Parse                      = har.Parse
	ParseFile                  = har.ParseFile
	NewStreamingParser         = har.NewStreamingParser
	NewStreamingParserFromFile = har.NewStreamingParserFromFile

	// Parse option functions
	WithLenient         = har.WithLenient
	WithSkipValidation  = har.WithSkipValidation
	WithCollectWarnings = har.WithCollectWarnings
	WithMaxWarnings     = har.WithMaxWarnings
	WithMemoryOptimized = har.WithMemoryOptimized
	WithLazyLoading     = har.WithLazyLoading
	WithStreaming       = har.WithStreaming
	WithHarVersion      = har.WithHarVersion
	WithAutoDetectVersion = har.WithAutoDetectVersion

	// Predefined option groups
	OptMemoryEfficient = har.OptMemoryEfficient
	OptFast            = har.OptFast
	OptLenient         = har.OptLenient
	OptPerformance     = har.OptPerformance

	// Filter functional options
	WithFilterURL               = har.WithFilterURL
	WithFilterMethod            = har.WithFilterMethod
	WithFilterStatusCode        = har.WithFilterStatusCode
	WithFilterStatusCodeRange   = har.WithFilterStatusCodeRange
	WithFilterContentType       = har.WithFilterContentType
	WithFilterTimeRange         = har.WithFilterTimeRange
	WithFilterDuration          = har.WithFilterDuration
	WithFilterResourceType      = har.WithFilterResourceType
	WithFilterHasError          = har.WithFilterHasError
	WithFilterHeader            = har.WithFilterHeader
	WithFilterResponseHeader    = har.WithFilterResponseHeader
	WithFilterRegex             = har.WithFilterRegex
	NewFilterOptions            = har.NewFilterOptions

	// Replay functional options
	WithReplayTimeout        = har.WithReplayTimeout
	WithReplayFollowRedirects = har.WithReplayFollowRedirects
	WithReplayMaxRedirects   = har.WithReplayMaxRedirects
	WithReplaySkipSSLVerify  = har.WithReplaySkipSSLVerify
	WithReplayOverrideHeader = har.WithReplayOverrideHeader
	WithReplayTransport      = har.WithReplayTransport

	// Convert functional options
	WithConvertIncludeHeaders       = har.WithConvertIncludeHeaders
	WithConvertIncludeTimings       = har.WithConvertIncludeTimings
	WithConvertIncludeBodies        = har.WithConvertIncludeBodies
	WithConvertIncludeCookies       = har.WithConvertIncludeCookies
	WithConvertIncludeQueryStrings  = har.WithConvertIncludeQueryStrings
	WithConvertIncludeStatus        = har.WithConvertIncludeStatus
	WithConvertIncludeSize          = har.WithConvertIncludeSize
	WithConvertIncludeURL           = har.WithConvertIncludeURL
	WithConvertIncludeMethod        = har.WithConvertIncludeMethod
	WithConvertIncludeTime          = har.WithConvertIncludeTime
	WithConvertIncludeMimeType      = har.WithConvertIncludeMimeType
	WithConvertHeaders              = har.WithConvertHeaders
	WithConvertFilter               = har.WithConvertFilter
	NewConvertOptions               = har.NewConvertOptions

	// Diff functional options
	WithDiffIgnoreHeaders = har.WithDiffIgnoreHeaders
	WithDiffIgnoreTimings = har.WithDiffIgnoreTimings
	WithDiffIgnoreDates   = har.WithDiffIgnoreDates
	WithDiffIgnoreCache   = har.WithDiffIgnoreCache
	WithDiffIgnoreComment = har.WithDiffIgnoreComment
	WithDiffNormalizeURL  = har.WithDiffNormalizeURL
	WithDiffCompareByURL  = har.WithDiffCompareByURL
	WithDiffIncludeBody   = har.WithDiffIncludeBody

	// Merge functional options
	WithMergeSortByTime  = har.WithMergeSortByTime
	WithMergeDeduplicate = har.WithMergeDeduplicate

	// Builder functional options
	WithBuilderVersion = har.WithBuilderVersion
	WithBuilderCreator = har.WithBuilderCreator
	WithBuilderBrowser = har.WithBuilderBrowser
	WithBuilderComment = har.WithBuilderComment
	NewHarBuilderWithOptions = har.NewHarBuilderWithOptions

	// Utility functions
	BuildQueryStringFromURL = har.BuildQueryStringFromURL
	ParseResponseHeaders    = har.ParseResponseHeaders
	EstimateHeaderSize      = har.EstimateHeaderSize
	FormatBytes             = har.FormatBytes
	ReadBody                = har.ReadBody
	WriteRequestToWriter    = har.WriteRequestToWriter
	CloneEntry              = har.CloneEntry
	WriteToWriter           = har.WriteToWriter
	WriteEntriesToWriter    = har.WriteEntriesToWriter
	ReadEntriesFromReader   = har.ReadEntriesFromReader
)

// MIME category constants for content analysis
const (
	MIMEImage      = har.MIMEImage
	MIMEScript     = har.MIMEScript
	MIMEStylesheet = har.MIMEStylesheet
	MIMEFont       = har.MIMEFont
	MIMEMedia      = har.MIMEMedia
	MIMEDocument   = har.MIMEDocument
	MIMEAPI        = har.MIMEAPI
	MIMEData       = har.MIMEData
	MIMEOther      = har.MIMEOther
)

// Dedup strategy constants
const (
	DedupExactURL     = har.DedupExactURL
	DedupURLPattern   = har.DedupURLPattern
	DedupContentHash  = har.DedupContentHash
)

// Transform type constants
const (
	TransformURLRewrite      = har.TransformURLRewrite
	TransformHostReplace     = har.TransformHostReplace
	TransformSchemeChange    = har.TransformSchemeChange
	TransformHeaderAdd       = har.TransformHeaderAdd
	TransformHeaderRemove    = har.TransformHeaderRemove
	TransformHeaderReplace   = har.TransformHeaderReplace
	TransformQueryParamRemove = har.TransformQueryParamRemove
	TransformQueryParamAdd   = har.TransformQueryParamAdd
	TransformCookieDomainRewrite = har.TransformCookieDomainRewrite
	TransformBodyReplace     = har.TransformBodyReplace
)

// New module functions
var (
	// Redaction
	DefaultRedactOptions = har.DefaultRedactOptions

	// Security audit
	DefaultSecurityAuditOptions = har.DefaultSecurityAuditOptions

	// Dedup
	IsCacheBusterParam         = har.IsCacheBusterParam
	IsCacheBusterParamWithValue = har.IsCacheBusterParamWithValue
	DefaultDeduplicateOptions  = har.DefaultDeduplicateOptions

	// Cache analysis
	ParseCacheControl = har.ParseCacheControl

	// Decode/Compress
	DecompressByEncoding   = har.DecompressByEncoding
	DecompressWithEncoding = har.DecompressWithEncoding
	CompressContent        = har.CompressContent
)
