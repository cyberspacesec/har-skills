package har

import (
	"regexp"
	"sort"
	"time"
)

// HarIndex HAR条目的索引，支持快速查找
type HarIndex struct {
	byURL      map[string][]int // URL -> 条目索引
	byMethod   map[string][]int // HTTP方法 -> 条目索引
	byStatus   map[int][]int    // 状态码 -> 条目索引
	byDomain   map[string][]int // 域名 -> 条目索引
	byMimeType map[string][]int // MIME类型 -> 条目索引
	har        *Har             // 关联的Har对象
}

// IndexStats 索引统计信息
type IndexStats struct {
	UniqueURLs    int      // 唯一URL数
	UniqueDomains int      // 唯一域名数
	StatusCodes   []int    // 出现过的状态码
	Methods       []string // 出现过的HTTP方法
}

// BuildIndex 为HAR构建所有索引
func (h *Har) BuildIndex() *HarIndex {
	if h == nil {
		return &HarIndex{
			byURL:      make(map[string][]int),
			byMethod:   make(map[string][]int),
			byStatus:   make(map[int][]int),
			byDomain:   make(map[string][]int),
			byMimeType: make(map[string][]int),
			har:        h,
		}
	}

	idx := &HarIndex{
		byURL:      make(map[string][]int),
		byMethod:   make(map[string][]int),
		byStatus:   make(map[int][]int),
		byDomain:   make(map[string][]int),
		byMimeType: make(map[string][]int),
		har:        h,
	}

	for i, entry := range h.Log.Entries {
		// URL索引
		idx.byURL[entry.Request.URL] = append(idx.byURL[entry.Request.URL], i)

		// 方法索引
		idx.byMethod[entry.Request.Method] = append(idx.byMethod[entry.Request.Method], i)

		// 状态码索引
		idx.byStatus[entry.Response.Status] = append(idx.byStatus[entry.Response.Status], i)

		// 域名索引
		domain := extractDomain(entry.Request.URL)
		if domain != "" {
			idx.byDomain[domain] = append(idx.byDomain[domain], i)
		}

		// MIME类型索引
		mime := entry.Response.Content.MimeType
		if mime != "" {
			idx.byMimeType[mime] = append(idx.byMimeType[mime], i)
		}
	}

	return idx
}

// ByURL 按精确URL查找条目
func (idx *HarIndex) ByURL(urlStr string) []*Entries {
	return idx.entriesByIndices(idx.byURL[urlStr])
}

// ByMethod 按HTTP方法查找条目
func (idx *HarIndex) ByMethod(method string) []*Entries {
	return idx.entriesByIndices(idx.byMethod[method])
}

// ByStatus 按状态码查找条目
func (idx *HarIndex) ByStatus(code int) []*Entries {
	return idx.entriesByIndices(idx.byStatus[code])
}

// ByDomain 按域名查找条目
func (idx *HarIndex) ByDomain(domain string) []*Entries {
	return idx.entriesByIndices(idx.byDomain[domain])
}

// ByMimeType 按MIME类型查找条目
func (idx *HarIndex) ByMimeType(mime string) []*Entries {
	return idx.entriesByIndices(idx.byMimeType[mime])
}

// ByURLPattern 按正则URL模式查找条目
func (idx *HarIndex) ByURLPattern(pattern string) []*Entries {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	var result []*Entries
	for urlStr, indices := range idx.byURL {
		if re.MatchString(urlStr) {
			result = append(result, idx.entriesByIndices(indices)...)
		}
	}
	return result
}

// ByTimeRange 按时间范围查找条目
func (idx *HarIndex) ByTimeRange(start, end time.Time) []*Entries {
	if idx.har == nil {
		return nil
	}

	var result []*Entries
	for i := range idx.har.Log.Entries {
		entry := &idx.har.Log.Entries[i]
		if !entry.StartedDateTime.Before(start) && !entry.StartedDateTime.After(end) {
			result = append(result, entry)
		}
	}
	return result
}

// Size 返回索引中的总条目数
func (idx *HarIndex) Size() int {
	if idx.har == nil || idx.har.Log.Entries == nil {
		return 0
	}
	return len(idx.har.Log.Entries)
}

// Stats 返回索引统计信息
func (idx *HarIndex) Stats() IndexStats {
	stats := IndexStats{
		UniqueURLs:    len(idx.byURL),
		UniqueDomains: len(idx.byDomain),
	}

	// 收集状态码
	for code := range idx.byStatus {
		stats.StatusCodes = append(stats.StatusCodes, code)
	}
	sort.Ints(stats.StatusCodes)

	// 收集方法
	for method := range idx.byMethod {
		stats.Methods = append(stats.Methods, method)
	}
	sort.Strings(stats.Methods)

	return stats
}

// entriesByIndices 根据索引列表获取条目指针
func (idx *HarIndex) entriesByIndices(indices []int) []*Entries {
	if idx.har == nil || len(indices) == 0 {
		return nil
	}

	result := make([]*Entries, 0, len(indices))
	for _, i := range indices {
		if i >= 0 && i < len(idx.har.Log.Entries) {
			result = append(result, &idx.har.Log.Entries[i])
		}
	}
	return result
}
