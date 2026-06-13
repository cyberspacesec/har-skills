package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// headersCmd 显示请求和响应头部
var headersCmd = &cobra.Command{
	Use:   "headers [url-pattern]",
	Short: "显示请求和响应头部",
	Long: `显示匹配条目的请求和响应头部信息。
可以只显示请求头部或响应头部，也可以按头部名称过滤。`,
	Example: `  har -f capture.har headers
  har -f capture.har headers "api/users"
  har -f capture.har headers --request
  har -f capture.har headers --response --name content-type`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)

		// 获取参数
		urlPattern := ""
		if len(args) > 0 {
			urlPattern = args[0]
		}
		showRequest, _ := cmd.Flags().GetBool("request")
		showResponse, _ := cmd.Flags().GetBool("response")
		headerName, _ := cmd.Flags().GetString("name")
		limit, _ := cmd.Flags().GetInt("limit")

		// 默认同时显示请求和响应头部
		if !showRequest && !showResponse {
			showRequest = true
			showResponse = true
		}

		// 过滤条目
		var entries []har.Entries
		for _, entry := range h.Log.Entries {
			if urlPattern != "" && !strings.Contains(entry.Request.URL, urlPattern) {
				continue
			}
			entries = append(entries, entry)
		}

		// 限制条数
		if limit > 0 && limit < len(entries) {
			entries = entries[:limit]
		}

		return internal.WriteOutput(cmd, buildHeadersJSON(entries, showRequest, showResponse, headerName), func() string {
			return formatHeadersText(entries, showRequest, showResponse, headerName)
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(headersCmd)

	headersCmd.Flags().Bool("request", false, "仅显示请求头部")
	headersCmd.Flags().Bool("response", false, "仅显示响应头部")
	headersCmd.Flags().String("name", "", "按头部名称过滤(不区分大小写)")
	headersCmd.Flags().IntP("limit", "n", 1, "显示的条目数 (默认1)")
}

// headerEntry 用于JSON输出的头部信息
type headerEntry struct {
	Index    int               `json:"index"`
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Status   int               `json:"status"`
	Request  map[string]string `json:"requestHeaders,omitempty"`
	Response map[string]string `json:"responseHeaders,omitempty"`
}

// buildHeadersJSON 构建JSON输出数据
func buildHeadersJSON(entries []har.Entries, showRequest, showResponse bool, headerName string) []headerEntry {
	result := make([]headerEntry, 0, len(entries))
	for i, entry := range entries {
		he := headerEntry{
			Index:  i,
			URL:    entry.Request.URL,
			Method: entry.Request.Method,
			Status: entry.Response.Status,
		}
		if showRequest {
			he.Request = filterHeaders(entry.Request.Headers, headerName)
		}
		if showResponse {
			he.Response = filterHeaders(entry.Response.Headers, headerName)
		}
		result = append(result, he)
	}
	return result
}

// filterHeaders 过滤头部，按名称筛选
func filterHeaders(headers []har.Headers, name string) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		if name != "" && !strings.EqualFold(h.Name, name) {
			continue
		}
		result[h.Name] = h.Value
	}
	return result
}

// formatHeadersText 格式化头部信息为文本输出
func formatHeadersText(entries []har.Entries, showRequest, showResponse bool, headerName string) string {
	var sb strings.Builder

	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf("=== 条目 #%d ===\n", i))
		sb.WriteString(fmt.Sprintf("URL: %s %s\n", entry.Request.Method, entry.Request.URL))
		sb.WriteString(fmt.Sprintf("状态: %d %s\n", entry.Response.Status, entry.Response.StatusText))

		if showRequest {
			sb.WriteString("\n请求头部:\n")
			for _, h := range entry.Request.Headers {
				if headerName != "" && !strings.EqualFold(h.Name, headerName) {
					continue
				}
				sb.WriteString(fmt.Sprintf("  %s: %s\n", h.Name, h.Value))
			}
		}

		if showResponse {
			sb.WriteString("\n响应头部:\n")
			for _, h := range entry.Response.Headers {
				if headerName != "" && !strings.EqualFold(h.Name, headerName) {
					continue
				}
				sb.WriteString(fmt.Sprintf("  %s: %s\n", h.Name, h.Value))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}