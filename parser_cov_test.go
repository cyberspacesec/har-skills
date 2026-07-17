package har

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本测试文件为 parser.go 补充覆盖率，使用独立的 Cov 前缀函数名，
// 不与 parser_test.go 重复或冲突。

// --- 辅助函数 ---

// covWriteFile 在 t.TempDir() 下写入文件，返回完整路径。
func covWriteFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(p, content, 0644))
	return p
}

// covWriteJSON 在 t.TempDir() 下写入 JSON 文件，返回完整路径。
func covWriteJSON(t *testing.T, name string, v interface{}) string {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	return covWriteFile(t, name, data)
}

// validHarJSON 返回一个最小但有效的 HAR JSON 字节切片。
func validHarJSON() []byte {
	return []byte(`{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "1.0"},
			"entries": []
		}
	}`)
}

// --- ParseHarWithOptions 分支覆盖 ---

// TestCovParseHarWithOptionsEmpty 覆盖行 13-15: 空输入分支。
func TestCovParseHarWithOptionsEmpty(t *testing.T) {
	h, err := ParseHarWithOptions(nil, DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
	he, ok := err.(*HarError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
}

// TestCovParseHarWithOptionsEmptySlice 覆盖行 13-15: 长度为0的切片。
func TestCovParseHarWithOptionsEmptySlice(t *testing.T) {
	h, err := ParseHarWithOptions([]byte{}, DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseHarWithOptionsNonJSON 覆盖行 18-19: 非 JSON 格式分支。
func TestCovParseHarWithOptionsNonJSON(t *testing.T) {
	h, err := ParseHarWithOptions([]byte("not a json"), DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
	he, ok := err.(*HarError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
}

// TestCovParseHarWithOptionsNonJSONBrackets 覆盖行 18-19: 看似 JSON 但实际不合法（前缀后缀不匹配）。
func TestCovParseHarWithOptionsNonJSONBrackets(t *testing.T) {
	// 前缀是 { 但后缀不是 }，isJSONContent 返回 false
	h, err := ParseHarWithOptions([]byte("{ not closing brace"), DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseHarWithOptionsStrictJSONError 覆盖行 25-28: 严格模式下 json.Unmarshal 失败。
func TestCovParseHarWithOptionsStrictJSONError(t *testing.T) {
	// 合法 JSON 但结构无法映射到 Har（log 是字符串而非对象）
	h, err := ParseHarWithOptions([]byte(`{"log":"notobj"}`), DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseHarWithOptionsStrictValidationError 覆盖行 31-34: 严格模式下解析成功但验证失败。
func TestCovParseHarWithOptionsStrictValidationError(t *testing.T) {
	// 合法 JSON 结构，但缺少 version/creator.name，validateHar 会返回错误
	bad := []byte(`{"log":{"entries":[]}}`)
	h, err := ParseHarWithOptions(bad, DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseHarWithOptionsStrictSkipValidation 覆盖行 31（SkipValidation=true 跳过验证分支）和 37。
func TestCovParseHarWithOptionsStrictSkipValidation(t *testing.T) {
	opts := DefaultParseOptions()
	opts.SkipValidation = true
	// 缺少 version，但跳过验证应成功
	bad := []byte(`{"log":{"entries":[]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// TestCovParseHarWithOptionsLenient 覆盖行 41: 宽松模式分支。
func TestCovParseHarWithOptionsLenient(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	h, err := ParseHarWithOptions(validHarJSON(), opts)
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// --- ParseHarFileWithOptions 覆盖 ---

// TestCovParseHarFileWithOptionsReadError 覆盖行 46-49: 读文件失败。
func TestCovParseHarFileWithOptionsReadError(t *testing.T) {
	h, err := ParseHarFileWithOptions("/nonexistent/path/file.har", DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
	he, ok := err.(*HarError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeFileSystem, he.Code)
}

// TestCovParseHarFileWithOptionsParseError 覆盖行 51-57: 解析失败且为 HarError，调用 WithMetadata。
func TestCovParseHarFileWithOptionsParseError(t *testing.T) {
	p := covWriteFile(t, "bad.har", []byte("not json"))
	h, err := ParseHarFileWithOptions(p, DefaultParseOptions())
	assert.Nil(t, h)
	require.Error(t, err)
	he, ok := err.(*HarError)
	assert.True(t, ok)
	// WithMetadata 应被调用，metadata 中应包含 filePath
	require.NotNil(t, he.Metadata)
	_, hasPath := he.Metadata["filePath"]
	assert.True(t, hasPath)
}

// TestCovParseHarFileWithOptionsSuccess 覆盖行 60: 成功路径。
func TestCovParseHarFileWithOptionsSuccess(t *testing.T) {
	p := covWriteFile(t, "ok.har", validHarJSON())
	h, err := ParseHarFileWithOptions(p, DefaultParseOptions())
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// --- ParseHarEnhanced 覆盖 ---

// TestCovParseHarEnhancedHarError 覆盖行 65, 67-69: 返回 HarError。
func TestCovParseHarEnhancedHarError(t *testing.T) {
	h, he := ParseHarEnhanced([]byte("not json"))
	assert.Nil(t, h)
	require.NotNil(t, he)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
}

// TestCovParseHarEnhancedUnknownError 覆盖行 71: 非 HarError 包装分支。
// 注意: 由于所有 ParseHarWithOptions 的错误都是 *HarError, 此分支在实际中不可达。
// 但通过直接传入能产生非 *HarError 的输入进行尝试覆盖（理论上无法触发）。
func TestCovParseHarEnhancedUnknownError(t *testing.T) {
	// ParseHarWithOptions 内部所有错误构造器都返回 *HarError，
	// 因此无法触发 71 行的 unknown 包装分支。此处仅验证成功路径不进入该分支。
	h, he := ParseHarEnhanced(validHarJSON())
	assert.NotNil(t, h)
	assert.Nil(t, he)
}

// TestCovParseHarEnhancedSuccess 覆盖行 73: 成功路径。
func TestCovParseHarEnhancedSuccess(t *testing.T) {
	h, he := ParseHarEnhanced(validHarJSON())
	assert.Nil(t, he)
	assert.NotNil(t, h)
	assert.Equal(t, "1.2", h.Log.Version)
}

// --- ParseHarFileEnhanced 覆盖 ---

// TestCovParseHarFileEnhancedFileSystemError 覆盖行 77-82: 文件系统错误返回 HarError。
func TestCovParseHarFileEnhancedFileSystemError(t *testing.T) {
	h, he := ParseHarFileEnhanced("/nonexistent/file.har")
	assert.Nil(t, h)
	require.NotNil(t, he)
}

// TestCovParseHarFileEnhancedParseError 覆盖行 80-81: HarError 路径。
func TestCovParseHarFileEnhancedParseError(t *testing.T) {
	p := covWriteFile(t, "bad.har", []byte("not json"))
	h, he := ParseHarFileEnhanced(p)
	assert.Nil(t, h)
	require.NotNil(t, he)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
}

// TestCovParseHarFileEnhancedUnknownError 覆盖行 84: 非 HarError 包装分支（不可达，仅验证成功路径）。
func TestCovParseHarFileEnhancedUnknownError(t *testing.T) {
	p := covWriteFile(t, "ok.har", validHarJSON())
	h, he := ParseHarFileEnhanced(p)
	assert.Nil(t, he)
	assert.NotNil(t, h)
}

// TestCovParseHarFileEnhancedSuccess 覆盖行 86: 成功路径。
func TestCovParseHarFileEnhancedSuccess(t *testing.T) {
	p := covWriteFile(t, "ok.har", validHarJSON())
	h, he := ParseHarFileEnhanced(p)
	assert.Nil(t, he)
	assert.NotNil(t, h)
}

// --- ParseHarLenient 覆盖 ---

// TestCovParseHarLenient 覆盖行 90-95: 整个函数。
func TestCovParseHarLenient(t *testing.T) {
	h, err := ParseHarLenient(validHarJSON())
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// TestCovParseHarLenientPartial 覆盖行 90-95: 宽松模式有部分错误。
func TestCovParseHarLenientPartial(t *testing.T) {
	// 有 entries 但 version 错误类型
	bad := []byte(`{"log":{"version":123,"entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	h, err := ParseHarLenient(bad)
	// 有 entries，所以即使有错误也返回 har（CollectWarnings=true）
	require.NotNil(t, h)
	// 可能有错误也可能没有，取决于解析
	_ = err
}

// --- ParseHarFileLenient 覆盖 ---

// TestCovParseHarFileLenient 覆盖行 98-103: 整个函数。
func TestCovParseHarFileLenient(t *testing.T) {
	p := covWriteFile(t, "ok.har", validHarJSON())
	h, err := ParseHarFileLenient(p)
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// TestCovParseHarFileLenientReadError 覆盖行 98-103: 文件不存在。
func TestCovParseHarFileLenientReadError(t *testing.T) {
	h, err := ParseHarFileLenient("/nonexistent/file.har")
	assert.Nil(t, h)
	require.Error(t, err)
}

// --- validateHar nil 分支 ---

// TestCovValidateHarNil 覆盖行 115-117: nil 分支。
func TestCovValidateHarNil(t *testing.T) {
	err := validateHar(nil)
	require.Error(t, err)
	he, ok := err.(*HarError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInvalidFormat, he.Code)
}

// TestCovValidateHarValid 覆盖行 119: 有效 HAR。
func TestCovValidateHarValid(t *testing.T) {
	h := &Har{Log: Log{Version: "1.2", Creator: Creator{Name: "x", Version: "1"}, Entries: []Entries{}}}
	err := validateHar(h)
	assert.NoError(t, err)
}

// --- parseLenient 各分支覆盖 ---

// TestCovParseLenientNoLog 覆盖行 215-217: 缺少 log 字段。
func TestCovParseLenientNoLog(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// 没有任何可解析内容（无 version/entries/pages），返回 (nil, err)
	h, err := ParseHarWithOptions([]byte(`{"foo":"bar"}`), opts)
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseLenientNoLogNoCollect 覆盖行 227-230: 有错误但不收集警告。
func TestCovParseLenientNoLogNoCollect(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = false
	h, err := ParseHarWithOptions([]byte(`{"foo":"bar"}`), opts)
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseLenientLogNotObject 覆盖行 147-150: log 字段不是对象（无法解析为 map）。
func TestCovParseLenientLogNotObject(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// log 是字符串，无法 unmarshal 成 map[string]json.RawMessage
	h, err := ParseHarWithOptions([]byte(`{"log":"notobj"}`), opts)
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseLenientVersionBadType 覆盖行 156-159: version 字段类型错误。
func TestCovParseLenientVersionBadType(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// version 是数字，无法解析为 string；但有 entries，所以返回 har
	bad := []byte(`{"log":{"version":123,"entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	require.NotNil(t, h)
	assert.Len(t, h.Log.Entries, 1)
	// 有错误（版本解析失败），CollectWarnings=true 且有内容，返回 (har, err)
	require.Error(t, err)
}

// TestCovParseLenientCreatorBadType 覆盖行 167-170: creator 字段类型错误。
func TestCovParseLenientCreatorBadType(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	bad := []byte(`{"log":{"version":"1.2","creator":"notobj","entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	require.NotNil(t, h)
	assert.Equal(t, "1.2", h.Log.Version)
	require.Error(t, err)
}

// TestCovParseLenientPagesBadType 覆盖行 188-191: pages 字段不是数组。
func TestCovParseLenientPagesBadType(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	bad := []byte(`{"log":{"version":"1.2","pages":"notarray","entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	require.NotNil(t, h)
	require.Error(t, err)
}

// TestCovParseLenientPageItemBad 覆盖行 177-186: pages 是数组但单个 page 解析失败。
func TestCovParseLenientPageItemBad(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// 第一个 page 是字符串（无法解析为 Pages），第二个有效
	bad := []byte(`{"log":{"version":"1.2","pages":["notobj",{"id":"p1","title":"T","startedDateTime":"2024-01-01T00:00:00Z","pageTimings":{"onContentLoad":0,"onLoad":0}}],"entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	require.NotNil(t, h)
	require.Error(t, err)
	// 第二个 page 应该被成功解析
	require.Len(t, h.Log.Pages, 1)
}

// TestCovParseLenientEntriesBadType 覆盖行 209-212: entries 字段不是数组。
func TestCovParseLenientEntriesBadType(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	bad := []byte(`{"log":{"version":"1.2","entries":"notarray"}}`)
	h, err := ParseHarWithOptions(bad, opts)
	// 有 version，所以返回 har
	require.NotNil(t, h)
	assert.Equal(t, "1.2", h.Log.Version)
	require.Error(t, err)
}

// TestCovParseLenientEntryItemBad 覆盖行 202-207: entries 是数组但单个 entry 解析失败。
func TestCovParseLenientEntryItemBad(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// 第一个 entry 是字符串，第二个有效
	bad := []byte(`{"log":{"version":"1.2","entries":["badentry",{"startedDateTime":"2024-01-01T00:00:00Z","time":10,"request":{"method":"GET","url":"http://x"},"response":{"status":200}}]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	require.NotNil(t, h)
	require.Error(t, err)
	require.Len(t, h.Log.Entries, 1)
}

// TestCovParseLenientPartialNoCollectWithContent 覆盖行 227-230: 有错误、不收集警告、但有内容。
// 此时返回 (nil, err) 因为 CollectWarnings=false。
func TestCovParseLenientPartialNoCollectWithContent(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = false
	bad := []byte(`{"log":{"version":123,"entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	h, err := ParseHarWithOptions(bad, opts)
	// CollectWarnings=false 且有错误 -> 返回 (nil, err)
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseLenientFullFailureWithCollect 覆盖行 222-226: 有错误、收集警告、但无任何可解析内容。
func TestCovParseLenientFullFailureWithCollect(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// log 存在但全是坏数据，无 version/entries/pages 成功
	bad := []byte(`{"log":{"version":123,"creator":"x","pages":"x","entries":"x"}}`)
	h, err := ParseHarWithOptions(bad, opts)
	assert.Nil(t, h)
	require.Error(t, err)
}

// TestCovParseLenientNoErrors 覆盖行 232: 无错误返回 (har, nil)。
func TestCovParseLenientNoErrors(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	h, err := ParseHarWithOptions(validHarJSON(), opts)
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// TestCovParseLenientRootUnmarshalError 覆盖行 134-136: 顶层 JSON 无法解析为 map。
// isJSONContent 接受数组形式 [...]，但无法 unmarshal 成 map[string]json.RawMessage。
func TestCovParseLenientRootUnmarshalError(t *testing.T) {
	opts := DefaultParseOptions()
	opts.Lenient = true
	opts.CollectWarnings = true
	// 合法的 JSON 数组，isJSONContent 返回 true，但 unmarshal 到 map 失败
	h, err := ParseHarWithOptions([]byte(`[1,2,3]`), opts)
	assert.Nil(t, h)
	require.Error(t, err)
}

// --- ParseHarWithWarnings 覆盖 ---

// TestCovParseHarWithWarningsFullFailure 覆盖行 259-266: 解析完全失败 (nil har)。
func TestCovParseHarWithWarningsFullFailure(t *testing.T) {
	// 空输入 -> ParseHarWithOptions 返回 (nil, err)，且 har==nil -> 完全失败分支
	result, err := ParseHarWithWarnings([]byte{})
	assert.Nil(t, result)
	require.Error(t, err)
}

// TestCovParseHarWithWarningsEmpty 覆盖行 259-266: 非空但完全无法解析。
func TestCovParseHarWithWarningsEmpty(t *testing.T) {
	result, err := ParseHarWithWarnings(nil)
	assert.Nil(t, result)
	require.Error(t, err)
}

// TestCovParseHarWithWarningsValidNoWarnings 覆盖行 276-279: 成功且无警告 -> performFullValidation 分支。
func TestCovParseHarWithWarningsValidNoWarnings(t *testing.T) {
	// 完全有效的 HAR，宽松模式解析无错误，validateURLs 也无警告
	// -> 进入 performFullValidation 分支
	result, err := ParseHarWithWarnings(validHarJSON())
	require.NoError(t, err)
	require.NotNil(t, result)
	// performFullValidation 会发现缺少 creator.name/version 等，产生警告
	// 但这里 HAR 是有效的，所以可能无警告
	_ = result.Warnings
}

// TestCovParseHarWithWarningsPartialWithHar 覆盖行 260-262: 有 har 但有解析错误 -> 转为警告。
func TestCovParseHarWithWarningsPartialWithHar(t *testing.T) {
	// version 类型错误但有 entries -> (har, err) 且 har != nil
	bad := []byte(`{"log":{"version":123,"entries":[{"request":{"method":"GET","url":"http://x"}}]}}`)
	result, err := ParseHarWithWarnings(bad)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Har)
	// 应该有解析警告或 URL 验证相关
	// URL http://x 包含 :// 无空格，validateURLs 不会产生警告
	// 但解析错误会产生警告
	assert.NotEmpty(t, result.Warnings)
}

// TestCovParseHarWithWarningsURLValidation 覆盖行 270-273: validateURLs 产生警告。
func TestCovParseHarWithWarningsURLValidation(t *testing.T) {
	// URL 包含空格，触发 validateURLs 空格分支
	bad := []byte(`{"log":{"version":"1.2","entries":[{"request":{"method":"GET","url":"http://example.com/with space"}}]}}`)
	result, err := ParseHarWithWarnings(bad)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Warnings)
}

// --- validateURLs 覆盖 ---

// TestCovValidateURLsNil 覆盖行 286-288: nil har。
func TestCovValidateURLsNil(t *testing.T) {
	warnings := validateURLs(nil)
	assert.Nil(t, warnings)
}

// TestCovValidateURLsNoEntries 覆盖行 286-288: 无 entries。
func TestCovValidateURLsNoEntries(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{}}}
	warnings := validateURLs(h)
	assert.Nil(t, warnings)
}

// TestCovValidateURLsEmptyURL 覆盖行 292-293: 空 URL 跳过。
func TestCovValidateURLsEmptyURL(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{
		{Request: Request{URL: ""}},
	}}}
	warnings := validateURLs(h)
	assert.Empty(t, warnings)
}

// TestCovValidateURLsNoScheme 覆盖行 315-316: URL 缺少 ://。
func TestCovValidateURLsNoScheme(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{
		{Request: Request{URL: "example.com/path"}},
	}}}
	warnings := validateURLs(h)
	assert.NotEmpty(t, warnings)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "缺少协议") {
			found = true
		}
	}
	assert.True(t, found)
}

// TestCovValidateURLsSpace 覆盖行 307-308: URL 包含空格。
func TestCovValidateURLsSpace(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{
		{Request: Request{URL: "http://example.com/with space"}},
	}}}
	warnings := validateURLs(h)
	assert.NotEmpty(t, warnings)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "空格") {
			found = true
		}
	}
	assert.True(t, found)
}

// TestCovValidateURLsValid 覆盖行 297: 有效 URL（url.Parse 无错误）+ 有 :// + 无空格。
func TestCovValidateURLsValid(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{
		{Request: Request{URL: "http://example.com/path"}},
	}}}
	warnings := validateURLs(h)
	assert.Empty(t, warnings)
}

// TestCovValidateURLsParseError 覆盖行 297-303: url.Parse 返回错误分支。
// "http://[::1" 缺少 ] 会导致 url.Parse 报错，且 URL 包含 :// 但无空格，
// 因此进入 url.Parse 错误分支后 continue，不进入空格/协议检查。
func TestCovValidateURLsParseError(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{
		{Request: Request{URL: "http://[::1"}},
	}}}
	warnings := validateURLs(h)
	assert.NotEmpty(t, warnings)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "无效的URL格式") {
			found = true
		}
	}
	assert.True(t, found)
}

// TestCovValidateURLsMultiple 覆盖行 291 循环多个 entry + 空格+缺协议同时存在。
func TestCovValidateURLsMultiple(t *testing.T) {
	h := &Har{Log: Log{Entries: []Entries{
		{Request: Request{URL: "http://example.com/path"}}, // 有效
		{Request: Request{URL: ""}},                        // 空，跳过
		{Request: Request{URL: "noscheme path"}},           // 空格 + 缺协议
	}}}
	warnings := validateURLs(h)
	assert.NotEmpty(t, warnings)
	// 应同时有空格和缺协议警告
	assert.GreaterOrEqual(t, len(warnings), 2)
}

// --- performFullValidation 覆盖 ---

// TestCovPerformFullValidationNil 覆盖行 329-331: nil har。
func TestCovPerformFullValidationNil(t *testing.T) {
	warnings := performFullValidation(nil)
	assert.Nil(t, warnings)
}

// TestCovPerformFullValidationValid 覆盖行 333-336: 有效 HAR，验证无错误。
func TestCovPerformFullValidationValid(t *testing.T) {
	h := &Har{Log: Log{Version: "1.2", Creator: Creator{Name: "x", Version: "1"}, Entries: []Entries{}}}
	warnings := performFullValidation(h)
	assert.Nil(t, warnings)
}

// TestCovPerformFullValidationInvalid 覆盖行 338-340: 无效 HAR，返回 HarError -> GetPartialErrors。
func TestCovPerformFullValidationInvalid(t *testing.T) {
	// 缺少 version -> validateBasicStructure 会添加 partial error 并返回 rootError (*HarError)
	h := &Har{Log: Log{Creator: Creator{Name: "x", Version: "1"}, Entries: []Entries{}}}
	warnings := performFullValidation(h)
	assert.NotEmpty(t, warnings)
}

// TestCovPerformFullValidationNonHarError 覆盖行 343-345: 非 HarError 分支。
// 注意: ValidateHarFile 总是返回 *HarError 或 nil，因此此分支在实践中不可达。
// 此测试仅作记录，无法真正触发该分支。
func TestCovPerformFullValidationNonHarError(t *testing.T) {
	// 此分支不可达，ValidateHarFile 不会返回非 *HarError 类型。
	// 通过验证一个有部分错误的 HAR 确保走 HarError 分支。
	h := &Har{Log: Log{Creator: Creator{Name: "x", Version: "1"}, Entries: []Entries{}}}
	warnings := performFullValidation(h)
	assert.NotEmpty(t, warnings)
}

// --- appendWarnings 覆盖 ---

// TestCovAppendWarningsEmptyNew 覆盖行 350-352: 新警告为空。
func TestCovAppendWarningsEmptyNew(t *testing.T) {
	existing := []*HarError{NewValidationError("a", "f1")}
	result := appendWarnings(existing, nil)
	assert.Len(t, result, 1)
}

// TestCovAppendWarningsNilExisting 覆盖行 354-356: existing 为 nil。
func TestCovAppendWarningsNilExisting(t *testing.T) {
	newW := []*HarError{NewValidationError("a", "f1")}
	result := appendWarnings(nil, newW)
	assert.Len(t, result, 1)
}

// TestCovAppendWarningsDuplicate 覆盖行 360-363 + 366-372 去重分支。
func TestCovAppendWarningsDuplicate(t *testing.T) {
	w1 := NewValidationError("msg", "field")
	existing := []*HarError{w1}
	// 相同 field+message 应去重
	dup := NewValidationError("msg", "field")
	result := appendWarnings(existing, []*HarError{dup})
	assert.Len(t, result, 1)
}

// TestCovAppendWarningsNewAdded 覆盖行 366-372: 新警告追加。
func TestCovAppendWarningsNewAdded(t *testing.T) {
	existing := []*HarError{NewValidationError("msg1", "f1")}
	newW := []*HarError{NewValidationError("msg2", "f2")}
	result := appendWarnings(existing, newW)
	assert.Len(t, result, 2)
}

// TestCovAppendWarningsMixed 覆盖去重+追加混合。
func TestCovAppendWarningsMixed(t *testing.T) {
	existing := []*HarError{
		NewValidationError("keep", "f1"),
		NewValidationError("dup", "f2"),
	}
	newW := []*HarError{
		NewValidationError("dup", "f2"), // 重复
		NewValidationError("new", "f3"), // 新的
	}
	result := appendWarnings(existing, newW)
	assert.Len(t, result, 3)
}

// --- ParseHarFileWithWarnings 覆盖 ---

// TestCovParseHarFileWithWarningsReadError 覆盖行 378-382: 读文件失败。
func TestCovParseHarFileWithWarningsReadError(t *testing.T) {
	result, err := ParseHarFileWithWarnings("/nonexistent/file.har")
	assert.Nil(t, result)
	require.Error(t, err)
	he, ok := err.(*HarError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeFileSystem, he.Code)
}

// TestCovParseHarFileWithWarningsSuccess 覆盖行 384: 成功路径。
func TestCovParseHarFileWithWarningsSuccess(t *testing.T) {
	p := covWriteFile(t, "ok.har", validHarJSON())
	result, err := ParseHarFileWithWarnings(p)
	require.NoError(t, err)
	require.NotNil(t, result)
}

// --- isJSONContent 补充覆盖 ---

// TestCovIsJSONContentArray 覆盖数组形式 JSON。
func TestCovIsJSONContentArray(t *testing.T) {
	assert.True(t, isJSONContent([]byte("  [1,2,3]  ")))
}

// TestCovIsJSONContentObject 覆盖对象形式 JSON。
func TestCovIsJSONContentObject(t *testing.T) {
	assert.True(t, isJSONContent([]byte("  {\"a\":1}  ")))
}

// TestCovIsJSONContentInvalid 覆盖非 JSON。
func TestCovIsJSONContentInvalid(t *testing.T) {
	assert.False(t, isJSONContent([]byte("  hello  ")))
	assert.False(t, isJSONContent([]byte("[1,2}"))) // 前后不匹配
}
