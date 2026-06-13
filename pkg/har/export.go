package har

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

// FormatJSON JSON格式常量
const FormatJSON ConvertFormat = "json"

// ---------------------------------------------------------------------------
// cURL 导出
// ---------------------------------------------------------------------------

// ToCurl 生成所有条目的cURL命令
func (h *Har) ToCurl() string {
	if h == nil || len(h.Log.Entries) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, entry := range h.Log.Entries {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(entryToCurl(&entry))
	}
	return sb.String()
}

// ToCurl 生成单条目的cURL命令
func (e *Entries) ToCurl() string {
	if e == nil {
		return ""
	}
	return entryToCurl(e)
}

// entryToCurl 将单条HAR条目转换为cURL命令
func entryToCurl(entry *Entries) string {
	var parts []string

	// curl 命令
	parts = append(parts, "curl")

	// 非GET方法使用 -X
	method := strings.ToUpper(entry.Request.Method)
	if method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", method))
	}

	// 请求头
	for _, h := range entry.Request.Headers {
		// 跳过 Host 头，curl会自动添加
		if strings.EqualFold(h.Name, "Host") {
			continue
		}
		escaped := escapeSingleQuotes(h.Value)
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", h.Name, escaped))
	}

	// POST数据
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		escaped := escapeSingleQuotes(entry.Request.PostData.Text)
		parts = append(parts, fmt.Sprintf("--data '%s'", escaped))
	}

	// 检查 Accept-Encoding 是否包含 gzip/deflate
	if hasAcceptEncoding(entry) {
		parts = append(parts, "--compressed")
	}

	// 检查是否跳过SSL验证（根据URL判断是否为HTTPS）
	parsedURL, err := url.Parse(entry.Request.URL)
	if err == nil && parsedURL.Scheme == "https" {
		// 如果存在 _error 字段或URL为自签名证书场景，添加 -k
		// 这里保守地检查响应中是否有SSL相关错误
		if entry.Response.Error != nil {
			parts = append(parts, "-k")
		}
	}

	// URL（单引号包裹）
	parts = append(parts, fmt.Sprintf("'%s'", entry.Request.URL))

	return strings.Join(parts, " \\\n  ")
}

// ---------------------------------------------------------------------------
// Wget 导出
// ---------------------------------------------------------------------------

// ToWget 生成所有条目的wget命令
func (h *Har) ToWget() string {
	if h == nil || len(h.Log.Entries) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, entry := range h.Log.Entries {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(entryToWget(&entry))
	}
	return sb.String()
}

// ToWget 生成单条目的wget命令
func (e *Entries) ToWget() string {
	if e == nil {
		return ""
	}
	return entryToWget(e)
}

// entryToWget 将单条HAR条目转换为wget命令
func entryToWget(entry *Entries) string {
	var parts []string

	parts = append(parts, "wget")

	method := strings.ToUpper(entry.Request.Method)
	// wget 默认是GET，对于非GET需要使用 --method
	if method != "GET" {
		parts = append(parts, fmt.Sprintf("--method=%s", method))
	}

	// 请求头
	for _, h := range entry.Request.Headers {
		if strings.EqualFold(h.Name, "Host") {
			continue
		}
		escaped := escapeSingleQuotes(h.Value)
		parts = append(parts, fmt.Sprintf("--header='%s: %s'", h.Name, escaped))
	}

	// POST数据
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		escaped := escapeSingleQuotes(entry.Request.PostData.Text)
		parts = append(parts, fmt.Sprintf("--post-data='%s'", escaped))
	}

	// 不验证SSL
	parsedURL, err := url.Parse(entry.Request.URL)
	if err == nil && parsedURL.Scheme == "https" {
		parts = append(parts, "--no-check-certificate")
	}

	// 静默模式 + 输出到stdout
	parts = append(parts, "-qO-")

	// URL
	parts = append(parts, fmt.Sprintf("'%s'", entry.Request.URL))

	return strings.Join(parts, " \\\n  ")
}

// ---------------------------------------------------------------------------
// Python Requests 导出
// ---------------------------------------------------------------------------

// ToPythonRequests 生成所有条目的Python requests代码
func (h *Har) ToPythonRequests() string {
	if h == nil || len(h.Log.Entries) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("import requests\n\n")
	for i, entry := range h.Log.Entries {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(entryToPythonRequests(&entry))
		sb.WriteString("\n")
	}
	return sb.String()
}

// ToPythonRequests 生成单条目的Python requests代码
func (e *Entries) ToPythonRequests() string {
	if e == nil {
		return ""
	}
	return entryToPythonRequests(e)
}

// entryToPythonRequests 将单条HAR条目转换为Python requests代码
func entryToPythonRequests(entry *Entries) string {
	var sb strings.Builder

	method := strings.ToLower(entry.Request.Method)

	// 构建headers字典
	headers := buildHeadersDict(entry)

	// 构建请求调用
	if len(headers) > 0 {
		sb.WriteString(fmt.Sprintf("headers = %s\n", headers))
	}

	// 构建请求参数
	args := []string{fmt.Sprintf("'%s'", entry.Request.URL)}
	if len(headers) > 0 {
		args = append(args, "headers=headers")
	}

	// POST数据
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		escaped := escapePythonString(entry.Request.PostData.Text)
		args = append(args, fmt.Sprintf("data='%s'", escaped))
	}

	sb.WriteString(fmt.Sprintf("response = requests.%s(%s)\n", method, strings.Join(args, ", ")))
	sb.WriteString(fmt.Sprintf("print(response.status_code)\n"))
	sb.WriteString(fmt.Sprintf("print(response.text)\n"))

	return sb.String()
}

// buildHeadersDict 构建Python字典格式的headers字符串
func buildHeadersDict(entry *Entries) string {
	if len(entry.Request.Headers) == 0 {
		return ""
	}
	var pairs []string
	for _, h := range entry.Request.Headers {
		key := escapePythonString(h.Name)
		val := escapePythonString(h.Value)
		pairs = append(pairs, fmt.Sprintf("'%s': '%s'", key, val))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// ---------------------------------------------------------------------------
// Postman Collection v2.1 导出
// ---------------------------------------------------------------------------

// PostmanCollection 表示Postman Collection v2.1格式
type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanInfo Postman Collection信息
type PostmanInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

// PostmanItem Postman Collection中的请求项
type PostmanItem struct {
	Name    string         `json:"name"`
	Request PostmanRequest `json:"request"`
}

// PostmanRequest Postman请求定义
type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header,omitempty"`
	URL    PostmanURL      `json:"url"`
	Body   *PostmanBody    `json:"body,omitempty"`
}

// PostmanHeader Postman请求头
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanURL Postman URL定义
type PostmanURL struct {
	Raw      string            `json:"raw"`
	Protocol string            `json:"protocol"`
	Host     []string          `json:"host"`
	Path     []string          `json:"path"`
	Query    []PostmanQuery    `json:"query,omitempty"`
}

// PostmanQuery Postman查询参数
type PostmanQuery struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody Postman请求体
type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

// ToPostmanCollection 将HAR转换为Postman Collection v2.1格式JSON
func (h *Har) ToPostmanCollection() ([]byte, error) {
	if h == nil {
		return nil, fmt.Errorf("HAR对象为空")
	}

	collection := PostmanCollection{
		Info: PostmanInfo{
			Name:   "HAR Export",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: make([]PostmanItem, 0, len(h.Log.Entries)),
	}

	for i := range h.Log.Entries {
		entry := &h.Log.Entries[i]
		item := entryToPostmanItem(entry)
		collection.Item = append(collection.Item, item)
	}

	return json.MarshalIndent(collection, "", "  ")
}

// SaveAsPostmanCollection 将HAR保存为Postman Collection文件
func (h *Har) SaveAsPostmanCollection(filePath string) error {
	data, err := h.ToPostmanCollection()
	if err != nil {
		return err
	}
	return writeToFile(filePath, data)
}

// entryToPostmanItem 将HAR条目转换为Postman请求项
func entryToPostmanItem(entry *Entries) PostmanItem {
	// 解析URL
	parsedURL, err := url.Parse(entry.Request.URL)
	name := entry.Request.URL
	if err == nil {
		if parsedURL.Path != "" {
			name = parsedURL.Path
		}
	}

	item := PostmanItem{
		Name: name,
		Request: PostmanRequest{
			Method: entry.Request.Method,
			URL:    buildPostmanURL(entry.Request.URL, parsedURL),
		},
	}

	// 请求头
	for _, h := range entry.Request.Headers {
		item.Request.Header = append(item.Request.Header, PostmanHeader{
			Key:   h.Name,
			Value: h.Value,
		})
	}

	// 请求体
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		item.Request.Body = &PostmanBody{
			Mode: "raw",
			Raw:  entry.Request.PostData.Text,
		}
	}

	return item
}

// buildPostmanURL 构建Postman URL结构
func buildPostmanURL(rawURL string, parsedURL *url.URL) PostmanURL {
	pmURL := PostmanURL{
		Raw: rawURL,
	}

	if parsedURL == nil {
		return pmURL
	}

	pmURL.Protocol = parsedURL.Scheme

	// Host拆分
	host := parsedURL.Host
	if h := strings.Split(host, "."); len(h) > 0 {
		pmURL.Host = h
	}

	// Path拆分
	path := parsedURL.Path
	if path != "" {
		segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
		for _, seg := range segments {
			if seg != "" {
				pmURL.Path = append(pmURL.Path, seg)
			}
		}
	}

	// 查询参数
	for key, values := range parsedURL.Query() {
		for _, v := range values {
			pmURL.Query = append(pmURL.Query, PostmanQuery{
				Key:   key,
				Value: v,
			})
		}
	}

	return pmURL
}

// ---------------------------------------------------------------------------
// XML 导出
// ---------------------------------------------------------------------------

// XMLElement 用于生成简单XML的辅助结构
type XMLElement struct {
	XMLName  xml.Name
	Attrs    []xml.Attr      `xml:",any,attr,omitempty"`
	Children []XMLElement    `xml:",any,omitempty"`
	Content  string          `xml:",chardata,omitempty"`
}

// HARXML HAR的XML表示
type HARXML struct {
	XMLName xml.Name  `xml:"har"`
	Log     LogXML    `xml:"log"`
}

// LogXML Log的XML表示
type LogXML struct {
	Version string      `xml:"version"`
	Creator CreatorXML  `xml:"creator"`
	Entries []EntryXML  `xml:"entries>entry"`
}

// CreatorXML Creator的XML表示
type CreatorXML struct {
	Name    string `xml:"name"`
	Version string `xml:"version"`
}

// EntryXML Entries的XML表示
type EntryXML struct {
	StartedDateTime string      `xml:"startedDateTime"`
	Time            float64     `xml:"time"`
	Request         RequestXML  `xml:"request"`
	Response        ResponseXML `xml:"response"`
}

// RequestXML Request的XML表示
type RequestXML struct {
	Method      string         `xml:"method"`
	URL         string         `xml:"url"`
	HTTPVersion string         `xml:"httpVersion"`
	Headers     []HeaderXML    `xml:"headers>header"`
	PostData    *PostDataXML   `xml:"postData,omitempty"`
}

// ResponseXML Response的XML表示
type ResponseXML struct {
	Status      int         `xml:"status"`
	StatusText  string      `xml:"statusText"`
	HTTPVersion string      `xml:"httpVersion"`
	Headers     []HeaderXML `xml:"headers>header"`
	Content     ContentXML  `xml:"content"`
}

// HeaderXML Headers的XML表示
type HeaderXML struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

// PostDataXML PostData的XML表示
type PostDataXML struct {
	MimeType string `xml:"mimeType"`
	Text     string `xml:"text"`
}

// ContentXML Content的XML表示
type ContentXML struct {
	Size     int    `xml:"size"`
	MimeType string `xml:"mimeType"`
	Text     string `xml:"text,omitempty"`
}

// ToXML 将HAR转换为XML格式
func (h *Har) ToXML() (string, error) {
	if h == nil {
		return "", nil
	}

	harXML := HARXML{
		Log: LogXML{
			Version: h.Log.Version,
			Creator: CreatorXML{
				Name:    h.Log.Creator.Name,
				Version: h.Log.Creator.Version,
			},
			Entries: make([]EntryXML, 0, len(h.Log.Entries)),
		},
	}

	for i := range h.Log.Entries {
		entry := &h.Log.Entries[i]
		entryXML := EntryXML{
			StartedDateTime: entry.StartedDateTime.Format("2006-01-02T15:04:05.000Z"),
			Time:            entry.Time,
			Request: RequestXML{
				Method:      entry.Request.Method,
				URL:         entry.Request.URL,
				HTTPVersion: entry.Request.HTTPVersion,
				Headers:     make([]HeaderXML, 0, len(entry.Request.Headers)),
			},
			Response: ResponseXML{
				Status:      entry.Response.Status,
				StatusText:  entry.Response.StatusText,
				HTTPVersion: entry.Response.HTTPVersion,
				Headers:     make([]HeaderXML, 0, len(entry.Response.Headers)),
				Content: ContentXML{
					Size:     entry.Response.Content.Size,
					MimeType: entry.Response.Content.MimeType,
					Text:     entry.Response.Content.Text,
				},
			},
		}

		// 请求头
		for _, hdr := range entry.Request.Headers {
			entryXML.Request.Headers = append(entryXML.Request.Headers, HeaderXML{
				Name:  hdr.Name,
				Value: hdr.Value,
			})
		}

		// 响应头
		for _, hdr := range entry.Response.Headers {
			entryXML.Response.Headers = append(entryXML.Response.Headers, HeaderXML{
				Name:  hdr.Name,
				Value: hdr.Value,
			})
		}

		// POST数据
		if entry.Request.PostData != nil {
			entryXML.Request.PostData = &PostDataXML{
				MimeType: entry.Request.PostData.MimeType,
				Text:     entry.Request.PostData.Text,
			}
		}

		harXML.Log.Entries = append(harXML.Log.Entries, entryXML)
	}

	data, err := xml.MarshalIndent(harXML, "", "  ")
	if err != nil {
		return "", fmt.Errorf("XML序列化失败: %w", err)
	}

	return xml.Header + string(data), nil
}

// SaveAsXML 将HAR保存为XML文件
func (h *Har) SaveAsXML(filePath string) error {
	xmlData, err := h.ToXML()
	if err != nil {
		return err
	}
	return writeToFile(filePath, []byte(xmlData))
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

// escapeSingleQuotes 转义单引号（用于shell命令中的单引号字符串）
// 在单引号内，不能转义单引号，需要结束单引号、添加转义的单引号、再重新开始单引号
func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

// escapePythonString 转义Python字符串中的特殊字符
func escapePythonString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// hasAcceptEncoding 检查请求头中Accept-Encoding是否包含gzip或deflate
func hasAcceptEncoding(entry *Entries) bool {
	for _, h := range entry.Request.Headers {
		if strings.EqualFold(h.Name, "Accept-Encoding") {
			val := strings.ToLower(h.Value)
			if strings.Contains(val, "gzip") || strings.Contains(val, "deflate") {
				return true
			}
		}
	}
	return false
}
