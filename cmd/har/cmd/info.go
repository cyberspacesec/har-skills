package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// infoCmd 显示HAR文件概要和统计信息
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "显示HAR文件概要和统计信息",
	Long: `显示HAR文件的详细概要信息，包括版本、创建者、页面数、
请求数、总传输量、时间百分位数、状态码分布、方法分布、
域名分布和内容类型分布等统计信息。`,
	Example: `  har -f capture.har info
  har -f capture.har info --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)
		stats := h.Statistics()

		return internal.WriteOutput(cmd, stats, func() string {
			return formatInfoText(h, stats)
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

// formatInfoText 格式化HAR概要信息为文本输出
func formatInfoText(h *har.Har, stats *har.HarStatistics) string {
	var sb strings.Builder

	// 基本信息
	sb.WriteString("HAR 文件概要\n")
	sb.WriteString("=============\n")
	sb.WriteString(fmt.Sprintf("版本:          %s\n", h.Log.Version))
	if h.Log.Creator.Name != "" {
		sb.WriteString(fmt.Sprintf("创建者:        %s %s\n", h.Log.Creator.Name, h.Log.Creator.Version))
	}
	if h.Log.Browser.Name != "" {
		sb.WriteString(fmt.Sprintf("浏览器:        %s %s\n", h.Log.Browser.Name, h.Log.Browser.Version))
	}
	sb.WriteString(fmt.Sprintf("页面数:        %d\n", len(h.Log.Pages)))
	sb.WriteString(fmt.Sprintf("请求数:        %d\n", stats.TotalRequests))

	// 传输量
	sb.WriteString(fmt.Sprintf("总传输量:      %s\n", internal.FormatBytes(int(stats.TotalTransferred))))
	sb.WriteString(fmt.Sprintf("总未压缩量:    %s\n", internal.FormatBytes(int(stats.TotalUncompressed))))

	// 时间统计
	sb.WriteString("\n时间统计\n")
	sb.WriteString("--------\n")
	sb.WriteString(fmt.Sprintf("总时间:        %s\n", internal.FormatDuration(stats.TotalTime)))
	sb.WriteString(fmt.Sprintf("平均请求时间:  %s\n", internal.FormatDuration(stats.AvgTime)))
	sb.WriteString(fmt.Sprintf("中位数时间:    %s\n", internal.FormatDuration(stats.MedianTime)))
	sb.WriteString(fmt.Sprintf("P95 时间:      %s\n", internal.FormatDuration(stats.P95Time)))
	sb.WriteString(fmt.Sprintf("P99 时间:      %s\n", internal.FormatDuration(stats.P99Time)))
	sb.WriteString(fmt.Sprintf("最慢请求:      %s\n", internal.FormatDuration(stats.MaxTime)))
	sb.WriteString(fmt.Sprintf("最快请求:      %s\n", internal.FormatDuration(stats.MinTime)))

	// 错误和重定向
	sb.WriteString(fmt.Sprintf("\n错误请求数:    %d\n", stats.ErrorCount))
	sb.WriteString(fmt.Sprintf("重定向数:      %d\n", stats.RedirectCount))

	// 状态码分布
	statusDist := h.StatusCodeDistribution()
	sb.WriteString("\n状态码分布\n")
	sb.WriteString("----------\n")
	for _, code := range sortedKeysInt(statusDist) {
		sb.WriteString(fmt.Sprintf("  %d: %d\n", code, statusDist[code]))
	}

	// 方法分布
	methodDist := h.MethodDistribution()
	sb.WriteString("\n方法分布\n")
	sb.WriteString("--------\n")
	for _, method := range sortedKeysStr(methodDist) {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", method, methodDist[method]))
	}

	// 域名分布（前10）
	sb.WriteString("\n域名分布（前10）\n")
	sb.WriteString("----------------\n")
	topDomains := topN(stats.Domains, 10)
	for _, d := range topDomains {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", d.key, d.count))
	}

	// 内容类型分布（前10）
	contentDist := h.ContentTypeDistribution()
	sb.WriteString("\n内容类型分布（前10）\n")
	sb.WriteString("--------------------\n")
	topContentTypes := topN(contentDist, 10)
	for _, ct := range topContentTypes {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", ct.key, ct.count))
	}

	return sb.String()
}

// keyValue 用于排序的键值对
type keyValue struct {
	key   string
	count int
}

// topN 获取map中值最大的前N个键
func topN(m map[string]int, n int) []keyValue {
	var items []keyValue
	for k, v := range m {
		items = append(items, keyValue{k, v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].count > items[j].count
	})
	if len(items) > n {
		items = items[:n]
	}
	return items
}

// sortedKeysInt 返回map中排序后的int键
func sortedKeysInt(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// sortedKeysStr 返回map中排序后的string键
func sortedKeysStr(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
