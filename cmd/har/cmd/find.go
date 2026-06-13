package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// findCmd 搜索HAR条目
var findCmd = &cobra.Command{
	Use:   "find [pattern]",
	Short: "搜索HAR条目",
	Long: `按URL、状态码或条件搜索HAR条目。支持正则表达式匹配URL、
按状态码范围、内容类型、域名、请求头、资源类型等条件过滤，
还可以快速查找错误请求、重定向请求和慢请求。`,
	Example: `  har -f capture.har find "api/users"
  har -f capture.har find --regex "api/v[0-9]+"
  har -f capture.har find --errors
  har -f capture.har find --redirects
  har -f capture.har find --slow 1000
  har -f capture.har find --status-code 404
  har -f capture.har find --method GET --domain example.com`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)

		// 获取所有过滤参数
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}
		useRegex, _ := cmd.Flags().GetBool("regex")
		method, _ := cmd.Flags().GetString("method")
		statusCode, _ := cmd.Flags().GetInt("status-code")
		statusMin, _ := cmd.Flags().GetInt("status-min")
		statusMax, _ := cmd.Flags().GetInt("status-max")
		contentType, _ := cmd.Flags().GetString("content-type")
		domain, _ := cmd.Flags().GetString("domain")
		headers, _ := cmd.Flags().GetStringSlice("header")
		resourceType, _ := cmd.Flags().GetString("resource-type")
		errors, _ := cmd.Flags().GetBool("errors")
		redirects, _ := cmd.Flags().GetBool("redirects")
		slow, _ := cmd.Flags().GetFloat64("slow")
		limit, _ := cmd.Flags().GetInt("limit")

		// 构建过滤选项
		opts := []har.FilterOption{}

		// URL模式匹配
		if pattern != "" {
			if useRegex {
				opts = append(opts, har.WithFilterURL(pattern))
				opts = append(opts, har.WithFilterRegex())
			} else {
				opts = append(opts, har.WithFilterURL(pattern))
			}
		}

		// 方法过滤
		if method != "" {
			opts = append(opts, har.WithFilterMethod(method))
		}

		// 状态码过滤
		if statusCode > 0 {
			opts = append(opts, har.WithFilterStatusCode(statusCode))
		}

		// 状态码范围过滤
		if statusMin > 0 || statusMax > 0 {
			opts = append(opts, har.WithFilterStatusCodeRange(statusMin, statusMax))
		}

		// 内容类型过滤
		if contentType != "" {
			opts = append(opts, har.WithFilterContentType(contentType))
		}

		// 错误过滤
		if errors {
			opts = append(opts, har.WithFilterHasError())
		}

		// 资源类型过滤
		if resourceType != "" {
			opts = append(opts, har.WithFilterResourceType(resourceType))
		}

		// 慢请求过滤
		if slow > 0 {
			opts = append(opts, har.WithFilterDuration(slow, 0))
		}

		// 请求头过滤
		for _, h := range headers {
			parts := strings.SplitN(h, ":", 2)
			name := parts[0]
			value := ""
			if len(parts) > 1 {
				value = strings.TrimSpace(parts[1])
			}
			opts = append(opts, har.WithFilterHeader(name, value))
		}

		// 执行过滤
		var result *har.FilterResult
		if len(opts) > 0 {
			result = h.FilterWith(opts...)
		} else {
			result = &har.FilterResult{Entries: h.Log.Entries}
		}

		// 按域名过滤（需单独处理）
		if domain != "" {
			var filtered []har.Entries
			for _, entry := range result.Entries {
				if d := har.ExtractDomain(entry.Request.URL); d == domain {
					filtered = append(filtered, entry)
				}
			}
			result = &har.FilterResult{Entries: filtered}
		}

		// 重定向过滤
		if redirects {
			redirectResult := h.FindRedirects()
			if len(opts) > 0 || domain != "" {
				// 与现有结果取交集
				redirectSet := make(map[string]bool)
				for _, e := range redirectResult.Entries {
					redirectSet[e.Request.URL+e.Request.Method] = true
				}
				var filtered []har.Entries
				for _, entry := range result.Entries {
					if redirectSet[entry.Request.URL+entry.Request.Method] {
						filtered = append(filtered, entry)
					}
				}
				result = &har.FilterResult{Entries: filtered}
			} else {
				result = redirectResult
			}
		}

		// 限制条数
		if limit > 0 {
			result.Limit(limit)
		}

		return internal.WriteOutput(cmd, buildListJSON(result), func() string {
			return formatFindTable(result)
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().Bool("regex", false, "使用正则表达式匹配URL")
	findCmd.Flags().String("method", "", "按HTTP方法过滤")
	findCmd.Flags().Int("status-code", 0, "按状态码过滤")
	findCmd.Flags().Int("status-min", 0, "最小状态码")
	findCmd.Flags().Int("status-max", 0, "最大状态码")
	findCmd.Flags().String("content-type", "", "按内容类型过滤")
	findCmd.Flags().String("domain", "", "按域名过滤")
	findCmd.Flags().StringSlice("header", nil, "按请求头过滤 (格式: name 或 name:value)")
	findCmd.Flags().String("resource-type", "", "按资源类型过滤")
	findCmd.Flags().Bool("errors", false, "查找所有错误请求(4xx/5xx)")
	findCmd.Flags().Bool("redirects", false, "查找所有重定向请求(3xx)")
	findCmd.Flags().Float64("slow", 0, "查找慢请求(最小毫秒数)")
	findCmd.Flags().IntP("limit", "n", 0, "限制输出条数 (0=全部)")
}

// formatFindTable 格式化搜索结果为tabwriter表格
func formatFindTable(result *har.FilterResult) string {
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