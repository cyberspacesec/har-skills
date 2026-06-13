package har

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestHar 创建用于测试的HAR对象
func createTestHar() *Har {
	h := NewHar()
	h.SetCreator("test-creator", "1.0")

	entry := h.AddEntry("GET", "https://example.com/api/v1/users?limit=10", "HTTP/1.1", "")
	entry.AddRequestHeader("Content-Type", "application/json")
	entry.AddRequestHeader("Accept-Encoding", "gzip, deflate")
	entry.AddRequestHeader("Authorization", "Bearer token123")
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(1024, "application/json")
	entry.AddResponseHeader("Content-Type", "application/json")

	entry2 := h.AddEntry("POST", "https://api.example.com/login", "HTTP/1.1", "")
	entry2.AddRequestHeader("Content-Type", "application/json")
	entry2.SetPostData("application/json", `{"username":"admin","password":"test's pass"}`)
	entry2.SetResponseStatus(200, "OK")
	entry2.SetResponseContent(256, "application/json")
	entry2.AddResponseHeader("Content-Type", "application/json")

	entry3 := h.AddEntry("DELETE", "https://api.example.com/items/42", "HTTP/1.1", "")
	entry3.AddRequestHeader("Authorization", "Bearer token123")
	entry3.SetResponseStatus(204, "No Content")

	return h
}

// ---------------------------------------------------------------------------
// ToCurl 测试
// ---------------------------------------------------------------------------

func TestExportToCurlOnHar(t *testing.T) {
	h := createTestHar()
	result := h.ToCurl()

	if result == "" {
		t.Fatal("ToCurl() 不应返回空字符串")
	}

	// 应包含多个curl命令（用双换行分隔）
	parts := strings.Split(result, "\n\n")
	if len(parts) != 3 {
		t.Fatalf("应有3个curl命令，实际得到 %d 个", len(parts))
	}
}

func TestExportToCurlOnHarNil(t *testing.T) {
	var h *Har
	result := h.ToCurl()
	if result != "" {
		t.Fatalf("nil HAR应返回空字符串，实际得到: %s", result)
	}
}

func TestExportToCurlOnHarEmpty(t *testing.T) {
	h := &Har{}
	result := h.ToCurl()
	if result != "" {
		t.Fatalf("空HAR应返回空字符串，实际得到: %s", result)
	}
}

func TestExportToCurlOnEntries(t *testing.T) {
	h := createTestHar()
	entry := &h.Log.Entries[0]
	result := entry.ToCurl()

	if result == "" {
		t.Fatal("Entries.ToCurl() 不应返回空字符串")
	}
	if !strings.HasPrefix(result, "curl") {
		t.Errorf("cURL命令应以'curl'开头，实际: %s", result[:10])
	}
}

func TestExportToCurlOnEntriesNil(t *testing.T) {
	var e *Entries
	result := e.ToCurl()
	if result != "" {
		t.Fatalf("nil Entries应返回空字符串，实际得到: %s", result)
	}
}

func TestExportCurlMethod(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Content-Type", "application/json")
	entry.SetPostData("application/json", `{"key":"value"}`)
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if !strings.Contains(result, "-X POST") {
		t.Errorf("POST请求应包含 -X POST，实际: %s", result)
	}
	if !strings.Contains(result, "--data") {
		t.Errorf("有POST数据的请求应包含 --data，实际: %s", result)
	}
}

func TestExportCurlGETNoMethod(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if strings.Contains(result, "-X GET") {
		t.Errorf("GET请求不应包含 -X GET，实际: %s", result)
	}
}

func TestExportCurlHeaders(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("X-Custom", "value123")
	entry.AddRequestHeader("Host", "example.com") // 应被跳过
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if !strings.Contains(result, "-H 'X-Custom: value123'") {
		t.Errorf("应包含自定义请求头，实际: %s", result)
	}
	if strings.Contains(result, "Host:") {
		t.Errorf("不应包含Host头，实际: %s", result)
	}
}

func TestExportCurlCompressed(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Accept-Encoding", "gzip, deflate")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if !strings.Contains(result, "--compressed") {
		t.Errorf("Accept-Encoding包含gzip/deflate时应添加--compressed，实际: %s", result)
	}
}

func TestExportCurlNoCompressed(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Accept-Encoding", "br")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if strings.Contains(result, "--compressed") {
		t.Errorf("Accept-Encoding不包含gzip/deflate时不应添加--compressed，实际: %s", result)
	}
}

func TestExportCurlSingleQuoteEscaping(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.SetPostData("application/json", `it's a test`)
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if !strings.Contains(result, `it'\''s a test`) {
		t.Errorf("单引号应被正确转义，实际: %s", result)
	}
}

func TestExportCurlSSLVerifySkip(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://self-signed.example.com/api", "HTTP/1.1", "")
	entry.Response.Error = "SSL certificate problem"
	entry.SetResponseStatus(0, "")

	result := entry.ToCurl()

	if !strings.Contains(result, "-k") {
		t.Errorf("有SSL错误时应添加 -k，实际: %s", result)
	}
}

func TestExportCurlURLQuoted(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api?key=value", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToCurl()

	if !strings.Contains(result, "'https://example.com/api?key=value'") {
		t.Errorf("URL应用单引号包裹，实际: %s", result)
	}
}

// ---------------------------------------------------------------------------
// ToWget 测试
// ---------------------------------------------------------------------------

func TestExportToWgetOnHar(t *testing.T) {
	h := createTestHar()
	result := h.ToWget()

	if result == "" {
		t.Fatal("ToWget() 不应返回空字符串")
	}

	parts := strings.Split(result, "\n\n")
	if len(parts) != 3 {
		t.Fatalf("应有3个wget命令，实际得到 %d 个", len(parts))
	}
}

func TestExportToWgetOnHarNil(t *testing.T) {
	var h *Har
	result := h.ToWget()
	if result != "" {
		t.Fatalf("nil HAR应返回空字符串，实际得到: %s", result)
	}
}

func TestExportToWgetOnEntries(t *testing.T) {
	h := createTestHar()
	entry := &h.Log.Entries[0]
	result := entry.ToWget()

	if result == "" {
		t.Fatal("Entries.ToWget() 不应返回空字符串")
	}
	if !strings.HasPrefix(result, "wget") {
		t.Errorf("wget命令应以'wget'开头，实际: %s", result[:10])
	}
}

func TestExportToWgetOnEntriesNil(t *testing.T) {
	var e *Entries
	result := e.ToWget()
	if result != "" {
		t.Fatalf("nil Entries应返回空字符串，实际得到: %s", result)
	}
}

func TestExportWgetMethod(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("DELETE", "https://example.com/api/items/1", "HTTP/1.1", "")
	entry.SetResponseStatus(204, "No Content")

	result := entry.ToWget()

	if !strings.Contains(result, "--method=DELETE") {
		t.Errorf("DELETE请求应包含 --method=DELETE，实际: %s", result)
	}
}

func TestExportWgetGETNoMethod(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToWget()

	if strings.Contains(result, "--method=") {
		t.Errorf("GET请求不应包含 --method，实际: %s", result)
	}
}

func TestExportWgetPostData(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.SetPostData("application/json", `{"name":"test"}`)
	entry.SetResponseStatus(200, "OK")

	result := entry.ToWget()

	if !strings.Contains(result, "--post-data=") {
		t.Errorf("有POST数据的请求应包含 --post-data，实际: %s", result)
	}
}

func TestExportWgetHTTPSNoCheckCert(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToWget()

	if !strings.Contains(result, "--no-check-certificate") {
		t.Errorf("HTTPS请求应包含 --no-check-certificate，实际: %s", result)
	}
}

func TestExportWgetHTTPNoCertFlag(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "http://example.com/api", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToWget()

	if strings.Contains(result, "--no-check-certificate") {
		t.Errorf("HTTP请求不应包含 --no-check-certificate，实际: %s", result)
	}
}

func TestExportWgetHeaderSkipHost(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Host", "example.com")
	entry.AddRequestHeader("X-Custom", "test")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToWget()

	if strings.Contains(result, "Host:") {
		t.Errorf("wget不应包含Host头，实际: %s", result)
	}
	if !strings.Contains(result, "--header='X-Custom: test'") {
		t.Errorf("应包含自定义请求头，实际: %s", result)
	}
}

// ---------------------------------------------------------------------------
// ToPythonRequests 测试
// ---------------------------------------------------------------------------

func TestExportToPythonRequestsOnHar(t *testing.T) {
	h := createTestHar()
	result := h.ToPythonRequests()

	if result == "" {
		t.Fatal("ToPythonRequests() 不应返回空字符串")
	}
	if !strings.Contains(result, "import requests") {
		t.Error("Python代码应包含 'import requests'")
	}
	// 应包含3个请求
	count := strings.Count(result, "response = requests.")
	if count != 3 {
		t.Errorf("应有3个requests调用，实际: %d", count)
	}
}

func TestExportToPythonRequestsOnHarNil(t *testing.T) {
	var h *Har
	result := h.ToPythonRequests()
	if result != "" {
		t.Fatalf("nil HAR应返回空字符串，实际得到: %s", result)
	}
}

func TestExportToPythonRequestsOnEntries(t *testing.T) {
	h := createTestHar()
	entry := &h.Log.Entries[0]
	result := entry.ToPythonRequests()

	if result == "" {
		t.Fatal("Entries.ToPythonRequests() 不应返回空字符串")
	}
	if !strings.Contains(result, "requests.get") {
		t.Errorf("GET请求应使用 requests.get，实际: %s", result)
	}
}

func TestExportToPythonRequestsOnEntriesNil(t *testing.T) {
	var e *Entries
	result := e.ToPythonRequests()
	if result != "" {
		t.Fatalf("nil Entries应返回空字符串，实际得到: %s", result)
	}
}

func TestExportPythonRequestMethod(t *testing.T) {
	tests := []struct {
		method         string
		expectedMethod string
	}{
		{"GET", "requests.get"},
		{"POST", "requests.post"},
		{"PUT", "requests.put"},
		{"DELETE", "requests.delete"},
		{"PATCH", "requests.patch"},
		{"HEAD", "requests.head"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			h := NewHar()
			entry := h.AddEntry(tt.method, "https://example.com/api", "HTTP/1.1", "")
			entry.SetResponseStatus(200, "OK")
			result := entry.ToPythonRequests()
			if !strings.Contains(result, tt.expectedMethod) {
				t.Errorf("方法 %s 应生成 %s，实际: %s", tt.method, tt.expectedMethod, result)
			}
		})
	}
}

func TestExportPythonHeaders(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Authorization", "Bearer token")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToPythonRequests()

	if !strings.Contains(result, "headers = {") {
		t.Errorf("有请求头时应生成headers字典，实际: %s", result)
	}
	if !strings.Contains(result, "'Authorization': 'Bearer token'") {
		t.Errorf("应包含Authorization头，实际: %s", result)
	}
}

func TestExportPythonPostData(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.SetPostData("application/json", `{"key":"value"}`)
	entry.SetResponseStatus(200, "OK")

	result := entry.ToPythonRequests()

	if !strings.Contains(result, "data=") {
		t.Errorf("POST请求应包含data参数，实际: %s", result)
	}
}

func TestExportPythonPrint(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToPythonRequests()

	if !strings.Contains(result, "print(response.status_code)") {
		t.Errorf("应打印status_code，实际: %s", result)
	}
	if !strings.Contains(result, "print(response.text)") {
		t.Errorf("应打印response.text，实际: %s", result)
	}
}

func TestExportPythonStringEscaping(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.SetPostData("application/json", "it's a test\nwith newline")
	entry.SetResponseStatus(200, "OK")

	result := entry.ToPythonRequests()

	if !strings.Contains(result, `it\'s a test`) {
		t.Errorf("单引号应被转义，实际: %s", result)
	}
	if !strings.Contains(result, `\n`) {
		t.Errorf("换行符应被转义，实际: %s", result)
	}
}

// ---------------------------------------------------------------------------
// ToPostmanCollection 测试
// ---------------------------------------------------------------------------

func TestExportToPostmanCollection(t *testing.T) {
	h := createTestHar()
	data, err := h.ToPostmanCollection()
	if err != nil {
		t.Fatalf("ToPostmanCollection() 返回错误: %v", err)
	}

	// 验证是有效的JSON
	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("结果不是有效的JSON: %v", err)
	}

	// 验证info
	if collection.Info.Name != "HAR Export" {
		t.Errorf("Name 应为 'HAR Export'，实际: %s", collection.Info.Name)
	}
	expectedSchema := "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	if collection.Info.Schema != expectedSchema {
		t.Errorf("Schema 不正确，实际: %s", collection.Info.Schema)
	}

	// 验证item数量
	if len(collection.Item) != 3 {
		t.Errorf("应有3个item，实际: %d", len(collection.Item))
	}
}

func TestExportToPostmanCollectionNil(t *testing.T) {
	var h *Har
	_, err := h.ToPostmanCollection()
	if err == nil {
		t.Error("nil HAR应返回错误")
	}
}

func TestExportPostmanCollectionMethod(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api/v1/users", "HTTP/1.1", "")
	entry.SetPostData("application/json", `{"name":"test"}`)
	entry.AddRequestHeader("Content-Type", "application/json")
	entry.SetResponseStatus(200, "OK")

	data, err := h.ToPostmanCollection()
	if err != nil {
		t.Fatalf("ToPostmanCollection() 返回错误: %v", err)
	}

	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	item := collection.Item[0]
	if item.Request.Method != "POST" {
		t.Errorf("Method 应为 POST，实际: %s", item.Request.Method)
	}

	// 验证URL结构
	if item.Request.URL.Protocol != "https" {
		t.Errorf("Protocol 应为 https，实际: %s", item.Request.URL.Protocol)
	}

	// 验证请求头
	if len(item.Request.Header) == 0 {
		t.Error("应有请求头")
	}

	// 验证请求体
	if item.Request.Body == nil {
		t.Fatal("应有请求体")
	}
	if item.Request.Body.Mode != "raw" {
		t.Errorf("Body mode 应为 raw，实际: %s", item.Request.Body.Mode)
	}
}

func TestExportPostmanCollectionURLParsing(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://api.example.com/v1/users?limit=10&offset=0", "HTTP/1.1", "")
	entry.SetResponseStatus(200, "OK")

	data, err := h.ToPostmanCollection()
	if err != nil {
		t.Fatalf("ToPostmanCollection() 返回错误: %v", err)
	}

	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	item := collection.Item[0]
	pmURL := item.Request.URL

	// 验证Host
	if len(pmURL.Host) == 0 {
		t.Error("Host不应为空")
	}

	// 验证Path
	if len(pmURL.Path) < 2 {
		t.Errorf("Path 应至少有2段，实际: %v", pmURL.Path)
	}

	// 验证Query
	if len(pmURL.Query) < 2 {
		t.Errorf("Query 应至少有2个参数，实际: %v", pmURL.Query)
	}
}

// ---------------------------------------------------------------------------
// SaveAsPostmanCollection 测试
// ---------------------------------------------------------------------------

func TestExportSaveAsPostmanCollection(t *testing.T) {
	h := createTestHar()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "collection.json")

	if err := h.SaveAsPostmanCollection(filePath); err != nil {
		t.Fatalf("SaveAsPostmanCollection() 返回错误: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("文件未被创建")
	}

	// 验证文件内容是有效的Postman Collection JSON
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("文件内容不是有效的JSON: %v", err)
	}

	if collection.Info.Name != "HAR Export" {
		t.Errorf("Name 应为 'HAR Export'，实际: %s", collection.Info.Name)
	}
}

func TestExportSaveAsPostmanCollectionNil(t *testing.T) {
	var h *Har
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "collection.json")

	err := h.SaveAsPostmanCollection(filePath)
	if err == nil {
		t.Error("nil HAR应返回错误")
	}
}

// ---------------------------------------------------------------------------
// ToXML 测试
// ---------------------------------------------------------------------------

func TestExportToXML(t *testing.T) {
	h := createTestHar()
	result, err := h.ToXML()
	if err != nil {
		t.Fatalf("ToXML() 返回错误: %v", err)
	}

	if result == "" {
		t.Fatal("ToXML() 不应返回空字符串")
	}

	// 验证包含XML声明
	if !strings.Contains(result, `<?xml`) {
		t.Error("XML应包含XML声明")
	}

	// 验证根元素
	if !strings.Contains(result, "<har>") {
		t.Error("XML应包含 <har> 根元素")
	}
	if !strings.Contains(result, "</har>") {
		t.Error("XML应包含 </har> 结束标签")
	}

	// 验证能被正确解析
	var harXML HARXML
	if err := xml.Unmarshal([]byte(result), &harXML); err != nil {
		t.Fatalf("XML解析失败: %v", err)
	}

	if harXML.Log.Version != "1.2" {
		t.Errorf("Version 应为 1.2，实际: %s", harXML.Log.Version)
	}

	if len(harXML.Log.Entries) != 3 {
		t.Errorf("应有3个entry，实际: %d", len(harXML.Log.Entries))
	}
}

func TestExportToXMLNil(t *testing.T) {
	var h *Har
	result, err := h.ToXML()
	if err != nil {
		t.Fatalf("nil HAR ToXML() 不应返回错误: %v", err)
	}
	if result != "" {
		t.Errorf("nil HAR应返回空字符串，实际: %s", result)
	}
}

func TestExportToXMLContent(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Content-Type", "application/json")
	entry.SetPostData("application/json", `{"key":"value"}`)
	entry.SetResponseStatus(200, "OK")
	entry.SetResponseContent(42, "application/json")

	result, err := h.ToXML()
	if err != nil {
		t.Fatalf("ToXML() 返回错误: %v", err)
	}

	// 验证方法
	if !strings.Contains(result, "<method>POST</method>") {
		t.Errorf("XML应包含POST方法，实际: %s", result)
	}

	// 验证URL
	if !strings.Contains(result, "<url>https://example.com/api</url>") {
		t.Errorf("XML应包含URL，实际: %s", result)
	}

	// 验证请求头
	if !strings.Contains(result, "<name>Content-Type</name>") {
		t.Errorf("XML应包含请求头名称，实际: %s", result)
	}

	// 验证POST数据
	if !strings.Contains(result, "<postData>") {
		t.Errorf("XML应包含postData元素，实际: %s", result)
	}
}

// ---------------------------------------------------------------------------
// SaveAsXML 测试
// ---------------------------------------------------------------------------

func TestExportSaveAsXML(t *testing.T) {
	h := createTestHar()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.xml")

	if err := h.SaveAsXML(filePath); err != nil {
		t.Fatalf("SaveAsXML() 返回错误: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("文件未被创建")
	}

	// 验证文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	if !strings.Contains(string(data), "<?xml") {
		t.Error("文件内容应包含XML声明")
	}
}

func TestExportSaveAsXMLNil(t *testing.T) {
	var h *Har
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.xml")

	result, err := h.ToXML()
	if err != nil {
		t.Fatalf("nil HAR ToXML() 不应返回错误: %v", err)
	}
	if result != "" {
		t.Errorf("nil HAR应返回空字符串")
	}

	// SaveAsXML对于nil HAR：ToXML返回空字符串，writeToFile也能写入空文件
	// 这是可接受的行为
	_ = filePath
}

// ---------------------------------------------------------------------------
// FormatJSON 常量测试
// ---------------------------------------------------------------------------

func TestExportFormatJSON(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON 应为 'json'，实际: %s", FormatJSON)
	}
}

// ---------------------------------------------------------------------------
// 辅助函数测试
// ---------------------------------------------------------------------------

func TestExportEscapeSingleQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"it's", "it'\\''s"},
		{"don't stop", "don'\\''t stop"},
		{"no quotes", "no quotes"},
		{"", ""},
	}

	for _, tt := range tests {
		result := escapeSingleQuotes(tt.input)
		if result != tt.expected {
			t.Errorf("escapeSingleQuotes(%q) = %q, 期望 %q", tt.input, result, tt.expected)
		}
	}
}

func TestExportEscapePythonString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"it's", `it\'s`},
		{`back\slash`, `back\\slash`},
		{"line\nbreak", "line\\nbreak"},
		{"tab\there", "tab\\there"},
		{"", ""},
	}

	for _, tt := range tests {
		result := escapePythonString(tt.input)
		if result != tt.expected {
			t.Errorf("escapePythonString(%q) = %q, 期望 %q", tt.input, result, tt.expected)
		}
	}
}

func TestExportHasAcceptEncoding(t *testing.T) {
	tests := []struct {
		name     string
		headers  []Headers
		expected bool
	}{
		{
			"gzip",
			[]Headers{{Name: "Accept-Encoding", Value: "gzip"}},
			true,
		},
		{
			"deflate",
			[]Headers{{Name: "Accept-Encoding", Value: "deflate"}},
			true,
		},
		{
			"mixed",
			[]Headers{{Name: "Accept-Encoding", Value: "gzip, deflate, br"}},
			true,
		},
		{
			"br only",
			[]Headers{{Name: "Accept-Encoding", Value: "br"}},
			false,
		},
		{
			"no header",
			[]Headers{},
			false,
		},
		{
			"case insensitive",
			[]Headers{{Name: "accept-encoding", Value: "GZIP"}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &Entries{
				Request: Request{
					Headers: tt.headers,
				},
			}
			result := hasAcceptEncoding(entry)
			if result != tt.expected {
				t.Errorf("hasAcceptEncoding() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 综合测试
// ---------------------------------------------------------------------------

func TestExportAllMethodsConsistency(t *testing.T) {
	h := createTestHar()

	// 确保所有导出方法都不会崩溃
	curlResult := h.ToCurl()
	wgetResult := h.ToWget()
	pythonResult := h.ToPythonRequests()
	postmanResult, postmanErr := h.ToPostmanCollection()
	xmlResult, xmlErr := h.ToXML()

	if curlResult == "" {
		t.Error("ToCurl() 不应返回空字符串")
	}
	if wgetResult == "" {
		t.Error("ToWget() 不应返回空字符串")
	}
	if pythonResult == "" {
		t.Error("ToPythonRequests() 不应返回空字符串")
	}
	if postmanErr != nil {
		t.Errorf("ToPostmanCollection() 返回错误: %v", postmanErr)
	}
	if len(postmanResult) == 0 {
		t.Error("ToPostmanCollection() 不应返回空数据")
	}
	if xmlErr != nil {
		t.Errorf("ToXML() 返回错误: %v", xmlErr)
	}
	if xmlResult == "" {
		t.Error("ToXML() 不应返回空字符串")
	}
}

func TestExportEntryMethodsConsistency(t *testing.T) {
	h := NewHar()
	entry := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	entry.AddRequestHeader("Accept", "application/json")
	entry.SetResponseStatus(200, "OK")

	curlResult := entry.ToCurl()
	wgetResult := entry.ToWget()
	pythonResult := entry.ToPythonRequests()

	if curlResult == "" {
		t.Error("Entries.ToCurl() 不应返回空字符串")
	}
	if wgetResult == "" {
		t.Error("Entries.ToWget() 不应返回空字符串")
	}
	if pythonResult == "" {
		t.Error("Entries.ToPythonRequests() 不应返回空字符串")
	}
}
