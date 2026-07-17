package har

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- WriteEntryToWriter 覆盖测试 ---

func TestWriteEntryToWriterNormal(t *testing.T) {
	entry := Entries{
		StartedDateTime: time.Now(),
		Time:            42,
		Request:         Request{URL: "https://example.com"},
	}

	var buf bytes.Buffer
	err := WriteEntryToWriter(&buf, entry)
	require.NoError(t, err)

	// 输出应为合法 JSON（json.Encoder.Encode 末尾会加换行，符合 JSONL 规范）
	line := strings.TrimRight(buf.String(), "\n")
	assert.True(t, json.Valid([]byte(line)), "输出应为合法 JSON")
	assert.NotContains(t, line, "\n", "JSON 内容应为单行")

	// 反序列化验证内容
	var decoded Entries
	require.NoError(t, json.Unmarshal([]byte(line), &decoded))
	assert.Equal(t, entry.Request.URL, decoded.Request.URL)
	assert.Equal(t, entry.Time, decoded.Time)
}

func TestWriteEntryToWriterNilWriter(t *testing.T) {
	entry := Entries{Request: Request{URL: "https://example.com"}}
	err := WriteEntryToWriter(nil, entry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "writer")
}

func TestWriteEntryToWriterEncodeError(t *testing.T) {
	// Entries 本身总是可 JSON 编码，无法触发 encode error
	// 改为测试 writer 返回 error（通过 errWriter）
	entry := Entries{Request: Request{URL: "https://example.com"}}
	err := WriteEntryToWriter(errWriter{}, entry)
	assert.Error(t, err)
}

func TestWriteEntryToWriterShortWrite(t *testing.T) {
	entry := Entries{Request: Request{URL: "https://example.com"}}
	err := WriteEntryToWriter(jsonlShortWriter{}, entry)
	assert.Error(t, err)
}

// --- AppendEntryToJSONLFile 覆盖测试 ---

func TestAppendEntryToJSONLFileCreate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "archive.jsonl")

	entry := Entries{
		StartedDateTime: time.Now(),
		Request:         Request{URL: "https://example.com/1"},
	}

	// 文件不存在时会自动创建
	err := AppendEntryToJSONLFile(path, entry)
	require.NoError(t, err)

	// 验证文件内容
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, json.Valid(data))
}

func TestAppendEntryToJSONLFileAppend(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "archive.jsonl")

	// 先写两条
	for i := 0; i < 2; i++ {
		entry := Entries{
			Request: Request{URL: "https://example.com/" + string(rune('a'+i))},
		}
		require.NoError(t, AppendEntryToJSONLFile(path, entry))
	}

	// 验证文件有两行
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 2)
}

func TestAppendEntryToJSONLFileInvalidPath(t *testing.T) {
	// 路径指向目录而非文件
	entry := Entries{Request: Request{URL: "https://example.com"}}
	err := AppendEntryToJSONLFile("/tmp/", entry)
	assert.Error(t, err)
}

// --- ForEachEntryFromReader 覆盖测试 ---

func TestForEachEntryFromReaderNormal(t *testing.T) {
	entries := []Entries{
		{Request: Request{URL: "https://example.com/1"}},
		{Request: Request{URL: "https://example.com/2"}},
		{Request: Request{URL: "https://example.com/3"}},
	}

	var buf bytes.Buffer
	for _, e := range entries {
		require.NoError(t, WriteEntryToWriter(&buf, e))
	}

	var collected []Entries
	err := ForEachEntryFromReader(&buf, func(e Entries) error {
		collected = append(collected, e)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, collected, 3)
}

func TestForEachEntryFromReaderCallbackError(t *testing.T) {
	entries := []Entries{
		{Request: Request{URL: "https://example.com/1"}},
		{Request: Request{URL: "https://example.com/2"}},
	}

	var buf bytes.Buffer
	for _, e := range entries {
		require.NoError(t, WriteEntryToWriter(&buf, e))
	}

	count := 0
	err := ForEachEntryFromReader(&buf, func(e Entries) error {
		count++
		if count >= 2 {
			return errors.New("stop")
		}
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stop")
}

func TestForEachEntryFromReaderNilReader(t *testing.T) {
	err := ForEachEntryFromReader(nil, func(e Entries) error { return nil })
	assert.Error(t, err)
}

func TestForEachEntryFromReaderNilCallback(t *testing.T) {
	var buf bytes.Buffer
	err := ForEachEntryFromReader(&buf, nil)
	assert.Error(t, err)
}

func TestForEachEntryFromReaderEmpty(t *testing.T) {
	err := ForEachEntryFromReader(strings.NewReader(""), func(e Entries) error { return nil })
	assert.NoError(t, err)
}

func TestForEachEntryFromReaderInvalidJSON(t *testing.T) {
	err := ForEachEntryFromReader(strings.NewReader("not json\n"), func(e Entries) error { return nil })
	assert.Error(t, err)
}

func TestForEachEntryFromReaderSkipInvalidContinue(t *testing.T) {
	// ForEachEntryFromReader 在遇到非法行时返回 error（当前实现不跳过）
	// 这个测试验证此行为
	valid := Entries{Request: Request{URL: "https://example.com"}}
	var buf bytes.Buffer
	require.NoError(t, WriteEntryToWriter(&buf, valid))
	buf.WriteString("not json\n")

	count := 0
	err := ForEachEntryFromReader(&buf, func(e Entries) error {
		count++
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, 1, count) // 第一条合法的已处理
}

// --- 边界：URL-encoded body 与 JSONL 往返 ---

func TestJSONLEntryWithURLEncodedBody(t *testing.T) {
	entry := Entries{
		Request: Request{
			URL:      "https://example.com/form",
			Method:   "POST",
			PostData: &PostData{Text: "key=value&secret=hidden"},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, WriteEntryToWriter(&buf, entry))

	var decoded Entries
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, "key=value&secret=hidden", decoded.Request.PostData.Text)
}

// --- 边界：大 entry 性能（不去重，只验证不 panic）---

func TestJSONLLargeEntry(t *testing.T) {
	// 构造含大 body 的 entry（模拟大响应）
	largeBody := strings.Repeat("x", 10000)
	entry := Entries{
		Response: Response{
			Content: Content{
				Text: largeBody,
				Size: len(largeBody),
			},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, WriteEntryToWriter(&buf, entry))
	assert.Greater(t, buf.Len(), 10000)
}

// --- 边界：entry 含特殊字符（非 ASCII）---

func TestJSONLEntryWithNonASCII(t *testing.T) {
	entry := Entries{
		Request: Request{
			URL: "https://example.com/测试?name=张三",
		},
	}

	var buf bytes.Buffer
	require.NoError(t, WriteEntryToWriter(&buf, entry))

	var decoded Entries
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Contains(t, decoded.Request.URL, "测试")
}

// --- 辅助类型 ---

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) { return 0, errors.New("write error") }

type jsonlShortWriter struct{}

func (jsonlShortWriter) Write(p []byte) (int, error) { return len(p) - 1, io.ErrShortWrite }
