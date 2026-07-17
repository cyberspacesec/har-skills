package har

import (
	"io"
	"net/http"
	"strings"
)

// EntryMeta 携带 HAR 条目的可选元数据。
// 上层测绘系统在持有 *http.Request / *http.Response 之外，往往还能拿到
// 服务器 IP、连接 ID、所属页面引用等额外信息，这些无法从 req/resp 直接推断，
// 通过本结构传入 AddEntryFromHTTPWithMeta。
type EntryMeta struct {
	// ServerIPAddress 请求对端的服务器 IP（HAR 字段 serverIPAddress）。
	ServerIPAddress string
	// Connection 连接标识，用于关联复用同一连接的条目（HAR 字段 connection）。
	Connection string
	// Pageref 所属页面引用，需与 HarBuilder.AddPage 注册的 id 对应。
	Pageref string
	// InitiatorType / InitiatorURL / InitiatorLine 描述请求发起来源
	//（Chrome 扩展字段 _initiator），如 "script"/"parser"/"other"。
	InitiatorType string
	InitiatorURL  string
	InitiatorLine int
	// Priority 资源优先级（Chrome 扩展字段 _priority），如 "High"/"Low"。
	Priority string
	// ResourceType 资源类型（Chrome 扩展字段 _resourceType），如 "xhr"/"script"。
	ResourceType string
	// Comment 条目注释。
	Comment string
}

// HeadersFromHTTP 把 net/http 的请求/响应头转换为 HAR 的 []Headers。
// 每个 header 值（含多值）展开为独立的 Headers 条目，保留原始大小写。
func HeadersFromHTTP(h http.Header) []Headers {
	if h == nil {
		return []Headers{}
	}
	out := make([]Headers, 0, len(h))
	for key, values := range h {
		for _, value := range values {
			out = append(out, Headers{Name: key, Value: value})
		}
	}
	return out
}

// CookiesFromHTTP 把 net/http 的 Cookie 切片转换为 HAR 的 []Cookie。
// 仅填充能从 http.Cookie 直接获得的字段：Name/Value/Path/Domain/HTTPOnly/Secure。
// Expires/SameSite 在 http.Cookie 中存在但此处不推断，如需可由调用方后置设置。
func CookiesFromHTTP(cookies []*http.Cookie) []Cookie {
	if len(cookies) == 0 {
		return []Cookie{}
	}
	out := make([]Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		out = append(out, Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
		})
	}
	return out
}

// PostDataFromRequest 读取并构造请求体。
// 会消费 req.Body（io.ReadAll 后 Close），因此调用方若仍需 body 必须自行缓存副本。
// 返回的 PostData 可能为 nil（无 body 或读取失败时）。
// 第二个返回值为 body 字节数，用于设置 Request.BodySize。
//
// 自动识别 Content-Type：
//   - application/x-www-form-urlencoded → PostData.Params 解析为表单参数，Text 留空
//   - 其他 → PostData.Text 存原始字符串
//
// mimeType 为空时回退为 req.Header.Get("Content-Type")。
func PostDataFromRequest(req *http.Request) (*PostData, int) {
	if req == nil || isNilReader(req.Body) {
		return nil, 0
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil || len(bodyBytes) == 0 {
		return nil, 0
	}
	contentType := req.Header.Get("Content-Type")

	var postData *PostData
	if strings.HasPrefix(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
		// 表单参数解析：保留原始值，与 form 行为一致但走独立实现避免改写 req.Form
		params := parseFormParams(string(bodyBytes))
		postData = &PostData{
			MimeType: contentType,
			Params:   params,
		}
	} else {
		postData = &PostData{
			MimeType: contentType,
			Text:     string(bodyBytes),
		}
	}
	return postData, len(bodyBytes)
}

// parseFormParams 解析 application/x-www-form-urlencoded 体为 []Param。
func parseFormParams(body string) []Param {
	if body == "" {
		return []Param{}
	}
	out := make([]Param, 0)
	for _, pair := range strings.Split(body, "&") {
		if pair == "" {
			continue
		}
		key, value, found := strings.Cut(pair, "=")
		if !found {
			out = append(out, Param{Name: key})
			continue
		}
		out = append(out, Param{Name: key, Value: value})
	}
	return out
}

// isTextContentType 粗略判断 Content-Type 是否为文本，用于决定 body 是否需 base64 编码。
// 非文本（图片/音频/视频/字体/二进制 octet-stream）应走 base64 以保证 JSON 往返无损。
func isTextContentType(mimeType string) bool {
	m := strings.ToLower(strings.TrimSpace(mimeType))
	if m == "" {
		return true // 未知类型按文本处理，向后兼容旧行为
	}
	switch {
	case strings.HasPrefix(m, "text/"),
		strings.Contains(m, "json"),
		strings.Contains(m, "xml"),
		strings.Contains(m, "javascript"),
		strings.Contains(m, "urlencoded"),
		strings.Contains(m, "form-data"),
		strings.HasPrefix(m, "application/"):
		// application/* 中除图片/字体等二进制外多为文本，这里对 application 子类型
		// 再排除已知二进制族，命中则按文本。
		if strings.Contains(m, "image") ||
			strings.Contains(m, "audio") ||
			strings.Contains(m, "video") ||
			strings.Contains(m, "font") ||
			strings.Contains(m, "octet-stream") ||
			strings.Contains(m, "pdf") ||
			strings.Contains(m, "zip") ||
			strings.Contains(m, "gzip") {
			return false
		}
		return true
	}
	// image/* audio/* video/* font/* 等顶层类型按二进制
	return false
}
