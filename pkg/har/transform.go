package har

import (
	"net/url"
	"regexp"
	"strings"
)

// TransformType 定义转换规则的类型
type TransformType int

const (
	TransformURLRewrite       TransformType = iota // URL重写
	TransformHostReplace                            // 主机替换
	TransformSchemeChange                           // 协议变更
	TransformHeaderAdd                              // 添加请求头
	TransformHeaderRemove                           // 删除请求头
	TransformHeaderReplace                          // 替换请求头
	TransformQueryParamRemove                       // 删除查询参数
	TransformQueryParamAdd                          // 添加查询参数
	TransformCookieDomainRewrite                    // Cookie域重写
	TransformBodyReplace                            // 请求体替换
)

// TransformRule 定义一条转换规则
type TransformRule struct {
	Type         TransformType // 转换类型
	Pattern      string        // 正则或字符串匹配模式
	Replacement  string        // 替换字符串
	HeaderName   string        // 用于请求头转换的头部名称
	HeaderValue  string        // 用于添加/替换请求头的头部值
}

// Transform 对HAR应用转换规则，返回新的Har对象（克隆+转换）
func (h *Har) Transform(rules []TransformRule) *Har {
	if h == nil {
		return nil
	}

	cloned := h.Clone()
	cloned.TransformInPlace(rules)
	return cloned
}

// TransformInPlace 对HAR原地应用转换规则
func (h *Har) TransformInPlace(rules []TransformRule) {
	if h == nil || len(rules) == 0 {
		return
	}

	for i := range h.Log.Entries {
		applyRules(&h.Log.Entries[i], rules)
	}
}

// RewriteURL 便捷方法：替换URL前缀，返回新的Har对象
// 例如：RewriteURL("http://localhost:8080", "https://prod.example.com")
func (h *Har) RewriteURL(from, to string) *Har {
	return h.Transform([]TransformRule{
		{
			Type:        TransformURLRewrite,
			Pattern:     from,
			Replacement: to,
		},
	})
}

// RemoveHeaders 从所有请求和响应中移除指定的头部，返回新的Har对象
func (h *Har) RemoveHeaders(names []string) *Har {
	rules := make([]TransformRule, len(names))
	for i, name := range names {
		rules[i] = TransformRule{
			Type:       TransformHeaderRemove,
			HeaderName: name,
		}
	}
	return h.Transform(rules)
}

// AddHeaders 向所有请求和/或响应中添加头部，返回新的Har对象
// target 取值为 "request"、"response" 或 "both"
func (h *Har) AddHeaders(headers map[string]string, target string) *Har {
	var rules []TransformRule
	for name, value := range headers {
		rules = append(rules, TransformRule{
			Type:        TransformHeaderAdd,
			HeaderName:  name,
			HeaderValue: value,
		})
	}

	cloned := h.Clone()

	for i := range cloned.Log.Entries {
		entry := &cloned.Log.Entries[i]
		for _, rule := range rules {
			if target == "request" || target == "both" {
				entry.Request.Headers = append(entry.Request.Headers, Headers{
					Name:  rule.HeaderName,
					Value: rule.HeaderValue,
				})
			}
			if target == "response" || target == "both" {
				entry.Response.Headers = append(entry.Response.Headers, Headers{
					Name:  rule.HeaderName,
					Value: rule.HeaderValue,
				})
			}
		}
	}

	return cloned
}

// applyRules 对单个条目应用所有转换规则
func applyRules(entry *Entries, rules []TransformRule) {
	for _, rule := range rules {
		switch rule.Type {
		case TransformURLRewrite:
			applyURLRewrite(entry, rule)
		case TransformHostReplace:
			applyHostReplace(entry, rule)
		case TransformSchemeChange:
			applySchemeChange(entry, rule)
		case TransformHeaderAdd:
			applyHeaderAdd(entry, rule)
		case TransformHeaderRemove:
			applyHeaderRemove(entry, rule)
		case TransformHeaderReplace:
			applyHeaderReplace(entry, rule)
		case TransformQueryParamRemove:
			applyQueryParamRemove(entry, rule)
		case TransformQueryParamAdd:
			applyQueryParamAdd(entry, rule)
		case TransformCookieDomainRewrite:
			applyCookieDomainRewrite(entry, rule)
		case TransformBodyReplace:
			applyBodyReplace(entry, rule)
		}
	}
}

// applyURLRewrite 替换URL前缀
func applyURLRewrite(entry *Entries, rule TransformRule) {
	if strings.HasPrefix(entry.Request.URL, rule.Pattern) {
		newURL := rule.Replacement + entry.Request.URL[len(rule.Pattern):]
		entry.Request.URL = newURL

		// 更新QueryString（如果URL解析发生变化）
		entry.Request.QueryString = BuildQueryStringFromURL(newURL)

		// 更新Host请求头
		if u, err := url.Parse(newURL); err == nil {
			updateHostHeader(entry, u.Host)
		}
	}
}

// applyHostReplace 替换主机名
func applyHostReplace(entry *Entries, rule TransformRule) {
	if u, err := url.Parse(entry.Request.URL); err == nil {
		if u.Host == rule.Pattern {
			u.Host = rule.Replacement
			newURL := u.String()
			entry.Request.URL = newURL
			entry.Request.QueryString = BuildQueryStringFromURL(newURL)
			updateHostHeader(entry, rule.Replacement)
		}
	}
}

// applySchemeChange 变更协议（http <-> https）
func applySchemeChange(entry *Entries, rule TransformRule) {
	if u, err := url.Parse(entry.Request.URL); err == nil {
		if u.Scheme == rule.Pattern {
			u.Scheme = rule.Replacement
			newURL := u.String()
			entry.Request.URL = newURL
			entry.Request.QueryString = BuildQueryStringFromURL(newURL)
		}
	}
}

// applyHeaderAdd 添加请求头到请求和响应
func applyHeaderAdd(entry *Entries, rule TransformRule) {
	entry.Request.Headers = append(entry.Request.Headers, Headers{
		Name:  rule.HeaderName,
		Value: rule.HeaderValue,
	})
	entry.Response.Headers = append(entry.Response.Headers, Headers{
		Name:  rule.HeaderName,
		Value: rule.HeaderValue,
	})
}

// applyHeaderRemove 移除请求头（从请求和响应中）
func applyHeaderRemove(entry *Entries, rule TransformRule) {
	entry.Request.Headers = removeHeaderByName(entry.Request.Headers, rule.HeaderName)
	entry.Response.Headers = removeHeaderByName(entry.Response.Headers, rule.HeaderName)
}

// applyHeaderReplace 替换请求头的值
func applyHeaderReplace(entry *Entries, rule TransformRule) {
	replaceHeaderValue(entry.Request.Headers, rule.HeaderName, rule.HeaderValue)
	replaceHeaderValue(entry.Response.Headers, rule.HeaderName, rule.HeaderValue)
}

// applyQueryParamRemove 移除指定的查询参数
func applyQueryParamRemove(entry *Entries, rule TransformRule) {
	// If QueryString is empty, parse from URL
	if len(entry.Request.QueryString) == 0 {
		entry.Request.QueryString = BuildQueryStringFromURL(entry.Request.URL)
	}

	newQS := make([]QueryString, 0, len(entry.Request.QueryString))
	for _, q := range entry.Request.QueryString {
		if q.Name != rule.Pattern {
			newQS = append(newQS, q)
		}
	}
	entry.Request.QueryString = newQS

	// 重建URL
	rebuildURLFromQueryString(entry)
}

// applyQueryParamAdd 添加查询参数
func applyQueryParamAdd(entry *Entries, rule TransformRule) {
	// If QueryString is empty, parse from URL
	if len(entry.Request.QueryString) == 0 {
		entry.Request.QueryString = BuildQueryStringFromURL(entry.Request.URL)
	}

	entry.Request.QueryString = append(entry.Request.QueryString, QueryString{
		Name:  rule.HeaderName,
		Value: rule.HeaderValue,
	})

	// 重建URL
	rebuildURLFromQueryString(entry)
}

// applyCookieDomainRewrite 重写Cookie域
func applyCookieDomainRewrite(entry *Entries, rule TransformRule) {
	for i := range entry.Request.Cookies {
		if entry.Request.Cookies[i].Domain == rule.Pattern {
			entry.Request.Cookies[i].Domain = rule.Replacement
		}
	}
	for i := range entry.Response.Cookies {
		if entry.Response.Cookies[i].Domain == rule.Pattern {
			entry.Response.Cookies[i].Domain = rule.Replacement
		}
	}
}

// applyBodyReplace 替换请求体文本
func applyBodyReplace(entry *Entries, rule TransformRule) {
	if entry.Request.PostData == nil {
		return
	}
	if rule.Pattern == "" {
		return
	}

	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		// 不是合法正则，按普通字符串替换
		entry.Request.PostData.Text = strings.ReplaceAll(
			entry.Request.PostData.Text, rule.Pattern, rule.Replacement,
		)
		return
	}
	entry.Request.PostData.Text = re.ReplaceAllString(
		entry.Request.PostData.Text, rule.Replacement,
	)
}

// updateHostHeader 更新Host请求头
func updateHostHeader(entry *Entries, host string) {
	for i := range entry.Request.Headers {
		if strings.EqualFold(entry.Request.Headers[i].Name, "Host") {
			entry.Request.Headers[i].Value = host
			return
		}
	}
	// 如果没有Host头部，添加一个
	entry.Request.Headers = append(entry.Request.Headers, Headers{
		Name:  "Host",
		Value: host,
	})
}

// removeHeaderByName 按名称移除头部（不区分大小写）
func removeHeaderByName(headers []Headers, name string) []Headers {
	result := make([]Headers, 0, len(headers))
	for _, h := range headers {
		if !strings.EqualFold(h.Name, name) {
			result = append(result, h)
		}
	}
	return result
}

// replaceHeaderValue 替换指定名称的头部值（不区分大小写）
func replaceHeaderValue(headers []Headers, name, value string) {
	for i := range headers {
		if strings.EqualFold(headers[i].Name, name) {
			headers[i].Value = value
		}
	}
}

// rebuildURLFromQueryString 根据当前URL和QueryString重建URL
func rebuildURLFromQueryString(entry *Entries) {
	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		return
	}

	// 清除现有参数，使用QueryString中的值
	q := url.Values{}
	for _, qs := range entry.Request.QueryString {
		q.Set(qs.Name, qs.Value)
	}
	u.RawQuery = q.Encode()
	entry.Request.URL = u.String()
}

