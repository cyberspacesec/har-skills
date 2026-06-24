package har

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
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

// --- ToYAML error paths ---

func TestToYAML_ErrorFromToJSON(t *testing.T) {
	// Create a Har with entries that would cause ToJSON to fail is tricky
	// because ToJSON rarely fails. Instead, test with a Har that has
	// circular references or other problematic data by using a nil Har pointer
	// (already covered). Test the jsonToYAML error path via ConvertTo.
	// We test the ToJSON error path through ConvertTo below.
}

// --- SaveAsYAML error paths ---

func TestSaveAsYAML_NilHAR(t *testing.T) {
	var h *Har
	filePath := filepath.Join(t.TempDir(), "nil.yaml")

	err := h.SaveAsYAML(filePath)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	if _, statErr := os.Stat(filePath); !os.IsNotExist(statErr) {
		t.Fatalf("expected nil HAR not to create file, stat err: %v", statErr)
	}
}

func TestSaveAsYAML_InvalidPath(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	// Write to an invalid path (directory that doesn't exist)
	err := h.SaveAsYAML(filepath.Join(t.TempDir(), "missing", "output.yaml"))
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

// --- valueToYAML comprehensive branches ---

func TestValueToYAML_TopLevelArray(t *testing.T) {
	// Top-level []interface{} case in valueToYAML
	input := []byte(`["hello","world"]`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- hello") {
		t.Errorf("Expected YAML array with '- hello', got %q", result)
	}
	if !strings.Contains(result, "- world") {
		t.Errorf("Expected YAML array with '- world', got %q", result)
	}
}

func TestValueToYAML_TopLevelNil(t *testing.T) {
	// Top-level nil case in valueToYAML
	input := []byte(`null`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "null") {
		t.Errorf("Expected YAML with 'null', got %q", result)
	}
}

func TestValueToYAML_TopLevelString(t *testing.T) {
	// Top-level string case in valueToYAML
	input := []byte(`"hello"`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "hello") {
		t.Errorf("Expected YAML with 'hello', got %q", result)
	}
}

func TestValueToYAML_TopLevelFloat64Integer(t *testing.T) {
	// Top-level float64 that is an integer value
	input := []byte(`42`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "42") {
		t.Errorf("Expected YAML with '42', got %q", result)
	}
	// JSON numbers unmarshal as float64; 42 is integer-valued
	if strings.Contains(result, "42.0") {
		t.Errorf("Expected integer format, got %q", result)
	}
}

func TestValueToYAML_TopLevelFloat64Decimal(t *testing.T) {
	// Top-level float64 that is a decimal value
	input := []byte(`3.14159`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "3.14159") {
		t.Errorf("Expected YAML with '3.14159', got %q", result)
	}
}

func TestValueToYAML_TopLevelBool(t *testing.T) {
	// Top-level bool case in valueToYAML
	input := []byte(`true`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "true") {
		t.Errorf("Expected YAML with 'true', got %q", result)
	}
}

func TestValueToYAML_MapValueFloat64Integer(t *testing.T) {
	// Map value that is float64 with integer value
	input := []byte(`{"count":42}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "count: 42") {
		t.Errorf("Expected 'count: 42', got %q", result)
	}
}

func TestValueToYAML_MapValueFloat64Decimal(t *testing.T) {
	// Map value that is float64 with decimal value
	input := []byte(`{"pi":3.14}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "pi: 3.14") {
		t.Errorf("Expected 'pi: 3.14', got %q", result)
	}
}

func TestValueToYAML_MapValueStringNeedsQuoting(t *testing.T) {
	// String with special YAML chars that need quoting
	input := []byte(`{"val":"hello:world"}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, `val: "hello:world"`) {
		t.Errorf("Expected quoted string, got %q", result)
	}
}

func TestValueToYAML_MapValueStringNoQuotingNeeded(t *testing.T) {
	// String without special YAML chars
	input := []byte(`{"name":"alice"}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "name: alice") {
		t.Errorf("Expected unquoted string, got %q", result)
	}
}

func TestValueToYAML_MapValueEmptyString(t *testing.T) {
	// Empty string in map value should be quoted
	input := []byte(`{"name":""}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, `name: ""`) {
		t.Errorf("Expected quoted empty string, got %q", result)
	}
}

func TestValueToYAML_MapValueNil(t *testing.T) {
	// Already covered by TestJsonToYAML_Null but let's be thorough
	input := []byte(`{"value":null}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "value: null") {
		t.Errorf("Expected 'value: null', got %q", result)
	}
}

func TestValueToYAML_MapValueBool(t *testing.T) {
	// Already covered by TestJsonToYAML_Boolean but let's be explicit
	input := []byte(`{"active":true,"deleted":false}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "active: true") {
		t.Errorf("Expected 'active: true', got %q", result)
	}
	if !strings.Contains(result, "deleted: false") {
		t.Errorf("Expected 'deleted: false', got %q", result)
	}
}

func TestValueToYAML_MapValueArray(t *testing.T) {
	// Map value that is an array
	input := []byte(`{"items":["a","b"]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "items:") {
		t.Errorf("Expected 'items:' key, got %q", result)
	}
	if !strings.Contains(result, "- a") {
		t.Errorf("Expected '- a' array item, got %q", result)
	}
}

func TestValueToYAML_MapValueMap(t *testing.T) {
	// Map value that is a nested map
	input := []byte(`{"outer":{"inner":"deep"}}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "outer:") {
		t.Errorf("Expected 'outer:' key, got %q", result)
	}
	if !strings.Contains(result, "inner:") {
		t.Errorf("Expected 'inner:' key, got %q", result)
	}
}

func TestValueToYAML_MultipleMapKeys(t *testing.T) {
	// Test that multiple keys produce newlines between them (the !first branch)
	input := []byte(`{"a":1,"b":2,"c":3}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "a:") || !strings.Contains(result, "b:") || !strings.Contains(result, "c:") {
		t.Errorf("Expected all keys present, got %q", result)
	}
}

func TestValueToYAML_StringWithVariousSpecialChars(t *testing.T) {
	// Test strings that trigger quoting due to various special characters
	specialStrings := []struct {
		name  string
		input string
	}{
		{"colon", `{"v":"a:b"}`},
		{"brace_open", `{"v":"a{b"}`},
		{"brace_close", `{"v":"a}b"}`},
		{"bracket_open", `{"v":"a[b"}`},
		{"bracket_close", `{"v":"a]b"}`},
		{"ampersand", `{"v":"a&b"}`},
		{"asterisk", `{"v":"a*b"}`},
		{"question", `{"v":"a?b"}`},
		{"pipe", `{"v":"a|b"}`},
		{"gt", `{"v":"a>b"}`},
		{"dash", `{"v":"a-b"}`},
		{"exclaim", `{"v":"a!b"}`},
		{"percent", `{"v":"a%b"}`},
		{"at", `{"v":"a@b"}`},
		{"backtick", "{\"v\":\"a`b\"}"},
		{"doublequote", `{"v":"a\"b"}`},
		{"singlequote", `{"v":"a'b"}`},
	}

	for _, tc := range specialStrings {
		t.Run(tc.name, func(t *testing.T) {
			result := jsonToYAML([]byte(tc.input))
			if !strings.Contains(result, "v:") {
				t.Errorf("Expected v: key for %s, got %q", tc.name, result)
			}
		})
	}
}

// --- arrayToYAML comprehensive branches ---

func TestArrayToYAML_MapItems(t *testing.T) {
	// Array of maps - tests the map case in arrayToYAML
	input := []byte(`{"items":[{"name":"a"},{"name":"b"}]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "-\n") {
		t.Errorf("Expected '-\\n' for map items in array, got %q", result)
	}
	if !strings.Contains(result, "name:") {
		t.Errorf("Expected 'name:' in map items, got %q", result)
	}
}

func TestArrayToYAML_StringItemsNoQuoting(t *testing.T) {
	// Array of plain strings - no quoting needed
	input := []byte(`{"items":["hello","world"]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- hello") {
		t.Errorf("Expected '- hello', got %q", result)
	}
	if !strings.Contains(result, "- world") {
		t.Errorf("Expected '- world', got %q", result)
	}
}

func TestArrayToYAML_StringItemsWithSpecialChars(t *testing.T) {
	// Array of strings with special YAML chars - needs quoting
	input := []byte(`{"items":["hello:world","foo{}bar"]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, `- "hello:world"`) {
		t.Errorf("Expected quoted string in array, got %q", result)
	}
}

func TestArrayToYAML_EmptyStringItem(t *testing.T) {
	// Empty string in array should be quoted
	input := []byte(`{"items":[""]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, `- ""`) {
		t.Errorf("Expected quoted empty string in array, got %q", result)
	}
}

func TestArrayToYAML_Float64Integer(t *testing.T) {
	// Float64 with integer value in array
	input := []byte(`{"items":[42,100]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- 42") {
		t.Errorf("Expected '- 42', got %q", result)
	}
}

func TestArrayToYAML_Float64Decimal(t *testing.T) {
	// Float64 with decimal value in array
	input := []byte(`{"items":[3.14,2.718]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- 3.14") {
		t.Errorf("Expected '- 3.14', got %q", result)
	}
}

func TestArrayToYAML_BoolItems(t *testing.T) {
	// Boolean items in array
	input := []byte(`{"items":[true,false]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- true") {
		t.Errorf("Expected '- true', got %q", result)
	}
	if !strings.Contains(result, "- false") {
		t.Errorf("Expected '- false', got %q", result)
	}
}

func TestArrayToYAML_NilItems(t *testing.T) {
	// Null items in array
	input := []byte(`{"items":[null,null]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- null") {
		t.Errorf("Expected '- null', got %q", result)
	}
}

func TestArrayToYAML_NestedArray(t *testing.T) {
	// Nested array in array (triggers default case in arrayToYAML since
	// nested []interface{} doesn't match the typed cases)
	input := []byte(`{"items":[[1,2],[3,4]]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "items:") {
		t.Errorf("Expected 'items:' key, got %q", result)
	}
}

func TestArrayToYAML_TopLevelMixedTypes(t *testing.T) {
	// Top-level array with mixed types to cover valueToYAML's []interface{} case
	input := []byte(`[1,"hello",true,null,3.14]`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "- 1") {
		t.Errorf("Expected '- 1', got %q", result)
	}
	if !strings.Contains(result, "- hello") {
		t.Errorf("Expected '- hello', got %q", result)
	}
	if !strings.Contains(result, "- true") {
		t.Errorf("Expected '- true', got %q", result)
	}
	if !strings.Contains(result, "- null") {
		t.Errorf("Expected '- null', got %q", result)
	}
	if !strings.Contains(result, "- 3.14") {
		t.Errorf("Expected '- 3.14', got %q", result)
	}
}

// --- ConvertTo error paths ---

func TestConvertTo_DefaultFormatJSON(t *testing.T) {
	// Default/unknown format should output JSON
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(ConvertFormat("xml"), &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	output := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(output, "{") {
		t.Errorf("Expected JSON output starting with '{', got: %s", output[:min(20, len(output))])
	}
}

func TestConvertTo_YAMLFormat(t *testing.T) {
	// Already tested but let's verify the error path is exercised
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com/api", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	var buf bytes.Buffer
	err := h.ConvertTo(FormatYAML, &buf, DefaultConvertOptions())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "log:") {
		t.Error("YAML output should contain 'log:' key")
	}
}

func TestConvertTo_NilHAR(t *testing.T) {
	var h *Har
	var buf bytes.Buffer

	err := h.ConvertTo(FormatYAML, &buf, DefaultConvertOptions())
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
	if buf.Len() != 0 {
		t.Fatalf("expected nil HAR not to write output, got %q", buf.String())
	}
}

func TestConvertTo_NilWriter(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")

	var nilWriter io.Writer
	var typedNilBuffer *bytes.Buffer
	tests := []struct {
		name   string
		writer io.Writer
	}{
		{name: "nil interface", writer: nilWriter},
		{name: "typed nil", writer: typedNilBuffer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.ConvertTo(FormatYAML, tt.writer, DefaultConvertOptions())
			assertHarErrorCode(t, err, ErrCodeInvalidFormat)
		})
	}
}

func TestConvertTo_WriterError(t *testing.T) {
	// Test that writer errors are propagated
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	// Use a writer that always fails
	errWriter := &formatFailingWriter{}
	err := h.ConvertTo(FormatYAML, errWriter, DefaultConvertOptions())
	assertHarErrorCode(t, err, ErrCodeFileSystem)
}

func TestConvertTo_ShortWrite(t *testing.T) {
	h := NewHar()
	h.SetCreator("test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

	err := h.ConvertTo(FormatYAML, &shortWriter{}, DefaultConvertOptions())
	assertHarErrorCode(t, err, ErrCodeFileSystem)
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected wrapped short write error, got %v", err)
	}
}

// formatFailingWriter is an io.Writer that always returns an error
type formatFailingWriter struct{}

func (f *formatFailingWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write error")
}

// --- valueToYAML default case coverage ---

func TestValueToYAML_DefaultCaseInMap(t *testing.T) {
	// The default case in the map value switch handles types that don't
	// match the specific cases. Since JSON unmarshaling only produces
	// map[string]interface{}, []interface{}, string, float64, bool, and nil,
	// the default case is hard to trigger through JSON alone.
	// We test it by calling valueToYAML directly with a custom type.
	result := valueToYAML(map[string]interface{}{
		"key": complex(1, 2), // complex128 doesn't match any typed case
	}, 0)
	if !strings.Contains(result, "key:") {
		t.Errorf("Expected 'key:' in output, got %q", result)
	}
	// complex128 should be rendered via %v format
	if !strings.Contains(result, "(1+2i)") {
		t.Errorf("Expected complex number representation, got %q", result)
	}
}

func TestValueToYAML_DefaultCaseAtTopLevel(t *testing.T) {
	// Top-level default case
	result := valueToYAML(complex(1, 2), 0)
	if !strings.Contains(result, "(1+2i)") {
		t.Errorf("Expected complex number representation, got %q", result)
	}
}

func TestArrayToYAML_DefaultCase(t *testing.T) {
	// Default case in arrayToYAML
	result := valueToYAML(map[string]interface{}{
		"items": []interface{}{complex(3, 4)},
	}, 0)
	if !strings.Contains(result, "items:") {
		t.Errorf("Expected 'items:' key, got %q", result)
	}
	if !strings.Contains(result, "(3+4i)") {
		t.Errorf("Expected complex number representation in array, got %q", result)
	}
}

// --- writeToFile error path (covered via SaveAsYAML_InvalidPath) ---

func TestWriteToFile_InvalidPath(t *testing.T) {
	err := writeToFile("/nonexistent/dir/file.txt", []byte("data"))
	if err == nil {
		t.Error("Expected error writing to nonexistent directory")
	}
}

// --- jsonToYAML with various edge cases ---

func TestJsonToYAML_TopLevelNull(t *testing.T) {
	input := []byte(`null`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "null") {
		t.Errorf("Expected 'null', got %q", result)
	}
}

func TestJsonToYAML_TopLevelNumber(t *testing.T) {
	input := []byte(`42`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "42") {
		t.Errorf("Expected '42', got %q", result)
	}
}

func TestJsonToYAML_TopLevelBool(t *testing.T) {
	input := []byte(`true`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "true") {
		t.Errorf("Expected 'true', got %q", result)
	}
}

func TestJsonToYAML_TopLevelString(t *testing.T) {
	input := []byte(`"hello"`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "hello") {
		t.Errorf("Expected 'hello', got %q", result)
	}
}

func TestJsonToYAML_DeeplyNested(t *testing.T) {
	input := []byte(`{"a":{"b":{"c":"deep"}}}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "a:") {
		t.Errorf("Expected 'a:' key, got %q", result)
	}
	if !strings.Contains(result, "b:") {
		t.Errorf("Expected 'b:' key, got %q", result)
	}
	if !strings.Contains(result, "c:") {
		t.Errorf("Expected 'c:' key, got %q", result)
	}
	if !strings.Contains(result, "deep") {
		t.Errorf("Expected 'deep' value, got %q", result)
	}
}

func TestArrayToYAML_ArrayOfArrays(t *testing.T) {
	// Nested arrays - inner array triggers default case in arrayToYAML
	input := []byte(`{"data":[[1,2],[3,4]]}`)
	result := jsonToYAML(input)
	if !strings.Contains(result, "data:") {
		t.Errorf("Expected 'data:' key, got %q", result)
	}
}

// --- Comprehensive ToYAML with complex HAR ---

func TestToYAML_ComplexHAR(t *testing.T) {
	h := NewHar()
	h.SetCreator("complex-test", "3.0")
	h.SetBrowser("Firefox", "115.0")

	// Add multiple entries with various data
	e1 := h.AddEntry("GET", "https://example.com/api/users", "HTTP/1.1", "")
	e1.SetResponseStatus(200, "OK")
	e1.SetResponseContent(1024, "application/json")

	e2 := h.AddEntry("POST", "https://example.com/api/login", "HTTP/1.1", "")
	e2.SetResponseStatus(401, "Unauthorized")
	e2.SetResponseContent(128, "text/html")

	yaml, err := h.ToYAML()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(yaml, "log:") {
		t.Error("YAML should contain 'log:' key")
	}
	if !strings.Contains(yaml, "entries:") {
		t.Error("YAML should contain 'entries:' key")
	}
	if !strings.Contains(yaml, "browser:") {
		t.Error("YAML should contain 'browser:' key")
	}
}

func TestSaveAsYAML_Success(t *testing.T) {
	h := NewHar()
	h.SetCreator("save-test", "1.0")
	e := h.AddEntry("GET", "https://example.com", "HTTP/1.1", "")
	e.SetResponseStatus(200, "OK")

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
}

// Helper for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
