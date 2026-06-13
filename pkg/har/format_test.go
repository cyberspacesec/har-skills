package har

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- ToYAML ---

func TestToYAML_BasicOutput(t *testing.T) {
	h := NewHar()
	h.SetCreator("test-app", "2.0")
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(512, "application/json")

	yaml, err := h.ToYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if yaml == "" {
		t.Fatal("Expected non-empty YAML output")
	}
	if !strings.Contains(yaml, "log:") {
		t.Error("YAML should contain 'log:' key")
	}
	if !strings.Contains(yaml, "version:") {
		t.Error("YAML should contain 'version:' key")
	}
	if !strings.Contains(yaml, "creator:") {
		t.Error("YAML should contain 'creator:' key")
	}
	if !strings.Contains(yaml, "entries:") {
		t.Error("YAML should contain 'entries:' key")
	}
	if !strings.Contains(yaml, "https://example.com/api") {
		t.Error("YAML should contain the URL")
	}
	if !strings.Contains(yaml, "GET") {
		t.Error("YAML should contain the HTTP method")
	}
}

func TestToYAML_NilHAR(t *testing.T) {
	var h *Har
	yaml, err := h.ToYAML()
	if err != nil {
		t.Fatalf("Unexpected error for nil HAR: %v", err)
	}
	if yaml != "" {
		t.Errorf("Expected empty string for nil HAR, got %q", yaml)
	}
}

func TestToYAML_WithBrowser(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	h.SetBrowser("Chrome", "120.0")

	yaml, err := h.ToYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(yaml, "browser:") {
		t.Error("YAML should contain 'browser:' key")
	}
	if !strings.Contains(yaml, "Chrome") {
		t.Error("YAML should contain browser name")
	}
}

// --- SaveAsYAML ---

func TestSaveAsYAML_WriteToTempFile(t *testing.T) {
	h := NewHar()
	h.SetCreator("save-test", "1.0")
	e := h.AddEntry("POST", "https://example.com/submit", "HTTP/1.1", "")
	e.SetResponseStatus(201, "Created")
	e.SetResponseContent(0, "text/plain")

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.yaml")

	err := h.SaveAsYAML(filePath)
	if err != nil {
		t.Fatalf("SaveAsYAML returned error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Fatal("Saved YAML file is empty")
	}
	if !strings.Contains(content, "log:") {
		t.Error("Saved YAML should contain 'log:' key")
	}
	if !strings.Contains(content, "POST") {
		t.Error("Saved YAML should contain HTTP method")
	}
}

// --- ConvertTo ---

func TestConvertTo_CSV(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(FormatCSV, &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected non-empty CSV output")
	}
	if !strings.Contains(buf.String(), "GET") {
		t.Error("CSV output should contain HTTP method")
	}
}

func TestConvertTo_Markdown(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(FormatMarkdown, &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected non-empty Markdown output")
	}
	output := buf.String()
	if !strings.Contains(output, "|") {
		t.Error("Markdown output should contain table delimiters")
	}
}

func TestConvertTo_HTML(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(FormatHTML, &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected non-empty HTML output")
	}
	output := buf.String()
	if !strings.Contains(output, "<table") {
		t.Error("HTML output should contain <table> tag")
	}
	if !strings.Contains(output, "</table>") {
		t.Error("HTML output should contain closing </table> tag")
	}
}

func TestConvertTo_Text(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(FormatText, &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected non-empty text output")
	}
	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Error("Text output should contain HTTP method")
	}
}

func TestConvertTo_YAML(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(FormatYAML, &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected non-empty YAML output")
	}
	output := buf.String()
	if !strings.Contains(output, "log:") {
		t.Error("YAML output should contain 'log:' key")
	}
}

func TestConvertTo_DefaultFormat(t *testing.T) {
	// Unknown format should fall back to JSON
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(ConvertFormat("unknown"), &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error for unknown format: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected non-empty JSON fallback output")
	}
	// JSON output should start with '{'
	output := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(output, "{") {
		preview := output
		if len(preview) > 20 {
			preview = preview[:20]
		}
		t.Errorf("Expected JSON fallback to start with '{', got: %s", preview)
	}
}

func TestConvertTo_AllFormats(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")
	e.SetResponseContent(256, "application/json")

	formats := []ConvertFormat{FormatCSV, FormatMarkdown, FormatHTML, FormatText, FormatYAML}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			var buf bytes.Buffer
			err := h.ConvertTo(format, &buf, DefaultConvertOptions())
			if err != nil {
				t.Fatalf("Unexpected error for format %s: %v", format, err)
			}
			if buf.Len() == 0 {
				t.Errorf("Expected non-empty output for format %s", format)
			}
		})
	}
}

// --- jsonToYAML ---

func TestJsonToYAML_Object(t *testing.T) {
	input := []byte(`{"name":"test","value":42}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "name:") {
		t.Error("Expected YAML to contain 'name:' key")
	}
	if !strings.Contains(result, "test") {
		t.Error("Expected YAML to contain 'test' value")
	}
	if !strings.Contains(result, "value:") {
		t.Error("Expected YAML to contain 'value:' key")
	}
}

func TestJsonToYAML_Array(t *testing.T) {
	input := []byte(`[1,2,3]`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "-") {
		t.Error("Expected YAML array to contain '-' items")
	}
}

func TestJsonToYAML_NestedObject(t *testing.T) {
	input := []byte(`{"outer":{"inner":"deep"}}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "outer:") {
		t.Error("Expected YAML to contain 'outer:' key")
	}
	if !strings.Contains(result, "inner:") {
		t.Error("Expected YAML to contain 'inner:' key")
	}
	if !strings.Contains(result, "deep") {
		t.Error("Expected YAML to contain 'deep' value")
	}
}

func TestJsonToYAML_Boolean(t *testing.T) {
	input := []byte(`{"active":true,"deleted":false}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "active: true") {
		t.Error("Expected YAML to contain 'active: true'")
	}
	if !strings.Contains(result, "deleted: false") {
		t.Error("Expected YAML to contain 'deleted: false'")
	}
}

func TestJsonToYAML_Null(t *testing.T) {
	input := []byte(`{"value":null}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "value: null") {
		t.Error("Expected YAML to contain 'value: null'")
	}
}

func TestJsonToYAML_Float(t *testing.T) {
	input := []byte(`{"pi":3.14}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "pi:") {
		t.Error("Expected YAML to contain 'pi:' key")
	}
	if !strings.Contains(result, "3.14") {
		t.Error("Expected YAML to contain '3.14' value")
	}
}

func TestJsonToYAML_InvalidJSON(t *testing.T) {
	// Invalid JSON should be returned as-is
	input := []byte(`not valid json`)
	result := jsonToYAML(input)
	if result != "not valid json" {
		t.Errorf("Expected invalid JSON to be returned as-is, got %q", result)
	}
}

func TestJsonToYAML_EmptyObject(t *testing.T) {
	input := []byte(`{}`)
	result := jsonToYAML(input)
	// Empty object produces empty YAML (just whitespace / nothing)
	result = strings.TrimSpace(result)
	if result != "" {
		t.Errorf("Expected empty YAML for empty object, got %q", result)
	}
}

func TestJsonToYAML_StringWithSpecialChars(t *testing.T) {
	input := []byte(`{"url":"https://example.com?a=1&b=2"}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "url:") {
		t.Error("Expected YAML to contain 'url:' key")
	}
}

// --- escapeYAMLString ---

func TestEscapeYAMLString_DoubleQuote(t *testing.T) {
	result := escapeYAMLString(`hello "world"`)
	if !strings.Contains(result, `\"`) {
		t.Errorf("Expected escaped double quote, got %q", result)
	}
}

func TestEscapeYAMLString_Newline(t *testing.T) {
	result := escapeYAMLString("line1\nline2")
	if !strings.Contains(result, `\n`) {
		t.Errorf("Expected escaped newline, got %q", result)
	}
}

func TestEscapeYAMLString_Tab(t *testing.T) {
	result := escapeYAMLString("col1\tcol2")
	if !strings.Contains(result, `\t`) {
		t.Errorf("Expected escaped tab, got %q", result)
	}
}

func TestEscapeYAMLString_Backslash(t *testing.T) {
	result := escapeYAMLString(`back\slash`)
	if !strings.Contains(result, `\\`) {
		t.Errorf("Expected escaped backslash, got %q", result)
	}
}

func TestEscapeYAMLString_NoEscapingNeeded(t *testing.T) {
	input := "hello world"
	result := escapeYAMLString(input)
	if result != input {
		t.Errorf("Expected %q unchanged, got %q", input, result)
	}
}

func TestEscapeYAMLString_AllSpecialChars(t *testing.T) {
	input := "a\tb\nc\\d\"e"
	result := escapeYAMLString(input)
	if !strings.Contains(result, `\t`) {
		t.Error("Expected \\t in result")
	}
	if !strings.Contains(result, `\n`) {
		t.Error("Expected \\n in result")
	}
	if !strings.Contains(result, `\\`) {
		t.Error("Expected \\\\ in result")
	}
	if !strings.Contains(result, `\"`) {
		t.Error("Expected \\\" in result")
	}
}

func TestEscapeYAMLString_Empty(t *testing.T) {
	result := escapeYAMLString("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}
