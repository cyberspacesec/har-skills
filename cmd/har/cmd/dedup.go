package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// dedupCmd 查找或移除重复请求
var dedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "查找或移除重复请求",
	Long: `查找HAR文件中的重复或近似重复请求，或去除重复请求生成新的HAR文件。

支持三种去重策略：
  - exact:       精确URL匹配
  - pattern:     忽略缓存破坏器等参数的URL模式匹配（默认）
  - content-hash: 基于内容哈希匹配

示例:
  har -f capture.har dedup
  har -f capture.har dedup --strategy exact
  har -f capture.har dedup --strategy content-hash --compare-headers --compare-body
  har -f capture.har dedup --remove -o cleaned.har
  har -f capture.har dedup --ignore-param "timestamp" --ignore-param "_"`,
	RunE: runDedup,
}

func init() {
	rootCmd.AddCommand(dedupCmd)

	dedupCmd.Flags().String("strategy", "pattern", "去重策略 (exact/pattern/content-hash)")
	dedupCmd.Flags().StringSlice("ignore-param", nil, "去重时忽略的查询参数")
	dedupCmd.Flags().Bool("compare-headers", false, "比较时包含请求头")
	dedupCmd.Flags().Bool("compare-body", false, "比较时包含请求体")
	dedupCmd.Flags().Bool("remove", false, "去除重复请求，输出清理后的HAR文件")
}

func runDedup(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	// 构建去重选项
	strategyName, _ := cmd.Flags().GetString("strategy")
	ignoreParams, _ := cmd.Flags().GetStringSlice("ignore-param")
	compareHeaders, _ := cmd.Flags().GetBool("compare-headers")
	compareBody, _ := cmd.Flags().GetBool("compare-body")
	doRemove, _ := cmd.Flags().GetBool("remove")

	// 解析策略
	strategy := parseDedupStrategy(strategyName)

	// 从默认选项开始，覆盖指定字段
	opts := har.DefaultDeduplicateOptions()
	opts.Strategy = strategy
	opts.CompareHeaders = compareHeaders
	opts.CompareBody = compareBody

	if len(ignoreParams) > 0 {
		opts.IgnoreParams = ignoreParams
	}

	if doRemove {
		// 去除重复，输出清理后的HAR
		deduped := h.Deduplicate(opts)
		data, err := deduped.ToJSON(true)
		if err != nil {
			return fmt.Errorf("序列化HAR失败: %w", err)
		}
		return internal.WriteStringOutput(cmd, string(data))
	}

	// 默认：查找重复
	groups := h.FindDuplicates(opts)

	return internal.WriteOutput(cmd, groups, func() string {
		return formatDuplicateGroups(groups)
	}, nil)
}

// parseDedupStrategy 解析去重策略字符串
func parseDedupStrategy(s string) har.DedupStrategy {
	switch strings.ToLower(s) {
	case "exact":
		return har.DedupExactURL
	case "pattern":
		return har.DedupURLPattern
	case "content-hash":
		return har.DedupContentHash
	default:
		return har.DedupURLPattern
	}
}

// formatDuplicateGroups 格式化重复分组为文本
func formatDuplicateGroups(groups []har.DuplicateGroup) string {
	if len(groups) == 0 {
		return "未发现重复请求。\n"
	}

	var sb strings.Builder
	sb.WriteString("重复请求分析\n")
	sb.WriteString("============\n\n")

	totalDuplicates := 0
	for _, g := range groups {
		totalDuplicates += g.Count - 1 // 减1因为第一个不算重复
	}

	sb.WriteString(fmt.Sprintf("发现 %d 组重复请求，共 %d 个重复条目\n\n",
		len(groups), totalDuplicates))

	sb.WriteString(fmt.Sprintf("%-60s %-6s %s\n", "去重键", "数量", "条目索引"))
	sb.WriteString(strings.Repeat("-", 90) + "\n")

	for _, g := range groups {
		keyDisplay := g.Key
		if len(keyDisplay) > 60 {
			keyDisplay = keyDisplay[:57] + "..."
		}
		indices := fmt.Sprintf("%v", g.EntryIndices)
		if len(indices) > 30 {
			indices = indices[:27] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-60s %-6d %s\n", keyDisplay, g.Count, indices))
	}

	return sb.String()
}