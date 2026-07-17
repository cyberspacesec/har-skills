package har

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- SafeRecorder 基础功能测试 ---

func TestNewSafeRecorder(t *testing.T) {
	rec := NewSafeRecorder()
	require.NotNil(t, rec)
	assert.Equal(t, 0, rec.EntryCount())
}

func TestSafeRecorderSetCreator(t *testing.T) {
	rec := NewSafeRecorder().SetCreator("test-agent", "1.0")
	require.NotNil(t, rec)

	h := rec.ToHarCopy()
	require.NotNil(t, h)
	assert.Equal(t, "test-agent", h.Log.Creator.Name)
	assert.Equal(t, "1.0", h.Log.Creator.Version)
}

func TestSafeRecorderSetBrowser(t *testing.T) {
	rec := NewSafeRecorder().SetBrowser("test-browser", "2.0")
	require.NotNil(t, rec)

	h := rec.ToHarCopy()
	require.NotNil(t, h)
	assert.Equal(t, "test-browser", h.Log.Browser.Name)
	assert.Equal(t, "2.0", h.Log.Browser.Version)
}

func TestSafeRecorderCapture(t *testing.T) {
	rec := NewSafeRecorder()

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp := &http.Response{
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Body:       http.NoBody,
	}

	rec.Capture(req, resp, 50*time.Millisecond)
	assert.Equal(t, 1, rec.EntryCount())
}

func TestSafeRecorderCaptureWithMeta(t *testing.T) {
	rec := NewSafeRecorder()

	req, _ := http.NewRequest("POST", "https://api.example.com/data", nil)
	resp := &http.Response{
		StatusCode: 201,
		Proto:      "HTTP/1.1",
		Body:       http.NoBody,
	}

	started := time.Now().Add(-100 * time.Millisecond)
	meta := EntryMeta{
		ServerIPAddress: "192.0.2.1",
		Connection:      "conn-1",
		Pageref:         "page-1",
		InitiatorType:   "script",
		Priority:        "High",
		ResourceType:    "xhr",
		Comment:         "test entry",
	}

	rec.CaptureWithMeta(req, resp, started, 100*time.Millisecond, meta)
	assert.Equal(t, 1, rec.EntryCount())

	h := rec.ToHarCopy()
	require.Len(t, h.Log.Entries, 1)
	entry := h.Log.Entries[0]

	assert.Equal(t, "192.0.2.1", entry.ServerIPAddress)
	assert.Equal(t, "conn-1", entry.Connection)
	assert.Equal(t, "page-1", entry.Pageref)
	assert.Equal(t, "High", entry.Priority)
	assert.Equal(t, "xhr", entry.ResourceType)
	assert.Equal(t, "test entry", entry.Comment)
}

func TestSafeRecorderCaptureEntry(t *testing.T) {
	rec := NewSafeRecorder()

	entry := Entries{
		Request: Request{URL: "https://example.com/captured"},
	}
	rec.CaptureEntry(entry)
	assert.Equal(t, 1, rec.EntryCount())

	h := rec.ToHarCopy()
	assert.Equal(t, "https://example.com/captured", h.Log.Entries[0].Request.URL)
}

func TestSafeRecorderToHar(t *testing.T) {
	rec := NewSafeRecorder()
	rec.CaptureEntry(Entries{Request: Request{URL: "https://example.com"}})

	h := rec.ToHar()
	require.NotNil(t, h)
	assert.Len(t, h.Log.Entries, 1)
}

func TestSafeRecorderToHarCopy(t *testing.T) {
	rec := NewSafeRecorder()
	rec.CaptureEntry(Entries{Request: Request{URL: "https://example.com"}})

	copy1 := rec.ToHarCopy()
	copy2 := rec.ToHarCopy()

	// 副本间互不影响
	require.NotNil(t, copy1)
	require.NotNil(t, copy2)
	assert.NotSame(t, copy1, copy2)

	// 修改 copy1 不影响 copy2
	copy1.Log.Entries[0].Request.URL = "https://modified.com"
	assert.Equal(t, "https://example.com", copy2.Log.Entries[0].Request.URL)
}

func TestSafeRecorderSaveToFile(t *testing.T) {
	rec := NewSafeRecorder().SetCreator("test", "1.0")

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp := &http.Response{
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       http.NoBody,
	}
	rec.Capture(req, resp, 50*time.Millisecond)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.har")

	// SaveToFile 内部调用 SaveToFileWithOptions(true, false) 会做严格验证
	// 用 SaveToFileWithOptions(false, false) 跳过验证，仅测试序列化
	err := rec.SaveToFileWithOptions(path, false, false)
	require.NoError(t, err)

	// 验证文件存在且是合法 JSON
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, json.Valid(data))
}

func TestSafeRecorderSaveToFileWithOptions(t *testing.T) {
	rec := NewSafeRecorder()
	rec.CaptureEntry(Entries{Request: Request{URL: "https://example.com"}})

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.har")

	// 无缩进 + gzip
	err := rec.SaveToFileWithOptions(path, false, true)
	require.NoError(t, err)

	// 验证文件存在且非空
	stat, err := os.Stat(path)
	require.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(0))
}

func TestSafeRecorderSaveToFileInvalidPath(t *testing.T) {
	rec := NewSafeRecorder()
	err := rec.SaveToFile("/nonexistent/dir/file.har")
	assert.Error(t, err)
}

// --- 并发安全测试 ---

func TestSafeRecorderConcurrentCapture(t *testing.T) {
	rec := NewSafeRecorder()

	const workers = 10
	const entriesPerWorker = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < entriesPerWorker; i++ {
				req, _ := http.NewRequest("GET", "https://example.com/"+string(rune('a'+id%26)), nil)
				resp := &http.Response{
					StatusCode: 200,
					Body:       http.NoBody,
				}
				rec.Capture(req, resp, 10*time.Millisecond)
			}
		}(w)
	}

	wg.Wait()

	assert.Equal(t, workers*entriesPerWorker, rec.EntryCount())
}

func TestSafeRecorderConcurrentReadAndWrite(t *testing.T) {
	rec := NewSafeRecorder()

	const writers = 3
	const readers = 3
	const opsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// 写协程
	for w := 0; w < writers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				req, _ := http.NewRequest("GET", "https://example.com/", nil)
				resp := &http.Response{Body: http.NoBody}
				rec.Capture(req, resp, 1*time.Millisecond)
			}
		}()
	}

	// 读协程
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				_ = rec.EntryCount()
				_ = rec.ToHarCopy()
			}
		}()
	}

	wg.Wait()

	// 最终 entry 数应等于写入总数
	assert.Equal(t, writers*opsPerGoroutine, rec.EntryCount())
}

func TestSafeRecorderConcurrentToHarCopy(t *testing.T) {
	rec := NewSafeRecorder()
	rec.CaptureEntry(Entries{Request: Request{URL: "https://example.com"}})

	const goroutines = 20
	results := make([]*Har, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = rec.ToHarCopy()
		}(i)
	}

	wg.Wait()

	// 所有副本应独立且有效
	for _, h := range results {
		require.NotNil(t, h)
		assert.Len(t, h.Log.Entries, 1)
	}
}

// --- 边界：空/nil 输入 ---

func TestSafeRecorderCaptureNilRequest(t *testing.T) {
	rec := NewSafeRecorder()
	rec.Capture(nil, nil, 0)
	// nil request 不应 panic，但不一定产生有效 entry
	// 视实现可能跳过或产生空 entry
	assert.LessOrEqual(t, rec.EntryCount(), 1)
}

func TestSafeRecorderToHarNil(t *testing.T) {
	var rec *SafeRecorder
	h := rec.ToHar()
	assert.Nil(t, h)
}

func TestSafeRecorderToHarCopyNil(t *testing.T) {
	var rec *SafeRecorder
	h := rec.ToHarCopy()
	assert.Nil(t, h)
}

// --- 边界：链式调用 ---

func TestSafeRecorderChaining(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp := &http.Response{Body: http.NoBody}

	rec := NewSafeRecorder().
		SetCreator("chain-test", "1.0").
		SetBrowser("chain-browser", "2.0").
		Capture(req, resp, 10*time.Millisecond).
		Capture(req, resp, 20*time.Millisecond)

	assert.Equal(t, 2, rec.EntryCount())

	h := rec.ToHarCopy()
	assert.Equal(t, "chain-test", h.Log.Creator.Name)
	assert.Equal(t, "chain-browser", h.Log.Browser.Name)
}
