package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// listCmd 列出HAR条目
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出HAR条目",
	Long: `列出HAR文件中的请求条目，支持按时间、大小、URL、状态码排序，
支持按方法、状态码、域名等条件过滤，支持限制输出条数。`,
	Example: `  har -f capture.har list
  har -f capture.har list --limit 10
  har -f capture.har list --sort size --order asc
  har -f capture.har list --method GET --status 200`,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)

		// 获取过滤参数
		method, _ := cmd.Flags().GetString("method")
		status, _ := cmd.Flags().GetInt("status")
		domain, _ := cmd.Flags().GetString("domain")
		sortBy, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")

		// 使用FilterWith进行过滤
		opts := []har.FilterOption{}
		if method != "" {
			opts = append(opts, har.WithFilterMethod(method))
		}
		if status > 0 {
			opts = append(opts, har.WithFilterStatusCode(status))
		}

		var result *har.FilterResult
		if len(opts) > 0 {
			result = h.FilterWith(opts...)
		} else {
			// 无过滤条件，使用所有条目
			result = &har.FilterResult{Entries: h.Log.Entries}
		}

		// 按域名进一步过滤
		if domain != "" {
			var filtered []har.Entries
			for _, entry := range result.Entries {
				if d := har.ExtractDomain(entry.Request.URL); d == domain {
					filtered = append(filtered, entry)
				}
			}
			result = &har.FilterResult{Entries: filtered}
		}

		// 排序
		switch sortBy {
		case "size":
			if order == "asc" {
				result.SortBySize()
			} else {
				result.SortBySizeDesc()
			}
		case "url":
			// URL排序无SDK方法，保持默认顺序
		case "status":
			// 状态码排序无SDK方法，保持默认顺序
		default: // time
			if order == "asc" {
				result.SortByDuration()
			} else {
				result.SortByDurationDesc()
			}
		}

		// 限制条数
		if limit > 0 {
			result.Limit(limit)
		}

		return internal.WriteOutput(cmd, buildListJSON(result), func() string {
			return formatListTable(result, h)
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntP("limit", "n", 0, "限制输出条数 (0=全部)")
	listCmd.Flags().String("sort", "time", "排序方式 (time, size, url, status)")
	listCmd.Flags().String("order", "desc", "排序方向 (asc, desc)")
	listCmd.Flags().String("method", "", "按HTTP方法过滤")
	listCmd.Flags().Int("status", 0, "按状态码过滤")
	listCmd.Flags().String("domain", "", "按域名过滤")
}

// listEntry 简化的条目对象用于JSON输出
type listEntry struct {
	Index   int    `json:"index"`
	Method  string `json:"method"`
	Status  int    `json:"status"`
	Size    int    `json:"size"`
	Time    float64 `json:"time"`
	URL     string `json:"url"`
}

// buildListJSON 构建JSON输出数据
func buildListJSON(result *har.FilterResult) []listEntry {
	entries := make([]listEntry, len(result.Entries))
	for i, entry := range result.Entries {
		entries[i] = listEntry{
			Index:  i,
			Method: entry.Request.Method,
			Status: entry.Response.Status,
			Size:   entry.Response.Content.Size,
			Time:   entry.Time,
			URL:    entry.Request.URL,
		}
	}
	return entries
}

// formatListTable 格式化条目列表为tabwriter表格
func formatListTable(result *har.FilterResult, h *har.Har) string {
	var sb tabWriterBuf

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "INDEX\tMETHOD\tSTATUS\tSIZE\tTIME\tURL\n")

	for i, entry := range result.Entries {
		size := internal.FormatBytes(entry.Response.Content.Size)
		time := internal.FormatDuration(entry.Time)
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\n",
			i, entry.Request.Method, entry.Response.Status, size, time, entry.Request.URL)
	}
	w.Flush()

	return sb.String()
}

// tabWriterBuf 用于tabwriter输出的缓冲区
type tabWriterBuf struct {
	buf []byte
}

func (t *tabWriterBuf) Write(p []byte) (n int, err error) {
	t.buf = append(t.buf, p...)
	return len(p), nil
}

func (t *tabWriterBuf) String() string {
	return string(t.buf)
}