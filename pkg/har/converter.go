package har

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

// ConvertFormat 支持的转换格式
type ConvertFormat string

const (
	FormatCSV      ConvertFormat = "csv"
	FormatMarkdown ConvertFormat = "markdown"
	FormatHTML     ConvertFormat = "html"
	FormatText     ConvertFormat = "text"
)

// ConvertOptions 转换选项
type ConvertOptions struct {
	// 包含的字段
	IncludeURL         bool
	IncludeMethod      bool
	IncludeStatus      bool
	IncludeContentType bool
	IncludeSize        bool
	IncludeTime        bool
	IncludeTimings     bool
	IncludeHeaders     bool
	IncludeDateTime    bool
	IncludePostData    bool // 是否包含POST数据
	IncludeQueryString bool // 是否包含查询参数

	// 自定义表头（可选，如果不指定则使用默认值）
	Headers []string

	// 过滤选项（可选，用于在转换前先过滤数据）
	Filter *FilterOptions
}

// DefaultConvertOptions 默认的转换选项
func DefaultConvertOptions() ConvertOptions {
	return ConvertOptions{
		IncludeURL:         true,
		IncludeMethod:      true,
		IncludeStatus:      true,
		IncludeContentType: true,
		IncludeSize:        true,
		IncludeTime:        true,
		IncludeTimings:     false,
		IncludeHeaders:     false,
		IncludeDateTime:    true,
	}
}

// Convert 将HAR转换为指定格式
func (h *Har) Convert(format ConvertFormat, options ConvertOptions) (string, error) {
	// 如果有过滤条件，先过滤
	entries := h.Log.Entries
	if options.Filter != nil {
		filterResult := h.Filter(*options.Filter)
		entries = filterResult.Entries
	}

	switch format {
	case FormatCSV:
		return convertToCSV(entries, options)
	case FormatMarkdown:
		return convertToMarkdown(entries, options)
	case FormatHTML:
		return convertToHTML(entries, options)
	case FormatText:
		return convertToText(entries, options)
	default:
		return "", fmt.Errorf("不支持的转换格式: %s", format)
	}
}

// 转换为CSV格式
func convertToCSV(entries []Entries, options ConvertOptions) (string, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// 写入表头
	headers := getHeaders(options)
	if err := writer.Write(headers); err != nil {
		return "", err
	}

	// 写入数据行
	for _, entry := range entries {
		row := createDataRow(entry, options)
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// 转换为Markdown表格
func convertToMarkdown(entries []Entries, options ConvertOptions) (string, error) {
	buf := &bytes.Buffer{}

	// 写入表头
	headers := getHeaders(options)
	fmt.Fprintf(buf, "| %s |\n", strings.Join(headers, " | "))

	// 写入分隔行
	fmt.Fprintf(buf, "|%s|\n", strings.Repeat(" --- |", len(headers)))

	// 写入数据行
	for _, entry := range entries {
		row := createDataRow(entry, options)
		for i, cell := range row {
			// 转义Markdown中的特殊字符
			row[i] = strings.ReplaceAll(cell, "|", "\\|")
		}
		fmt.Fprintf(buf, "| %s |\n", strings.Join(row, " | "))
	}

	return buf.String(), nil
}

// 转换为HTML表格
func convertToHTML(entries []Entries, options ConvertOptions) (string, error) {
	buf := &bytes.Buffer{}

	// 开始表格
	fmt.Fprintln(buf, "<table border=\"1\">")

	// 写入表头
	headers := getHeaders(options)
	fmt.Fprintln(buf, "  <thead>")
	fmt.Fprintln(buf, "    <tr>")
	for _, header := range headers {
		fmt.Fprintf(buf, "      <th>%s</th>\n", escapeHTML(header))
	}
	fmt.Fprintln(buf, "    </tr>")
	fmt.Fprintln(buf, "  </thead>")

	// 写入数据行
	fmt.Fprintln(buf, "  <tbody>")
	for _, entry := range entries {
		row := createDataRow(entry, options)
		fmt.Fprintln(buf, "    <tr>")
		for _, cell := range row {
			fmt.Fprintf(buf, "      <td>%s</td>\n", escapeHTML(cell))
		}
		fmt.Fprintln(buf, "    </tr>")
	}
	fmt.Fprintln(buf, "  </tbody>")

	// 结束表格
	fmt.Fprintln(buf, "</table>")

	return buf.String(), nil
}

// 转换为纯文本格式
func convertToText(entries []Entries, options ConvertOptions) (string, error) {
	buf := &bytes.Buffer{}

	// 写入表头
	headers := getHeaders(options)
	fmt.Fprintln(buf, strings.Join(headers, "\t"))
	fmt.Fprintln(buf, strings.Repeat("-", 80))

	// 写入数据行
	for _, entry := range entries {
		row := createDataRow(entry, options)
		fmt.Fprintln(buf, strings.Join(row, "\t"))
	}

	return buf.String(), nil
}

// 获取表头
func getHeaders(options ConvertOptions) []string {
	if len(options.Headers) > 0 {
		return options.Headers
	}

	var headers []string

	if options.IncludeDateTime {
		headers = append(headers, "日期时间")
	}
	if options.IncludeMethod {
		headers = append(headers, "方法")
	}
	if options.IncludeURL {
		headers = append(headers, "URL")
	}
	if options.IncludeStatus {
		headers = append(headers, "状态码")
	}
	if options.IncludeContentType {
		headers = append(headers, "内容类型")
	}
	if options.IncludeSize {
		headers = append(headers, "大小(字节)")
	}
	if options.IncludeTime {
		headers = append(headers, "时间(ms)")
	}
	if options.IncludeTimings {
		headers = append(headers, "阻塞(ms)", "DNS(ms)", "连接(ms)", "发送(ms)", "等待(ms)", "接收(ms)")
	}
	if options.IncludePostData {
		headers = append(headers, "POST数据类型", "POST数据")
	}
	if options.IncludeQueryString {
		headers = append(headers, "查询参数")
	}
	if options.IncludeHeaders {
		headers = append(headers, "请求头", "响应头")
	}

	return headers
}

// 创建数据行
func createDataRow(entry Entries, options ConvertOptions) []string {
	var row []string

	// 日期时间
	if options.IncludeDateTime {
		row = append(row, entry.StartedDateTime.Format(time.RFC3339))
	}

	// 请求方法
	if options.IncludeMethod {
		row = append(row, entry.Request.Method)
	}

	// URL
	if options.IncludeURL {
		row = append(row, entry.Request.URL)
	}

	// 状态码
	if options.IncludeStatus {
		row = append(row, fmt.Sprintf("%d %s", entry.Response.Status, entry.Response.StatusText))
	}

	// 内容类型
	if options.IncludeContentType {
		row = append(row, entry.Response.Content.MimeType)
	}

	// 大小
	if options.IncludeSize {
		row = append(row, fmt.Sprintf("%d", entry.Response.Content.Size))
	}

	// 总时间
	if options.IncludeTime {
		row = append(row, fmt.Sprintf("%.2f", entry.Time))
	}

	// 详细时间
	if options.IncludeTimings {
		row = append(row,
			fmt.Sprintf("%.2f", entry.Timings.Blocked),
			fmt.Sprintf("%.2f", entry.Timings.DNS),
			fmt.Sprintf("%.2f", entry.Timings.Connect),
			fmt.Sprintf("%.2f", entry.Timings.Send),
			fmt.Sprintf("%.2f", entry.Timings.Wait),
			fmt.Sprintf("%.2f", entry.Timings.Receive),
		)
	}

	// POST数据
	if options.IncludePostData {
		if entry.Request.PostData != nil {
			row = append(row, entry.Request.PostData.MimeType, entry.Request.PostData.Text)
		} else {
			row = append(row, "", "")
		}
	}

	// 查询参数
	if options.IncludeQueryString {
		var qs []string
		for _, param := range entry.Request.QueryString {
			qs = append(qs, fmt.Sprintf("%s=%s", param.Name, param.Value))
		}
		row = append(row, strings.Join(qs, "&"))
	}

	// 请求头和响应头
	if options.IncludeHeaders {
		var reqHeaders []string
		for _, h := range entry.Request.Headers {
			reqHeaders = append(reqHeaders, fmt.Sprintf("%s: %s", h.Name, h.Value))
		}
		var respHeaders []string
		for _, h := range entry.Response.Headers {
			respHeaders = append(respHeaders, fmt.Sprintf("%s: %s", h.Name, h.Value))
		}
		row = append(row, strings.Join(reqHeaders, "; "), strings.Join(respHeaders, "; "))
	}

	return row
}

// 转义HTML特殊字符
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
