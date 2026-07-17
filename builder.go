package har

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
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

func (b *HarBuilder) ensureHar() *Har {
	if b == nil {
		return nil
	}
	if b.har == nil {
		b.har = NewHar()
	}
	return b.har
}

func (r *Recorder) ensureBuilder() *HarBuilder {
	if r == nil {
		return nil
	}
	if r.builder == nil {
		r.builder = NewHarBuilder().SetCreator("go-har-recorder", "1.0")
	}
	return r.builder
}

// NewHarBuilder 创建一个新的HAR Builder
func NewHarBuilder() *HarBuilder {
	return &HarBuilder{
		har: NewHar(),
	}
}

// SetVersion 设置HAR规范版本
func (b *HarBuilder) SetVersion(version string) *HarBuilder {
	if har := b.ensureHar(); har != nil {
		har.Log.Version = version
	}
	return b
}

// SetCreator 设置创建者信息
func (b *HarBuilder) SetCreator(name, version string) *HarBuilder {
	if har := b.ensureHar(); har != nil {
		har.Log.Creator = Creator{
			Name:    name,
			Version: version,
		}
	}
	return b
}

// SetBrowser 设置浏览器信息
func (b *HarBuilder) SetBrowser(name, version string) *HarBuilder {
	if har := b.ensureHar(); har != nil {
		har.Log.Browser = Browser{
			Name:    name,
			Version: version,
		}
	}
	return b
}

// SetComment 设置注释
func (b *HarBuilder) SetComment(comment string) *HarBuilder {
	if har := b.ensureHar(); har != nil {
		har.Log.Comment = comment
	}
	return b
}

// AddPage 添加页面信息
func (b *HarBuilder) AddPage(id, title string) *HarBuilder {
	if har := b.ensureHar(); har != nil {
		har.AddPage(id, title)
	}
	return b
}

// AddEntry 添加一个条目并返回EntryBuilder用于进一步配置
func (b *HarBuilder) AddEntry(method, url string) *EntryBuilder {
	har := b.ensureHar()
	if har == nil {
		return nil
	}
	entry := har.AddEntry(method, url, "HTTP/1.1", "")
	return &EntryBuilder{
		entry:  entry,
		parent: b,
	}
}

// AddEntryWithHTTPVersion 添加一个条目（指定HTTP版本）
func (b *HarBuilder) AddEntryWithHTTPVersion(method, url, httpVersion string) *EntryBuilder {
	har := b.ensureHar()
	if har == nil {
		return nil
	}
	entry := har.AddEntry(method, url, httpVersion, "")
	return &EntryBuilder{
		entry:  entry,
		parent: b,
	}
}

// AddEntryForPage 添加一个条目并关联到指定页面
func (b *HarBuilder) AddEntryForPage(method, url, pageref string) *EntryBuilder {
	har := b.ensureHar()
	if har == nil {
		return nil
	}
	entry := har.AddEntry(method, url, "HTTP/1.1", pageref)
	return &EntryBuilder{
		entry:  entry,
		parent: b,
	}
}

// AddEntryFromHTTP 从HTTP请求/响应创建条目。
//
// 兼容入口：startedDateTime 取当下时间、不带元数据。
// 如需传入真实请求发起时间或服务器 IP/连接 ID 等元数据，请用 AddEntryFromHTTPWithMeta。
//
// 注意：本方法会消费并关闭 req.Body 与 resp.Body。若调用方仍需响应体，
// 请先自行缓存副本（如 io.ReadAll 后用 io.NopCloser(bytes.NewReader(...)) 回填）。
func (b *HarBuilder) AddEntryFromHTTP(req *http.Request, resp *http.Response, duration time.Duration) *HarBuilder {
	return b.addEntryFromHTTPImpl(req, resp, time.Now(), duration, EntryMeta{})
}

// AddEntryFromHTTPWithMeta 从HTTP请求/响应创建条目，并携带真实开始时间与可选元数据。
//
// 适用于被上层网络安全/网络空间测绘系统作为底层库封装的场景：
//   - startedAt 为请求真正发起的时刻（写入 HAR 的 startedDateTime），
//     解决旧入口写死 time.Now() 导致时序错乱的问题；
//   - meta 携带 serverIP / connection / pageref / initiator / priority / resourceType 等
//     无法从 req/resp 推断的元数据；
//   - 二进制响应体（图片/字体/视频等）自动 base64 编码，保证 JSON 往返无损；
//   - HeadersSize 自动估算填充。
//
// 返回 *EntryBuilder 以便对生成的条目做后置定制（追加头、Cookie 等），调用 EndEntry() 回到 HarBuilder。
// 与 AddEntryFromHTTP 一样，会消费并关闭 req.Body / resp.Body。
func (b *HarBuilder) AddEntryFromHTTPWithMeta(req *http.Request, resp *http.Response, startedAt time.Time, duration time.Duration, meta EntryMeta) *EntryBuilder {
	b.addEntryFromHTTPImpl(req, resp, startedAt, duration, meta)
	har := b.ensureHar()
	if har == nil || len(har.Log.Entries) == 0 {
		return nil
	}
	// 返回指向刚追加的 entry 的 EntryBuilder，便于后置定制
	return &EntryBuilder{
		entry:  &har.Log.Entries[len(har.Log.Entries)-1],
		parent: b,
	}
}

// addEntryFromHTTPImpl 是 AddEntryFromHTTP / AddEntryFromHTTPWithMeta 的共享实现。
func (b *HarBuilder) addEntryFromHTTPImpl(req *http.Request, resp *http.Response, startedAt time.Time, duration time.Duration, meta EntryMeta) *HarBuilder {
	har := b.ensureHar()
	if har == nil {
		return b
	}
	if req == nil {
		return b
	}

	requestURL := ""
	if req.URL != nil {
		requestURL = req.URL.String()
	}

	entry := Entries{
		StartedDateTime: startedAt,
		Time:            float64(duration.Milliseconds()),
		Request: Request{
			Method:      req.Method,
			URL:         requestURL,
			HTTPVersion: req.Proto,
			Headers:     HeadersFromHTTP(req.Header),
			Cookies:     CookiesFromHTTP(req.Cookies()),
			QueryString: BuildQueryStringFromURL(requestURL),
			HeadersSize: EstimateHeaderSize(HeadersFromHTTP(req.Header)),
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

	// 读取请求体（会消费 req.Body）
	postData, bodySize := PostDataFromRequest(req)
	if postData != nil {
		entry.Request.PostData = postData
		entry.Request.BodySize = bodySize
	}

	// 转换响应
	if resp != nil {
		entry.Response.Status = resp.StatusCode
		entry.Response.StatusText = resp.Status
		entry.Response.HTTPVersion = resp.Proto
		entry.Response.Headers = HeadersFromHTTP(resp.Header)
		entry.Response.HeadersSize = EstimateHeaderSize(entry.Response.Headers)

		entry.Response.Cookies = CookiesFromHTTP(resp.Cookies())

		if !isNilReader(resp.Body) {
			bodyBytes, readErr, closeErr := readAndCloseResponseBody(resp.Body)
			if bodyErr := responseBodyErrorMessage(readErr, closeErr); bodyErr != "" {
				entry.Response.Error = bodyErr
			}
			if readErr == nil {
				mimeType := resp.Header.Get("Content-Type")
				content := Content{
					Size:     len(bodyBytes),
					MimeType: mimeType,
				}
				// 非文本 body 走 base64 编码，保证 JSON 往返无损
				if isTextContentType(mimeType) {
					content.Text = string(bodyBytes)
				} else {
					content.Text = base64.StdEncoding.EncodeToString(bodyBytes)
					content.Encoding = "base64"
				}
				entry.Response.Content = content
				entry.Response.BodySize = len(bodyBytes)
			}
		}
	}

	// 应用可选元数据
	applyEntryMeta(&entry, meta)

	har.Log.Entries = append(har.Log.Entries, entry)
	return b
}

// applyEntryMeta 把 EntryMeta 中的非零字段写入 entry。
func applyEntryMeta(entry *Entries, meta EntryMeta) {
	if entry == nil {
		return
	}
	if meta.ServerIPAddress != "" {
		entry.ServerIPAddress = meta.ServerIPAddress
	}
	if meta.Connection != "" {
		entry.Connection = meta.Connection
	}
	if meta.Pageref != "" {
		entry.Pageref = meta.Pageref
	}
	if meta.Comment != "" {
		entry.Comment = meta.Comment
	}
	if meta.InitiatorType != "" || meta.InitiatorURL != "" {
		entry.Initiator = Initiator{
			Type: meta.InitiatorType,
			URL:  meta.InitiatorURL,
		}
		if meta.InitiatorLine > 0 {
			entry.Initiator.LineNumber = meta.InitiatorLine
		}
	}
	if meta.Priority != "" {
		entry.Priority = meta.Priority
	}
	if meta.ResourceType != "" {
		entry.ResourceType = meta.ResourceType
	}
}

// Build 构建并返回HAR对象
func (b *HarBuilder) Build() *Har {
	return b.ensureHar()
}

// BuildJSON 构建HAR并返回JSON
func (b *HarBuilder) BuildJSON(indent bool) ([]byte, error) {
	har := b.ensureHar()
	if har == nil {
		return nil, NewInvalidFormatError("HAR Builder为空")
	}
	return har.ToJSON(indent)
}

// BuildAndSave 构建HAR并保存到文件
func (b *HarBuilder) BuildAndSave(filePath string, indent bool) error {
	har := b.ensureHar()
	if har == nil {
		return NewInvalidFormatError("HAR Builder为空")
	}
	return har.SaveToFile(filePath, indent)
}

// Entry Builder 方法

// WithHTTPVersion 设置HTTP版本
func (eb *EntryBuilder) WithHTTPVersion(version string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Request.HTTPVersion = version
	eb.entry.Response.HTTPVersion = version
	return eb
}

// WithStartedDateTime 设置请求开始时间
func (eb *EntryBuilder) WithStartedDateTime(t time.Time) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.StartedDateTime = t
	return eb
}

// WithPageref 设置页面引用
func (eb *EntryBuilder) WithPageref(ref string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Pageref = ref
	return eb
}

// WithServerIP 设置服务器IP
func (eb *EntryBuilder) WithServerIP(ip string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.ServerIPAddress = ip
	return eb
}

// WithConnection 设置连接ID
func (eb *EntryBuilder) WithConnection(id string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Connection = id
	return eb
}

// WithComment 设置注释
func (eb *EntryBuilder) WithComment(comment string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Comment = comment
	return eb
}

// AddRequestHeader 添加请求头
func (eb *EntryBuilder) AddRequestHeader(name, value string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.AddRequestHeader(name, value)
	return eb
}

// AddResponseHeader 添加响应头
func (eb *EntryBuilder) AddResponseHeader(name, value string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.AddResponseHeader(name, value)
	return eb
}

// AddCookie 添加请求Cookie
func (eb *EntryBuilder) AddCookie(name, value string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.AddCookie(name, value)
	return eb
}

// AddResponseCookie 添加响应Cookie
func (eb *EntryBuilder) AddResponseCookie(name, value string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.AddResponseCookie(name, value)
	return eb
}

// AddQueryParam 添加查询参数
func (eb *EntryBuilder) AddQueryParam(name, value string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.AddQueryParameter(name, value)
	return eb
}

// WithPostData 设置POST数据
func (eb *EntryBuilder) WithPostData(mimeType, text string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.SetPostData(mimeType, text)
	return eb
}

// WithPostDataParams 设置POST表单参数
func (eb *EntryBuilder) WithPostDataParams(mimeType string, params []Param) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.SetPostDataParams(mimeType, params)
	return eb
}

// WithResponseStatus 设置响应状态
func (eb *EntryBuilder) WithResponseStatus(status int, statusText string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.SetResponseStatus(status, statusText)
	return eb
}

// WithResponseContent 设置响应内容
func (eb *EntryBuilder) WithResponseContent(size int, mimeType string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.SetResponseContent(size, mimeType)
	return eb
}

// WithResponseContentText 设置响应内容（含文本）
func (eb *EntryBuilder) WithResponseContentText(size int, mimeType, text string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.SetResponseContent(size, mimeType)
	eb.entry.Response.Content.Text = text
	return eb
}

// WithTimings 设置时间数据
func (eb *EntryBuilder) WithTimings(blocked, dns, connect, send, wait, receive, ssl float64) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.SetTimings(blocked, dns, connect, send, wait, receive, ssl)
	return eb
}

// WithCache 设置缓存数据
func (eb *EntryBuilder) WithCache(cache Cache) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Cache = cache
	return eb
}

// WithInitiator 设置请求发起者
func (eb *EntryBuilder) WithInitiator(initiatorType, initiatorURL string, lineNumber int) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Initiator = Initiator{
		Type:       initiatorType,
		URL:        initiatorURL,
		LineNumber: lineNumber,
	}
	return eb
}

// WithPriority 设置请求优先级
func (eb *EntryBuilder) WithPriority(priority string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.Priority = priority
	return eb
}

// WithResourceType 设置资源类型
func (eb *EntryBuilder) WithResourceType(resourceType string) *EntryBuilder {
	if eb == nil || eb.entry == nil {
		return eb
	}
	eb.entry.ResourceType = resourceType
	return eb
}

// EndEntry 结束条目构建，返回HarBuilder
func (eb *EntryBuilder) EndEntry() *HarBuilder {
	if eb == nil {
		return nil
	}
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
	if builder := r.ensureBuilder(); builder != nil {
		builder.SetCreator(name, version)
	}
	return r
}

// SetBrowser 设置浏览器信息
func (r *Recorder) SetBrowser(name, version string) *Recorder {
	if builder := r.ensureBuilder(); builder != nil {
		builder.SetBrowser(name, version)
	}
	return r
}

// Capture 捕获一个HTTP请求/响应
func (r *Recorder) Capture(req *http.Request, resp *http.Response, duration time.Duration) *Recorder {
	if builder := r.ensureBuilder(); builder != nil {
		builder.AddEntryFromHTTP(req, resp, duration)
	}
	return r
}

// CaptureEntry 捕获一个预构建的HAR条目
func (r *Recorder) CaptureEntry(entry Entries) *Recorder {
	if builder := r.ensureBuilder(); builder != nil {
		builder.ensureHar().Log.Entries = append(builder.ensureHar().Log.Entries, entry)
	}
	return r
}

// EntryCount 返回已录制的条目数
func (r *Recorder) EntryCount() int {
	builder := r.ensureBuilder()
	if builder == nil {
		return 0
	}
	// builder 非 nil 时 ensureHar 必返回非 nil 的 Har (NewHarBuilder
	// 已初始化 b.har)。
	return len(builder.ensureHar().Log.Entries)
}

// ToHar 生成HAR对象
func (r *Recorder) ToHar() *Har {
	builder := r.ensureBuilder()
	if builder == nil {
		return nil
	}
	return builder.Build()
}

// SaveToFile 保存录制结果到文件
func (r *Recorder) SaveToFile(path string) error {
	builder := r.ensureBuilder()
	if builder == nil {
		return NewInvalidFormatError("Recorder为空")
	}
	return builder.BuildAndSave(path, true)
}

// ToJSON 生成JSON格式
func (r *Recorder) ToJSON(indent bool) ([]byte, error) {
	builder := r.ensureBuilder()
	if builder == nil {
		return nil, NewInvalidFormatError("Recorder为空")
	}
	return builder.BuildJSON(indent)
}

// WriteToWriter 将HAR写入指定的Writer
func WriteToWriter(har *Har, w io.Writer, indent bool) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}
	if isNilWriter(w) {
		return NewInvalidFormatError("writer is nil")
	}

	data, err := har.ToJSON(indent)
	if err != nil {
		return err
	}

	return writeAllToWriter(w, data, "failed to write HAR data")
}

// WriteEntriesToWriter 将HAR条目以JSON Lines格式写入Writer
// 每行一个条目的JSON对象，适用于流式处理
func WriteEntriesToWriter(har *Har, w io.Writer) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}
	if isNilWriter(w) {
		return NewInvalidFormatError("writer is nil")
	}

	for _, entry := range har.Log.Entries {
		var buf bytes.Buffer
		entryEncoder := json.NewEncoder(&buf)
		if err := entryEncoder.Encode(entry); err != nil {
			return NewJSONParseError("failed to encode HAR entry", err)
		}
		if err := writeAllToWriter(w, buf.Bytes(), "failed to write HAR entries"); err != nil {
			return err
		}
	}

	return nil
}

// ReadEntriesFromReader 从Reader中读取JSON Lines格式的条目
func ReadEntriesFromReader(r io.Reader) ([]Entries, error) {
	if isNilReader(r) {
		return nil, NewInvalidFormatError("reader is nil")
	}

	var entries []Entries
	decoder := json.NewDecoder(r)

	for {
		var entry Entries
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return entries, WrapJSONUnmarshalError(err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// ToJSONLines 将HAR条目转换为JSON Lines格式字符串
func (h *Har) ToJSONLines() (string, error) {
	if h == nil {
		return "", NewInvalidFormatError("HAR对象为空")
	}

	var buf bytes.Buffer
	err := WriteEntriesToWriter(h, &buf)
	return buf.String(), err
}

// WriteEntryToWriter 将单条 HAR 条目以 JSON Lines 格式写入 Writer（一行一个 JSON 对象）。
// 适用于"抓一条写一条"的低内存长期归档场景：无需在内存中攒成完整 *Har，
// 上层测绘系统每抓到一个请求即可立即落盘。
func WriteEntryToWriter(w io.Writer, entry Entries) error {
	if isNilWriter(w) {
		return NewInvalidFormatError("writer is nil")
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(entry); err != nil {
		return NewJSONParseError("failed to encode HAR entry", err)
	}
	return writeAllToWriter(w, buf.Bytes(), "failed to write HAR entry")
}

// AppendEntryToJSONLFile 将单条 HAR 条目以 JSON Lines 形式追加到文件。
// 文件不存在时自动创建；存在时 O_APPEND 追加，不会读入既有内容，内存占用恒定。
// 适合长期持续归档：每条请求一行，文件可后续用 ForEachEntryFromReader 或
// ReadEntriesFromReader 读回，也可用 split --by 等命令分片。
func AppendEntryToJSONLFile(path string, entry Entries) error {
	if path == "" {
		return NewInvalidFormatError("path is empty")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return NewFileSystemError("failed to open HAR JSONL file", err)
	}
	defer f.Close()
	return WriteEntryToWriter(f, entry)
}

// ForEachEntryFromReader 流式读取 JSON Lines 格式的条目，对每条条目调用 fn。
// 与 ReadEntriesFromReader 不同，本函数不会把所有条目一次性读入内存，
// 而是逐条 Decode 后立即交给回调，适合处理超大归档文件。
// fn 返回非 nil error 时立即停止迭代并返回该错误。
func ForEachEntryFromReader(r io.Reader, fn func(entry Entries) error) error {
	if isNilReader(r) {
		return NewInvalidFormatError("reader is nil")
	}
	if fn == nil {
		return NewInvalidFormatError("callback is nil")
	}
	decoder := json.NewDecoder(r)
	for {
		var entry Entries
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				return nil
			}
			return WrapJSONUnmarshalError(err)
		}
		if err := fn(entry); err != nil {
			return err
		}
	}
}

// SafeRecorder 是并发安全的 Recorder，适合上层测绘系统多协程并发抓包归档。
// 内部用 sync.Mutex 保护所有读写操作。对于"持续累积 + 一次性导出"或
// "并发 Capture 后定期 SaveToFile"的场景，直接使用即可，无需调用方自行加锁。
type SafeRecorder struct {
	mu       sync.Mutex
	recorder *Recorder
}

// NewSafeRecorder 创建一个新的并发安全 Recorder。
func NewSafeRecorder() *SafeRecorder {
	return &SafeRecorder{recorder: NewRecorder()}
}

// SetCreator 设置录制器的创建者信息。
func (s *SafeRecorder) SetCreator(name, version string) *SafeRecorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recorder.SetCreator(name, version)
	return s
}

// SetBrowser 设置浏览器信息。
func (s *SafeRecorder) SetBrowser(name, version string) *SafeRecorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recorder.SetBrowser(name, version)
	return s
}

// Capture 并发安全地捕获一个 HTTP 请求/响应。
// 注意：会消费并关闭 req.Body / resp.Body，调用方若仍需响应体请先缓存副本。
func (s *SafeRecorder) Capture(req *http.Request, resp *http.Response, duration time.Duration) *SafeRecorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recorder.Capture(req, resp, duration)
	return s
}

// CaptureWithMeta 并发安全地捕获一个 HTTP 请求/响应，携带真实开始时间与元数据。
// 详见 HarBuilder.AddEntryFromHTTPWithMeta。
func (s *SafeRecorder) CaptureWithMeta(req *http.Request, resp *http.Response, startedAt time.Time, duration time.Duration, meta EntryMeta) *SafeRecorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	builder := s.recorder.ensureBuilder()
	if builder != nil {
		builder.AddEntryFromHTTPWithMeta(req, resp, startedAt, duration, meta)
	}
	return s
}

// CaptureEntry 并发安全地捕获一个预构建的 HAR 条目（不碰任何 body）。
func (s *SafeRecorder) CaptureEntry(entry Entries) *SafeRecorder {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recorder.CaptureEntry(entry)
	return s
}

// EntryCount 返回已录制的条目数。
func (s *SafeRecorder) EntryCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recorder.EntryCount()
}

// ToHarCopy 返回内部 HAR 的深拷贝，调用方持有的副本不会因后续 Capture 而变化。
// 适合在并发归档过程中定期导出快照。
func (s *SafeRecorder) ToHarCopy() *Har {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h := s.recorder.ToHar()
	if h == nil {
		return nil
	}
	return h.Clone()
}

// ToHar 返回内部 HAR 指针（在锁内取，但返回的指针指向的内存可能被后续 Capture 修改）。
// 如需稳定快照请用 ToHarCopy。
func (s *SafeRecorder) ToHar() *Har {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recorder.ToHar()
}

// SaveToFile 并发安全地保存录制结果到文件（缩进 JSON）。
func (s *SafeRecorder) SaveToFile(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recorder.SaveToFile(path)
}

// SaveToFileWithOptions 保存录制结果，可选择是否缩进及是否 gzip 压缩。
func (s *SafeRecorder) SaveToFileWithOptions(path string, indent, gzip bool) error {
	s.mu.Lock()
	h := s.recorder.ToHar()
	s.mu.Unlock()
	if h == nil {
		return NewInvalidFormatError("Recorder为空")
	}
	if gzip {
		return SaveToFileGzipped(h, path, indent)
	}
	return h.SaveToFile(path, indent)
}
