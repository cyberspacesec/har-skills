package har

import (
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// RedactURLRule defines a rule for redacting URL path segments.
type RedactURLRule struct {
	Pattern     string // regex pattern to match URL path segments
	Replacement string // replacement for matched segments
}

// RedactValuePattern defines a rule for redacting by *value content* rather than
// by field name. The regex is matched against the full string value of any
// string field (header/cookie/query/post-data/JSON string value); a match
// triggers redaction regardless of which key it lives under.
//
// This catches secrets that hide in fields whose names are not on the
// name-based redaction list — e.g. a custom header "X-Trace-Id" whose value
// happens to be a Bearer token, or a JSON field "notes" containing a pasted
// API key.
type RedactValuePattern struct {
	// Name is a human label for the pattern, used in logs/debugging. Optional.
	Name string
	// Pattern is a regex matched (unanchored) against string field values.
	Pattern string
	// Replacement overrides the global Replacement for this pattern. If empty,
	// the global Replacement is used.
	Replacement string
}

// RedactOptions configures how sensitive data is redacted from a HAR file.
type RedactOptions struct {
	Headers        []string                                                 // header names to redact (case-insensitive)
	Cookies        []string                                                 // cookie names to redact (case-insensitive)
	QueryParams    []string                                                 // query parameter names to redact (case-insensitive)
	PostDataFields []string                                                 // POST form field names to redact (case-insensitive)
	Replacement    string                                                   // replacement text (default: "[REDACTED]")
	RedactIPs      bool                                                     // whether to anonymize IP addresses
	RedactURLs     []RedactURLRule                                          // URL path segment redaction rules
	CustomRedactor func(fieldType string, name string, value string) string // custom redaction function
	// ValuePatterns redacts by *value content* regardless of field name.
	// Applied to every string field after name-based redaction. A value that
	// matches any pattern is replaced. See RedactValuePattern for details.
	ValuePatterns []RedactValuePattern
}

// DefaultRedactOptions returns a RedactOptions with sensible defaults for
// common sensitive fields found in HTTP traffic.
func DefaultRedactOptions() RedactOptions {
	return RedactOptions{
		Headers: []string{
			"Authorization",
			"Proxy-Authorization",
			"WWW-Authenticate",
			"Cookie",
			"Set-Cookie",
			"X-Api-Key",
			"X-Auth-Token",
			"X-CSRF-Token",
		},
		Cookies: []string{
			"session",
			"token",
			"auth",
			"password",
			"secret",
			"api_key",
			"access_token",
			"refresh_token",
		},
		QueryParams: []string{
			"password",
			"token",
			"api_key",
			"secret",
			"access_token",
			"refresh_token",
			"private_key",
			"client_secret",
		},
		PostDataFields: []string{
			"password",
			"token",
			"api_key",
			"secret",
			"access_token",
			"refresh_token",
			"private_key",
			"client_secret",
		},
		Replacement: "[REDACTED]",
		RedactIPs:   false,
		// 默认按值内容脱敏：捕获藏在任意字段里的常见密钥格式，
		// 不依赖字段名匹配（应对自定义 header / JSON 任意键）。
		ValuePatterns: DefaultRedactValuePatterns(),
	}
}

// DefaultRedactValuePatterns returns a set of common secret-shaped patterns
// for use as RedactOptions.ValuePatterns. Each matches a recognizable secret
// format (Bearer/JWT tokens, long hex/base64 secrets, AWS-style keys, etc.)
// so that secrets hiding under arbitrary field names still get redacted.
//
// Patterns are intentionally conservative: minimum lengths are set high
// enough to avoid flagging ordinary short strings.
func DefaultRedactValuePatterns() []RedactValuePattern {
	return []RedactValuePattern{
		{Name: "bearer-token", Pattern: `(?i)\bBearer\s+[A-Za-z0-9\-\._~+/]+=*`},
		{Name: "jwt", Pattern: `\beyJ[A-Za-z0-9_\-=]+\.[A-Za-z0-9_\-=]+\.?[A-Za-z0-9_\-=]*`},
		{Name: "aws-access-key", Pattern: `\bAKIA[0-9A-Z]{16}\b`},
		{Name: "aws-secret", Pattern: `\b[A-Za-z0-9/+=]{40}\b`},
		{Name: "github-pat", Pattern: `\bgh[pousr]_[A-Za-z0-9]{36,}\b`},
		{Name: "slack-token", Pattern: `\bxox[bp]-[A-Za-z0-9-]+\b`},
		{Name: "google-api-key", Pattern: `\bAIza[0-9A-Za-z\-_]{35}\b`},
		{Name: "hex-secret", Pattern: `\b[0-9a-fA-F]{64}\b`},
	}
}

// Redact returns a new Har with sensitive data redacted.
// It deep-clones the Har first, then redacts the clone so the original is unchanged.
func (h *Har) Redact(opts RedactOptions) *Har {
	if h == nil {
		return nil
	}

	clone := h.Clone()
	if clone == nil {
		return nil
	}
	clone.RedactInPlace(opts)
	return clone
}

// RedactInPlace mutates the Har in place, redacting sensitive data without cloning.
func (h *Har) RedactInPlace(opts RedactOptions) {
	if h == nil {
		return
	}

	replacement := opts.Replacement
	if replacement == "" {
		replacement = "[REDACTED]"
	}

	// 预编译值模式正则，避免每个字段重复编译
	valueRes := compileValuePatterns(opts.ValuePatterns)

	for i := range h.Log.Entries {
		entry := &h.Log.Entries[i]

		// Redact request headers
		for j := range entry.Request.Headers {
			header := &entry.Request.Headers[j]
			if matchesAny(header.Name, opts.Headers) {
				header.Value = redactHeaderValue(header.Name, header.Value, opts, replacement)
			} else if header.Value != "" {
				header.Value = applyValuePatterns(header.Value, valueRes, replacement)
			}
		}

		// Redact response headers
		for j := range entry.Response.Headers {
			header := &entry.Response.Headers[j]
			if matchesAny(header.Name, opts.Headers) {
				header.Value = redactHeaderValue(header.Name, header.Value, opts, replacement)
			} else if header.Value != "" {
				header.Value = applyValuePatterns(header.Value, valueRes, replacement)
			}
		}

		// Redact request cookies
		for j := range entry.Request.Cookies {
			cookie := &entry.Request.Cookies[j]
			if matchesAny(cookie.Name, opts.Cookies) {
				cookie.Value = redactCookieValue(cookie.Name, cookie.Value, opts, replacement)
			} else if cookie.Value != "" {
				cookie.Value = applyValuePatterns(cookie.Value, valueRes, replacement)
			}
		}

		// Redact response cookies
		for j := range entry.Response.Cookies {
			cookie := &entry.Response.Cookies[j]
			if matchesAny(cookie.Name, opts.Cookies) {
				cookie.Value = redactCookieValue(cookie.Name, cookie.Value, opts, replacement)
			} else if cookie.Value != "" {
				cookie.Value = applyValuePatterns(cookie.Value, valueRes, replacement)
			}
		}

		// Redact query string parameters
		for j := range entry.Request.QueryString {
			qs := &entry.Request.QueryString[j]
			if matchesAny(qs.Name, opts.QueryParams) {
				qs.Value = redactQueryParamValue(qs.Name, qs.Value, opts, replacement)
			} else if qs.Value != "" {
				qs.Value = applyValuePatterns(qs.Value, valueRes, replacement)
			}
		}

		// Redact URL (query params in URL string and path segment rules)
		if entry.Request.URL != "" {
			entry.Request.URL = redactURLString(entry.Request.URL, opts, replacement, valueRes)
		}

		// Redact POST data
		if entry.Request.PostData != nil {
			pd := entry.Request.PostData
			// Redact POST params
			for j := range pd.Params {
				param := &pd.Params[j]
				if matchesAny(param.Name, opts.PostDataFields) {
					param.Value = redactPostDataFieldValue(param.Name, param.Value, opts, replacement)
				} else if param.Value != "" {
					param.Value = applyValuePatterns(param.Value, valueRes, replacement)
				}
			}
			// Redact POST text body (form key=value 或 JSON)
			if pd.Text != "" {
				pd.Text = redactPostDataText(pd.Text, opts, replacement, valueRes)
			}
		}

		// Anonymize IP addresses
		if opts.RedactIPs && entry.ServerIPAddress != "" {
			entry.ServerIPAddress = anonymizeIP(entry.ServerIPAddress)
		}
	}
}

// matchesAny checks whether name matches any of the patterns (case-insensitive).
func matchesAny(name string, patterns []string) bool {
	nameLower := strings.ToLower(name)
	for _, p := range patterns {
		if strings.ToLower(p) == nameLower {
			return true
		}
	}
	return false
}

// compiledValuePattern 是预编译的 RedactValuePattern 正则。
type compiledValuePattern struct {
	re          *regexp.Regexp
	replacement string // per-pattern override, empty means use global
}

// compileValuePatterns 预编译值模式正则，避免每个字段重复编译。
// 若 Pattern 非法则跳过（不中断整体脱敏）。
func compileValuePatterns(patterns []RedactValuePattern) []compiledValuePattern {
	if len(patterns) == 0 {
		return nil
	}
	out := make([]compiledValuePattern, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p.Pattern)
		if err != nil {
			continue // 跳过非法正则
		}
		out = append(out, compiledValuePattern{re: re, replacement: p.Replacement})
	}
	return out
}

// applyValuePatterns 对字符串值应用所有值模式正则。
// 第一个匹配的 pattern 生效（短路后续）。返回替换后的字符串；
// 无匹配则返回原值。
func applyValuePatterns(value string, patterns []compiledValuePattern, globalReplacement string) string {
	for _, p := range patterns {
		if p.re.MatchString(value) {
			repl := p.replacement
			if repl == "" {
				repl = globalReplacement
			}
			return p.re.ReplaceAllString(value, repl)
		}
	}
	return value
}

// redactHeaderValue redacts a header value.
func redactHeaderValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("header", name, value)
	}
	return replacement
}

// redactCookieValue redacts a cookie value.
func redactCookieValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("cookie", name, value)
	}
	return replacement
}

// redactQueryParamValue redacts a query parameter value.
func redactQueryParamValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("queryparam", name, value)
	}
	return replacement
}

// redactPostDataFieldValue redacts a POST form field value.
func redactPostDataFieldValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("postdatafield", name, value)
	}
	return replacement
}

// redactPostDataText redacts sensitive data in POST body text.
// It auto-detects JSON vs URL-encoded form bodies and dispatches to the
// appropriate redactor. JSON bodies are parsed and walked recursively so that
// non-string values (numbers, booleans, null, nested objects/arrays) are
// handled correctly — a sensitive key like "secret": 12345 is redacted even
// though its value is not a string.
func redactPostDataText(text string, opts RedactOptions, replacement string, valueRes []compiledValuePattern) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return text
	}

	// JSON body：以 { 或 [ 开头才尝试 JSON 解析（避免误把表单当 JSON）
	if (trimmed[0] == '{' || trimmed[0] == '[') && looksLikeJSON(trimmed) {
		if out, ok := redactJSONBody(text, opts, replacement, valueRes); ok {
			return out
		}
		// JSON 解析失败则回退到正则方式
	}

	// URL-encoded form bodies (key=value&key=value)
	if strings.Contains(text, "=") {
		return redactKeyValuePairs(text, opts, replacement, valueRes)
	}

	// 无法识别格式：仍跑一遍值模式，捕获明文密钥
	if len(valueRes) > 0 {
		return applyValuePatterns(text, valueRes, replacement)
	}
	return text
}

// looksLikeJSON 粗略判断字符串是否为合法 JSON：尝试 Unmarshal 到 interface{}。
func looksLikeJSON(s string) bool {
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}

// redactJSONBody 解析 JSON 文本，递归遍历所有键值对：
//   - 名字匹配 PostDataFields 的字段，其值（无论何种类型）整体替换为 replacement；
//   - 字符串值额外跑值模式匹配（捕获藏在任意键下的密钥）；
//   - 保留原始结构（对象/数组/嵌套）与缩进风格。
//
// 调用前已由 looksLikeJSON 保证 text 是合法 JSON，故 Unmarshal 必成功；
// data 是已解码的 JSON 值，Marshal/MarshalIndent 对其必成功。
// 第二个返回值表示是否成功解析并重写（false 时调用方应回退到正则方式）。
func redactJSONBody(text string, opts RedactOptions, replacement string, valueRes []compiledValuePattern) (string, bool) {
	var data interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		// looksLikeJSON 已过滤，理论不可达；保留防御
		return text, false
	}
	data = redactJSONValue(data, "", opts, replacement, valueRes)

	// 保持原始缩进：多行 JSON（含换行）用 MarshalIndent 重建
	if strings.Contains(text, "\n") {
		indent := detectJSONIndent(text)
		out, _ := json.MarshalIndent(data, "", indent)
		return string(out), true
	}

	// 单行 JSON：紧凑输出，再按原文是否带空格补回 ": " / ", " 风格
	out, _ := json.Marshal(data)
	if strings.Contains(text, ": ") || strings.Contains(text, ", ") {
		return prettifySingleLineJSON(string(out)), true
	}
	return string(out), true
}

// detectJSONIndent 探测多行 JSON 的缩进单位（默认两空格）。
func detectJSONIndent(text string) string {
	for _, line := range strings.Split(text, "\n") {
		// 跳过首行（无缩进）
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == line || trimmed == "" {
			continue
		}
		return line[:len(line)-len(trimmed)]
	}
	return "  "
}

// prettifySingleLineJSON 把紧凑 JSON {"k":"v","a":1} 补成 {"k": "v", "a": 1}。
// 仅作用于单行、无嵌套结构的简单场景；嵌套结构会被紧凑输出（功能仍正确，仅风格略变）。
func prettifySingleLineJSON(s string) string {
	s = strings.ReplaceAll(s, `":`, `": `)
	s = strings.ReplaceAll(s, `","`, `", "`)
	return s
}

// redactJSONValue 递归处理解码后的 JSON 值（map/slice/string/number/bool/nil）。
// parentKey 用于 CustomRedactor 上下文。
func redactJSONValue(data interface{}, parentKey string, opts RedactOptions, replacement string, valueRes []compiledValuePattern) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if matchesAny(key, opts.PostDataFields) {
				// 名字匹配：整体替换（不论值类型）
				if opts.CustomRedactor != nil {
					if s, ok := val.(string); ok {
						v[key] = opts.CustomRedactor("postdatafield", key, s)
					} else {
						v[key] = opts.CustomRedactor("postdatafield", key, fmtJSON(val))
					}
				} else {
					v[key] = replacement
				}
				continue
			}
			v[key] = redactJSONValue(val, key, opts, replacement, valueRes)
		}
		return v
	case []interface{}:
		for i := range v {
			v[i] = redactJSONValue(v[i], parentKey, opts, replacement, valueRes)
		}
		return v
	case string:
		// 字符串值跑值模式匹配
		if v != "" && len(valueRes) > 0 {
			return applyValuePatterns(v, valueRes, replacement)
		}
		return v
	default:
		// number/bool/nil：值模式不适用
		return v
	}
}

// fmtJSON 把任意值转成紧凑 JSON 字符串（供 CustomRedactor 接收非字符串值）。
// v 来自已解码的 JSON 值（number/bool/nil），Marshal 必成功。
func fmtJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// redactKeyValuePairs redacts sensitive fields in URL-encoded body text.
// Handles both key=value& and key=value at end of string.
func redactKeyValuePairs(text string, opts RedactOptions, replacement string, valueRes []compiledValuePattern) string {
	// Match key=value patterns (URL-encoded or plain)
	re := regexp.MustCompile(`([^&=]+)=([^&]*)`)
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			key := parts[1]
			value := parts[2]
			if matchesAny(key, opts.PostDataFields) {
				if opts.CustomRedactor != nil {
					return key + "=" + opts.CustomRedactor("postdatafield", key, value)
				}
				return key + "=" + replacement
			}
			// 名字不匹配：仍对值跑值模式
			if value != "" && len(valueRes) > 0 {
				return key + "=" + applyValuePatterns(value, valueRes, replacement)
			}
		}
		return match
	})
	return result
}

// （redactJSONKeys 已删除：原先用正则匹配 "key":"value" 脱敏 JSON，
// 已由 redactJSONBody 取代，后者真正解析 JSON、支持非字符串值与嵌套。）

// anonymizeIP replaces the last octet of an IPv4 address with .0.
// For IPv6, it replaces the last segment with :0.
// If the string is not a valid IP, it returns the replacement text.
func anonymizeIP(ip string) string {
	// Try IPv4 first
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// Not a valid IP, just return the original
		return ip
	}

	if parsedIP.To4() != nil {
		// IPv4: replace last octet with 0
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return parts[0] + "." + parts[1] + "." + parts[2] + ".0"
		}
	}

	// IPv6: replace last hextet with :0
	// net.IP.String() 对 IPv6 地址必含 ":"，故 LastIndex 必 >= 0。
	str := parsedIP.String()
	lastColon := strings.LastIndex(str, ":")
	return str[:lastColon] + ":0"
}

// redactURLString redacts sensitive query parameters in a URL string
// and applies URL path segment redaction rules.
func redactURLString(rawURL string, opts RedactOptions, replacement string, valueRes []compiledValuePattern) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// If URL can't be parsed, try simple string-based redaction
		return redactQueryStringSimple(rawURL, opts.QueryParams, replacement, opts, valueRes)
	}

	// Redact query parameters in the URL
	if parsed.RawQuery != "" {
		parsed.RawQuery = redactURLQuery(parsed.RawQuery, opts, replacement, valueRes)
	}

	// Apply URL path segment redaction rules
	if len(opts.RedactURLs) > 0 {
		parsed.Path = redactURLPath(parsed.Path, opts.RedactURLs)
	}

	return parsed.String()
}

// redactURLQuery redacts sensitive query parameters in a URL query string.
func redactURLQuery(query string, opts RedactOptions, replacement string, valueRes []compiledValuePattern) string {
	params := strings.Split(query, "&")
	var result []string
	for _, param := range params {
		if param == "" {
			continue
		}
		parts := strings.SplitN(param, "=", 2)
		key := parts[0]
		if len(parts) == 2 {
			value := parts[1]
			if matchesAny(key, opts.QueryParams) {
				if opts.CustomRedactor != nil {
					result = append(result, key+"="+opts.CustomRedactor("queryparam", key, value))
				} else {
					result = append(result, key+"="+replacement)
				}
			} else if value != "" && len(valueRes) > 0 {
				// 名字不匹配：仍对值跑值模式
				result = append(result, key+"="+applyValuePatterns(value, valueRes, replacement))
			} else {
				result = append(result, param)
			}
		} else {
			// No value, just a key
			if matchesAny(key, opts.QueryParams) {
				if opts.CustomRedactor != nil {
					result = append(result, key+"="+opts.CustomRedactor("queryparam", key, ""))
				} else {
					result = append(result, key+"="+replacement)
				}
			} else {
				result = append(result, param)
			}
		}
	}
	return strings.Join(result, "&")
}

// redactURLPath applies URL path segment redaction rules.
func redactURLPath(path string, rules []RedactURLRule) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		for _, rule := range rules {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue
			}
			if re.MatchString(segment) {
				segments[i] = re.ReplaceAllString(segment, rule.Replacement)
			}
		}
	}
	return strings.Join(segments, "/")
}

// redactQueryStringSimple redacts query parameters in a raw URL string
// without fully parsing it. Used as a fallback when url.Parse fails.
func redactQueryStringSimple(rawURL string, paramNames []string, replacement string, opts RedactOptions, valueRes []compiledValuePattern) string {
	for _, name := range paramNames {
		// Match name=value pattern
		re := regexp.MustCompile(`(?i)(` + regexp.QuoteMeta(name) + `)=([^&]*)`)
		rawURL = re.ReplaceAllStringFunc(rawURL, func(match string) string {
			// match 由正则匹配产生，FindStringSubmatch 必返回 3 个组
			// (完整匹配 + 2 个捕获组)。
			parts := re.FindStringSubmatch(match)
			if opts.CustomRedactor != nil {
				return parts[1] + "=" + opts.CustomRedactor("queryparam", parts[1], parts[2])
			}
			return parts[1] + "=" + replacement
		})
	}
	// 名字不匹配的参数仍跑值模式
	if len(valueRes) > 0 {
		re := regexp.MustCompile(`([^&=]+)=([^&]*)`)
		rawURL = re.ReplaceAllStringFunc(rawURL, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) >= 3 && parts[2] != "" {
				return parts[1] + "=" + applyValuePatterns(parts[2], valueRes, replacement)
			}
			return match
		})
	}
	return rawURL
}
