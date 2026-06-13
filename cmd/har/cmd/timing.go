package cmd

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// timingCmd 分析请求计时分解
var timingCmd = &cobra.Command{
	Use:   "timing",
	Short: "分析请求计时分解",
	Long: `分析HAR文件中请求的计时分解信息，包括阻塞、DNS解析、
TCP连接、SSL握手、发送、等待和接收各阶段耗时。
支持排序、限制条数和显示汇总统计。`,
	Example: `  har -f capture.har timing
  har -f capture.har timing --sort wait --limit 10
  har -f capture.har timing --summary
  har -f capture.har timing --filter "api/users"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)

		// 获取参数
		filter, _ := cmd.Flags().GetString("filter")
		sortBy, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")
		showSummary, _ := cmd.Flags().GetBool("summary")

		// 过滤条目
		var entries []har.Entries
		for _, entry := range h.Log.Entries {
			if filter != "" && !strings.Contains(entry.Request.URL, filter) {
				continue
			}
			entries = append(entries, entry)
		}

		// 排序
		sortEntries(entries, sortBy)

		// 限制条数
		if limit > 0 && limit < len(entries) {
			entries = entries[:limit]
		}

		// 汇总模式
		if showSummary {
			timingsSummary := h.TimingStatistics()
			return internal.WriteOutput(cmd, timingsSummary, func() string {
				return formatTimingSummary(timingsSummary)
			}, nil)
		}

		return internal.WriteOutput(cmd, buildTimingJSON(entries), func() string {
			return formatTimingTable(entries)
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(timingCmd)

	timingCmd.Flags().String("filter", "", "URL过滤字符串")
	timingCmd.Flags().String("sort", "time", "排序方式 (time, wait, dns, connect)")
	timingCmd.Flags().IntP("limit", "n", 0, "限制输出条数 (0=全部)")
	timingCmd.Flags().Bool("summary", false, "显示汇总统计")
}

// timingEntry 用于JSON输出的计时信息
type timingEntry struct {
	URL     string  `json:"url"`
	Total   float64 `json:"total"`
	Blocked float64 `json:"blocked"`
	DNS     float64 `json:"dns"`
	Connect float64 `json:"connect"`
	SSL     float64 `json:"ssl"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

// sortEntries 按指定字段排序条目
func sortEntries(entries []har.Entries, sortBy string) {
	switch sortBy {
	case "wait":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timings.Wait > entries[j].Timings.Wait
		})
	case "dns":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timings.DNS > entries[j].Timings.DNS
		})
	case "connect":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timings.Connect > entries[j].Timings.Connect
		})
	default: // time
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Time > entries[j].Time
		})
	}
}

// buildTimingJSON 构建JSON输出数据
func buildTimingJSON(entries []har.Entries) []timingEntry {
	result := make([]timingEntry, len(entries))
	for i, entry := range entries {
		result[i] = timingEntry{
			URL:     entry.Request.URL,
			Total:   entry.Time,
			Blocked: entry.Timings.Blocked,
			DNS:     entry.Timings.DNS,
			Connect: entry.Timings.Connect,
			SSL:     entry.Timings.Ssl,
			Send:    entry.Timings.Send,
			Wait:    entry.Timings.Wait,
			Receive: entry.Timings.Receive,
		}
	}
	return result
}

// formatTimingTable 格式化计时信息为tabwriter表格
func formatTimingTable(entries []har.Entries) string {
	var sb tabWriterBuf

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "URL\tTOTAL\tBLOCKED\tDNS\tCONNECT\tSSL\tSEND\tWAIT\tRECEIVE\n")

	for _, entry := range entries {
		url := entry.Request.URL
		// 截断过长的URL
		if len(url) > 60 {
			url = url[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			url,
			internal.FormatDuration(entry.Time),
			formatTimingValue(entry.Timings.Blocked),
			formatTimingValue(entry.Timings.DNS),
			formatTimingValue(entry.Timings.Connect),
			formatTimingValue(entry.Timings.Ssl),
			formatTimingValue(entry.Timings.Send),
			formatTimingValue(entry.Timings.Wait),
			formatTimingValue(entry.Timings.Receive),
		)
	}
	w.Flush()

	return sb.String()
}

// formatTimingSummary 格式化计时汇总信息
func formatTimingSummary(summary *har.TimingsSummary) string {
	var sb strings.Builder

	sb.WriteString("计时汇总\n")
	sb.WriteString("========\n")
	sb.WriteString(fmt.Sprintf("平均阻塞时间:   %s\n", internal.FormatDuration(summary.AvgBlocked)))
	sb.WriteString(fmt.Sprintf("平均DNS时间:    %s\n", internal.FormatDuration(summary.AvgDNS)))
	sb.WriteString(fmt.Sprintf("平均连接时间:   %s\n", internal.FormatDuration(summary.AvgConnect)))
	sb.WriteString(fmt.Sprintf("平均SSL时间:    %s\n", internal.FormatDuration(summary.AvgSSL)))
	sb.WriteString(fmt.Sprintf("平均发送时间:   %s\n", internal.FormatDuration(summary.AvgSend)))
	sb.WriteString(fmt.Sprintf("平均等待时间:   %s\n", internal.FormatDuration(summary.AvgWait)))
	sb.WriteString(fmt.Sprintf("平均接收时间:   %s\n", internal.FormatDuration(summary.AvgReceive)))

	sb.WriteString("\n最大值\n")
	sb.WriteString("------\n")
	sb.WriteString(fmt.Sprintf("最大阻塞时间:   %s\n", internal.FormatDuration(summary.MaxBlocked)))
	sb.WriteString(fmt.Sprintf("最大DNS时间:    %s\n", internal.FormatDuration(summary.MaxDNS)))
	sb.WriteString(fmt.Sprintf("最大连接时间:   %s\n", internal.FormatDuration(summary.MaxConnect)))
	sb.WriteString(fmt.Sprintf("最大SSL时间:    %s\n", internal.FormatDuration(summary.MaxSSL)))
	sb.WriteString(fmt.Sprintf("最大发送时间:   %s\n", internal.FormatDuration(summary.MaxSend)))
	sb.WriteString(fmt.Sprintf("最大等待时间:   %s\n", internal.FormatDuration(summary.MaxWait)))
	sb.WriteString(fmt.Sprintf("最大接收时间:   %s\n", internal.FormatDuration(summary.MaxReceive)))

	return sb.String()
}

// formatTimingValue 格式化计时值，负值显示为"-"
func formatTimingValue(v float64) string {
	if v < 0 {
		return "-"
	}
	return internal.FormatDuration(v)
}