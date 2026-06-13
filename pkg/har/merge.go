package har

import (
	"fmt"
	"sort"
	"time"
)

// MergeOptions 合并选项
type MergeOptions struct {
	SortByTime  bool // 按时间排序合并后的条目
	Deduplicate bool // 去重（按Method+URL去重，保留最新的）
}

// DefaultMergeOptions 返回默认的合并选项
func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		SortByTime:  true,
		Deduplicate: false,
	}
}

// Merge 合并多个HAR文件
//
// 将多个HAR文件的条目合并到一个HAR文件中。
// 合并后的HAR文件使用第一个HAR的版本和创建者信息。
func Merge(hars ...*Har) *Har {
	return MergeWithOptions(DefaultMergeOptions(), hars...)
}

// MergeWithOptions 使用选项合并多个HAR文件
func MergeWithOptions(options MergeOptions, hars ...*Har) *Har {
	if len(hars) == 0 {
		return NewHar()
	}

	result := NewHar()

	// 使用第一个HAR的元信息
	result.Log.Version = hars[0].Log.Version
	result.Log.Creator = hars[0].Log.Creator
	result.Log.Browser = hars[0].Log.Browser

	// 合并所有条目和页面
	for _, h := range hars {
		if h == nil {
			continue
		}
		result.Log.Entries = append(result.Log.Entries, h.Log.Entries...)
		result.Log.Pages = append(result.Log.Pages, h.Log.Pages...)
	}

	// 去重
	if options.Deduplicate {
		result.Log.Entries = deduplicateEntries(result.Log.Entries)
	}

	// 排序
	if options.SortByTime {
		sortEntriesByTime(result.Log.Entries)
	}

	return result
}

// deduplicateEntries 按Method+URL去重，保留最新的
func deduplicateEntries(entries []Entries) []Entries {
	seen := make(map[string]int) // key -> index in result
	var result []Entries

	for _, entry := range entries {
		key := entry.Request.Method + " " + entry.Request.URL
		if idx, ok := seen[key]; ok {
			// 保留较新的条目
			if entry.StartedDateTime.After(result[idx].StartedDateTime) {
				result[idx] = entry
			}
		} else {
			seen[key] = len(result)
			result = append(result, entry)
		}
	}

	return result
}

// sortEntriesByTime 按时间排序条目
func sortEntriesByTime(entries []Entries) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartedDateTime.Before(entries[j].StartedDateTime)
	})
}

// SplitByPage 按页面引用拆分HAR文件
//
// 将HAR条目按pageref分组，返回以pageref为键的HAR映射。
// 没有pageref的条目归入空字符串键。
func (h *Har) SplitByPage() map[string]*Har {
	result := make(map[string]*Har)

	if h == nil {
		return result
	}

	// 收集所有页面
	pagesMap := make(map[string]Pages)
	for _, page := range h.Log.Pages {
		pagesMap[page.ID] = page
	}

	// 按pageref分组
	groups := make(map[string][]Entries)
	for _, entry := range h.Log.Entries {
		ref := entry.Pageref
		groups[ref] = append(groups[ref], entry)
	}

	// 为每个分组创建HAR
	for ref, entries := range groups {
		har := NewHar()
		har.Log.Version = h.Log.Version
		har.Log.Creator = h.Log.Creator

		if page, ok := pagesMap[ref]; ok {
			har.Log.Pages = []Pages{page}
		}

		har.Log.Entries = entries
		result[ref] = har
	}

	return result
}

// SplitByDomain 按域名拆分HAR文件
//
// 将HAR条目按请求域名分组，返回以域名为键的HAR映射。
func (h *Har) SplitByDomain() map[string]*Har {
	result := make(map[string]*Har)

	if h == nil {
		return result
	}

	// 按域名分组
	groups := make(map[string][]Entries)
	for _, entry := range h.Log.Entries {
		domain := extractDomain(entry.Request.URL)
		groups[domain] = append(groups[domain], entry)
	}

	// 为每个分组创建HAR
	for domain, entries := range groups {
		har := NewHar()
		har.Log.Version = h.Log.Version
		har.Log.Creator = h.Log.Creator
		har.Log.Entries = entries
		result[domain] = har
	}

	return result
}

// SplitByTimeRange 按时间范围拆分HAR文件
//
// 将HAR条目按指定的时间间隔分组。
// 例如，如果interval为1小时，则每个HAR文件包含该小时内的所有条目。
func (h *Har) SplitByTimeRange(interval time.Duration) []*Har {
	if h == nil || len(h.Log.Entries) == 0 {
		return nil
	}

	// 按时间排序
	sorted := make([]Entries, len(h.Log.Entries))
	copy(sorted, h.Log.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartedDateTime.Before(sorted[j].StartedDateTime)
	})

	// 按时间间隔分组
	var result []*Har
	var currentGroup []Entries
	var groupStart time.Time

	for i, entry := range sorted {
		if i == 0 {
			groupStart = entry.StartedDateTime
			currentGroup = append(currentGroup, entry)
			continue
		}

		if entry.StartedDateTime.Sub(groupStart) > interval {
			// 创建当前分组的HAR
			har := NewHar()
			har.Log.Version = h.Log.Version
			har.Log.Creator = h.Log.Creator
			har.Log.Entries = currentGroup
			result = append(result, har)

			// 开始新分组
			currentGroup = []Entries{entry}
			groupStart = entry.StartedDateTime
		} else {
			currentGroup = append(currentGroup, entry)
		}
	}

	// 处理最后一组
	if len(currentGroup) > 0 {
		har := NewHar()
		har.Log.Version = h.Log.Version
		har.Log.Creator = h.Log.Creator
		har.Log.Entries = currentGroup
		result = append(result, har)
	}

	return result
}

// SplitBySize 按条目数量拆分HAR文件
//
// 将HAR条目按指定数量分组，每个HAR文件最多包含maxEntries个条目。
func (h *Har) SplitBySize(maxEntries int) []*Har {
	if h == nil || maxEntries <= 0 {
		return nil
	}

	total := len(h.Log.Entries)
	if total == 0 {
		return nil
	}

	numGroups := (total + maxEntries - 1) / maxEntries
	result := make([]*Har, 0, numGroups)

	for i := 0; i < total; i += maxEntries {
		end := i + maxEntries
		if end > total {
			end = total
		}

		har := NewHar()
		har.Log.Version = h.Log.Version
		har.Log.Creator = h.Log.Creator

		entries := make([]Entries, end-i)
		copy(entries, h.Log.Entries[i:end])
		har.Log.Entries = entries

		result = append(result, har)
	}

	return result
}

// SplitByStatusCode 按状态码范围拆分HAR文件
//
// 将HAR条目按状态码范围分组：2xx, 3xx, 4xx, 5xx
func (h *Har) SplitByStatusCode() map[string]*Har {
	result := make(map[string]*Har)

	if h == nil {
		return result
	}

	groups := make(map[string][]Entries)
	for _, entry := range h.Log.Entries {
		var group string
		status := entry.Response.Status
		switch {
		case status >= 200 && status < 300:
			group = "2xx"
		case status >= 300 && status < 400:
			group = "3xx"
		case status >= 400 && status < 500:
			group = "4xx"
		case status >= 500 && status < 600:
			group = "5xx"
		default:
			group = fmt.Sprintf("%dxx", status/100)
		}
		groups[group] = append(groups[group], entry)
	}

	for group, entries := range groups {
		har := NewHar()
		har.Log.Version = h.Log.Version
		har.Log.Creator = h.Log.Creator
		har.Log.Entries = entries
		result[group] = har
	}

	return result
}

// SplitByMethod 按HTTP方法拆分HAR文件
func (h *Har) SplitByMethod() map[string]*Har {
	result := make(map[string]*Har)

	if h == nil {
		return result
	}

	groups := make(map[string][]Entries)
	for _, entry := range h.Log.Entries {
		method := entry.Request.Method
		groups[method] = append(groups[method], entry)
	}

	for method, entries := range groups {
		har := NewHar()
		har.Log.Version = h.Log.Version
		har.Log.Creator = h.Log.Creator
		har.Log.Entries = entries
		result[method] = har
	}

	return result
}
