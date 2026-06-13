package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// performanceCmd 对HAR文件进行性能评分
var performanceCmd = &cobra.Command{
	Use:   "performance",
	Short: "对HAR文件进行性能评分",
	Long: `对HAR文件中的请求进行综合性能评分，生成评级和优化建议。

评分维度包括：TTFB（首字节时间）、总加载时间、请求数量、
传输大小、缓存效率和压缩率。

示例:
  har -f capture.har performance
  har -f capture.har performance --format json`,
	RunE: runPerformance,
}

func init() {
	rootCmd.AddCommand(performanceCmd)
}

func runPerformance(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	report := h.PerformanceScore()

	return internal.WriteOutput(cmd, report, func() string {
		return formatPerformanceReport(report)
	}, nil)
}

// formatPerformanceReport 格式化性能评分报告为文本
func formatPerformanceReport(report *har.PerformanceReport) string {
	var sb strings.Builder

	sb.WriteString("性能评分报告\n")
	sb.WriteString("============\n")

	// 总分和等级
	sb.WriteString(fmt.Sprintf("总分: %.1f/100  等级: %s\n\n", report.OverallScore, report.Grade()))

	// 分类评分
	sb.WriteString("分类评分:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-20s %-10s %-10s %s\n", "类别", "分数", "权重", "状态"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, cat := range report.Categories {
		status := "优秀"
		if cat.Score < 50 {
			status = "差"
		} else if cat.Score < 70 {
			status = "一般"
		} else if cat.Score < 90 {
			status = "良好"
		}
		sb.WriteString(fmt.Sprintf("%-20s %-10.1f %-10.1f %s\n",
			cat.Name, cat.Score, cat.Weight, status))
	}

	// 发现的问题
	for _, cat := range report.Categories {
		if len(cat.Findings) > 0 {
			sb.WriteString(fmt.Sprintf("\n%s 详情:\n", cat.Name))
			for _, f := range cat.Findings {
				sb.WriteString(fmt.Sprintf("  - [%s] %s\n", f.Type, f.Title))
				if f.Description != "" {
					sb.WriteString(fmt.Sprintf("    %s\n", f.Description))
				}
				if f.Impact != "" {
					sb.WriteString(fmt.Sprintf("    影响: %s\n", f.Impact))
				}
			}
		}
	}

	// 优化建议
	if len(report.Recommendations) > 0 {
		sb.WriteString("\n优化建议:\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		for i, rec := range report.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
	}

	return sb.String()
}
