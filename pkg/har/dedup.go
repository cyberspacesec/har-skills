package har

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// DedupStrategy 去重策略
type DedupStrategy int

const (
	DedupExactURL    DedupStrategy = iota // 精确URL匹配
	DedupURLPattern                       // 忽略指定参数的URL模式匹配
	DedupContentHash                      // 基于内容哈希
)

// DeduplicateOptions 去重选项
type DeduplicateOptions struct {
	Strategy       DedupStrategy // 去重策略
	IgnoreParams   []string      // 忽略的查询参数（缓存破坏器等）
	CompareHeaders bool          // 比较时是否包含头部
	CompareBody    bool          // 比较时是否包含请求体
}

// DuplicateGroup 表示一组重复请求
type DuplicateGroup struct {
	Key          string // 去重键（URL模式、哈希等）
	EntryIndices []int  // 重复条目的索引
	Count        int    // 重复数量
}

// DefaultDeduplicateOptions 返回默认的去重选项
// 使用DedupURLPattern策略，并忽略常见缓存破坏器参数
func DefaultDeduplicateOptions() DeduplicateOptions {
	return DeduplicateOptions{
		Strategy:     DedupURLPattern,
		IgnoreParams: defaultCacheBusterParams(),
	}
}

// defaultCacheBusterParams 返回常见缓存破坏器参数名列表
func defaultCacheBusterParams() []string {
	return []string{"_", "cb", "cachebuster", "timestamp", "t", "rand", "random", "v"}
}

// IsCacheBusterParam 检查参数名是否看起来像缓存破坏器
// 判断规则：参数名为常见缓存破坏器名称，或参数名为"v"且值为纯数字
func IsCacheBusterParam(name string) bool {
	lower := strings.ToLower(name)
	commonBusters := map[string]bool{
		"_":           true,
		"cb":          true,
		"cachebuster": true,
		"timestamp":   true,
		"t":           true,
		"rand":        true,
		"random":      true,
	}
	return commonBusters[lower]
}

// IsCacheBusterParamWithValue 检查参数名和值是否看起来像缓存破坏器
// 对于"v"参数，仅在值为纯数字时判定为缓存破坏器
func IsCacheBusterParamWithValue(name, value string) bool {
	lower := strings.ToLower(name)
	if lower == "v" {
		_, err := strconv.Atoi(value)
		return err == nil
	}
	return IsCacheBusterParam(name)
}

// FindDuplicates 查找重复/近似重复的请求
func (h *Har) FindDuplicates(opts DeduplicateOptions) []DuplicateGroup {
	if h == nil || len(h.Log.Entries) == 0 {
		return nil
	}

	groups := make(map[string][]int) // key -> entry indices

	for i, entry := range h.Log.Entries {
		key := computeDedupKey(entry, opts)
		groups[key] = append(groups[key], i)
	}

	var result []DuplicateGroup
	for key, indices := range groups {
		if len(indices) > 1 {
			result = append(result, DuplicateGroup{
				Key:          key,
				EntryIndices: indices,
				Count:        len(indices),
			})
		}
	}

	return result
}

// Deduplicate 去除重复请求，保留第一次出现的条目
func (h *Har) Deduplicate(opts DeduplicateOptions) *Har {
	if h == nil {
		return nil
	}

	cloned := h.Clone()
	if len(cloned.Log.Entries) == 0 {
		return cloned
	}

	seen := make(map[string]bool)
	var deduped []Entries

	for _, entry := range cloned.Log.Entries {
		key := computeDedupKey(entry, opts)
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, entry)
		}
	}

	cloned.Log.Entries = deduped
	return cloned
}

// computeDedupKey 根据策略计算去重键
func computeDedupKey(entry Entries, opts DeduplicateOptions) string {
	switch opts.Strategy {
	case DedupExactURL:
		return computeExactURLKey(entry, opts)
	case DedupURLPattern:
		return computeURLPatternKey(entry, opts)
	case DedupContentHash:
		return computeContentHashKey(entry, opts)
	default:
		return computeURLPatternKey(entry, opts)
	}
}

// computeExactURLKey 精确URL匹配的键
func computeExactURLKey(entry Entries, opts DeduplicateOptions) string {
	key := entry.Request.Method + " " + entry.Request.URL
	if opts.CompareHeaders {
		key += " " + headersKey(entry.Request.Headers)
	}
	if opts.CompareBody && entry.Request.PostData != nil {
		key += " " + entry.Request.PostData.Text
	}
	return key
}

// computeURLPatternKey 忽略指定参数的URL模式匹配键
func computeURLPatternKey(entry Entries, opts DeduplicateOptions) string {
	normalizedURL := normalizeURL(entry.Request.URL, opts.IgnoreParams)
	key := entry.Request.Method + " " + normalizedURL
	if opts.CompareHeaders {
		key += " " + headersKey(entry.Request.Headers)
	}
	if opts.CompareBody && entry.Request.PostData != nil {
		key += " " + entry.Request.PostData.Text
	}
	return key
}

// computeContentHashKey 基于内容哈希的键
// 内容哈希策略关注响应内容的相似性，不包含URL（因为不同URL可能返回相同内容）
func computeContentHashKey(entry Entries, opts DeduplicateOptions) string {
	h := sha256.New()

	// 方法
	h.Write([]byte(entry.Request.Method))

	// 响应状态码
	h.Write([]byte(fmt.Sprintf("%d", entry.Response.Status)))

	// 响应MIME类型
	h.Write([]byte(entry.Response.Content.MimeType))

	// 请求体
	if opts.CompareBody && entry.Request.PostData != nil {
		h.Write([]byte(entry.Request.PostData.Text))
	}

	// 请求头
	if opts.CompareHeaders {
		h.Write([]byte(headersKey(entry.Request.Headers)))
	}

	// 响应体（内容哈希通常关注响应内容）
	if entry.Response.Content.Text != "" {
		h.Write([]byte(entry.Response.Content.Text))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// normalizeURL 忽略指定参数，规范化URL
// 当ignoreParams为nil时，对查询参数进行排序
// 当ignoreParams不为nil时，移除指定参数并对剩余参数排序
func normalizeURL(rawURL string, ignoreParams []string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()

	if len(ignoreParams) > 0 {
		ignoreSet := make(map[string]bool, len(ignoreParams))
		for _, p := range ignoreParams {
			ignoreSet[strings.ToLower(p)] = true
		}

		newQ := make(url.Values)
		for name, values := range q {
			if !ignoreSet[strings.ToLower(name)] {
				newQ[name] = values
			}
		}
		u.RawQuery = newQ.Encode()
	} else {
		// Sort query parameters for consistent normalization
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// headersKey 将头部列表序列化为可比较的字符串
func headersKey(headers []Headers) string {
	var sb strings.Builder
	for _, h := range headers {
		sb.WriteString(h.Name)
		sb.WriteString(":")
		sb.WriteString(h.Value)
		sb.WriteString(";")
	}
	return sb.String()
}

// cacheBusterRegex 用于匹配纯数字值的缓存破坏器参数
var cacheBusterRegex = regexp.MustCompile(`^\d+$`)
