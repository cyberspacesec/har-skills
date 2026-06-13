package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// waterfallCmd 生成瀑布流时间线分析
var waterfallCmd = &cobra.Command{
	Use:   "waterfall",
	Short: "生成瀑布流时间线分析",
	Long: `生成HAR文件的请求瀑布流时间线分析，展示各请求的
时间关系和详细计时阶段。

支持关键路径分析、并发度分析、SLA合规检查和页面计时指标。

示例:
  har -f capture.har waterfall
  har -f capture.har waterfall --critical-path
  har -f capture.har waterfall --concurrency
  har -f capture.har waterfall --sla "首页:/:2000" "API:/api:500"
  har -f capture.har waterfall --page-timings`,
	RunE: runWaterfall,
}

func init() {
	rootCmd.AddCommand(waterfallCmd)

	waterfallCmd.Flags().Bool("critical-path", false, "显示关键路径（最长的请求依赖链）")
	waterfallCmd.Flags().Bool("concurrency", false, "显示并发度时间线")
	waterfallCmd.Flags().StringSlice("sla", nil, "SLA规则 (格式: name:urlPattern:maxDurationMs)")
	waterfallCmd.Flags().Bool("page-timings", false, "显示页面计时指标")
}

func runWaterfall(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	showCriticalPath, _ := cmd.Flags().GetBool("critical-path")
	showConcurrency, _ := cmd.Flags().GetBool("concurrency")
	slaRules, _ := cmd.Flags().GetStringSlice("sla")
	showPageTimings, _ := cmd.Flags().GetBool("page-timings")

	// 根据标志决定输出内容
	if showCriticalPath {
		path := h.CriticalPath()
		return internal.WriteOutput(cmd, path, func() string {
			return formatCriticalPath(path)
		}, nil)
	}

	if showConcurrency {
		timeline := h.ConcurrencyTimeline()
		return internal.WriteOutput(cmd, timeline, func() string {
			return formatConcurrencyTimeline(timeline)
		}, nil)
	}

	if len(slaRules) > 0 {
		rules, err := parseSLARules(slaRules)
		if err != nil {
			return fmt.Errorf("解析SLA规则失败: %w", err)
		}
		results := h.SLACheck(rules)
		return internal.WriteOutput(cmd, results, func() string {
			return formatSLAResults(results)
		}, nil)
	}

	if showPageTimings {
		metrics := h.PageTimingMetrics()
		return internal.WriteOutput(cmd, metrics, func() string {
			return formatPageTimings(metrics)
		}, nil)
	}

	// 默认：显示瀑布流
	entries := h.Waterfall()
	return internal.WriteOutput(cmd, entries, func() string {
		return formatWaterfall(entries)
	}, nil)
}

// parseSLARules 解析SLA规则字符串
func parseSLARules(rules []string) ([]har.SLARule, error) {
	var result []har.SLARule
	for _, r := range rules {
		parts := strings.SplitN(r, ":", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("SLA规则格式应为 name:urlPattern:maxDurationMs, 实际: '%s'", r)
		}
		maxMs, err := time.ParseDuration(parts[2])
		if err != nil {
			// 尝试解析为毫秒数值
			var msInt int
			n, _ := fmt.Sscanf(parts[2], "%d", &msInt)
			if n != 1 {
				return nil, fmt.Errorf("无法解析SLA最大时长 '%s': %w", parts[2], err)
			}
			maxMs = time.Duration(msInt) * time.Millisecond
		}
		result = append(result, har.SLARule{
			Name:       strings.TrimSpace(parts[0]),
			URLPattern: strings.TrimSpace(parts[1]),
			MaxTime:    maxMs,
		})
	}
	return result, nil
}

// formatWaterfall 格式化瀑布流为ASCII文本
func formatWaterfall(entries []har.WaterfallEntry) string {
	if len(entries) == 0 {
		return "无瀑布流数据。\n"
	}

	var sb strings.Builder
	sb.WriteString("请求瀑布流\n")
	sb.WriteString("==========\n\n")

	// 计算总时间范围
	maxEnd := time.Duration(0)
	for _, e := range entries {
		if e.EndTime > maxEnd {
			maxEnd = e.EndTime
		}
	}

	// 每个条目一行，使用ASCII字符表示时间范围
	barWidth := 50 // ASCII条宽度
	scale := float64(barWidth) / float64(maxEnd.Milliseconds())

	for _, e := range entries {
		startPos := int(float64(e.StartTime.Milliseconds()) * scale)
		endPos := int(float64(e.EndTime.Milliseconds()) * scale)
		barLen := endPos - startPos
		if barLen < 1 {
			barLen = 1
		}

		urlDisplay := e.URL
		if len(urlDisplay) > 40 {
			urlDisplay = urlDisplay[:37] + "..."
		}

		// 构建时间条
		bar := strings.Repeat(" ", startPos) + strings.Repeat("#", barLen)

		sb.WriteString(fmt.Sprintf("#%2d %s %-6dms [%s]\n",
			e.Index, urlDisplay, e.Duration.Milliseconds(), bar))
	}

	sb.WriteString(fmt.Sprintf("\n总时长: %.1fms\n", float64(maxEnd.Milliseconds())))

	return sb.String()
}

// formatCriticalPath 格式化关键路径为文本
func formatCriticalPath(path []har.WaterfallEntry) string {
	if len(path) == 0 {
		return "无关键路径数据。\n"
	}

	var sb strings.Builder
	sb.WriteString("关键路径分析\n")
	sb.WriteString("============\n\n")

	totalDuration := time.Duration(0)
	for i, e := range path {
		totalDuration += e.Duration
		urlDisplay := e.URL
		if len(urlDisplay) > 60 {
			urlDisplay = urlDisplay[:57] + "..."
		}
		sb.WriteString(fmt.Sprintf("%d. #%d %s %s (%dms)\n",
			i+1, e.Index, e.Method, urlDisplay, e.Duration.Milliseconds()))
	}

	sb.WriteString(fmt.Sprintf("\n关键路径总耗时: %dms\n", totalDuration.Milliseconds()))
	return sb.String()
}

// formatConcurrencyTimeline 格式化并发度时间线为文本
func formatConcurrencyTimeline(timeline []har.ConcurrencyPoint) string {
	if len(timeline) == 0 {
		return "无并发度数据。\n"
	}

	var sb strings.Builder
	sb.WriteString("并发度时间线\n")
	sb.WriteString("============\n\n")

	sb.WriteString(fmt.Sprintf("%-12s %-6s %s\n", "时间", "并发数", "活跃条目"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, p := range timeline {
		indices := fmt.Sprintf("%v", p.ActiveEntries)
		if len(indices) > 40 {
			indices = indices[:37] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-12s %-6d %s\n",
			fmt.Sprintf("%.0fms", float64(p.Time.Milliseconds())), p.ActiveCount, indices))
	}

	return sb.String()
}

// formatSLAResults 格式化SLA检查结果为文本
func formatSLAResults(results []har.SLAResult) string {
	if len(results) == 0 {
		return "无SLA检查结果。\n"
	}

	var sb strings.Builder
	sb.WriteString("SLA合规检查\n")
	sb.WriteString("============\n\n")

	sb.WriteString(fmt.Sprintf("%-15s %-6s %-10s %-10s %s\n",
		"规则", "通过", "实际耗时", "最大允许", "超时"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		overshoot := ""
		if r.Overshoot > 0 {
			overshoot = fmt.Sprintf("+%dms", r.Overshoot.Milliseconds())
		}
		sb.WriteString(fmt.Sprintf("%-15s %-6s %-10s %-10s %s\n",
			r.Rule.Name, status,
			fmt.Sprintf("%dms", r.Actual.Milliseconds()),
			fmt.Sprintf("%dms", r.Rule.MaxTime.Milliseconds()),
			overshoot))
	}

	return sb.String()
}

// formatPageTimings 格式化页面计时指标为文本
func formatPageTimings(metrics *har.PageTimingMetrics) string {
	if metrics == nil {
		return "无页面计时数据。\n"
	}

	var sb strings.Builder
	sb.WriteString("页面计时指标\n")
	sb.WriteString("============\n\n")

	sb.WriteString(fmt.Sprintf("TTFB:             %dms\n", metrics.TTFB.Milliseconds()))
	sb.WriteString(fmt.Sprintf("DOMContentLoaded: %dms\n", metrics.DOMContentLoaded.Milliseconds()))
	sb.WriteString(fmt.Sprintf("OnLoad:           %dms\n", metrics.OnLoad.Milliseconds()))
	sb.WriteString(fmt.Sprintf("总时间:           %dms\n", metrics.TotalTime.Milliseconds()))
	sb.WriteString(fmt.Sprintf("DNS查询:          %dms\n", metrics.DNSLookup.Milliseconds()))
	sb.WriteString(fmt.Sprintf("连接时间:         %dms\n", metrics.ConnectTime.Milliseconds()))
	sb.WriteString(fmt.Sprintf("SSL时间:          %dms\n", metrics.SSLTime.Milliseconds()))

	return sb.String()
}