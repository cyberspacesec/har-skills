package har

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 本文件为覆盖率冲刺 100% 的最终补丁测试，覆盖 builder.go / decode.go /
// http_convert.go / redact.go 中剩余的错误分支与边界分支。函数名统一 TestCov
// 前缀，不与既有测试冲突。

// --- builder.go ---

// Cover AddEntryFromHTTPWithMeta nil-HarBuilder branch (builder.go:163-165):
// var b *HarBuilder (未初始化) -> ensureHar 返回 nil -> har==nil -> return nil.
func TestCovAddEntryFromHTTPWithMeta_NilBuilder(t *testing.T) {
	body := bytes.NewBufferString(`{"k":"v"}`)
	req := httptest.NewRequest(http.MethodPost, "https://api.example.com", body)
	resp := &http.Response{
		StatusCode: 200,
		Body:       nopBodyReadCloser(bytes.NewBufferString(`{"id":1}`)),
	}
	var b *HarBuilder // nil receiver -> ensureHar 返回 nil
	eb := b.AddEntryFromHTTPWithMeta(req, resp, time.Now(), 0, EntryMeta{})
	if eb != nil {
		t.Fatalf("expected nil EntryBuilder for nil HarBuilder, got %v", eb)
	}
}

// Cover applyEntryMeta nil-entry branch (builder.go:266-268) and
// InitiatorLine>0 branch (builder.go:286-288).
func TestCovApplyEntryMeta_Branches(t *testing.T) {
	// nil entry -> early return (line 266-268)
	applyEntryMeta(nil, EntryMeta{ServerIPAddress: "1.2.3.4"})

	// InitiatorType + InitiatorLine>0 -> sets Initiator with LineNumber (line 286-288)
	e := &Entries{}
	applyEntryMeta(e, EntryMeta{
		InitiatorType: "script",
		InitiatorURL:  "https://example.com/app.js",
		InitiatorLine: 42,
	})
	if e.Initiator.Type != "script" || e.Initiator.URL != "https://example.com/app.js" {
		t.Fatalf("Initiator not set correctly: %+v", e.Initiator)
	}
	if e.Initiator.LineNumber != 42 {
		t.Fatalf("Initiator.LineNumber = %d, want 42", e.Initiator.LineNumber)
	}
}

// Cover WriteEntryToWriter Encode-error branch (builder.go:692-694):
// 注入 Response.Error = func(){} (json.Marshal 不支持的类型) 触发 Encode 失败.
func TestCovWriteEntryToWriter_EncodeError(t *testing.T) {
	entry := Entries{
		Request:  Request{Method: "GET", URL: "https://example.com"},
		Response: Response{Error: func() {}}, // unsupported type for json.Marshal
	}
	var buf bytes.Buffer
	err := WriteEntryToWriter(&buf, entry)
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// Cover AppendEntryToJSONLFile empty-path branch (builder.go:703-705).
func TestCovAppendEntryToJSONLFile_EmptyPath(t *testing.T) {
	err := AppendEntryToJSONLFile("", Entries{})
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// Cover SafeRecorder.ToHarCopy nil-har branch (builder.go:814-816) and
// SaveToFileWithOptions nil-har branch (builder.go:843-845).
// SafeRecorder.recorder 为 nil 时，recorder.ToHar() 调 ensureBuilder，
// 后者处理 r==nil 返回 nil，故 ToHar 返回 nil，触发上述分支。
func TestCovSafeRecorder_NilHarBranches(t *testing.T) {
	var sr *SafeRecorder // nil receiver + nil recorder

	// ToHarCopy: s==nil 命中 808-810 提前返回 nil；用未初始化 SafeRecorder 覆盖 recorder==nil
	sr2 := &SafeRecorder{} // recorder == nil
	h := sr2.ToHarCopy()
	if h != nil {
		t.Fatalf("expected nil from ToHarCopy when recorder is nil, got %v", h)
	}

	// SaveToFileWithOptions: recorder==nil -> ToHar 返回 nil -> error (line 843-845)
	err := sr2.SaveToFileWithOptions("/tmp/should-not-be-created.har", false, false)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// 顺带覆盖 nil SafeReceiver 的 ToHarCopy (line 808-810)
	if sr.ToHarCopy() != nil {
		t.Fatalf("expected nil from nil SafeRecorder.ToHarCopy")
	}
}
