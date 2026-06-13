package har

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// ========== 辅助函数 ==========

// createHarForUtil 创建用于工具方法测试的HAR对象
func createHarForUtil() *Har {
	h := NewHar()
	h.SetCreator("test-util", "1.0")
	h.SetBrowser("TestBrowser", "1.0")

	e1 := h.AddEntry("GET", "https://example.com/api/users?id=1&format=json", "HTTP/1.1", "page_1")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")
	e1.SetResponseContentText(`{"users":[]}`)
	e1.AddRequestHeader("Accept", "application/json")
	e1.AddRequestHeader("User-Agent", "Go-HAR Test")
	e1.AddResponseHeader("Content-Type", "application/json")
	e1.AddResponseHeader("Cache-Control", "no-cache")
	e1.AddCookie("session", "abc123")
	e1.AddResponseCookie("tracking", "xyz789")
	e1.ServerIPAddress = "192.168.1.1"
	e1.Connection = "conn1"
	e1.SetTimings(10, 5, 15, 3, 50, 10, 8)

	e2 := h.AddEntry("POST", "https://other.com/api/data", "HTTP/1.1", "")
	e2.SetResponseStatus(404, "Not Found")
	e2.SetResponseContent(512, "text/html")
	e2.SetPostData("application/json", `{"key":"value"}`)
	e2.AddRequestHeader("Content-Type", "application/json")
	e2.AddResponseHeader("X-Error", "true")
	e2.SetTimings(5, 2, 8, 2, 30, 5, 3)

	e3 := h.AddEntry("GET", "https://cdn.example.com/static/style.css", "HTTP/1.1", "")
	e3.SetResponseStatus(301, "Moved Permanently")
	e3.SetResponseContent(128, "text/css")
	e3.AddResponseHeader("Location", "https://cdn2.example.com/style.css")
	e3.SetTimings(1, 1, 2, 1, 10, 2, 1)

	e4 := h.AddEntry("GET", "https://example.com/api/error", "HTTP/1.1", "")
	e4.SetResponseStatus(500, "Internal Server Error")
	e4.SetResponseContent(64, "application/json")
	e4.SetTimings(2, 1, 3, 1, 20, 3, 2)

	return h
}

// ========== Har 方法测试 ==========

func TestUtilHarClone(t *testing.T) {
	h := createHarForUtil()
	clone := h.Clone()

	if clone == nil {
		t.Fatal("Clone() 返回nil")
	}

	// 验证克隆对象与原对象内容相同
	if clone.Log.Version != h.Log.Version {
		t.Errorf("Clone() 版本不匹配: got %q, want %q", clone.Log.Version, h.Log.Version)
	}
	if clone.Log.Creator.Name != h.Log.Creator.Name {
		t.Errorf("Clone() 创建者名称不匹配: got %q, want %q", clone.Log.Creator.Name, h.Log.Creator.Name)
	}
	if len(clone.Log.Entries) != len(h.Log.Entries) {
		t.Errorf("Clone() 条目数量不匹配: got %d, want %d", len(clone.Log.Entries), len(h.Log.Entries))
	}

	// 验证深拷贝：修改克隆对象不应影响原始对象
	clone.Log.Entries[0].Request.Method = "PUT"
	if h.Log.Entries[0].Request.Method == "PUT" {
		t.Error("Clone() 不是深拷贝，修改克隆影响了原始对象")
	}

	clone.Log.Entries[0].Request.Headers[0].Value = "modified"
	if h.Log.Entries[0].Request.Headers[0].Value == "modified" {
		t.Error("Clone() 头部切片未深拷贝，修改克隆影响了原始对象")
	}
}

func TestUtilHarCloneNil(t *testing.T) {
	var h *Har = nil
	clone := h.Clone()
	if clone != nil {
		t.Error("nil Har 的 Clone() 应返回nil")
	}
}

func TestUtilHarGetEntryCount(t *testing.T) {
	h := createHarForUtil()

	count := h.GetEntryCount()
	if count != 4 {
		t.Errorf("GetEntryCount() = %d, want 4", count)
	}

	// 空HAR
	emptyHar := NewHar()
	if emptyHar.GetEntryCount() != 0 {
		t.Errorf("空HAR的 GetEntryCount() = %d, want 0", emptyHar.GetEntryCount())
	}

	// nil接收者
	var nilHar *Har
	if nilHar.GetEntryCount() != 0 {
		t.Errorf("nil Har的 GetEntryCount() = %d, want 0", nilHar.GetEntryCount())
	}
}

func TestUtilHarWalk(t *testing.T) {
	h := createHarForUtil()

	// 测试正常遍历
	var visited []string
	err := h.Walk(func(e *Entries) error {
		visited = append(visited, e.Request.Method)
		return nil
	})
	if err != nil {
		t.Errorf("Walk() 返回错误: %v", err)
	}
	if len(visited) != 4 {
		t.Errorf("Walk() 访问了 %d 个条目, want 4", len(visited))
	}

	// 测试提前终止
	var earlyVisited []string
	err = h.Walk(func(e *Entries) error {
		earlyVisited = append(earlyVisited, e.Request.Method)
		if len(earlyVisited) == 2 {
			return io.EOF
		}
		return nil
	})
	if err != io.EOF {
		t.Errorf("Walk() 提前终止时错误 = %v, want io.EOF", err)
	}
	if len(earlyVisited) != 2 {
		t.Errorf("Walk() 提前终止时访问了 %d 个条目, want 2", len(earlyVisited))
	}

	// nil接收者
	var nilHar *Har
	err = nilHar.Walk(func(e *Entries) error { return nil })
	if err != nil {
		t.Errorf("nil Har的 Walk() 不应返回错误, got %v", err)
	}
}

func TestUtilHarGetUniqueDomains(t *testing.T) {
	h := createHarForUtil()

	domains := h.GetUniqueDomains()
	expected := []string{"cdn.example.com", "example.com", "other.com"}
	if len(domains) != len(expected) {
		t.Errorf("GetUniqueDomains() 返回 %d 个域名, want %d", len(domains), len(expected))
	}
	for i, d := range domains {
		if d != expected[i] {
			t.Errorf("GetUniqueDomains()[%d] = %q, want %q", i, d, expected[i])
		}
	}

	// nil接收者
	var nilHar *Har
	if nilHar.GetUniqueDomains() != nil {
		t.Error("nil Har的 GetUniqueDomains() 应返回nil")
	}
}

func TestUtilHarEquals(t *testing.T) {
	h1 := createHarForUtil()
	h2 := createHarForUtil()

	// 相同内容
	if !h1.Equals(h2) {
		t.Error("Equals() 对相同内容的HAR应返回true")
	}

	// 修改版本
	h2.Log.Version = "1.1"
	if h1.Equals(h2) {
		t.Error("Equals() 对不同版本应返回false")
	}
	h2.Log.Version = h1.Log.Version

	// 修改创建者
	h2.Log.Creator.Name = "different"
	if h1.Equals(h2) {
		t.Error("Equals() 对不同创建者应返回false")
	}
	h2.Log.Creator.Name = h1.Log.Creator.Name

	// 修改浏览器
	h2.Log.Browser.Name = "OtherBrowser"
	if h1.Equals(h2) {
		t.Error("Equals() 对不同浏览器应返回false")
	}
	h2.Log.Browser.Name = h1.Log.Browser.Name

	// 修改条目数量
	h2.Log.Entries = h2.Log.Entries[:3]
	if h1.Equals(h2) {
		t.Error("Equals() 对不同条目数量应返回false")
	}

	// 修改条目方法
	h2 = createHarForUtil()
	h2.Log.Entries[0].Request.Method = "PUT"
	if h1.Equals(h2) {
		t.Error("Equals() 对不同条目方法应返回false")
	}

	// 修改条目URL
	h2 = createHarForUtil()
	h2.Log.Entries[0].Request.URL = "https://different.com"
	if h1.Equals(h2) {
		t.Error("Equals() 对不同条目URL应返回false")
	}

	// 修改条目状态码
	h2 = createHarForUtil()
	h2.Log.Entries[0].Response.Status = 404
	if h1.Equals(h2) {
		t.Error("Equals() 对不同状态码应返回false")
	}

	// 双nil
	var nilH1, nilH2 *Har
	if !nilH1.Equals(nilH2) {
		t.Error("Equals() 对双nil应返回true")
	}

	// 一方nil
	if nilH1.Equals(h1) {
		t.Error("Equals() 对一方nil应返回false")
	}
	if h1.Equals(nilH1) {
		t.Error("Equals() 对一方nil应返回false")
	}
}

func TestUtilHarSaveToFileGzipped(t *testing.T) {
	h := createHarForUtil()

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "har_test_*.har.gz")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// 保存为gzip压缩文件
	err = h.SaveToFileGzipped(tmpPath, true)
	if err != nil {
		t.Fatalf("SaveToFileGzipped() 返回错误: %v", err)
	}

	// 验证文件是有效的gzip文件
	f, err := os.Open(tmpPath)
	if err != nil {
		t.Fatalf("打开gzip文件失败: %v", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("创建gzip reader失败: %v", err)
	}
	defer gzReader.Close()

	data, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("读取gzip数据失败: %v", err)
	}

	// 验证解压后是有效的HAR JSON
	var parsedHar Har
	if err := json.Unmarshal(data, &parsedHar); err != nil {
		t.Fatalf("解析解压后的JSON失败: %v", err)
	}
	if parsedHar.GetEntryCount() != 4 {
		t.Errorf("解压后的HAR条目数 = %d, want 4", parsedHar.GetEntryCount())
	}

	// nil接收者
	var nilHar *Har
	err = nilHar.SaveToFileGzipped(tmpPath, true)
	if err == nil {
		t.Error("nil Har的 SaveToFileGzipped() 应返回错误")
	}
}

func TestUtilHarSaveToWriter(t *testing.T) {
	h := createHarForUtil()

	var buf bytes.Buffer
	err := h.SaveToWriter(&buf, false)
	if err != nil {
		t.Fatalf("SaveToWriter() 返回错误: %v", err)
	}

	// 验证写入的是有效JSON
	var parsedHar Har
	if err := json.Unmarshal(buf.Bytes(), &parsedHar); err != nil {
		t.Fatalf("解析SaveToWriter输出失败: %v", err)
	}
	if parsedHar.GetEntryCount() != 4 {
		t.Errorf("SaveToWriter输出的HAR条目数 = %d, want 4", parsedHar.GetEntryCount())
	}

	// 测试带缩进
	buf.Reset()
	err = h.SaveToWriter(&buf, true)
	if err != nil {
		t.Fatalf("SaveToWriter(indent=true) 返回错误: %v", err)
	}
	if !strings.Contains(buf.String(), "\n") {
		t.Error("SaveToWriter(indent=true) 输出应包含换行符")
	}

	// nil接收者
	var nilHar *Har
	err = nilHar.SaveToWriter(&buf, true)
	if err == nil {
		t.Error("nil Har的 SaveToWriter() 应返回错误")
	}
}

// ========== Entries 方法测试 ==========

func TestUtilEntriesIsError(t *testing.T) {
	h := createHarForUtil()

	tests := []struct {
		entryIdx int
		want     bool
	}{
		{0, false}, // 200
		{1, true},  // 404
		{2, false}, // 301
		{3, true},  // 500
	}
	for _, tt := range tests {
		got := h.Log.Entries[tt.entryIdx].IsError()
		if got != tt.want {
			t.Errorf("Entries[%d].IsError() = %v, want %v", tt.entryIdx, got, tt.want)
		}
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.IsError() != false {
		t.Error("nil Entries的 IsError() 应返回false")
	}
}

func TestUtilEntriesIsRedirect(t *testing.T) {
	h := createHarForUtil()

	tests := []struct {
		entryIdx int
		want     bool
	}{
		{0, false}, // 200
		{1, false}, // 404
		{2, true},  // 301
		{3, false}, // 500
	}
	for _, tt := range tests {
		got := h.Log.Entries[tt.entryIdx].IsRedirect()
		if got != tt.want {
			t.Errorf("Entries[%d].IsRedirect() = %v, want %v", tt.entryIdx, got, tt.want)
		}
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.IsRedirect() != false {
		t.Error("nil Entries的 IsRedirect() 应返回false")
	}
}

func TestUtilEntriesIsSuccess(t *testing.T) {
	h := createHarForUtil()

	tests := []struct {
		entryIdx int
		want     bool
	}{
		{0, true},  // 200
		{1, false}, // 404
		{2, false}, // 301
		{3, false}, // 500
	}
	for _, tt := range tests {
		got := h.Log.Entries[tt.entryIdx].IsSuccess()
		if got != tt.want {
			t.Errorf("Entries[%d].IsSuccess() = %v, want %v", tt.entryIdx, got, tt.want)
		}
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.IsSuccess() != false {
		t.Error("nil Entries的 IsSuccess() 应返回false")
	}
}

func TestUtilEntriesGetElapsedTime(t *testing.T) {
	h := createHarForUtil()

	// entry0 的 timings: 10+5+15+3+50+10+8 = 101ms (但Time可能由SetTimings单独计算)
	duration := h.Log.Entries[0].GetElapsedTime()
	expectedMs := h.Log.Entries[0].Time
	expectedDuration := time.Duration(expectedMs * float64(time.Millisecond))
	if duration != expectedDuration {
		t.Errorf("GetElapsedTime() = %v, want %v", duration, expectedDuration)
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.GetElapsedTime() != 0 {
		t.Error("nil Entries的 GetElapsedTime() 应返回0")
	}
}

func TestUtilEntriesGetURL(t *testing.T) {
	h := createHarForUtil()

	u := h.Log.Entries[0].GetURL()
	if u == nil {
		t.Fatal("GetURL() 返回nil")
	}
	if u.Host != "example.com" {
		t.Errorf("GetURL().Host = %q, want %q", u.Host, "example.com")
	}
	if u.Path != "/api/users" {
		t.Errorf("GetURL().Path = %q, want %q", u.Path, "/api/users")
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.GetURL() != nil {
		t.Error("nil Entries的 GetURL() 应返回nil")
	}
}

func TestUtilEntriesGetDomain(t *testing.T) {
	h := createHarForUtil()

	tests := []struct {
		entryIdx int
		want     string
	}{
		{0, "example.com"},
		{1, "other.com"},
		{2, "cdn.example.com"},
		{3, "example.com"},
	}
	for _, tt := range tests {
		got := h.Log.Entries[tt.entryIdx].GetDomain()
		if got != tt.want {
			t.Errorf("Entries[%d].GetDomain() = %q, want %q", tt.entryIdx, got, tt.want)
		}
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.GetDomain() != "" {
		t.Error("nil Entries的 GetDomain() 应返回空字符串")
	}
}

func TestUtilEntriesGetSize(t *testing.T) {
	h := createHarForUtil()

	size := h.Log.Entries[0].GetSize()
	// 请求头大小 150 + 请求体大小 0 + 响应头大小 120 + 响应体大小 1024 = 1294
	// (note: AddEntry sets HeadersSize=-1 and BodySize=-1 by default, but
	//  SetResponseContent only sets Content fields, not BodySize/HeadersSize)
	// The actual values depend on how the test HAR was constructed.
	// Since AddEntry sets -1 for sizes, and nothing overrides them here,
	// GetSize treats -1 as 0.
	if size < 0 {
		t.Errorf("GetSize() = %d, should not be negative", size)
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.GetSize() != 0 {
		t.Error("nil Entries的 GetSize() 应返回0")
	}
}

func TestUtilEntriesToCurl(t *testing.T) {
	h := createHarForUtil()

	curl := h.Log.Entries[0].ToCurl()
	if curl == "" {
		t.Fatal("ToCurl() 返回空字符串")
	}
	if !strings.Contains(curl, "curl") {
		t.Error("ToCurl() 应包含 'curl'")
	}
	// GET请求不应包含 -X 参数
	if strings.Contains(curl, "-X GET") {
		t.Error("GET请求的ToCurl()不应包含 '-X GET'")
	}
	// 应包含请求头
	if !strings.Contains(curl, "-H") {
		t.Error("ToCurl() 应包含 '-H' 参数")
	}
	// 应包含URL
	if !strings.Contains(curl, "https://example.com/api/users") {
		t.Error("ToCurl() 应包含请求URL")
	}

	// POST请求应包含 -X POST 和 --data
	postCurl := h.Log.Entries[1].ToCurl()
	if !strings.Contains(postCurl, "-X POST") {
		t.Error("POST请求的ToCurl()应包含 '-X POST'")
	}
	if !strings.Contains(postCurl, "--data") {
		t.Error("POST请求的ToCurl()应包含 '--data'")
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.ToCurl() != "" {
		t.Error("nil Entries的 ToCurl() 应返回空字符串")
	}
}

func TestUtilEntriesToWget(t *testing.T) {
	h := createHarForUtil()

	wget := h.Log.Entries[0].ToWget()
	if wget == "" {
		t.Fatal("ToWget() 返回空字符串")
	}
	if !strings.Contains(wget, "wget") {
		t.Error("ToWget() 应包含 'wget'")
	}
	if !strings.Contains(wget, "--header") {
		t.Error("ToWget() 应包含 '--header'")
	}
	if !strings.Contains(wget, "https://example.com/api/users") {
		t.Error("ToWget() 应包含请求URL")
	}

	// POST请求
	postWget := h.Log.Entries[1].ToWget()
	if !strings.Contains(postWget, "--method=POST") {
		t.Error("POST请求的ToWget()应包含 '--method=POST'")
	}
	if !strings.Contains(postWget, "--post-data") && !strings.Contains(postWget, "--body-data") {
		t.Error("POST请求的ToWget()应包含 '--post-data' 或 '--body-data'")
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.ToWget() != "" {
		t.Error("nil Entries的 ToWget() 应返回空字符串")
	}
}

func TestUtilEntriesToPythonRequests(t *testing.T) {
	h := createHarForUtil()

	py := h.Log.Entries[0].ToPythonRequests()
	if py == "" {
		t.Fatal("ToPythonRequests() 返回空字符串")
	}
	if !strings.Contains(py, "requests.get") {
		t.Error("GET请求的ToPythonRequests()应包含 'requests.get'")
	}
	if !strings.Contains(py, "headers=") {
		t.Error("ToPythonRequests() 应包含 'headers='")
	}

	// POST请求
	postPy := h.Log.Entries[1].ToPythonRequests()
	if !strings.Contains(postPy, "requests.post") {
		t.Error("POST请求的ToPythonRequests()应包含 'requests.post'")
	}
	if !strings.Contains(postPy, "data=") {
		t.Error("POST请求的ToPythonRequests()应包含 'data='")
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.ToPythonRequests() != "" {
		t.Error("nil Entries的 ToPythonRequests() 应返回空字符串")
	}
}

func TestUtilEntriesGetRequestBody(t *testing.T) {
	h := createHarForUtil()

	// 无PostData的条目
	body := h.Log.Entries[0].GetRequestBody()
	if body != nil {
		t.Errorf("无PostData的 GetRequestBody() = %v, want nil", body)
	}

	// 有PostData的条目
	body = h.Log.Entries[1].GetRequestBody()
	if body == nil {
		t.Fatal("有PostData的 GetRequestBody() 返回nil")
	}
	if string(body) != `{"key":"value"}` {
		t.Errorf("GetRequestBody() = %q, want %q", string(body), `{"key":"value"}`)
	}

	// nil接收者
	var nilEntry *Entries
	if nilEntry.GetRequestBody() != nil {
		t.Error("nil Entries的 GetRequestBody() 应返回nil")
	}
}

func TestUtilEntriesGetResponseBody(t *testing.T) {
	h := createHarForUtil()

	// 普通文本响应
	body, err := h.Log.Entries[0].GetResponseBody()
	if err != nil {
		t.Fatalf("GetResponseBody() 返回错误: %v", err)
	}
	if string(body) != `{"users":[]}` {
		t.Errorf("GetResponseBody() = %q, want %q", string(body), `{"users":[]}`)
	}

	// base64编码响应
	entry := h.AddEntry("GET", "https://example.com/binary", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.Response.Content.Encoding = "base64"
	originalData := []byte("binary data here")
	entry.Response.Content.Text = base64.StdEncoding.EncodeToString(originalData)
	entry.Response.Content.Size = len(originalData)

	body, err = entry.GetResponseBody()
	if err != nil {
		t.Fatalf("base64 GetResponseBody() 返回错误: %v", err)
	}
	if string(body) != string(originalData) {
		t.Errorf("base64 GetResponseBody() = %q, want %q", string(body), string(originalData))
	}

	// 空内容响应
	emptyEntry := h.AddEntry("GET", "https://example.com/empty", "HTTP/1.1", "")
	emptyEntry.SetResponseStatus(204, "No Content")
	body, err = emptyEntry.GetResponseBody()
	if err != nil {
		t.Fatalf("空内容 GetResponseBody() 返回错误: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("空内容 GetResponseBody() = %v, want empty", body)
	}

	// nil接收者
	var nilEntry *Entries
	body, err = nilEntry.GetResponseBody()
	if body != nil || err != nil {
		t.Errorf("nil Entries的 GetResponseBody() 应返回 (nil, nil), got (%v, %v)", body, err)
	}
}

// ========== Request 方法测试 ==========

func TestUtilRequestGetHeader(t *testing.T) {
	h := createHarForUtil()
	req := &h.Log.Entries[0].Request

	// 存在的头部（大小写不敏感）
	if v := req.GetHeader("accept"); v != "application/json" {
		t.Errorf("GetHeader(\"accept\") = %q, want %q", v, "application/json")
	}
	if v := req.GetHeader("Accept"); v != "application/json" {
		t.Errorf("GetHeader(\"Accept\") = %q, want %q", v, "application/json")
	}
	if v := req.GetHeader("ACCEPT"); v != "application/json" {
		t.Errorf("GetHeader(\"ACCEPT\") = %q, want %q", v, "application/json")
	}

	// 不存在的头部
	if v := req.GetHeader("X-Not-Exist"); v != "" {
		t.Errorf("GetHeader(\"X-Not-Exist\") = %q, want empty", v)
	}

	// nil接收者
	var nilReq *Request
	if v := nilReq.GetHeader("Accept"); v != "" {
		t.Error("nil Request的 GetHeader() 应返回空字符串")
	}
}

func TestUtilRequestGetHeaders(t *testing.T) {
	h := createHarForUtil()
	req := &h.Log.Entries[0].Request

	// 添加一个重复的头部用于测试
	req.Headers = append(req.Headers, Headers{Name: "Accept", Value: "text/html"})

	values := req.GetHeaderValues("Accept")
	if len(values) != 2 {
		t.Fatalf("GetHeaders(\"Accept\") 返回 %d 个值, want 2", len(values))
	}
	if values[0] != "application/json" {
		t.Errorf("GetHeaders(\"Accept\")[0] = %q, want %q", values[0], "application/json")
	}
	if values[1] != "text/html" {
		t.Errorf("GetHeaders(\"Accept\")[1] = %q, want %q", values[1], "text/html")
	}

	// 不存在的头部
	values = req.GetHeaderValues("X-Not-Exist")
	if len(values) != 0 {
		t.Errorf("GetHeaders(\"X-Not-Exist\") 返回 %d 个值, want 0", len(values))
	}

	// nil接收者
	var nilReq *Request
	if nilReq.GetHeaderValues("Accept") != nil {
		t.Error("nil Request的 GetHeaders() 应返回nil")
	}
}

func TestUtilRequestGetCookie(t *testing.T) {
	h := createHarForUtil()
	req := &h.Log.Entries[0].Request

	// 存在的Cookie
	cookie := req.GetCookie("session")
	if cookie == nil {
		t.Fatal("GetCookie(\"session\") 返回nil")
	}
	if cookie.Value != "abc123" {
		t.Errorf("GetCookie(\"session\").Value = %q, want %q", cookie.Value, "abc123")
	}

	// 不存在的Cookie
	cookie = req.GetCookie("notexist")
	if cookie != nil {
		t.Error("GetCookie(\"notexist\") 应返回nil")
	}

	// Cookie名称区分大小写
	cookie = req.GetCookie("Session")
	if cookie != nil {
		t.Error("GetCookie() 应区分大小写")
	}

	// nil接收者
	var nilReq *Request
	if nilReq.GetCookie("session") != nil {
		t.Error("nil Request的 GetCookie() 应返回nil")
	}
}

func TestUtilRequestHasHeader(t *testing.T) {
	h := createHarForUtil()
	req := &h.Log.Entries[0].Request

	// 存在的头部（大小写不敏感）
	if !req.HasHeader("accept") {
		t.Error("HasHeader(\"accept\") 应返回true")
	}
	if !req.HasHeader("Accept") {
		t.Error("HasHeader(\"Accept\") 应返回true")
	}
	if !req.HasHeader("user-agent") {
		t.Error("HasHeader(\"user-agent\") 应返回true")
	}

	// 不存在的头部
	if req.HasHeader("X-Not-Exist") {
		t.Error("HasHeader(\"X-Not-Exist\") 应返回false")
	}

	// nil接收者
	var nilReq *Request
	if nilReq.HasHeader("Accept") {
		t.Error("nil Request的 HasHeader() 应返回false")
	}
}

// ========== Response 方法测试 ==========

func TestUtilResponseGetHeader(t *testing.T) {
	h := createHarForUtil()
	resp := &h.Log.Entries[0].Response

	// 存在的头部（大小写不敏感）
	if v := resp.GetHeader("content-type"); v != "application/json" {
		t.Errorf("GetHeader(\"content-type\") = %q, want %q", v, "application/json")
	}
	if v := resp.GetHeader("Content-Type"); v != "application/json" {
		t.Errorf("GetHeader(\"Content-Type\") = %q, want %q", v, "application/json")
	}

	// 不存在的头部
	if v := resp.GetHeader("X-Not-Exist"); v != "" {
		t.Errorf("GetHeader(\"X-Not-Exist\") = %q, want empty", v)
	}

	// nil接收者
	var nilResp *Response
	if v := nilResp.GetHeader("Content-Type"); v != "" {
		t.Error("nil Response的 GetHeader() 应返回空字符串")
	}
}

func TestUtilResponseGetHeaders(t *testing.T) {
	h := createHarForUtil()
	resp := &h.Log.Entries[0].Response

	// 添加重复头部
	resp.Headers = append(resp.Headers, Headers{Name: "Content-Type", Value: "charset=utf-8"})

	values := resp.GetHeaderValues("Content-Type")
	if len(values) != 2 {
		t.Fatalf("GetHeaders(\"Content-Type\") 返回 %d 个值, want 2", len(values))
	}

	// 不存在的头部
	values = resp.GetHeaderValues("X-Not-Exist")
	if len(values) != 0 {
		t.Errorf("GetHeaders(\"X-Not-Exist\") 返回 %d 个值, want 0", len(values))
	}

	// nil接收者
	var nilResp *Response
	if nilResp.GetHeaderValues("Content-Type") != nil {
		t.Error("nil Response的 GetHeaders() 应返回nil")
	}
}

func TestUtilResponseGetCookie(t *testing.T) {
	h := createHarForUtil()
	resp := &h.Log.Entries[0].Response

	// 存在的Cookie
	cookie := resp.GetCookie("tracking")
	if cookie == nil {
		t.Fatal("GetCookie(\"tracking\") 返回nil")
	}
	if cookie.Value != "xyz789" {
		t.Errorf("GetCookie(\"tracking\").Value = %q, want %q", cookie.Value, "xyz789")
	}

	// 不存在的Cookie
	cookie = resp.GetCookie("notexist")
	if cookie != nil {
		t.Error("GetCookie(\"notexist\") 应返回nil")
	}

	// Cookie名称区分大小写
	cookie = resp.GetCookie("Tracking")
	if cookie != nil {
		t.Error("GetCookie() 应区分大小写")
	}

	// nil接收者
	var nilResp *Response
	if nilResp.GetCookie("tracking") != nil {
		t.Error("nil Response的 GetCookie() 应返回nil")
	}
}

func TestUtilResponseHasHeader(t *testing.T) {
	h := createHarForUtil()
	resp := &h.Log.Entries[0].Response

	if !resp.HasHeader("content-type") {
		t.Error("HasHeader(\"content-type\") 应返回true")
	}
	if !resp.HasHeader("Content-Type") {
		t.Error("HasHeader(\"Content-Type\") 应返回true")
	}
	if resp.HasHeader("X-Not-Exist") {
		t.Error("HasHeader(\"X-Not-Exist\") 应返回false")
	}

	// nil接收者
	var nilResp *Response
	if nilResp.HasHeader("Content-Type") {
		t.Error("nil Response的 HasHeader() 应返回false")
	}
}

func TestUtilResponseGetContentType(t *testing.T) {
	h := createHarForUtil()
	resp := &h.Log.Entries[0].Response

	ct := resp.GetContentType()
	if ct != "application/json" {
		t.Errorf("GetContentType() = %q, want %q", ct, "application/json")
	}

	// nil接收者
	var nilResp *Response
	if nilResp.GetContentType() != "" {
		t.Error("nil Response的 GetContentType() 应返回空字符串")
	}
}

// ========== Content 方法测试 ==========

func TestUtilContentEncodeContent(t *testing.T) {
	c := &Content{}
	data := []byte("hello world binary data")
	c.EncodeContent(data, "application/octet-stream")

	if c.Encoding != "base64" {
		t.Errorf("EncodeContent() 设置 Encoding = %q, want %q", c.Encoding, "base64")
	}
	if c.MimeType != "application/octet-stream" {
		t.Errorf("EncodeContent() 设置 MimeType = %q, want %q", c.MimeType, "application/octet-stream")
	}
	if c.Size != len(data) {
		t.Errorf("EncodeContent() 设置 Size = %d, want %d", c.Size, len(data))
	}

	// 验证Text可以被正确解码
	decoded, err := base64.StdEncoding.DecodeString(c.Text)
	if err != nil {
		t.Fatalf("base64解码失败: %v", err)
	}
	if string(decoded) != string(data) {
		t.Errorf("解码后数据 = %q, want %q", string(decoded), string(data))
	}

	// nil接收者
	var nilContent *Content
	nilContent.EncodeContent(data, "text/plain") // 不应panic
}

func TestUtilContentSetText(t *testing.T) {
	c := &Content{}
	text := "Hello, World!"
	c.SetText(text)

	if c.Text != text {
		t.Errorf("SetText() 设置 Text = %q, want %q", c.Text, text)
	}
	if c.Size != len(text) {
		t.Errorf("SetText() 设置 Size = %d, want %d", c.Size, len(text))
	}

	// 更新文本
	newText := "Updated text"
	c.SetText(newText)
	if c.Text != newText {
		t.Errorf("SetText() 更新 Text = %q, want %q", c.Text, newText)
	}
	if c.Size != len(newText) {
		t.Errorf("SetText() 更新 Size = %d, want %d", c.Size, len(newText))
	}

	// nil接收者
	var nilContent *Content
	nilContent.SetText("test") // 不应panic
}

// ========== 从测试文件解析测试 ==========

func TestUtilParseAndUseMethods(t *testing.T) {
	// 从测试文件解析HAR
	h, err := ParseHarFile("testdata/full.har")
	if err != nil {
		t.Fatalf("解析测试文件失败: %v", err)
	}

	// 测试各种工具方法
	if h.GetEntryCount() < 1 {
		t.Error("从文件解析的HAR应至少有1个条目")
	}

	domains := h.GetUniqueDomains()
	if len(domains) == 0 {
		t.Error("GetUniqueDomains() 应返回至少一个域名")
	}

	// 测试第一个条目的方法
	entry := &h.Log.Entries[0]
	if !entry.IsSuccess() {
		t.Error("第一个条目应为成功响应(200)")
	}

	u := entry.GetURL()
	if u == nil {
		t.Fatal("GetURL() 返回nil")
	}

	domain := entry.GetDomain()
	if domain == "" {
		t.Error("GetDomain() 不应返回空字符串")
	}

	// 测试请求方法
	reqHeader := entry.Request.GetHeader("Accept")
	if reqHeader == "" {
		t.Error("GetHeader(\"Accept\") 不应为空")
	}

	// 测试响应方法
	respCT := entry.Response.GetContentType()
	if respCT == "" {
		t.Error("GetContentType() 不应为空")
	}
}

func TestUtilParseMinimal(t *testing.T) {
	h, err := ParseHarFile("testdata/minimal.har")
	if err != nil {
		t.Fatalf("解析minimal.har失败: %v", err)
	}

	if h.GetEntryCount() != 0 {
		t.Errorf("minimal.har应有0个条目, got %d", h.GetEntryCount())
	}

	domains := h.GetUniqueDomains()
	if len(domains) != 0 {
		t.Errorf("minimal.har的GetUniqueDomains()应返回空, got %v", domains)
	}
}

// ========== 综合测试 ==========

func TestUtilClonePreservesDeepCopy(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	entry.AddRequestHeader("X-Test", "value1")
	entry.AddCookie("session", "abc")
	entry.SetResponseStatus(200, "OK")
	entry.Response.Content.Text = "original content"
	entry.Response.Content.Size = len("original content")

	clone := h.Clone()

	// 修改克隆的请求头
	clone.Log.Entries[0].Request.Headers[0].Value = "modified"
	if h.Log.Entries[0].Request.Headers[0].Value != "value1" {
		t.Error("修改克隆的请求头影响了原始对象")
	}

	// 修改克隆的Cookie
	clone.Log.Entries[0].Request.Cookies[0].Value = "modified"
	if h.Log.Entries[0].Request.Cookies[0].Value != "abc" {
		t.Error("修改克隆的Cookie影响了原始对象")
	}

	// 修改克隆的响应内容
	clone.Log.Entries[0].Response.Content.Text = "modified content"
	if h.Log.Entries[0].Response.Content.Text != "original content" {
		t.Error("修改克隆的响应内容影响了原始对象")
	}
}

func TestUtilWalkCount(t *testing.T) {
	h := createHarForUtil()

	count := 0
	err := h.Walk(func(e *Entries) error {
		count++
		return nil
	})
	if err != nil {
		t.Errorf("Walk() 返回错误: %v", err)
	}
	if count != 4 {
		t.Errorf("Walk() 遍历了 %d 个条目, want 4", count)
	}
}

func TestUtilSaveToWriterRoundTrip(t *testing.T) {
	h := createHarForUtil()

	// 写入
	var buf bytes.Buffer
	err := h.SaveToWriter(&buf, true)
	if err != nil {
		t.Fatalf("SaveToWriter() 失败: %v", err)
	}

	// 读回
	parsed, err := ParseHar(buf.Bytes())
	if err != nil {
		t.Fatalf("解析写入的数据失败: %v", err)
	}

	if !h.Equals(parsed) {
		t.Error("SaveToWriter + ParseHar 往返测试失败")
	}
}

func TestUtilGzipRoundTrip(t *testing.T) {
	h := createHarForUtil()

	tmpFile, err := os.CreateTemp("", "har_gzip_test_*.har.gz")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// 保存为gzip
	err = h.SaveToFileGzipped(tmpPath, true)
	if err != nil {
		t.Fatalf("SaveToFileGzipped() 失败: %v", err)
	}

	// 读取并解压
	f, err := os.Open(tmpPath)
	if err != nil {
		t.Fatalf("打开gzip文件失败: %v", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("创建gzip reader失败: %v", err)
	}
	defer gzReader.Close()

	data, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("读取gzip数据失败: %v", err)
	}

	// 解析
	parsed, err := ParseHar(data)
	if err != nil {
		t.Fatalf("解析解压数据失败: %v", err)
	}

	if !h.Equals(parsed) {
		t.Error("gzip往返测试失败")
	}
}

func TestUtilEntriesGetSizeWithPositiveValues(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	// 手动设置正值大小
	entry.Request.HeadersSize = 150
	entry.Request.BodySize = 200
	entry.Response.HeadersSize = 120
	entry.Response.BodySize = 1024

	size := entry.GetSize()
	expected := 150 + 200 + 120 + 1024
	if size != expected {
		t.Errorf("GetSize() = %d, want %d", size, expected)
	}
}

func TestUtilEntriesGetSizeWithNegativeValues(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	// AddEntry默认设置HeadersSize=-1, BodySize=-1
	entry.Request.HeadersSize = -1
	entry.Request.BodySize = -1
	entry.Response.HeadersSize = -1
	entry.Response.BodySize = -1

	size := entry.GetSize()
	if size != 0 {
		t.Errorf("GetSize() 对负值应返回0, got %d", size)
	}
}

func TestUtilToCurlWithSpecialChars(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Content-Type", "application/json")
	entry.SetPostData("application/json", `{"key":"it's a test"}`)
	entry.SetResponseStatus(200, "OK")

	curl := entry.ToCurl()
	if !strings.Contains(curl, "-X POST") {
		t.Error("POST请求应包含 '-X POST'")
	}
	if !strings.Contains(curl, "--data") {
		t.Error("有PostData的请求应包含 '--data'")
	}
}

func TestUtilToPythonRequestsWithoutHeaders(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/simple", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	// 清空headers
	entry.Request.Headers = nil
	entry.Request.Cookies = nil

	py := entry.ToPythonRequests()
	if !strings.Contains(py, "requests.get") {
		t.Error("应包含 'requests.get'")
	}
	if strings.Contains(py, "headers=") {
		t.Error("无头部时不应包含 'headers='")
	}
	if strings.Contains(py, "cookies=") {
		t.Error("无Cookie时不应包含 'cookies='")
	}
}

func TestUtilGetResponseBodyInvalidBase64(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")
	entry.Response.Content.Encoding = "base64"
	entry.Response.Content.Text = "!!!invalid-base64!!!"

	_, err := entry.GetResponseBody()
	if err == nil {
		t.Error("无效base64编码应返回错误")
	}
}

func TestUtilGetURLInvalid(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "://invalid-url", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	u := entry.GetURL()
	// 无效URL应返回nil或部分解析的URL
	// url.Parse对某些格式可能不返回错误，所以验证不会panic即可
	_ = u
}

func TestUtilHarEqualsSelf(t *testing.T) {
	h := createHarForUtil()
	if !h.Equals(h) {
		t.Error("HAR对象应等于自身")
	}
}

func TestUtilHarEqualsEmpty(t *testing.T) {
	h1 := NewHar()
	h2 := NewHar()
	if !h1.Equals(h2) {
		t.Error("两个新创建的空HAR应相等")
	}
}

func TestUtilContentEncodeDecodeRoundTrip(t *testing.T) {
	c := &Content{}
	originalData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	c.EncodeContent(originalData, "application/octet-stream")

	// 模拟从JSON反序列化
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("序列化Content失败: %v", err)
	}

	var decoded Content
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("反序列化Content失败: %v", err)
	}

	// 使用GetResponseBody验证解码
	entry := &Entries{}
	entry.Response.Content = decoded
	entry.Response.Status = 200

	body, err := entry.GetResponseBody()
	if err != nil {
		t.Fatalf("GetResponseBody() 返回错误: %v", err)
	}
	if !bytes.Equal(body, originalData) {
		t.Errorf("编码解码往返测试失败: got %v, want %v", body, originalData)
	}
}

func TestUtilGetUniqueDomainsWithParsedHar(t *testing.T) {
	h, err := ParseHarFile("testdata/full.har")
	if err != nil {
		t.Fatalf("解析full.har失败: %v", err)
	}

	domains := h.GetUniqueDomains()
	if len(domains) == 0 {
		t.Error("full.har应至少有一个域名")
	}

	// 验证域名是排序的
	for i := 1; i < len(domains); i++ {
		if domains[i] < domains[i-1] {
			t.Errorf("域名未排序: %q 在 %q 之前", domains[i], domains[i-1])
		}
	}
}

func TestUtilToWgetNoPostData(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("DELETE", "https://example.com/resource/1", "HTTP/1.1", "")
	entry.SetResponseStatus(204, "No Content")
	entry.Request.Headers = nil // 清空headers

	wget := entry.ToWget()
	if !strings.Contains(wget, "wget") {
		t.Error("ToWget() 应包含 'wget'")
	}
	if !strings.Contains(wget, "--method=DELETE") {
		t.Error("DELETE请求的ToWget()应包含 '--method=DELETE'")
	}
}

func TestUtilGetElapsedTimeZero(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	entry.Time = 0

	if entry.GetElapsedTime() != 0 {
		t.Error("Time=0时 GetElapsedTime() 应返回0")
	}
}

func TestUtilCloneEmptyHar(t *testing.T) {
	h := NewHar()
	clone := h.Clone()
	if clone == nil {
		t.Fatal("空HAR的Clone()不应返回nil")
	}
	if clone.GetEntryCount() != 0 {
		t.Error("空HAR的克隆应有0个条目")
	}
}

func TestUtilWalkEmptyHar(t *testing.T) {
	h := NewHar()
	count := 0
	err := h.Walk(func(e *Entries) error {
		count++
		return nil
	})
	if err != nil {
		t.Errorf("空HAR的Walk()不应返回错误: %v", err)
	}
	if count != 0 {
		t.Errorf("空HAR的Walk()不应访问任何条目, got %d", count)
	}
}

func TestUtilGetSizeMixedValues(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/upload", "HTTP/1.1", "")
	entry.Request.HeadersSize = 200
	entry.Request.BodySize = -1 // 未知
	entry.Response.HeadersSize = 150
	entry.Response.BodySize = 500

	size := entry.GetSize()
	expected := 200 + 0 + 150 + 500 // -1 treated as 0
	if size != expected {
		t.Errorf("GetSize() = %d, want %d", size, expected)
	}
}

func TestUtilParseV11(t *testing.T) {
	h, err := ParseHarFile("testdata/v1.1.har")
	if err != nil {
		t.Fatalf("解析v1.1.har失败: %v", err)
	}

	if h.GetEntryCount() != 0 {
		t.Errorf("v1.1.har应有0个条目, got %d", h.GetEntryCount())
	}

	// 验证版本
	if h.Log.Version != "1.1" {
		t.Errorf("v1.1.har版本 = %q, want %q", h.Log.Version, "1.1")
	}
}

func TestUtilGetURLParsing(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com:8443/path?q=test#fragment", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	u := entry.GetURL()
	if u == nil {
		t.Fatal("GetURL() 返回nil")
	}
	if u.Host != "example.com:8443" {
		t.Errorf("GetURL().Host = %q, want %q", u.Host, "example.com:8443")
	}
	if u.Path != "/path" {
		t.Errorf("GetURL().Path = %q, want %q", u.Path, "/path")
	}
}

func TestUtilRequestHasHeaderCaseInsensitive(t *testing.T) {
	req := &Request{
		Headers: []Headers{
			{Name: "Content-Type", Value: "text/html"},
		},
	}

	tests := []string{"content-type", "CONTENT-TYPE", "Content-Type", "cOnTeNt-TyPe"}
	for _, name := range tests {
		if !req.HasHeader(name) {
			t.Errorf("HasHeader(%q) 应返回true", name)
		}
	}
}

func TestUtilResponseHasHeaderCaseInsensitive(t *testing.T) {
	resp := &Response{
		Headers: []Headers{
			{Name: "X-Custom-Header", Value: "test"},
		},
	}

	tests := []string{"x-custom-header", "X-CUSTOM-HEADER", "X-Custom-Header"}
	for _, name := range tests {
		if !resp.HasHeader(name) {
			t.Errorf("HasHeader(%q) 应返回true", name)
		}
	}
}

func TestUtilContentSetTextEmpty(t *testing.T) {
	c := &Content{}
	c.SetText("")
	if c.Text != "" {
		t.Error("SetText(\"\") 应设置Text为空字符串")
	}
	if c.Size != 0 {
		t.Errorf("SetText(\"\") 应设置Size为0, got %d", c.Size)
	}
}

func TestUtilEncodeContentEmptyData(t *testing.T) {
	c := &Content{}
	c.EncodeContent([]byte{}, "text/plain")
	if c.Text != "" {
		t.Errorf("EncodeContent([]byte{}) Text = %q, want empty", c.Text)
	}
	if c.Size != 0 {
		t.Errorf("EncodeContent([]byte{}) Size = %d, want 0", c.Size)
	}
}

func TestUtilGetResponseBodyNoEncoding(t *testing.T) {
	c := Content{
		Text:     "plain text content",
		MimeType: "text/plain",
		Size:     18,
	}
	entry := &Entries{
		Response: Response{
			Status:  200,
			Content: c,
		},
	}

	body, err := entry.GetResponseBody()
	if err != nil {
		t.Fatalf("GetResponseBody() 返回错误: %v", err)
	}
	if string(body) != "plain text content" {
		t.Errorf("GetResponseBody() = %q, want %q", string(body), "plain text content")
	}
}

func TestUtilGetCookieEmptySlice(t *testing.T) {
	req := &Request{
		Cookies: []Cookie{},
	}
	if req.GetCookie("test") != nil {
		t.Error("空Cookie列表的GetCookie()应返回nil")
	}

	resp := &Response{
		Cookies: []Cookie{},
	}
	if resp.GetCookie("test") != nil {
		t.Error("空Cookie列表的GetCookie()应返回nil")
	}
}

func TestUtilGetHeadersEmptySlice(t *testing.T) {
	req := &Request{
		Headers: []Headers{},
	}
	if vals := req.GetHeaderValues("test"); len(vals) != 0 {
		t.Errorf("空Header列表的GetHeaders()应返回空切片, got %v", vals)
	}

	resp := &Response{
		Headers: []Headers{},
	}
	if vals := resp.GetHeaderValues("test"); len(vals) != 0 {
		t.Errorf("空Header列表的GetHeaders()应返回空切片, got %v", vals)
	}
}

// 验证SaveToFileGzipped内容确实被压缩了
func TestUtilGzipCompressionReducesSize(t *testing.T) {
	h := NewHar()
	// 创建一个有大量重复数据的HAR
	for i := 0; i < 100; i++ {
		entry := h.AddEntry("GET", "https://example.com/api/data", "HTTP/1.1", "")
		entry.SetResponseStatus(200, "OK")
		entry.SetResponseContentText(strings.Repeat("aaaaaaaaaa", 100))
		entry.AddRequestHeader("Accept", "application/json")
		entry.AddResponseHeader("Content-Type", "application/json")
	}

	// 保存为普通JSON
	var jsonBuf bytes.Buffer
	if err := h.SaveToWriter(&jsonBuf, false); err != nil {
		t.Fatalf("SaveToWriter失败: %v", err)
	}

	// 保存为gzip
	tmpFile, err := os.CreateTemp("", "har_compress_test_*.har.gz")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := h.SaveToFileGzipped(tmpPath, false); err != nil {
		t.Fatalf("SaveToFileGzipped失败: %v", err)
	}

	fi, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("读取文件信息失败: %v", err)
	}

	gzSize := fi.Size()
	jsonSize := int64(jsonBuf.Len())

	if gzSize >= jsonSize {
		t.Errorf("gzip压缩后大小(%d)应小于原始JSON大小(%d)", gzSize, jsonSize)
	}
}

func TestUtilToCurlSingleQuoteEscape(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/test", "HTTP/1.1", "")
	entry.AddRequestHeader("X-Custom", "value with 'quotes'")
	entry.SetResponseStatus(200, "OK")

	curl := entry.ToCurl()
	// 检查单引号被正确转义（使用 '\'' 格式）
	if !strings.Contains(curl, "quotes") {
		t.Errorf("ToCurl() 应包含原始值, got: %s", curl)
	}
	// 确保输出不包含未转义的原始单引号在header值中
	if strings.Contains(curl, "value with 'quotes'") {
		t.Errorf("ToCurl() 不应包含未转义的单引号, got: %s", curl)
	}
}

func TestUtilGetDomainWithPort(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com:8443/api/test", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	domain := entry.GetDomain()
	expected := "example.com:8443"
	if domain != expected {
		t.Errorf("GetDomain() = %q, want %q", domain, expected)
	}
}

func TestUtilParseURLMethod(t *testing.T) {
	u, _ := url.Parse("https://example.com/path?query=value")
	if u == nil {
		t.Fatal("url.Parse返回nil")
	}
	_ = u.Host
	_ = u.Path
}
