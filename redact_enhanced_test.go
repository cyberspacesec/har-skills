package har

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- 非字符串 JSON 值脱敏（Task 19）---

func TestRedactJSONNonStringValues(t *testing.T) {
	// JSON 里敏感键的值是数字、布尔、null、嵌套对象——都应被整体替换
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Method: "POST",
						PostData: &PostData{
							MimeType: "application/json",
							Text:     `{"secret_code": 12345, "is_admin": true, "hidden": null, "nested": {"token": "abc123"}}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"secret_code", "is_admin", "hidden", "token"},
		Replacement:    "[REDACTED]",
		ValuePatterns:  nil, // 禁用默认值模式，只测名字匹配
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	// 数字、布尔、null 值被整体替换为字符串 "[REDACTED]"
	assert.Contains(t, text, `"secret_code": "[REDACTED]"`)
	assert.Contains(t, text, `"is_admin": "[REDACTED]"`)
	assert.Contains(t, text, `"hidden": "[REDACTED]"`)
	// 嵌套对象里的 token 也脱敏
	assert.Contains(t, text, `"token": "[REDACTED]"`)
}

func TestRedactJSONArrayValues(t *testing.T) {
	// 敏感键的值是数组
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `{"tokens": ["abc", "def"], "ids": [1, 2, 3]}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"tokens", "ids"},
		Replacement:    "***",
		ValuePatterns:  nil,
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	assert.Contains(t, text, `"tokens": "***"`)
	assert.Contains(t, text, `"ids": "***"`)
}

// --- 值模式脱敏（按内容而非名字）---

func TestRedactValuePatternsBearerToken(t *testing.T) {
	// Bearer token 藏在自定义 header 值里
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "X-Custom-Auth", Value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.abc.def"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:       []string{}, // 不按名字匹配
		Replacement:   "[REDACTED]",
		ValuePatterns: DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)
	headerVal := result.Log.Entries[0].Request.Headers[0].Value

	assert.Equal(t, "[REDACTED]", headerVal)
}

func TestRedactValuePatternsInCookie(t *testing.T) {
	// cookie 值里藏 JWT
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Cookies: []Cookie{
							{Name: "session", Value: "eyJhbGciOiJIUzI1NiJ9.payload.sig"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Cookies:       []string{}, // 不按名字匹配
		Replacement:   "***",
		ValuePatterns: DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)
	cookieVal := result.Log.Entries[0].Request.Cookies[0].Value

	assert.Equal(t, "***", cookieVal)
}

func TestRedactValuePatternsInQueryParam(t *testing.T) {
	// query param 值是 AWS access key
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						URL: "https://api.example.com/data?key=AKIAIOSFODNN7EXAMPLE",
					},
				},
			},
		},
	}

	opts := RedactOptions{
		QueryParams:   []string{}, // 不按名字匹配
		Replacement:   "[SCRUBBED]",
		ValuePatterns: DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)

	assert.Contains(t, result.Log.Entries[0].Request.URL, "key=[SCRUBBED]")
}

func TestRedactValuePatternsInJSONBody(t *testing.T) {
	// JSON body 里有个字段值是 GitHub PAT，字段名不在 PostDataFields 列表里
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `{"note": "use this token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx for testing"}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{}, // 不按名字匹配
		Replacement:    "[REDACTED]",
		ValuePatterns:  DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	assert.Contains(t, text, "[REDACTED]")
	assert.NotContains(t, text, "ghp_")
}

func TestRedactValuePatternsNoMatch(t *testing.T) {
	// 正常值不触发值模式
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "X-Request-Id", Value: "abc-123-def"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:       []string{},
		Replacement:   "[REDACTED]",
		ValuePatterns: DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)

	// 正常 request-id 不被脱敏
	assert.Equal(t, "abc-123-def", result.Log.Entries[0].Request.Headers[0].Value)
}

func TestRedactValuePatternsCustomReplacement(t *testing.T) {
	// 单独 pattern 有自定义 replacement
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "X-Token", Value: "Bearer secret-token-here"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Replacement: "[GLOBAL]",
		ValuePatterns: []RedactValuePattern{
			{Name: "bearer", Pattern: `(?i)\bBearer\s+\S+`, Replacement: "[TOKEN_SCRUBBED]"},
		},
	}
	result := h.Redact(opts)

	assert.Equal(t, "[TOKEN_SCRUBBED]", result.Log.Entries[0].Request.Headers[0].Value)
}

func TestRedactValuePatternsDisabled(t *testing.T) {
	// 显式禁用值模式（ValuePatterns=nil 或空）
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "X-Auth", Value: "Bearer secret"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:       []string{}, // 不按名字匹配
		Replacement:   "[REDACTED]",
		ValuePatterns: nil, // 禁用
	}
	result := h.Redact(opts)

	// 禁用值模式时不脱敏（名字也不匹配）
	assert.Equal(t, "Bearer secret", result.Log.Entries[0].Request.Headers[0].Value)
}

// --- 边界：JSON body 无空格紧凑格式保留 ---

func TestRedactJSONBodyCompactFormat(t *testing.T) {
	// 输入无空格，输出也应无空格
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `{"password":"hunter2","username":"bob"}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"password"},
		Replacement:    "[REDACTED]",
		ValuePatterns:  nil,
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	assert.Contains(t, text, `"password":"[REDACTED]"`)
	assert.NotContains(t, text, `"password": "[REDACTED]"`) // 不应有空格
}

func TestRedactJSONBodyPrettifiedFormat(t *testing.T) {
	// 输入多行带缩进，输出应保持多行
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `{
  "password": "hunter2",
  "username": "bob"
}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"password"},
		Replacement:    "***",
		ValuePatterns:  nil,
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	require.Contains(t, text, `"password": "***"`)
	require.Contains(t, text, "\n") // 保持多行
}

// --- 边界：非法值模式正则跳过不中断脱敏 ---

func TestRedactInvalidValuePatternSkipped(t *testing.T) {
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "X-Token", Value: "Bearer secret"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Replacement: "[REDACTED]",
		ValuePatterns: []RedactValuePattern{
			{Name: "bad", Pattern: `[`}, // 非法正则
			{Name: "bearer", Pattern: `(?i)Bearer\s+\S+`},
		},
	}
	result := h.Redact(opts)

	// 非法 pattern 跳过，合法 pattern 生效
	assert.Equal(t, "[REDACTED]", result.Log.Entries[0].Request.Headers[0].Value)
}

// --- 边界：名字匹配优先于值模式 ---

func TestRedactNameMatchOverridesValuePattern(t *testing.T) {
	// 名字匹配触发整值替换，值模式不再跑（已替换为 replacement）
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer secret"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:       []string{"Authorization"},
		Replacement:   "***",
		ValuePatterns: DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)

	// 名字匹配触发，CustomRedactor 未设，整体替换为 ***
	assert.Equal(t, "***", result.Log.Entries[0].Request.Headers[0].Value)
}

// --- 覆盖率：CustomRedactor 处理非字符串 JSON 值 ---

func TestRedactCustomRedactorNonStringValue(t *testing.T) {
	// CustomRedactor 接收非字符串值（数字/布尔/null）时走 fmtJSON 分支
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `{"secret_code": 12345, "is_admin": true, "hidden": null}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"secret_code", "is_admin", "hidden"},
		ValuePatterns:  nil,
		CustomRedactor: func(fieldType, name, value string) string {
			return "[C:" + name + ":" + value + "]"
		},
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	// 非字符串值经 fmtJSON 转成字符串后传给 CustomRedactor
	assert.Contains(t, text, `"secret_code": "[C:secret_code:12345]"`)
	assert.Contains(t, text, `"is_admin": "[C:is_admin:true]"`)
	assert.Contains(t, text, `"hidden": "[C:hidden:null]"`)
}

// --- 覆盖率：JSON body 解析失败回退到正则 ---

func TestRedactJSONBodyFallbackToRegex(t *testing.T) {
	// 文本含 = 但不是合法 JSON，走 redactKeyValuePairs；token 作为键被脱敏
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `token=secret_value&keep=1`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"token"},
		Replacement:    "[R]",
		ValuePatterns:  nil,
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text
	// token 字段被脱敏
	assert.Contains(t, text, "token=[R]")
	assert.Contains(t, text, "keep=1")
}

// --- 覆盖率：JSON body 多行缩进探测 ---

func TestRedactJSONBodyFourSpaceIndent(t *testing.T) {
	// 4 空格缩进的多行 JSON
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: "{\n    \"password\": \"hunter2\"\n}",
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"password"},
		Replacement:    "***",
		ValuePatterns:  nil,
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	assert.Contains(t, text, `"password": "***"`)
	assert.Contains(t, text, "\n") // 保持多行
}

// --- 覆盖率：redactQueryStringSimple 的 valueRes 分支 ---

func TestRedactQueryStringSimpleValuePatternFallback(t *testing.T) {
	// url.Parse 会成功的 URL 不走 simple fallback；这里直接调用 simple 函数
	out := redactQueryStringSimple(
		"https://example.com/?trace=Bearer%20secret",
		[]string{}, // 不按名字匹配
		"[X]",
		RedactOptions{},
		compileValuePatterns(DefaultRedactValuePatterns()),
	)
	// 值模式匹配 Bearer，被替换
	assert.Contains(t, out, "trace=")
}

// --- 覆盖率：redactKeyValuePairs 的值模式分支 ---

func TestRedactKeyValuePairsValuePattern(t *testing.T) {
	// 表单里某字段值藏 Bearer token（明文，非 URL 编码），但字段名不在 PostDataFields
	text := `note=Bearer eyJhbGciOiJIUzI1NiJ9.abc.def&keep=1`
	opts := RedactOptions{
		PostDataFields: []string{},
		Replacement:    "[R]",
		ValuePatterns:  DefaultRedactValuePatterns(),
	}
	result := redactKeyValuePairs(text, opts, "[R]", compileValuePatterns(opts.ValuePatterns))
	// note 的值被值模式脱敏
	assert.Contains(t, result, "note=")
	assert.NotContains(t, result, "Bearer eyJ")
	assert.Contains(t, result, "keep=1")
}

// --- 覆盖率：JSON 数组内对象的字符串值跑值模式 ---

func TestRedactJSONValuePatternInArray(t *testing.T) {
	// JSON 数组里每个对象的字符串值都跑值模式
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `{"items":[{"note":"Bearer token1"},{"note":"Bearer token2"}]}`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{}, // 不按名字
		Replacement:    "[R]",
		ValuePatterns:  DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text

	assert.NotContains(t, text, "Bearer token1")
	assert.NotContains(t, text, "Bearer token2")
	assert.Contains(t, text, "[R]")
}

// --- 边界：CustomRedactor 与值模式共存 ---

func TestRedactCustomRedactorWithValuePatterns(t *testing.T) {
	// CustomRedactor 用于名字匹配；值模式仍作用于名字不匹配的字符串值
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						Headers: []Headers{
							{Name: "Authorization", Value: "Bearer named-match"},
							{Name: "X-Custom", Value: "Bearer value-pattern-match"},
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		Headers:       []string{"Authorization"},
		Replacement:   "[REDACTED]",
		ValuePatterns: DefaultRedactValuePatterns(),
		CustomRedactor: func(fieldType, name, value string) string {
			return "[CUSTOM:" + name + "]"
		},
	}
	result := h.Redact(opts)

	// 名字匹配走 CustomRedactor
	assert.Equal(t, "[CUSTOM:Authorization]", result.Log.Entries[0].Request.Headers[0].Value)
	// 名字不匹配但值模式命中，走值模式（用 global replacement）
	assert.Equal(t, "[REDACTED]", result.Log.Entries[0].Request.Headers[1].Value)
}

// --- 覆盖率：detectJSONIndent 无缩进回退默认两空格 ---

func TestRedactJSONBodyMultilineNoIndent(t *testing.T) {
	// 多行 JSON 但各行无前置空格——detectJSONIndent 遍历后回退默认 "  "
	// （json.Unmarshal 能解析这种不规范但合法的多行紧凑 JSON）
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: "{\n\"password\":\"x\"\n}",
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{"password"},
		Replacement:    "[R]",
		ValuePatterns:  nil,
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text
	// 走 MarshalIndent（含换行），password 被脱敏
	assert.Contains(t, text, "[R]")
	assert.Contains(t, text, "\n")
}

// --- 覆盖率：redactPostDataText 无 = 非 JSON 的值模式分支 ---

func TestRedactPostDataTextPlainValuePattern(t *testing.T) {
	// 既不是 JSON 也不含 =，但有值模式命中——走最后的值模式分支
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{
							Text: `raw text with Bearer eyJhbGciOiJIUzI1NiJ9.abc.def inside`,
						},
					},
				},
			},
		},
	}

	opts := RedactOptions{
		PostDataFields: []string{},
		Replacement:    "[R]",
		ValuePatterns:  DefaultRedactValuePatterns(),
	}
	result := h.Redact(opts)
	text := result.Log.Entries[0].Request.PostData.Text
	assert.NotContains(t, text, "Bearer eyJ")
	assert.Contains(t, text, "[R]")
}

// --- 覆盖率：redactPostDataText 空文本与非识别格式分支 ---

func TestRedactPostDataTextEmpty(t *testing.T) {
	// 空白文本直接返回
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{Text: "   "},
					},
				},
			},
		},
	}
	opts := RedactOptions{Replacement: "[R]", ValuePatterns: DefaultRedactValuePatterns()}
	result := h.Redact(opts)
	assert.Equal(t, "   ", result.Log.Entries[0].Request.PostData.Text)
}

func TestRedactPostDataTextUnrecognizedNoValuePattern(t *testing.T) {
	// 非 JSON、无 =、无值模式——原样返回
	h := &Har{
		Log: Log{
			Entries: []Entries{
				{
					Request: Request{
						PostData: &PostData{Text: "just plain text no equals"},
					},
				},
			},
		},
	}
	opts := RedactOptions{Replacement: "[R]", ValuePatterns: nil}
	result := h.Redact(opts)
	assert.Equal(t, "just plain text no equals", result.Log.Entries[0].Request.PostData.Text)
}

// --- 覆盖率：redactQueryStringSimple 空值参数 ---

func TestRedactQueryStringSimpleEmptyValue(t *testing.T) {
	// 参数值为空时，值模式分支应 return match（不替换）
	out := redactQueryStringSimple(
		"https://example.com/?key=&keep=1",
		[]string{},
		"[R]",
		RedactOptions{},
		compileValuePatterns(DefaultRedactValuePatterns()),
	)
	// key= 为空值，不被替换
	assert.Contains(t, out, "key=&")
	assert.Contains(t, out, "keep=1")
}
