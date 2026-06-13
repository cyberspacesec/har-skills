package har

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// MIMECategory 表示MIME类型的分类
type MIMECategory string

const (
	MIMEImage      MIMECategory = "image"
	MIMEScript     MIMECategory = "script"
	MIMEStylesheet MIMECategory = "stylesheet"
	MIMEFont       MIMECategory = "font"
	MIMEMedia      MIMECategory = "media"
	MIMEDocument   MIMECategory = "document"
	MIMEAPI        MIMECategory = "api"
	MIMEData       MIMECategory = "data"
	MIMEOther      MIMECategory = "other"
)

// ContentSummary 表示HAR中所有内容的摘要信息
type ContentSummary struct {
	TotalSize      int                    // 总大小
	TextSize       int                    // 文本内容大小
	BinarySize     int                    // 二进制内容大小
	CompressedSize int                    // 压缩后大小
	ByCategory     map[MIMECategory]int   // 按分类统计大小
	ByMIMEType     map[string]int         // 按具体MIME类型统计大小
}

// MIMECategory 返回内容的MIME分类
//
// 根据Content.MimeType字段判断内容属于哪个分类。
// 支持常见的MIME类型分类，无法识别的类型归类为MIMEOther。
func (c *Content) MIMECategory() MIMECategory {
	if c == nil {
		return MIMEOther
	}

	mime := strings.ToLower(c.MimeType)
	// Remove parameters (e.g., "text/html; charset=utf-8" -> "text/html")
	if idx := strings.Index(mime, ";"); idx >= 0 {
		mime = strings.TrimSpace(mime[:idx])
	}

	// Image
	if strings.HasPrefix(mime, "image/") {
		return MIMEImage
	}

	// Script
	if mime == "application/javascript" || mime == "application/x-javascript" ||
		mime == "text/javascript" || mime == "text/x-javascript" ||
		mime == "application/ecmascript" || mime == "text/ecmascript" ||
		strings.HasPrefix(mime, "application/vnd.dart") {
		return MIMEScript
	}

	// Stylesheet
	if mime == "text/css" || mime == "text/x-css" ||
		strings.HasPrefix(mime, "application/x-css") {
		return MIMEStylesheet
	}

	// Font
	if strings.HasPrefix(mime, "font/") ||
		mime == "application/x-font-ttf" || mime == "application/x-font-woff" ||
		mime == "application/font-woff" || mime == "application/font-woff2" ||
		mime == "application/x-font-opentype" || mime == "application/vnd.ms-fontobject" ||
		strings.HasPrefix(mime, "application/x-font-") {
		return MIMEFont
	}

	// Media (audio/video)
	if strings.HasPrefix(mime, "audio/") || strings.HasPrefix(mime, "video/") {
		return MIMEMedia
	}

	// Document
	if mime == "text/html" || mime == "application/xhtml+xml" ||
		mime == "text/xml" || mime == "application/xml" ||
		mime == "text/plain" || mime == "text/richtext" ||
		mime == "application/pdf" || mime == "application/msword" ||
		strings.HasPrefix(mime, "application/vnd.") && !strings.Contains(mime, "json") {
		return MIMEDocument
	}

	// API
	if mime == "application/json" || strings.HasSuffix(mime, "+json") ||
		mime == "text/json" || mime == "application/graphql" {
		return MIMEAPI
	}

	// Data
	if mime == "text/csv" || mime == "text/tab-separated-values" ||
		mime == "application/x-www-form-urlencoded" ||
		mime == "multipart/form-data" ||
		mime == "application/octet-stream" ||
		strings.HasPrefix(mime, "application/vnd.") && strings.Contains(mime, "json") {
		return MIMEData
	}

	// Other text types not yet classified
	if strings.HasPrefix(mime, "text/") {
		return MIMEDocument
	}

	return MIMEOther
}

// IsBinary 检测内容是否为二进制
//
// 通过MIME类型和内容字节检测来判断内容是否为二进制。
// 文本类型的MIME（text/*, application/json, application/xml, application/javascript等）
// 被视为非二进制。如果内容文本可用，还会使用http.DetectContentType()进行检测。
func (c *Content) IsBinary() bool {
	if c == nil {
		return false
	}

	// Check declared MIME type first
	if isTextMIME(c.MimeType) {
		return false
	}

	// If content text is available, use http.DetectContentType for further detection
	data, err := c.DecodeContent()
	if err == nil && len(data) > 0 {
		detected := http.DetectContentType(data)
		if isTextMIME(detected) {
			return false
		}
	}

	return true
}

// IsText 检测内容是否为文本
//
// IsText是IsBinary的相反判断，文本内容返回true。
func (c *Content) IsText() bool {
	return !c.IsBinary()
}

// DetectMIMEType 使用http.DetectContentType检测内容的实际MIME类型
//
// 如果内容文本可用，会先解码内容字节再检测MIME类型。
// 如果无法检测（内容为空或解码失败），则回退到Content.MimeType字段。
func (c *Content) DetectMIMEType() string {
	if c == nil {
		return ""
	}

	data, err := c.DecodeContent()
	if err == nil && len(data) > 0 {
		detected := http.DetectContentType(data)
		// http.DetectContentType returns "application/octet-stream" when it can't detect
		if detected != "application/octet-stream" {
			return detected
		}
	}

	return c.MimeType
}

// Hash 计算内容的SHA-256哈希值
//
// 对解码后的内容字节计算SHA-256哈希，返回十六进制编码的哈希字符串。
func (c *Content) Hash() (string, error) {
	if c == nil {
		return "", NewInvalidFormatError("内容为空")
	}

	data, err := c.DecodeContent()
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", NewInvalidFormatError("内容数据为空")
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// ParseJSON 将内容文本解析为JSON值
//
// 返回解析后的JSON值（可以是对象、数组、字符串等）。
// 如果内容为空或不是有效的JSON，返回错误。
func (c *Content) ParseJSON() (interface{}, error) {
	if c == nil {
		return nil, NewInvalidFormatError("内容为空")
	}

	data, err := c.DecodeContent()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, NewInvalidFormatError("内容数据为空")
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, NewHarError(ErrCodeJSONParse,
			fmt.Sprintf("JSON解析失败: %v", err), err)
	}

	return result, nil
}

// ParseAsMap 将内容文本解析为JSON对象(map)
//
// 尝试将内容解析为JSON对象(map[string]interface{})。
// 如果内容不是JSON对象（如数组或字符串），返回错误。
func (c *Content) ParseAsMap() (map[string]interface{}, error) {
	if c == nil {
		return nil, NewInvalidFormatError("内容为空")
	}

	data, err := c.DecodeContent()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, NewInvalidFormatError("内容数据为空")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, NewHarError(ErrCodeJSONParse,
			fmt.Sprintf("JSON对象解析失败: %v", err), err)
	}

	return result, nil
}

// ContentLength 返回响应的Content-Length头部值
//
// 从Response.Headers中查找Content-Length头部并返回其整数值。
// 如果头部不存在，返回-1。
func (e *Entries) ContentLength() int {
	if e == nil {
		return -1
	}

	for _, header := range e.Response.Headers {
		if strings.EqualFold(header.Name, "Content-Length") {
			val, err := strconv.Atoi(strings.TrimSpace(header.Value))
			if err != nil {
				return -1
			}
			return val
		}
	}

	return -1
}

// HasContentLengthMismatch 检查Content-Length头部值与实际内容大小是否不匹配
//
// 比较Content-Length头部值与Response.Content.Size，如果不一致返回true。
func (e *Entries) HasContentLengthMismatch() bool {
	if e == nil {
		return false
	}

	contentLen := e.ContentLength()
	if contentLen < 0 {
		return false
	}

	return contentLen != e.Response.Content.Size
}

// EstimateTransferSize 估算实际传输大小
//
// 考虑压缩等因素，估算内容的实际网络传输大小。
// 优先使用Response.TransferSize（Chrome扩展字段），
// 其次使用Response.BodySize减去Compression，最后使用Content.Size。
func (e *Entries) EstimateTransferSize() int {
	if e == nil {
		return 0
	}

	// Prefer TransferSize (Chrome extension) if available
	if e.Response.TransferSize > 0 {
		return e.Response.TransferSize
	}

	// Use BodySize minus Compression savings
	if e.Response.BodySize > 0 {
		compression := e.Response.Content.Compression
		if compression > 0 && e.Response.BodySize > compression {
			return e.Response.BodySize - compression
		}
		return e.Response.BodySize
	}

	// Fall back to Content.Size
	return e.Response.Content.Size
}

// ContentSummary 返回HAR中所有内容的摘要信息
//
// 统计所有条目的内容类型和大小，包括总大小、文本/二进制大小、
// 压缩大小，以及按MIME分类和具体MIME类型的大小统计。
func (h *Har) ContentSummary() *ContentSummary {
	if h == nil {
		return nil
	}

	summary := &ContentSummary{
		ByCategory: make(map[MIMECategory]int),
		ByMIMEType: make(map[string]int),
	}

	for _, entry := range h.Log.Entries {
		content := entry.Response.Content
		size := content.Size
		if size < 0 {
			size = 0
		}

		summary.TotalSize += size

		category := content.MIMECategory()
		summary.ByCategory[category] += size

		mimeKey := content.MimeType
		if mimeKey == "" {
			mimeKey = "unknown"
		}
		summary.ByMIMEType[mimeKey] += size

		if content.IsText() {
			summary.TextSize += size
		} else {
			summary.BinarySize += size
		}

		compression := content.Compression
		if compression > 0 {
			summary.CompressedSize += compression
		}
	}

	return summary
}

// SaveToFile 将解码后的内容保存到文件
//
// 自动处理base64解码和内容解压缩，将最终的原始数据写入指定路径的文件。
func (c *Content) SaveToFile(path string) error {
	if c == nil {
		return NewInvalidFormatError("内容为空")
	}

	data, err := c.DecodeContent()
	if err != nil {
		return err
	}
	if data == nil {
		// Empty content — write zero-length file
		data = []byte{}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return NewFileSystemError(fmt.Sprintf("写入文件 '%s' 失败", path), err)
	}

	return nil
}

// isTextMIME 检查MIME类型是否为文本类型
func isTextMIME(mime string) bool {
	if mime == "" {
		return false
	}

	lower := strings.ToLower(mime)
	// Remove parameters
	if idx := strings.Index(lower, ";"); idx >= 0 {
		lower = strings.TrimSpace(lower[:idx])
	}

	// text/* is always text
	if strings.HasPrefix(lower, "text/") {
		return true
	}

	// Common application types that are actually text
	textApplicationTypes := []string{
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-javascript",
		"application/ecmascript",
		"application/graphql",
		"application/xhtml+xml",
		"application/atom+xml",
		"application/rss+xml",
		"application/soap+xml",
		"application/x-yaml",
		"application/yaml",
		"application/toml",
		"application/ld+json",
		"application/manifest+json",
		"application/schema+json",
		"application/vnd.api+json",
	}

	for _, t := range textApplicationTypes {
		if lower == t {
			return true
		}
	}

	// Any type ending with +json or +xml is text
	if strings.HasSuffix(lower, "+json") || strings.HasSuffix(lower, "+xml") {
		return true
	}

	return false
}