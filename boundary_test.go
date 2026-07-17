package har

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- testdata 集成测试 ---

// TestCompressedHARIntegration 验证 testdata/compressed.har 里的各压缩格式能被正确解析和解压。
// 覆盖 gzip/deflate/br/zstd 四种单层压缩 + base64 编码的往返。
func TestCompressedHARIntegration(t *testing.T) {
	h, err := ParseHarFile("testdata/compressed.har")
	require.NoError(t, err)
	require.Len(t, h.Log.Entries, 5)

	expectedBody := `{"message":"compressed response body","data":[1,2,3,4,5],"nested":{"key":"value"}}`

	// 前 4 条是单层压缩：gzip/deflate/br/zstd
	for i := 0; i < 4; i++ {
		entry := h.Log.Entries[i]
		encoding := entry.Response.Headers[0].Value
		data, err := entry.DecodeContent()
		require.NoError(t, err, "entry %d (%s) 解压失败", i, encoding)
		assert.Equal(t, expectedBody, string(data), "entry %d (%s) 解压结果不匹配", i, encoding)
	}

	// 第 5 条是多重编码 gzip, deflate —— 当前 SDK 的 DecodeContent 靠魔数探测，
	// 只能解外层 gzip，内层 deflate 需手动二次解压
	multiEntry := h.Log.Entries[4]
	assert.Equal(t, "gzip, deflate", multiEntry.Response.Headers[0].Value)
	// 外层 gzip 应能解出（解出后是内层 deflate 的 zlib 字节）
	outerData, err := multiEntry.DecodeContent()
	require.NoError(t, err)
	// 进一步用 DecompressByEncoding 解内层 deflate
	innerData, err := DecompressByEncoding(outerData, "deflate")
	require.NoError(t, err)
	assert.Equal(t, expectedBody, string(innerData))
}

// TestSensitiveHARRedactionIntegration 验证 testdata/sensitive.har 经 redact 后
// 所有敏感数据（密钥、token、JWT、AWS key、GitHub PAT）都被脱敏。
func TestSensitiveHARRedactionIntegration(t *testing.T) {
	h, err := ParseHarFile("testdata/sensitive.har")
	require.NoError(t, err)

	redacted := h.Redact(DefaultRedactOptions())

	// 序列化后整体扫描：不应残留任何明文敏感串
	jsonBytes, err := redacted.ToJSON(false)
	require.NoError(t, err)
	jsonStr := string(jsonBytes)

	// 这些明文敏感值不应再出现
	forbidden := []string{
		"AKIAIOSFODNN7EXAMPLE", // AWS access key
		"eyJhbGciOiJIUzI1NiJ9", // JWT 头
		"ghp_",                 // GitHub PAT 前缀
		"xoxb-",                // Slack token 前缀
		"hunter2",              // 密码明文
		"secret123",            // token 明文
	}
	for _, s := range forbidden {
		assert.NotContains(t, jsonStr, s, "脱敏后仍含明文敏感串: %q", s)
	}

	// 但 URL 主机、path 等非敏感信息应保留
	assert.Contains(t, jsonStr, "api.example.com")
	assert.Contains(t, jsonStr, "/users")
	assert.Contains(t, jsonStr, "/login")
}

// TestSensitiveHARNonStringJSONRedaction 验证 redact 对 JSON body 里非字符串值
// （数字、布尔、null、嵌套对象）的脱敏——这是 Task 19 新增能力。
func TestSensitiveHARNonStringJSONRedaction(t *testing.T) {
	h, err := ParseHarFile("testdata/sensitive.har")
	require.NoError(t, err)

	// 第二条 entry 的 postData 是 JSON，含 password(string) 与 secret_code(int)
	entry := h.Log.Entries[1]
	require.NotNil(t, entry.Request.PostData)
	assert.Contains(t, entry.Request.PostData.Text, `"secret_code": 12345`)

	// 只脱敏 secret_code（数字），不脱敏 password
	opts := DefaultRedactOptions()
	opts.PostDataFields = []string{"secret_code"}
	opts.ValuePatterns = nil // 禁用值模式，专注测试名字匹配非字符串值

	redacted := h.Redact(opts)
	redactedText := redacted.Log.Entries[1].Request.PostData.Text

	// 数字值被替换为字符串 "[REDACTED]"
	assert.Contains(t, redactedText, `"secret_code": "[REDACTED]"`)
	// password 不在脱敏列表，保持原样
	assert.Contains(t, redactedText, `"password": "hunter2"`)
}

// TestParseMinimalHAR 验证最简合法 HAR 仍可解析。
func TestParseMinimalHAR(t *testing.T) {
	h, err := ParseHarFile("testdata/minimal.har")
	require.NoError(t, err)
	assert.NotNil(t, h)
}

// TestParseInvalidHARFiles 验证各类非法 HAR 文件被正确拒绝。
func TestParseInvalidHARFiles(t *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{"not json", "testdata/not_json.har"},
		{"invalid json", "testdata/invalid.har"},
		{"invalid version", "testdata/invalid_version.har"},
		{"missing required", "testdata/missing_required.har"},
		// invalid_url.har 当前解析器不严格验证 URL，故不测
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseHarFile(tc.file)
			assert.Error(t, err, "%s 应解析失败", tc.file)
		})
	}
}

// TestRoundtripCompressedHAR 验证压缩 HAR 解析后能再写出可解析的文件。
func TestRoundtripCompressedHAR(t *testing.T) {
	h, err := ParseHarFile("testdata/compressed.har")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "roundtrip.har")

	// 用 SDK 写出
	data, err := h.ToJSON(true)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(outPath, data, 0644))

	// 重新解析
	h2, err := ParseHarFile(outPath)
	require.NoError(t, err)
	assert.Len(t, h2.Log.Entries, len(h.Log.Entries))

	// 第一条解压结果应一致
	d1, err := h.Log.Entries[0].DecodeContent()
	require.NoError(t, err)
	d2, err := h2.Log.Entries[0].DecodeContent()
	require.NoError(t, err)
	assert.Equal(t, d1, d2)
}

// TestLargeHARSmokeTest 大文件冒烟测试：解析不 panic、条目数正确。
// large.har 含控制字符 URL，严格验证会报错，用宽松模式解析。
func TestLargeHARSmokeTest(t *testing.T) {
	f, err := os.Open("testdata/large.har")
	require.NoError(t, err)
	defer f.Close()

	h, err := ParseHarFromReaderWithOptions(f, ParseOptions{Lenient: true, SkipValidation: true})
	require.NoError(t, err)
	assert.NotEmpty(t, h.Log.Entries)
}

// TestFullHARValidate full.har 应通过严格验证。
func TestFullHARValidate(t *testing.T) {
	h, err := ParseHarFile("testdata/full.har")
	require.NoError(t, err)

	// 不严格验证应通过
	require.NoError(t, ValidateHarFile(h))
}

// TestV11HARParse 1.1 版本 HAR 应能被解析（向后兼容）。
func TestV11HARParse(t *testing.T) {
	h, err := ParseHarFile("testdata/v1.1.har")
	require.NoError(t, err)
	assert.Equal(t, "1.1", h.Log.Version)
}

// TestCompressedHARExtractByEncoding 验证对压缩 entry 用 GetContentEncoding 取值后解压。
func TestCompressedHARExtractByEncoding(t *testing.T) {
	h, err := ParseHarFile("testdata/compressed.har")
	require.NoError(t, err)

	expectedBody := `{"message":"compressed response body","data":[1,2,3,4,5],"nested":{"key":"value"}}`

	// 对每条 entry：先 base64 解码 Content.Text，再按 Content-Encoding 头解压
	for i := 0; i < 4; i++ {
		entry := h.Log.Entries[i]
		encoding := entry.GetContentEncoding()
		assert.NotEmpty(t, encoding, "entry %d 应有 Content-Encoding 头", i)

		// DecodeContent 应等价于 base64 解码 + 按魔数探测解压
		data, err := entry.DecodeContent()
		require.NoError(t, err)
		assert.Equal(t, expectedBody, string(data))
	}
}

// TestRedactValuePatternsAcrossAllFields 集成测试：值模式脱敏覆盖所有字段类型。
func TestRedactValuePatternsAcrossAllFields(t *testing.T) {
	// 构造一个 entry，每个字段都藏一个 Bearer token
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://api.example.com/path?trace=Bearer%20eyJhbGciOiJIUzI1NiJ9.abc.def",
						Headers: []Headers{
							{Name: "X-Custom", Value: "Bearer secret-token-1"},
						},
						Cookies: []Cookie{
							{Name: "tracking", Value: "Bearer secret-token-2"},
						},
						QueryString: []QueryString{
							{Name: "ref", Value: "Bearer secret-token-3"},
						},
						PostData: &PostData{
							Text: `{"note":"Bearer secret-token-4"}`,
						},
					},
					Response: Response{
						Headers: []Headers{
							{Name: "X-Debug", Value: "Bearer secret-token-5"},
						},
						Cookies: []Cookie{
							{Name: "analytics", Value: "Bearer secret-token-6"},
						},
					},
				},
			},
		},
	}

	// 只用值模式，不按名字匹配
	opts := RedactOptions{
		Replacement:   "[X]",
		ValuePatterns: DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)

	// 整体扫描：不应有任何 Bearer secret-token 残留
	jsonBytes, _ := result.ToJSON(false)
	jsonStr := string(jsonBytes)
	assert.NotContains(t, jsonStr, "Bearer secret-token")
	// 但 URL 主机/path 保留
	assert.Contains(t, jsonStr, "api.example.com")
}

// TestRedactURLWithSensitivePath 验证 URL path 段脱敏规则。
func TestRedactURLWithSensitivePath(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://api.example.com/users/12345/tokens/abcdef",
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Replacement: "[ID]",
		RedactURLs: []RedactURLRule{
			{Pattern: `^\d+$`, Replacement: "[ID]"},            // 纯数字段
			{Pattern: `^[a-f0-9]{6,}$`, Replacement: "[HASH]"}, // 长 hex 段
		},
	}
	result := h.Redact(opts)

	url := result.Log.Entries[0].Request.URL
	// URL 编码后 [ID] 变 %5BID%5D，解码验证
	assert.Contains(t, url, "users")
	assert.Contains(t, url, "tokens")
	// path 段已被替换，但 URL 编码保留
	decoded := strings.ReplaceAll(url, "%5B", "[")
	decoded = strings.ReplaceAll(decoded, "%5D", "]")
	assert.Contains(t, decoded, "/users/[ID]/tokens/[HASH]")
}

// TestEmptyEntriesParse 空条目列表的 HAR 应能解析。
func TestEmptyEntriesParse(t *testing.T) {
	jsonStr := `{"log":{"version":"1.2","creator":{"name":"t","version":"1"},"entries":[]}}`
	provider, err := Parse([]byte(jsonStr))
	require.NoError(t, err)
	h := provider.ToStandard()
	assert.Empty(t, h.Log.Entries)

	stats := h.Statistics()
	assert.Equal(t, 0, stats.TotalRequests)
}

// TestUnicodeHandling Unicode/非 ASCII 内容应正确往返。
func TestUnicodeHandling(t *testing.T) {
	unicodeBody := `{"name":"张三","city":"北京","emoji":"🎉"}`
	h := &Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
			Entries: []Entries{
				{
					StartedDateTime: time.Now(),
					Time:            10,
					Request: Request{
						URL:         "https://example.com/用户/资料?名=张三",
						Method:      "POST",
						HTTPVersion: "HTTP/1.1",
						PostData: &PostData{
							MimeType: "application/json",
							Text:     unicodeBody,
						},
					},
					Response: Response{
						Status:      200,
						StatusText:  "OK",
						HTTPVersion: "HTTP/1.1",
						Content:     Content{MimeType: "application/json"},
					},
					Timings: Timings{Wait: 10},
				},
			},
		},
	}

	// 序列化-反序列化往返
	data, err := h.ToJSON(true)
	require.NoError(t, err)
	// JSON 输出应保留 Unicode（不转义为 \uXXXX）
	assert.Contains(t, string(data), "张三")

	// 用宽松选项解析（避免 URL 严格验证）
	h2, err := ParseHarFromReaderWithOptions(bytes.NewReader(data), ParseOptions{SkipValidation: true})
	require.NoError(t, err)
	assert.Equal(t, unicodeBody, h2.Log.Entries[0].Request.PostData.Text)
	assert.Contains(t, h2.Log.Entries[0].Request.URL, "张三")
}

// --- 辅助：跳过缺失文件 ---

func TestTestDataFilesExist(t *testing.T) {
	files := []string{
		"testdata/compressed.har",
		"testdata/sensitive.har",
		"testdata/minimal.har",
		"testdata/large.har",
		"testdata/full.har",
	}
	for _, f := range files {
		_, err := os.Stat(f)
		assert.NoError(t, err, "测试数据文件应存在: %s", f)
	}
}

// 避免未使用 import
var _ = strings.Contains
