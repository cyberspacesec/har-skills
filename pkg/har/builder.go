package har

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// HarBuilder 提供流式API构建HAR文件
type HarBuilder struct {
	har *Har
}

// EntryBuilder 提供流式API构建HAR条目
type EntryBuilder struct {
	entry  *Entries
	parent *HarBuilder
}

// NewHarBuilder 创建一个新的HAR Builder
func NewHarBuilder() *HarBuilder {
	return &HarBuilder{
		har: NewHar(),
	}
}

// SetVersion 设置HAR规范版本
func (b *HarBuilder) SetVersion(version string) *HarBuilder {
	b.har.Log.Version = version
	return b
}

// SetCreator 设置创建者信息
func (b *HarBuilder) SetCreator(name, version string) *HarBuilder {
	b.har.Log.Creator = Creator{
		Name:    name,
		Version: version,
	}
	return b
}

// SetBrowser 设置浏览器信息
func (b *HarBuilder) SetBrowser(name, version string) *HarBuilder {
	b.har.Log.Browser = Browser{
		Name:    name,
		Version: version,
	}
	return b
}

// SetComment 设置注释
func (b *HarBuilder) SetComment(comment string) *HarBuilder {
	b.har.Log.Comment = comment
	return b
}

// AddPage 添加页面信息
func (b *HarBuilder) AddPage(id, title string) *HarBuilder {
	b.har.AddPage(id, title)
	return b
}

// AddEntry 添加一个条目并返回EntryBuilder用于进一步配置
func (b *HarBuilder) AddEntry(method, url string) *EntryBuilder {
	entry := b.har.AddEntry(method, url, "HTTP/1.1", "")
	return &EntryBuilder{
		entry:  entry,
		parent: b,
	}
}

// AddEntryWithHTTPVersion 添加一个条目（指定HTTP版本）
func (b *HarBuilder) AddEntryWithHTTPVersion(method, url, httpVersion string) *EntryBuilder {
	entry := b.har.AddEntry(method, url, httpVersion, "")
	return &EntryBuilder{
		entry:  entry,
		parent: b,
	}
}

// AddEntryForPage 添加一个条目并关联到指定页面
func (b *HarBuilder) AddEntryForPage(method, url, pageref string) *EntryBuilder {
	entry := b.har.AddEntry(method, url, "HTTP/1.1", pageref)
	return &EntryBuilder{
		entry:  entry,
		parent: b,
	}
}

// AddEntryFromHTTP 从HTTP请求/响应创建条目
func (b *HarBuilder) AddEntryFromHTTP(req *http.Request, resp *http.Response, duration time.Duration) *HarBuilder {
	if req == nil {
		return b
	}

	entry := Entries{
		StartedDateTime: time.Now(),
		Time:            float64(duration.Milliseconds()),
		Request: Request{
			Method:      req.Method,
			URL:         req.URL.String(),
			HTTPVersion: req.Proto,
			Headers:     make([]Headers, 0),
			Cookies:     make([]Cookie, 0),
			QueryString: BuildQueryStringFromURL(req.URL.String()),
			HeadersSize: -1,
			BodySize:    -1,
		},
		Response: Response{
			HeadersSize: -1,
			BodySize:    -1,
		},
		Timings: Timings{
			Blocked: -1,
			DNS:     -1,
			Connect: -1,
			Send:    -1,
			Wait:    float64(duration.Milliseconds()),
			Receive: -1,
			Ssl:     -1,
		},
	}

	// 转换请求头
	for key, values := range req.Header {
		for _, value := range values {
			entry.Request.Headers = append(entry.Request.Headers, Headers{
				Name:  key,
				Value: value,
			})
		}
	}

	// 转换请求Cookie
	for _, cookie := range req.Cookies() {
		entry.Request.Cookies = append(entry.Request.Cookies, Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
		})
	}

	// 读取请求体
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		if len(bodyBytes) > 0 {
			contentType := req.Header.Get("Content-Type")
			entry.Request.PostData = &PostData{
				MimeType: contentType,
				Text:     string(bodyBytes),
			}
			entry.Request.BodySize = len(bodyBytes)
		}
	}

	// 转换响应
	if resp != nil {
		entry.Response.Status = resp.StatusCode
		entry.Response.StatusText = resp.Status
		entry.Response.HTTPVersion = resp.Proto

		for key, values := range resp.Header {
			for _, value := range values {
				entry.Response.Headers = append(entry.Response.Headers, Headers{
					Name:  key,
					Value: value,
				})
			}
		}

		for _, cookie := range resp.Cookies() {
			entry.Response.Cookies = append(entry.Response.Cookies, Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Domain:   cookie.Domain,
				HTTPOnly: cookie.HttpOnly,
				Secure:   cookie.Secure,
			})
		}

		if resp.Body != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			entry.Response.Content = Content{
				Size:     len(bodyBytes),
				MimeType: resp.Header.Get("Content-Type"),
				Text:     string(bodyBytes),
			}
			entry.Response.BodySize = len(bodyBytes)
		}
	}

	b.har.Log.Entries = append(b.har.Log.Entries, entry)
	return b
}

// Build 构建并返回HAR对象
func (b *HarBuilder) Build() *Har {
	return b.har
}

// BuildJSON 构建HAR并返回JSON
func (b *HarBuilder) BuildJSON(indent bool) ([]byte, error) {
	return b.har.ToJSON(indent)
}

// BuildAndSave 构建HAR并保存到文件
func (b *HarBuilder) BuildAndSave(filePath string, indent bool) error {
	return b.har.SaveToFile(filePath, indent)
}

// Entry Builder 方法

// WithHTTPVersion 设置HTTP版本
func (eb *EntryBuilder) WithHTTPVersion(version string) *EntryBuilder {
	eb.entry.Request.HTTPVersion = version
	eb.entry.Response.HTTPVersion = version
	return eb
}

// WithStartedDateTime 设置请求开始时间
func (eb *EntryBuilder) WithStartedDateTime(t time.Time) *EntryBuilder {
	eb.entry.StartedDateTime = t
	return eb
}

// WithPageref 设置页面引用
func (eb *EntryBuilder) WithPageref(ref string) *EntryBuilder {
	eb.entry.Pageref = ref
	return eb
}

// WithServerIP 设置服务器IP
func (eb *EntryBuilder) WithServerIP(ip string) *EntryBuilder {
	eb.entry.ServerIPAddress = ip
	return eb
}

// WithConnection 设置连接ID
func (eb *EntryBuilder) WithConnection(id string) *EntryBuilder {
	eb.entry.Connection = id
	return eb
}

// WithComment 设置注释
func (eb *EntryBuilder) WithComment(comment string) *EntryBuilder {
	eb.entry.Comment = comment
	return eb
}

// AddRequestHeader 添加请求头
func (eb *EntryBuilder) AddRequestHeader(name, value string) *EntryBuilder {
	eb.entry.AddRequestHeader(name, value)
	return eb
}

// AddResponseHeader 添加响应头
func (eb *EntryBuilder) AddResponseHeader(name, value string) *EntryBuilder {
	eb.entry.AddResponseHeader(name, value)
	return eb
}

// AddCookie 添加请求Cookie
func (eb *EntryBuilder) AddCookie(name, value string) *EntryBuilder {
	eb.entry.AddCookie(name, value)
	return eb
}

// AddResponseCookie 添加响应Cookie
func (eb *EntryBuilder) AddResponseCookie(name, value string) *EntryBuilder {
	eb.entry.AddResponseCookie(name, value)
	return eb
}

// AddQueryParam 添加查询参数
func (eb *EntryBuilder) AddQueryParam(name, value string) *EntryBuilder {
	eb.entry.AddQueryParameter(name, value)
	return eb
}

// WithPostData 设置POST数据
func (eb *EntryBuilder) WithPostData(mimeType, text string) *EntryBuilder {
	eb.entry.SetPostData(mimeType, text)
	return eb
}

// WithPostDataParams 设置POST表单参数
func (eb *EntryBuilder) WithPostDataParams(mimeType string, params []Param) *EntryBuilder {
	eb.entry.SetPostDataParams(mimeType, params)
	return eb
}

// WithResponseStatus 设置响应状态
func (eb *EntryBuilder) WithResponseStatus(status int, statusText string) *EntryBuilder {
	eb.entry.SetResponseStatus(status, statusText)
	return eb
}

// WithResponseContent 设置响应内容
func (eb *EntryBuilder) WithResponseContent(size int, mimeType string) *EntryBuilder {
	eb.entry.SetResponseContent(size, mimeType)
	return eb
}

// WithResponseContentText 设置响应内容（含文本）
func (eb *EntryBuilder) WithResponseContentText(size int, mimeType, text string) *EntryBuilder {
	eb.entry.SetResponseContent(size, mimeType)
	eb.entry.Response.Content.Text = text
	return eb
}

// WithTimings 设置时间数据
func (eb *EntryBuilder) WithTimings(blocked, dns, connect, send, wait, receive, ssl float64) *EntryBuilder {
	eb.entry.SetTimings(blocked, dns, connect, send, wait, receive, ssl)
	return eb
}

// WithCache 设置缓存数据
func (eb *EntryBuilder) WithCache(cache Cache) *EntryBuilder {
	eb.entry.Cache = cache
	return eb
}

// WithInitiator 设置请求发起者
func (eb *EntryBuilder) WithInitiator(initiatorType, initiatorURL string, lineNumber int) *EntryBuilder {
	eb.entry.Initiator = Initiator{
		Type:       initiatorType,
		URL:        initiatorURL,
		LineNumber: lineNumber,
	}
	return eb
}

// WithPriority 设置请求优先级
func (eb *EntryBuilder) WithPriority(priority string) *EntryBuilder {
	eb.entry.Priority = priority
	return eb
}

// WithResourceType 设置资源类型
func (eb *EntryBuilder) WithResourceType(resourceType string) *EntryBuilder {
	eb.entry.ResourceType = resourceType
	return eb
}

// EndEntry 结束条目构建，返回HarBuilder
func (eb *EntryBuilder) EndEntry() *HarBuilder {
	return eb.parent
}

// Recorder 用于录制HTTP交互并生成HAR文件
type Recorder struct {
	builder *HarBuilder
}

// NewRecorder 创建一个新的Recorder
func NewRecorder() *Recorder {
	return &Recorder{
		builder: NewHarBuilder().SetCreator("go-har-recorder", "1.0"),
	}
}

// SetCreator 设置录制器的创建者信息
func (r *Recorder) SetCreator(name, version string) *Recorder {
	r.builder.SetCreator(name, version)
	return r
}

// SetBrowser 设置浏览器信息
func (r *Recorder) SetBrowser(name, version string) *Recorder {
	r.builder.SetBrowser(name, version)
	return r
}

// Capture 捕获一个HTTP请求/响应
func (r *Recorder) Capture(req *http.Request, resp *http.Response, duration time.Duration) *Recorder {
	r.builder.AddEntryFromHTTP(req, resp, duration)
	return r
}

// CaptureEntry 捕获一个预构建的HAR条目
func (r *Recorder) CaptureEntry(entry Entries) *Recorder {
	r.builder.har.Log.Entries = append(r.builder.har.Log.Entries, entry)
	return r
}

// EntryCount 返回已录制的条目数
func (r *Recorder) EntryCount() int {
	return len(r.builder.har.Log.Entries)
}

// ToHar 生成HAR对象
func (r *Recorder) ToHar() *Har {
	return r.builder.Build()
}

// SaveToFile 保存录制结果到文件
func (r *Recorder) SaveToFile(path string) error {
	return r.builder.BuildAndSave(path, true)
}

// ToJSON 生成JSON格式
func (r *Recorder) ToJSON(indent bool) ([]byte, error) {
	return r.builder.BuildJSON(indent)
}

// WriteToWriter 将HAR写入指定的Writer
func WriteToWriter(har *Har, w io.Writer, indent bool) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	data, err := har.ToJSON(indent)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// WriteEntriesToWriter 将HAR条目以JSON Lines格式写入Writer
// 每行一个条目的JSON对象，适用于流式处理
func WriteEntriesToWriter(har *Har, w io.Writer) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	encoder := json.NewEncoder(w)
	for _, entry := range har.Log.Entries {
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}

	return nil
}

// ReadEntriesFromReader 从Reader中读取JSON Lines格式的条目
func ReadEntriesFromReader(r io.Reader) ([]Entries, error) {
	var entries []Entries
	decoder := json.NewDecoder(r)

	for decoder.More() {
		var entry Entries
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return entries, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// ToJSONLines 将HAR条目转换为JSON Lines格式字符串
func (h *Har) ToJSONLines() (string, error) {
	if h == nil {
		return "", nil
	}

	var buf bytes.Buffer
	err := WriteEntriesToWriter(h, &buf)
	return buf.String(), err
}
