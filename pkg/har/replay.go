package har

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ReplayOptions 重放请求的选项
type ReplayOptions struct {
	Timeout         time.Duration     // 请求超时时间
	FollowRedirects bool              // 是否跟随重定向
	MaxRedirects    int               // 最大重定向次数
	SkipSSLVerify   bool              // 跳过SSL证书验证
	OverrideHeaders map[string]string // 覆盖指定请求头
	Transport       http.RoundTripper // 自定义Transport
}

// DefaultReplayOptions 返回默认的重放选项
func DefaultReplayOptions() ReplayOptions {
	return ReplayOptions{
		Timeout:         30 * time.Second,
		FollowRedirects: true,
		MaxRedirects:    10,
		SkipSSLVerify:   false,
	}
}

// ReplayResult 表示重放单个请求的结果
type ReplayResult struct {
	Entry     *Entries      // 原始HAR条目
	Response  *http.Response // HTTP响应
	Duration  time.Duration  // 请求耗时
	Error     error          // 错误信息
	Index     int            // 条目索引
}

// ToHTTPRequest 将HAR条目转换为标准库的http.Request对象
//
// 该方法根据HAR条目中的请求信息构建一个完整的http.Request，
// 包括方法、URL、头部、Cookie和请求体。
func (e *Entries) ToHTTPRequest() (*http.Request, error) {
	if e == nil {
		return nil, NewInvalidFormatError("条目为空")
	}

	// 解析URL
	parsedURL, err := url.Parse(e.Request.URL)
	if err != nil {
		return nil, NewInvalidValueError("request.url", e.Request.URL,
			fmt.Sprintf("URL解析失败: %v", err))
	}

	// 构建请求体
	var body io.Reader
	if e.Request.PostData != nil && e.Request.PostData.Text != "" {
		body = strings.NewReader(e.Request.PostData.Text)
	}

	// 创建请求
	req, err := http.NewRequest(e.Request.Method, parsedURL.String(), body)
	if err != nil {
		return nil, NewHarError(ErrCodeInvalidFormat,
			fmt.Sprintf("创建HTTP请求失败: %v", err), err)
	}

	// 设置请求头
	for _, header := range e.Request.Headers {
		req.Header.Set(header.Name, header.Value)
	}

	// 设置Cookie
	for _, cookie := range e.Request.Cookies {
		req.AddCookie(&http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			HttpOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
		})
	}

	// 设置Content-Type（如果有PostData）
	if e.Request.PostData != nil && e.Request.PostData.MimeType != "" {
		req.Header.Set("Content-Type", e.Request.PostData.MimeType)
	}

	return req, nil
}

// Replay 重放单个HAR条目的HTTP请求
//
// 该方法将HAR条目转换为HTTP请求并执行，返回重放结果。
func (e *Entries) Replay(options ReplayOptions) (*ReplayResult, error) {
	if e == nil {
		return nil, NewInvalidFormatError("条目为空")
	}

	// 构建HTTP请求
	req, err := e.ToHTTPRequest()
	if err != nil {
		return nil, err
	}

	// 应用头部覆盖
	for name, value := range options.OverrideHeaders {
		req.Header.Set(name, value)
	}

	// 创建HTTP客户端
	client := createHTTPClient(options)

	// 执行请求并计时
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return &ReplayResult{
			Entry:    e,
			Duration: duration,
			Error:    err,
		}, err
	}

	return &ReplayResult{
		Entry:    e,
		Response: resp,
		Duration: duration,
	}, nil
}

// ReplayAll 重放HAR文件中所有条目的HTTP请求
//
// 该方法依次执行所有条目的请求，返回每个条目的重放结果。
func (h *Har) ReplayAll(options ReplayOptions) ([]*ReplayResult, error) {
	if h == nil {
		return nil, NewInvalidFormatError("HAR对象为空")
	}

	results := make([]*ReplayResult, len(h.Log.Entries))
	var firstErr error

	for i, entry := range h.Log.Entries {
		result, err := entry.Replay(options)
		result.Index = i
		results[i] = result
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return results, firstErr
}

// ReplaySelective 选择性重放符合条件的条目
func (h *Har) ReplaySelective(options ReplayOptions, filterOptions FilterOptions) ([]*ReplayResult, error) {
	filtered := h.Filter(filterOptions)
	if filtered.Count() == 0 {
		return nil, nil
	}

	results := make([]*ReplayResult, filtered.Count())
	var firstErr error

	for i := range filtered.Entries {
		result, err := filtered.Entries[i].Replay(options)
		result.Index = i
		results[i] = result
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return results, firstErr
}

// HTTPResponseToEntries 将http.Response转换为HAR Entries
//
// 这是一个辅助函数，将标准库的HTTP响应转换为HAR格式的条目，
// 方便与重放功能配合使用。
func HTTPResponseToEntries(req *Entries, resp *http.Response, duration time.Duration) *Entries {
	if resp == nil {
		return nil
	}

	entry := &Entries{
		StartedDateTime: time.Now(),
		Time:            float64(duration.Milliseconds()),
		Request:         req.Request,
	}

	// 构建响应
	entry.Response = Response{
		Status:      resp.StatusCode,
		StatusText:  resp.Status,
		HTTPVersion: resp.Proto,
		HeadersSize: -1,
		BodySize:    -1,
	}

	// 读取响应头
	for key, values := range resp.Header {
		for _, value := range values {
			entry.Response.Headers = append(entry.Response.Headers, Headers{
				Name:  key,
				Value: value,
			})
		}
	}

	// 读取响应Cookie
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

	// 读取响应体
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			entry.Response.Content = Content{
				Size:     len(bodyBytes),
				MimeType: resp.Header.Get("Content-Type"),
				Text:     string(bodyBytes),
			}
			entry.Response.BodySize = len(bodyBytes)
		}
	}

	return entry
}

// createHTTPClient 根据选项创建HTTP客户端
func createHTTPClient(options ReplayOptions) *http.Client {
	transport := options.Transport
	if transport == nil {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: options.SkipSSLVerify,
			},
		}
	}

	client := &http.Client{
		Timeout:   options.Timeout,
		Transport: transport,
	}

	if !options.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if options.MaxRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= options.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", options.MaxRedirects)
			}
			return nil
		}
	}

	return client
}

// ReplayResultToHar 将重放结果转换回HAR对象
func ReplayResultsToHar(results []*ReplayResult) *Har {
	h := NewHar()
	h.SetCreator("go-har-replay", "1.0")

	for _, result := range results {
		if result == nil {
			continue
		}

		if result.Response != nil {
			entry := HTTPResponseToEntries(result.Entry, result.Response, result.Duration)
			h.Log.Entries = append(h.Log.Entries, *entry)
		} else if result.Entry != nil {
			// 即使请求失败，也保留原始条目
			h.Log.Entries = append(h.Log.Entries, *result.Entry)
		}
	}

	return h
}

// BuildQueryStringFromURL 从URL字符串解析查询参数
func BuildQueryStringFromURL(rawURL string) []QueryString {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}

	var params []QueryString
	for key, values := range parsedURL.Query() {
		for _, value := range values {
			params = append(params, QueryString{
				Name:  key,
				Value: value,
			})
		}
	}

	return params
}

// ParseResponseHeaders 解析原始HTTP响应头字符串
func ParseResponseHeaders(headerStr string) []Headers {
	var headers []Headers
	lines := strings.Split(headerStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers = append(headers, Headers{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}
	}
	return headers
}

// EstimateHeaderSize 估算HTTP头部大小
func EstimateHeaderSize(headers []Headers) int {
	size := 0
	for _, h := range headers {
		size += len(h.Name) + len(h.Value) + 4 // name: value\r\n
	}
	return size
}

// FormatBytes 格式化字节数
func FormatBytes(size int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case size >= GB:
		return strconv.FormatFloat(float64(size)/float64(GB), 'f', 2, 64) + " GB"
	case size >= MB:
		return strconv.FormatFloat(float64(size)/float64(MB), 'f', 2, 64) + " MB"
	case size >= KB:
		return strconv.FormatFloat(float64(size)/float64(KB), 'f', 2, 64) + " KB"
	default:
		return strconv.Itoa(size) + " B"
	}
}

// ReadBody 读取请求体内容
func ReadBody(entry *Entries) ([]byte, error) {
	if entry == nil || entry.Request.PostData == nil {
		return nil, nil
	}
	return []byte(entry.Request.PostData.Text), nil
}

// WriteToWriter 将HTTP请求写入io.Writer（用于调试）
func WriteRequestToWriter(entry *Entries, w io.Writer) error {
	if entry == nil {
		return NewInvalidFormatError("条目为空")
	}

	// 写入请求行
	fmt.Fprintf(w, "%s %s %s\r\n", entry.Request.Method, entry.Request.URL, entry.Request.HTTPVersion)

	// 写入请求头
	for _, h := range entry.Request.Headers {
		fmt.Fprintf(w, "%s: %s\r\n", h.Name, h.Value)
	}

	// 写入空行
	fmt.Fprintln(w)

	// 写入请求体
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		fmt.Fprint(w, entry.Request.PostData.Text)
	}

	return nil
}

// CloneEntry 深度复制HAR条目
func CloneEntry(entry *Entries) *Entries {
	if entry == nil {
		return nil
	}

	cloned := *entry

	// 复制切片
	cloned.Request.Headers = make([]Headers, len(entry.Request.Headers))
	copy(cloned.Request.Headers, entry.Request.Headers)

	cloned.Request.Cookies = make([]Cookie, len(entry.Request.Cookies))
	copy(cloned.Request.Cookies, entry.Request.Cookies)

	cloned.Request.QueryString = make([]QueryString, len(entry.Request.QueryString))
	copy(cloned.Request.QueryString, entry.Request.QueryString)

	if entry.Request.PostData != nil {
		pd := *entry.Request.PostData
		if len(entry.Request.PostData.Params) > 0 {
			pd.Params = make([]Param, len(entry.Request.PostData.Params))
			copy(pd.Params, entry.Request.PostData.Params)
		}
		cloned.Request.PostData = &pd
	}

	cloned.Response.Headers = make([]Headers, len(entry.Response.Headers))
	copy(cloned.Response.Headers, entry.Response.Headers)

	cloned.Response.Cookies = make([]Cookie, len(entry.Response.Cookies))
	copy(cloned.Response.Cookies, entry.Response.Cookies)

	return &cloned
}
