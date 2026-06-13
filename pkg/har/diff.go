package har

import (
	"fmt"
	"strings"
)

// HarDiff 表示两个HAR文件的差异
type HarDiff struct {
	Added    []DiffEntry     // 新增的请求
	Removed  []DiffEntry     // 删除的请求
	Modified []ModifiedEntry // 修改的请求
	Unchanged int            // 未变更的请求数
}

// DiffEntry 表示差异中的单个条目
type DiffEntry struct {
	Method string // HTTP方法
	URL    string // 请求URL
	Status int    // 响应状态码
	Index  int    // 在原HAR中的索引
}

// ModifiedEntry 表示修改的条目
type ModifiedEntry struct {
	Method  string        // HTTP方法
	URL     string        // 请求URL
	Changes []FieldChange // 字段变更列表
	Old     *Entries      // 旧条目
	New     *Entries      // 新条目
}

// FieldChange 表示单个字段的变更
type FieldChange struct {
	Field    string      // 字段名
	OldValue interface{} // 旧值
	NewValue interface{} // 新值
}

// DiffOptions 差异比较选项
type DiffOptions struct {
	IgnoreHeaders     []string // 忽略的头部字段名
	IgnoreTimings     bool     // 忽略时间差异
	IgnoreDates       bool     // 忽略日期差异
	IgnoreCache       bool     // 忽略缓存差异
	IgnoreComment     bool     // 忽略注释差异
	NormalizeURL      bool     // URL归一化（排序查询参数）
	CompareByURL      bool     // 按URL匹配（默认按索引+URL）
	IncludeBody       bool     // 比较响应体内容
}

// DefaultDiffOptions 返回默认的差异比较选项
func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		IgnoreTimings: true,
		IgnoreDates:   true,
		IgnoreCache:   true,
	}
}

// Diff 比较两个HAR文件的差异
func Diff(har1, har2 *Har, options DiffOptions) *HarDiff {
	result := &HarDiff{}

	if har1 == nil && har2 == nil {
		return result
	}
	if har1 == nil {
		for i, entry := range har2.Log.Entries {
			result.Added = append(result.Added, DiffEntry{
				Method: entry.Request.Method,
				URL:    entry.Request.URL,
				Status: entry.Response.Status,
				Index:  i,
			})
		}
		return result
	}
	if har2 == nil {
		for i, entry := range har1.Log.Entries {
			result.Removed = append(result.Removed, DiffEntry{
				Method: entry.Request.Method,
				URL:    entry.Request.URL,
				Status: entry.Response.Status,
				Index:  i,
			})
		}
		return result
	}

	// 构建键值映射
	entries1 := buildEntryMap(har1, options)
	entries2 := buildEntryMap(har2, options)

	// 查找新增和修改
	for key, entry2 := range entries2 {
		if entry1, ok := entries1[key]; ok {
			// 比较条目差异
			changes := compareEntries(entry1, entry2, options)
			if len(changes) > 0 {
				result.Modified = append(result.Modified, ModifiedEntry{
					Method:  entry2.Request.Method,
					URL:     entry2.Request.URL,
					Changes: changes,
					Old:     entry1,
					New:     entry2,
				})
			} else {
				result.Unchanged++
			}
		} else {
			result.Added = append(result.Added, DiffEntry{
				Method: entry2.Request.Method,
				URL:    entry2.Request.URL,
				Status: entry2.Response.Status,
			})
		}
	}

	// 查找删除
	for key, entry1 := range entries1 {
		if _, ok := entries2[key]; !ok {
			result.Removed = append(result.Removed, DiffEntry{
				Method: entry1.Request.Method,
				URL:    entry1.Request.URL,
				Status: entry1.Response.Status,
			})
		}
	}

	return result
}

// entryKey 生成条目的唯一键
func entryKey(entry *Entries, options DiffOptions) string {
	method := entry.Request.Method
	u := entry.Request.URL

	if options.NormalizeURL {
		u = normalizeURL(u, nil)
	}

	return method + " " + u
}

// buildEntryMap 构建条目映射
func buildEntryMap(har *Har, options DiffOptions) map[string]*Entries {
	result := make(map[string]*Entries)

	for i := range har.Log.Entries {
		key := entryKey(&har.Log.Entries[i], options)
		// 处理重复键
		if _, exists := result[key]; exists {
			key = fmt.Sprintf("%s_%d", key, i)
		}
		result[key] = &har.Log.Entries[i]
	}

	return result
}

// compareEntries 比较两个条目的差异
func compareEntries(entry1, entry2 *Entries, options DiffOptions) []FieldChange {
	var changes []FieldChange

	// 比较响应状态码
	if entry1.Response.Status != entry2.Response.Status {
		changes = append(changes, FieldChange{
			Field:    "response.status",
			OldValue: entry1.Response.Status,
			NewValue: entry2.Response.Status,
		})
	}

	// 比较响应状态文本
	if entry1.Response.StatusText != entry2.Response.StatusText {
		changes = append(changes, FieldChange{
			Field:    "response.statusText",
			OldValue: entry1.Response.StatusText,
			NewValue: entry2.Response.StatusText,
		})
	}

	// 比较总时间
	if !options.IgnoreTimings && entry1.Time != entry2.Time {
		changes = append(changes, FieldChange{
			Field:    "time",
			OldValue: entry1.Time,
			NewValue: entry2.Time,
		})
	}

	// 比较响应内容类型
	if entry1.Response.Content.MimeType != entry2.Response.Content.MimeType {
		changes = append(changes, FieldChange{
			Field:    "response.content.mimeType",
			OldValue: entry1.Response.Content.MimeType,
			NewValue: entry2.Response.Content.MimeType,
		})
	}

	// 比较响应内容大小
	if entry1.Response.Content.Size != entry2.Response.Content.Size {
		changes = append(changes, FieldChange{
			Field:    "response.content.size",
			OldValue: entry1.Response.Content.Size,
			NewValue: entry2.Response.Content.Size,
		})
	}

	// 比较响应体内容
	if options.IncludeBody && entry1.Response.Content.Text != entry2.Response.Content.Text {
		changes = append(changes, FieldChange{
			Field:    "response.content.text",
			OldValue: entry1.Response.Content.Text,
			NewValue: entry2.Response.Content.Text,
		})
	}

	// 比较请求头
	changes = append(changes, compareHeaders(entry1.Request.Headers, entry2.Request.Headers, "request.headers", options)...)

	// 比较响应头
	changes = append(changes, compareHeaders(entry1.Response.Headers, entry2.Response.Headers, "response.headers", options)...)

	return changes
}

// compareHeaders 比较头部差异
func compareHeaders(headers1, headers2 []Headers, prefix string, options DiffOptions) []FieldChange {
	var changes []FieldChange

	// 构建忽略头部的集合
	ignoreSet := make(map[string]bool)
	for _, h := range options.IgnoreHeaders {
		ignoreSet[strings.ToLower(h)] = true
	}

	// 转为map便于查找
	map1 := headersToMap(headers1, ignoreSet)
	map2 := headersToMap(headers2, ignoreSet)

	// 查找新增和修改的头部
	for name, value2 := range map2 {
		if value1, ok := map1[name]; ok {
			if value1 != value2 {
				changes = append(changes, FieldChange{
					Field:    fmt.Sprintf("%s.%s", prefix, name),
					OldValue: value1,
					NewValue: value2,
				})
			}
		} else {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("%s.%s", prefix, name),
				OldValue: nil,
				NewValue: value2,
			})
		}
	}

	// 查找删除的头部
	for name, value1 := range map1 {
		if _, ok := map2[name]; !ok {
			changes = append(changes, FieldChange{
				Field:    fmt.Sprintf("%s.%s", prefix, name),
				OldValue: value1,
				NewValue: nil,
			})
		}
	}

	return changes
}

// headersToMap 将头部列表转为map，并过滤忽略的头部
func headersToMap(headers []Headers, ignoreSet map[string]bool) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		if ignoreSet[strings.ToLower(h.Name)] {
			continue
		}
		result[h.Name] = h.Value
	}
	return result
}

// HasChanges 检查是否存在差异
func (d *HarDiff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Modified) > 0
}

// TotalChanges 返回总变更数
func (d *HarDiff) TotalChanges() int {
	return len(d.Added) + len(d.Removed) + len(d.Modified)
}

// Report 生成差异报告
func (d *HarDiff) Report(format ConvertFormat) string {
	var sb strings.Builder

	switch format {
	case FormatText:
		d.writeTextReport(&sb)
	case FormatMarkdown:
		d.writeMarkdownReport(&sb)
	case FormatCSV:
		d.writeCSVReport(&sb)
	default:
		d.writeTextReport(&sb)
	}

	return sb.String()
}

// writeTextReport 写入文本格式报告
func (d *HarDiff) writeTextReport(sb *strings.Builder) {
	sb.WriteString("HAR 差异报告\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("总变更: %d (新增: %d, 删除: %d, 修改: %d, 未变: %d)\n\n",
		d.TotalChanges(), len(d.Added), len(d.Removed), len(d.Modified), d.Unchanged))

	if len(d.Added) > 0 {
		sb.WriteString("新增请求:\n")
		for _, a := range d.Added {
			sb.WriteString(fmt.Sprintf("  + [%d] %s %s (状态: %d)\n", a.Index, a.Method, a.URL, a.Status))
		}
		sb.WriteString("\n")
	}

	if len(d.Removed) > 0 {
		sb.WriteString("删除请求:\n")
		for _, r := range d.Removed {
			sb.WriteString(fmt.Sprintf("  - [%d] %s %s (状态: %d)\n", r.Index, r.Method, r.URL, r.Status))
		}
		sb.WriteString("\n")
	}

	if len(d.Modified) > 0 {
		sb.WriteString("修改请求:\n")
		for _, m := range d.Modified {
			sb.WriteString(fmt.Sprintf("  ~ %s %s\n", m.Method, m.URL))
			for _, c := range m.Changes {
				sb.WriteString(fmt.Sprintf("      %s: %v -> %v\n", c.Field, c.OldValue, c.NewValue))
			}
		}
	}
}

// writeMarkdownReport 写入Markdown格式报告
func (d *HarDiff) writeMarkdownReport(sb *strings.Builder) {
	sb.WriteString("# HAR 差异报告\n\n")
	sb.WriteString(fmt.Sprintf("**总变更**: %d | **新增**: %d | **删除**: %d | **修改**: %d | **未变**: %d\n\n",
		d.TotalChanges(), len(d.Added), len(d.Removed), len(d.Modified), d.Unchanged))

	if len(d.Added) > 0 {
		sb.WriteString("## 新增请求\n\n")
		sb.WriteString("| 方法 | URL | 状态码 |\n")
		sb.WriteString("| --- | --- | --- |\n")
		for _, a := range d.Added {
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", a.Method, a.URL, a.Status))
		}
		sb.WriteString("\n")
	}

	if len(d.Removed) > 0 {
		sb.WriteString("## 删除请求\n\n")
		sb.WriteString("| 方法 | URL | 状态码 |\n")
		sb.WriteString("| --- | --- | --- |\n")
		for _, r := range d.Removed {
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", r.Method, r.URL, r.Status))
		}
		sb.WriteString("\n")
	}

	if len(d.Modified) > 0 {
		sb.WriteString("## 修改请求\n\n")
		for _, m := range d.Modified {
			sb.WriteString(fmt.Sprintf("### %s %s\n\n", m.Method, m.URL))
			sb.WriteString("| 字段 | 旧值 | 新值 |\n")
			sb.WriteString("| --- | --- | --- |\n")
			for _, c := range m.Changes {
				sb.WriteString(fmt.Sprintf("| %s | %v | %v |\n", c.Field, c.OldValue, c.NewValue))
			}
			sb.WriteString("\n")
		}
	}
}

// writeCSVReport 写入CSV格式报告
func (d *HarDiff) writeCSVReport(sb *strings.Builder) {
	sb.WriteString("type,method,url,field,old_value,new_value\n")

	for _, a := range d.Added {
		sb.WriteString(fmt.Sprintf("added,%s,%s,,,\"%d\"\n", a.Method, a.URL, a.Status))
	}
	for _, r := range d.Removed {
		sb.WriteString(fmt.Sprintf("removed,%s,%s,,,\"%d\"\n", r.Method, r.URL, r.Status))
	}
	for _, m := range d.Modified {
		for _, c := range m.Changes {
			sb.WriteString(fmt.Sprintf("modified,%s,%s,%s,%v,%v\n", m.Method, m.URL, c.Field, c.OldValue, c.NewValue))
		}
	}
}
