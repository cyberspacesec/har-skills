package cmd

import (
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
	"github.com/spf13/cobra"
)

// diffCmd 比较两个HAR文件的差异
var diffCmd = &cobra.Command{
	Use:   "diff <file1> <file2>",
	Short: "比较两个HAR文件",
	Long: `比较两个HAR文件的差异，找出新增、删除和修改的请求。

示例:
  har diff capture1.har capture2.har              # 比较两个HAR文件
  har diff a.har b.har --ignore-headers=Cookie    # 忽略Cookie头部差异
  har diff a.har b.har --compare-by-url           # 按URL匹配而非索引
  har diff a.har b.har --include-body             # 比较响应体内容`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringSlice("ignore-headers", nil, "忽略的头部字段名（逗号分隔）")
	diffCmd.Flags().Bool("ignore-timings", true, "忽略时间差异")
	diffCmd.Flags().Bool("ignore-dates", true, "忽略日期差异")
	diffCmd.Flags().Bool("include-body", false, "比较响应体内容")
	diffCmd.Flags().Bool("compare-by-url", false, "按URL匹配（默认按索引+URL）")
}

// runDiff 执行差异比较命令
func runDiff(cmd *cobra.Command, args []string) error {
	// 加载两个HAR文件
	har1 := internal.LoadHarFromArg(args[0])
	har2 := internal.LoadHarFromArg(args[1])

	// 构建差异比较选项
	options := har.DefaultDiffOptions()

	// 从命令行标志读取选项
	ignoreHeaders, _ := cmd.Flags().GetStringSlice("ignore-headers")
	ignoreTimings, _ := cmd.Flags().GetBool("ignore-timings")
	ignoreDates, _ := cmd.Flags().GetBool("ignore-dates")
	includeBody, _ := cmd.Flags().GetBool("include-body")
	compareByURL, _ := cmd.Flags().GetBool("compare-by-url")

	options.IgnoreHeaders = ignoreHeaders
	options.IgnoreTimings = ignoreTimings
	options.IgnoreDates = ignoreDates
	options.IncludeBody = includeBody
	options.CompareByURL = compareByURL

	// 执行差异比较
	diffResult := har.Diff(har1, har2, options)

	// 根据输出格式输出结果
	return internal.WriteOutput(cmd, diffResult,
		func() string { return diffResult.Report(har.FormatText) },
		func() string { return diffResult.Report(har.FormatCSV) },
	)
}