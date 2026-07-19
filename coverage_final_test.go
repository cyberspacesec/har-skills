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

// --- decode.go ---

// Cover DecompressByEncoding brotli error branch (decode.go:294-297).
// 用户显式传 "br" 绕过 isBrotliData 探测，损坏 brotli 数据使 io.ReadAll 失败。
func TestCovDecompressByEncoding_BrotliError(t *testing.T) {
	// 0x21 是 brotli 常见起始字节，但后续损坏使解码立即失败
	badBrotli := []byte{0x21, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	_, err := DecompressByEncoding(badBrotli, "br")
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// Cover DecompressByEncoding zstd DecodeAll error branch (decode.go:308-311)
// and decompressIfNeeded zstd error branch (decode.go after edit).
// 合法 zstd magic + 损坏 frame 使 DecodeAll 报 reserved block type。
func TestCovDecompress_ZstdDecodeError(t *testing.T) {
	badZstd := []byte{0x28, 0xB5, 0x2F, 0xFD, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00}

	// DecompressByEncoding 显式传 "zstd"
	_, err := DecompressByEncoding(badZstd, "zstd")
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// decompressIfNeeded 通过 isZstdData(magic 校验) 进入分支，DecodeAll 报错
	_, err = decompressIfNeeded(badZstd, "")
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// --- http_convert.go ---

// Cover parseFormParams empty-body branch (http_convert.go:108-110) and
// empty-pair continue branch (http_convert.go:113-114).
func TestCovParseFormParams_Branches(t *testing.T) {
	// 空 body -> 返回空切片 (line 108-110)
	if got := parseFormParams(""); len(got) != 0 {
		t.Fatalf("expected empty slice for empty body, got %v", got)
	}
	// 含空 pair（"&" 分割出空段 -> continue, line 113-114）
	// 含无 "=" 的 pair -> Param{Name: key} (line 118-119)
	// 含 key=value -> Param{Name, Value} (line 121)
	// "&&" 产生两个空段；"a&" 末尾产生一个空段；"noeq" 无 "="
	params := parseFormParams("&&key=&noeq&k=v&")
	if len(params) != 3 {
		t.Fatalf("expected 3 params (empty pairs skipped), got %d (%v)", len(params), params)
	}
	if params[0].Name != "key" || params[0].Value != "" {
		t.Errorf("params[0] = %+v, want Name=key Value=''", params[0])
	}
	if params[1].Name != "noeq" {
		t.Errorf("params[1].Name = %q, want noeq", params[1].Name)
	}
	if params[2].Name != "k" || params[2].Value != "v" {
		t.Errorf("params[2] = %+v, want Name=k Value=v", params[2])
	}
}

// Cover isTextContentType binary application subtype branch (http_convert.go:150-152).
func TestCovIsTextContentType_BinaryApplication(t *testing.T) {
	binary := []string{
		"application/pdf",
		"application/zip",
		"application/gzip",
		"application/octet-stream",
		"application/font-woff",
		"image/png",
		"audio/mpeg",
		"video/mp4",
	}
	for _, m := range binary {
		if isTextContentType(m) {
			t.Errorf("isTextContentType(%q) = true, want false", m)
		}
	}
}

// --- redact.go ---

// Cover redactJSONBody Unmarshal-error defensive branch (redact.go:366-369).
// 直接调用 redactJSONBody，传入 Unmarshal 失败的输入以触发防御分支。
func TestCovRedactJSONBody_UnmarshalError(t *testing.T) {
	// 非法 JSON：值缺失
	text := `{"key": }`
	opts := DefaultRedactOptions()
	out, ok := redactJSONBody(text, opts, "***", nil)
	// 走防御分支应返回 (text, false)
	if ok != false {
		t.Errorf("expected ok=false for invalid JSON, got %v", ok)
	}
	if out != text {
		t.Errorf("expected output==input for defensive branch, got %q", out)
	}
}
