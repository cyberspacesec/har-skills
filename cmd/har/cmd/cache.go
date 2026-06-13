package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// cacheCmd 分析HAR文件中的缓存头部
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "分析HAR文件中的缓存头部",
	Long: `分析HAR文件中响应的缓存相关头部，评估每个请求的可缓存性。

检查 Cache-Control、ETag、Last-Modified、Vary 等头部，
输出每个条目的缓存评估结果。

示例:
  har -f capture.har cache
  har -f capture.har cache --non-cacheable
  har -f capture.har cache --url "https://api.example.com/data"
  har -f capture.har cache --format json`,
	RunE: runCache,
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	cacheCmd.Flags().Bool("non-cacheable", false, "仅显示不可缓存的条目")
	cacheCmd.Flags().String("url", "", "仅显示指定URL的缓存评估")
}

func runCache(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	report := h.CacheAnalysis()

	// 过滤
	showNonCacheable, _ := cmd.Flags().GetBool("non-cacheable")
	specificURL, _ := cmd.Flags().GetString("url")

	if showNonCacheable {
		report.Assessments = report.NonCacheableEntries()
	}

	if specificURL != "" {
		assessment := report.FindByURL(specificURL)
		if assessment == nil {
			return fmt.Errorf("未找到URL '%s' 的缓存评估", specificURL)
		}
		// 替换为单个评估
		report.Assessments = []har.CacheEntryAssessment{*assessment}
	}

	return internal.WriteOutput(cmd, report, func() string {
		return formatCacheReport(report)
	}, nil)
}

// formatCacheReport 格式化缓存分析报告为文本
func formatCacheReport(report *har.CacheReport) string {
	var sb strings.Builder

	sb.WriteString("缓存分析报告\n")
	sb.WriteString("============\n")
	sb.WriteString(fmt.Sprintf("可缓存: %d / 不可缓存: %d / 缓存效率: %.1f%%\n\n",
		report.CacheableCount, report.NonCacheableCount, report.CacheEfficiency*100))

	if len(report.Assessments) == 0 {
		sb.WriteString("无缓存评估数据。\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("%-4s %-50s %-10s %-10s %-10s %-6s\n",
		"#", "URL", "可缓存", "类型", "Max-Age", "ETag"))
	sb.WriteString(strings.Repeat("-", 90) + "\n")

	for _, a := range report.Assessments {
		cacheable := "否"
		if a.Cacheable {
			cacheable = "是"
		}
		hasETag := "无"
		if a.HasETag {
			hasETag = "有"
		}
		maxAge := "N/A"
		if a.MaxAge > 0 {
			maxAge = fmt.Sprintf("%.0fs", a.MaxAge.Seconds())
		}
		urlDisplay := a.URL
		if len(urlDisplay) > 50 {
			urlDisplay = urlDisplay[:47] + "..."
		}

		sb.WriteString(fmt.Sprintf("%-4d %-50s %-10s %-10s %-10s %-6s\n",
			a.EntryIndex, urlDisplay, cacheable, a.CacheType, maxAge, hasETag))
	}

	return sb.String()
}