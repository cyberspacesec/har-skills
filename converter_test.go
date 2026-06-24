package har

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// --- Comprehensive branch coverage tests for converter.go ---

func createTestHarForConvert() *Har {
	h := NewHar()

	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContentText("response body")
	e1.Response.Content.MimeType = "application/json"
	e1.Response.Content.Size = 100
	e1.Time = 150.5
	e1.Timings = Timings{
		Blocked: 10.0,
		DNS:     5.0,
		Connect: 20.0,
		Send:    1.0,
		Wait:    100.0,
		Receive: 14.5,
		Ssl:     15.0,
	}

	e2 := h.AddEntry("POST", "https://example.com/api/items", "HTTP/1.1", "")
	e2.SetResponseStatus(201, "Created")
	e2.SetResponseContentText("created")
	e2.Response.Content.MimeType = "text/html"
	e2.Response.Content.Size = 50
	e2.Time = 200.0
	e2.Timings = Timings{
		Blocked: 5.0,
		DNS:     2.0,
		Connect: 10.0,
		Send:    2.0,
		Wait:    150.0,
		Receive: 31.0,
		Ssl:     8.0,
	}

	return h
}

func TestConvertUnsupportedFormat(t *testing.T) {
	// Branch: default case in Convert switch
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()

	_, err := h.Convert("xml", opts)
	assertHarErrorCode(t, err, ErrCodeUnsupported)
	if !strings.Contains(err.Error(), "不支持") && !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("Error should mention unsupported format, got: %s", err.Error())
	}
}

func TestConvertNilHAR(t *testing.T) {
	var h *Har

	result, err := h.Convert(FormatCSV, DefaultConvertOptions())
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if result != "" {
		t.Errorf("Expected empty result for nil HAR, got %q", result)
	}
}

func TestConvertWithFilter(t *testing.T) {
	// Branch: options.Filter != nil
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()
	opts.Filter = &FilterOptions{
		Method: "GET",
	}

	result, err := h.Convert(FormatCSV, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only contain the GET request, not the POST
	if strings.Contains(result, "POST") {
		t.Error("Filtered result should not contain POST method")
	}
	if !strings.Contains(result, "GET") {
		t.Error("Filtered result should contain GET method")
	}
}

func TestConvertCSVEmptyEntries(t *testing.T) {
	// Branch: convertToCSV with no entries
	h := NewHar()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatCSV, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have header but no data rows
	if !strings.Contains(result, "方法") && !strings.Contains(result, "URL") {
		t.Error("CSV should contain header row")
	}
	// Count lines - should be just header
	lines := strings.Count(result, "\n")
	if lines < 1 {
		t.Error("CSV should have at least a header line")
	}
}

func TestConvertCSVWithEntries(t *testing.T) {
	// Branch: convertToCSV with entries
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatCSV, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "GET") {
		t.Error("CSV should contain GET method")
	}
	if !strings.Contains(result, "POST") {
		t.Error("CSV should contain POST method")
	}
}

func TestConvertMarkdownWithEntries(t *testing.T) {
	// Branch: convertToMarkdown with entries
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatMarkdown, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "|") {
		t.Error("Markdown should contain table pipes")
	}
	if !strings.Contains(result, "---") {
		t.Error("Markdown should contain separator row")
	}
}

func TestConvertMarkdownEmptyEntries(t *testing.T) {
	// Branch: convertToMarkdown with no entries
	h := NewHar()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatMarkdown, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "|") {
		t.Error("Markdown should contain header pipes")
	}
}

func TestConvertHTMLWithEntries(t *testing.T) {
	// Branch: convertToHTML with entries
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatHTML, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "<table") {
		t.Error("HTML should contain table tag")
	}
	if !strings.Contains(result, "<th>") {
		t.Error("HTML should contain header cells")
	}
	if !strings.Contains(result, "<td>") {
		t.Error("HTML should contain data cells")
	}
}

func TestConvertHTMLEmptyEntries(t *testing.T) {
	// Branch: convertToHTML with no entries
	h := NewHar()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatHTML, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "<table") {
		t.Error("HTML should contain table tag")
	}
	if !strings.Contains(result, "<thead>") {
		t.Error("HTML should contain thead")
	}
}

func TestConvertTextWithEntries(t *testing.T) {
	// Branch: convertToText with entries
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatText, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "GET") {
		t.Error("Text should contain GET method")
	}
	if !strings.Contains(result, "---") {
		t.Error("Text should contain separator line")
	}
}

func TestConvertTextEmptyEntries(t *testing.T) {
	// Branch: convertToText with no entries
	h := NewHar()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatText, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have header and separator
	if !strings.Contains(result, "---") {
		t.Error("Text should contain separator line")
	}
}

func TestGetHeadersCustomHeaders(t *testing.T) {
	// Branch: len(options.Headers) > 0 returns custom headers
	opts := ConvertOptions{
		Headers: []string{"Custom1", "Custom2", "Custom3"},
	}

	headers := getHeaders(opts)
	if len(headers) != 3 {
		t.Fatalf("Expected 3 custom headers, got %d", len(headers))
	}
	if headers[0] != "Custom1" {
		t.Errorf("Expected 'Custom1', got '%s'", headers[0])
	}
	if headers[1] != "Custom2" {
		t.Errorf("Expected 'Custom2', got '%s'", headers[1])
	}
	if headers[2] != "Custom3" {
		t.Errorf("Expected 'Custom3', got '%s'", headers[2])
	}
}

func TestGetHeadersDefaultHeaders(t *testing.T) {
	// Branch: default headers from include flags
	opts := DefaultConvertOptions()

	headers := getHeaders(opts)
	if len(headers) == 0 {
		t.Fatal("Expected non-empty default headers")
	}

	// Default options include: URL, Method, Status, ContentType, Size, Time, DateTime
	expectedHeaders := []string{"日期时间", "方法", "URL", "状态码", "内容类型", "大小(字节)", "时间(ms)"}
	for i, expected := range expectedHeaders {
		if i < len(headers) && headers[i] != expected {
			t.Errorf("Header[%d]: expected '%s', got '%s'", i, expected, headers[i])
		}
	}
}

func TestGetHeadersWithTimings(t *testing.T) {
	// Branch: IncludeTimings adds timing sub-headers
	opts := ConvertOptions{
		IncludeTimings: true,
	}

	headers := getHeaders(opts)
	timingHeaders := []string{"阻塞(ms)", "DNS(ms)", "连接(ms)", "发送(ms)", "等待(ms)", "接收(ms)"}
	for _, th := range timingHeaders {
		found := false
		for _, h := range headers {
			if h == th {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected timing header '%s' not found in %v", th, headers)
		}
	}
}

func TestGetHeadersWithPostData(t *testing.T) {
	// Branch: IncludePostData adds POST data headers
	opts := ConvertOptions{
		IncludePostData: true,
	}

	headers := getHeaders(opts)
	postHeaders := []string{"POST数据类型", "POST数据"}
	for _, ph := range postHeaders {
		found := false
		for _, h := range headers {
			if h == ph {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected POST header '%s' not found in %v", ph, headers)
		}
	}
}

func TestGetHeadersWithQueryString(t *testing.T) {
	// Branch: IncludeQueryString adds query parameter header
	opts := ConvertOptions{
		IncludeQueryString: true,
	}

	headers := getHeaders(opts)
	found := false
	for _, h := range headers {
		if h == "查询参数" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected '查询参数' header not found in %v", headers)
	}
}

func TestGetHeadersWithHeaders(t *testing.T) {
	// Branch: IncludeHeaders adds request/response header columns
	opts := ConvertOptions{
		IncludeHeaders: true,
	}

	headers := getHeaders(opts)
	headerCols := []string{"请求头", "响应头"}
	for _, hc := range headerCols {
		found := false
		for _, h := range headers {
			if h == hc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected header column '%s' not found in %v", hc, headers)
		}
	}
}

func TestGetHeadersEmptyOptions(t *testing.T) {
	// Branch: no include flags set, returns empty headers
	opts := ConvertOptions{}

	headers := getHeaders(opts)
	if len(headers) != 0 {
		t.Errorf("Expected empty headers for empty options, got %v", headers)
	}
}

func TestCreateDataRowWithPostData(t *testing.T) {
	// Branch: IncludePostData with PostData present
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api",
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"key": "value"}`,
			},
		},
		Response: Response{
			Status:     201,
			StatusText: "Created",
			Content: Content{
				MimeType: "text/html",
				Size:     50,
			},
		},
		Time: 100.0,
	}

	opts := ConvertOptions{
		IncludePostData: true,
	}

	row := createDataRow(entry, opts)
	// Should have mimeType and text from PostData
	if len(row) < 2 {
		t.Fatalf("Expected at least 2 cells, got %d", len(row))
	}
	if row[0] != "application/json" {
		t.Errorf("Expected 'application/json', got '%s'", row[0])
	}
	if row[1] != `{"key": "value"}` {
		t.Errorf("Expected POST data text, got '%s'", row[1])
	}
}

func TestCreateDataRowWithPostDataNil(t *testing.T) {
	// Branch: IncludePostData with nil PostData
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method:   "GET",
			URL:      "https://example.com/api",
			PostData: nil, // nil PostData
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				MimeType: "text/html",
				Size:     100,
			},
		},
		Time: 50.0,
	}

	opts := ConvertOptions{
		IncludePostData: true,
	}

	row := createDataRow(entry, opts)
	// Should have two empty strings for nil PostData
	if len(row) < 2 {
		t.Fatalf("Expected at least 2 cells, got %d", len(row))
	}
	if row[0] != "" {
		t.Errorf("Expected empty string for nil PostData mimeType, got '%s'", row[0])
	}
	if row[1] != "" {
		t.Errorf("Expected empty string for nil PostData text, got '%s'", row[1])
	}
}

func TestCreateDataRowWithQueryString(t *testing.T) {
	// Branch: IncludeQueryString with query params
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/api?foo=bar&baz=qux",
			QueryString: []QueryString{
				{Name: "foo", Value: "bar"},
				{Name: "baz", Value: "qux"},
			},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				MimeType: "text/html",
				Size:     100,
			},
		},
		Time: 50.0,
	}

	opts := ConvertOptions{
		IncludeQueryString: true,
	}

	row := createDataRow(entry, opts)
	if len(row) < 1 {
		t.Fatalf("Expected at least 1 cell, got %d", len(row))
	}
	qs := row[0]
	if !strings.Contains(qs, "foo=bar") {
		t.Errorf("Expected query string to contain 'foo=bar', got '%s'", qs)
	}
	if !strings.Contains(qs, "baz=qux") {
		t.Errorf("Expected query string to contain 'baz=qux', got '%s'", qs)
	}
}

func TestCreateDataRowWithQueryStringEmpty(t *testing.T) {
	// Branch: IncludeQueryString with no query params
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method:      "GET",
			URL:         "https://example.com/api",
			QueryString: []QueryString{},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				MimeType: "text/html",
				Size:     100,
			},
		},
		Time: 50.0,
	}

	opts := ConvertOptions{
		IncludeQueryString: true,
	}

	row := createDataRow(entry, opts)
	if len(row) < 1 {
		t.Fatalf("Expected at least 1 cell, got %d", len(row))
	}
	if row[0] != "" {
		t.Errorf("Expected empty string for no query params, got '%s'", row[0])
	}
}

func TestCreateDataRowWithHeaders(t *testing.T) {
	// Branch: IncludeHeaders with request and response headers
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/api",
			Headers: []Headers{
				{Name: "Accept", Value: "application/json"},
				{Name: "Authorization", Value: "Bearer token"},
			},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Headers: []Headers{
				{Name: "Content-Type", Value: "application/json"},
				{Name: "Cache-Control", Value: "no-cache"},
			},
			Content: Content{
				MimeType: "application/json",
				Size:     100,
			},
		},
		Time: 50.0,
	}

	opts := ConvertOptions{
		IncludeHeaders: true,
	}

	row := createDataRow(entry, opts)
	if len(row) < 2 {
		t.Fatalf("Expected at least 2 cells for headers, got %d", len(row))
	}

	// First cell: request headers joined with "; "
	if !strings.Contains(row[0], "Accept: application/json") {
		t.Errorf("Expected request header 'Accept: application/json' in '%s'", row[0])
	}
	if !strings.Contains(row[0], "Authorization: Bearer token") {
		t.Errorf("Expected request header 'Authorization: Bearer token' in '%s'", row[0])
	}

	// Second cell: response headers joined with "; "
	if !strings.Contains(row[1], "Content-Type: application/json") {
		t.Errorf("Expected response header 'Content-Type: application/json' in '%s'", row[1])
	}
	if !strings.Contains(row[1], "Cache-Control: no-cache") {
		t.Errorf("Expected response header 'Cache-Control: no-cache' in '%s'", row[1])
	}
}

func TestCreateDataRowWithEmptyHeaders(t *testing.T) {
	// Branch: IncludeHeaders with empty header lists
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method:  "GET",
			URL:     "https://example.com/api",
			Headers: []Headers{},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Headers:    []Headers{},
			Content: Content{
				MimeType: "text/html",
				Size:     100,
			},
		},
		Time: 50.0,
	}

	opts := ConvertOptions{
		IncludeHeaders: true,
	}

	row := createDataRow(entry, opts)
	if len(row) < 2 {
		t.Fatalf("Expected at least 2 cells for headers, got %d", len(row))
	}
	if row[0] != "" {
		t.Errorf("Expected empty string for no request headers, got '%s'", row[0])
	}
	if row[1] != "" {
		t.Errorf("Expected empty string for no response headers, got '%s'", row[1])
	}
}

func TestCreateDataRowWithDateTime(t *testing.T) {
	// Branch: IncludeDateTime
	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	entry := Entries{
		StartedDateTime: now,
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/api",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				MimeType: "text/html",
				Size:     100,
			},
		},
		Time: 50.0,
	}

	opts := ConvertOptions{
		IncludeDateTime: true,
	}

	row := createDataRow(entry, opts)
	if len(row) < 1 {
		t.Fatalf("Expected at least 1 cell, got %d", len(row))
	}
	if !strings.Contains(row[0], "2024-06-15") {
		t.Errorf("Expected date in datetime cell, got '%s'", row[0])
	}
}

func TestCreateDataRowWithTimings(t *testing.T) {
	// Branch: IncludeTimings
	entry := Entries{
		StartedDateTime: time.Now(),
		Request: Request{
			Method: "GET",
			URL:    "https://example.com/api",
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Content: Content{
				MimeType: "text/html",
				Size:     100,
			},
		},
		Time: 150.0,
		Timings: Timings{
			Blocked: 10.0,
			DNS:     5.0,
			Connect: 20.0,
			Send:    1.0,
			Wait:    100.0,
			Receive: 14.0,
		},
	}

	opts := ConvertOptions{
		IncludeTimings: true,
	}

	row := createDataRow(entry, opts)
	if len(row) < 6 {
		t.Fatalf("Expected at least 6 timing cells, got %d", len(row))
	}

	// Check the timing values are formatted as "%.2f"
	expectedTimings := []string{"10.00", "5.00", "20.00", "1.00", "100.00", "14.00"}
	for i, expected := range expectedTimings {
		if row[i] != expected {
			t.Errorf("Timing[%d]: expected '%s', got '%s'", i, expected, row[i])
		}
	}
}

func TestCreateDataRowAllFields(t *testing.T) {
	// Branch: all include flags set
	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	entry := Entries{
		StartedDateTime: now,
		Request: Request{
			Method: "POST",
			URL:    "https://example.com/api?foo=bar",
			Headers: []Headers{
				{Name: "Accept", Value: "application/json"},
			},
			QueryString: []QueryString{
				{Name: "foo", Value: "bar"},
			},
			PostData: &PostData{
				MimeType: "application/json",
				Text:     `{"key":"value"}`,
			},
		},
		Response: Response{
			Status:     200,
			StatusText: "OK",
			Headers: []Headers{
				{Name: "Content-Type", Value: "application/json"},
			},
			Content: Content{
				MimeType: "application/json",
				Size:     256,
			},
		},
		Time: 150.0,
		Timings: Timings{
			Blocked: 10.0,
			DNS:     5.0,
			Connect: 20.0,
			Send:    1.0,
			Wait:    100.0,
			Receive: 14.0,
		},
	}

	opts := ConvertOptions{
		IncludeURL:         true,
		IncludeMethod:      true,
		IncludeStatus:      true,
		IncludeContentType: true,
		IncludeSize:        true,
		IncludeTime:        true,
		IncludeTimings:     true,
		IncludeHeaders:     true,
		IncludeDateTime:    true,
		IncludePostData:    true,
		IncludeQueryString: true,
	}

	row := createDataRow(entry, opts)

	// We expect a row with all fields populated
	// DateTime, Method, URL, Status, ContentType, Size, Time, 6 timings, PostDataMimeType, PostDataText, QueryString, ReqHeaders, RespHeaders
	expectedLen := len(getHeaders(opts))
	if len(row) != expectedLen {
		t.Errorf("Expected %d cells, got %d", expectedLen, len(row))
	}

	// Spot check some values
	if row[1] != "POST" {
		t.Errorf("Expected 'POST' at index 1, got '%s'", row[1])
	}
	if row[2] != "https://example.com/api?foo=bar" {
		t.Errorf("Expected URL at index 2, got '%s'", row[2])
	}
	if row[3] != "200 OK" {
		t.Errorf("Expected '200 OK' at index 3, got '%s'", row[3])
	}
}

func TestConvertCSVFullRoundTrip(t *testing.T) {
	// End-to-end CSV conversion test
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatCSV, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// CSV should be parseable
	if !strings.Contains(result, ",") {
		t.Error("CSV output should contain commas")
	}
	if !strings.Contains(result, "GET") {
		t.Error("CSV should contain GET")
	}
	if !strings.Contains(result, "POST") {
		t.Error("CSV should contain POST")
	}
}

func TestWriteCSVToWriterWrapsWriterErrors(t *testing.T) {
	rootErr := errors.New("csv write failed")
	err := writeCSVToWriter(&csvErrorWriter{err: rootErr}, []Entries{{}}, DefaultConvertOptions())
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, rootErr) {
		t.Fatalf("expected wrapped root error, got %v", err)
	}
}

func TestWriteCSVToWriterShortWrite(t *testing.T) {
	err := writeCSVToWriter(&shortWriter{}, []Entries{{}}, DefaultConvertOptions())
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected wrapped short write error, got %v", err)
	}
}

func TestWriteCSVToWriterNilWriter(t *testing.T) {
	var nilWriter io.Writer
	assertHarErrorCode(t, writeCSVToWriter(nilWriter, nil, DefaultConvertOptions()), ErrCodeInvalidFormat)

	var typedNilWriter *bytes.Buffer
	assertHarErrorCode(t, writeCSVToWriter(typedNilWriter, nil, DefaultConvertOptions()), ErrCodeInvalidFormat)
}

func TestConvertMarkdownWithPipeEscape(t *testing.T) {
	// Branch: markdown cell with pipe character should be escaped
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api?a=1|2", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.MimeType = "text/html"
	e.Response.Content.Size = 100
	e.Time = 50.0

	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatMarkdown, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Pipe in URL should be escaped as \|
	if !strings.Contains(result, "\\|") {
		t.Error("Markdown should escape pipe characters")
	}
}

type csvErrorWriter struct {
	err error
}

func (w *csvErrorWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestConvertHTMLWithSpecialChars(t *testing.T) {
	// Branch: HTML special chars like <, >, & should be escaped
	h := NewHar()
	e := h.AddEntry("GET", "https://example.com/api?a=<script>&b=1&2", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.MimeType = "text/html"
	e.Response.Content.Size = 100
	e.Time = 50.0

	opts := DefaultConvertOptions()

	result, err := h.Convert(FormatHTML, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// < should be escaped as &lt;
	if !strings.Contains(result, "&lt;") {
		t.Error("HTML should escape < characters")
	}
	// & should be escaped as &amp;
	if !strings.Contains(result, "&amp;") {
		t.Error("HTML should escape & characters")
	}
}

func TestConvertAllFormatsWithAllOptions(t *testing.T) {
	// Test all formats with all include options enabled
	h := NewHar()
	e := h.AddEntry("POST", "https://example.com/api?key=val", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.Response.Content.MimeType = "application/json"
	e.Response.Content.Size = 100
	e.Time = 150.0
	e.AddRequestHeader("Accept", "application/json")
	e.AddResponseHeader("Content-Type", "application/json")
	e.AddQueryParameter("key", "val")
	e.SetPostData("application/json", `{"name":"test"}`)
	e.SetTimings(10, 5, 20, 1, 100, 14, 15)

	opts := ConvertOptions{
		IncludeURL:         true,
		IncludeMethod:      true,
		IncludeStatus:      true,
		IncludeContentType: true,
		IncludeSize:        true,
		IncludeTime:        true,
		IncludeTimings:     true,
		IncludeHeaders:     true,
		IncludeDateTime:    true,
		IncludePostData:    true,
		IncludeQueryString: true,
	}

	formats := []ConvertFormat{FormatCSV, FormatMarkdown, FormatHTML, FormatText}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			result, err := h.Convert(format, opts)
			if err != nil {
				t.Fatalf("Convert %s failed: %v", format, err)
			}
			if result == "" {
				t.Errorf("Convert %s returned empty result", format)
			}
		})
	}
}

func TestConvertFilterResultEmpty(t *testing.T) {
	// Branch: Filter matches no entries
	h := createTestHarForConvert()
	opts := DefaultConvertOptions()
	opts.Filter = &FilterOptions{
		Method: "DELETE", // No DELETE requests in test data
	}

	result, err := h.Convert(FormatCSV, opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have headers but no data rows (only header line)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected only header line for empty filter result, got %d lines", len(lines))
	}
}

func TestEscapeHTML(t *testing.T) {
	// Test escapeHTML function directly
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"a&b", "a&amp;b"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"normal text", "normal text"},
		{"<a href=\"test\">", "&lt;a href=&quot;test&quot;&gt;"},
	}

	for _, tt := range tests {
		result := escapeHTML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeHTML(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
