package cmd

import (
	"encoding/json"
	"fmt"

	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
	"github.com/spf13/cobra"
)

// mergeCmd 合并多个HAR文件
var mergeCmd = &cobra.Command{
	Use:   "merge <file1> [file2] ...",
	Short: "合并多个HAR文件",
	Long: `将多个HAR文件的条目合并到一个HAR文件中。
合并后的HAR文件使用第一个文件的版本和创建者信息。

示例:
  har merge capture1.har capture2.har               # 合并两个HAR文件
  har merge a.har b.har c.har --deduplicate         # 合并并去重
  har merge a.har b.har --sort-by-time=false        # 不按时间排序
  har merge a.har b.har -o merged.har               # 输出到文件`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMerge,
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().Bool("sort-by-time", true, "按时间排序合并后的条目")
	mergeCmd.Flags().Bool("deduplicate", false, "去重（按Method+URL去重，保留最新的）")
}

// runMerge 执行合并命令
func runMerge(cmd *cobra.Command, args []string) error {
	// 加载所有HAR文件
	hars := make([]*har.Har, 0, len(args))
	for _, path := range args {
		h := internal.LoadHarFromArg(path)
		hars = append(hars, h)
	}

	// 读取合并选项
	sortByTime, _ := cmd.Flags().GetBool("sort-by-time")
	deduplicate, _ := cmd.Flags().GetBool("deduplicate")

	options := har.MergeOptions{
		SortByTime:  sortByTime,
		Deduplicate: deduplicate,
	}

	// 执行合并
	merged := har.MergeWithOptions(options, hars...)

	// 序列化为JSON
	output, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}
	output = append(output, '\n')

	// 写入输出
	outputPath := internal.GetOutputPath(cmd)
	return internal.WriteToFileOrStdout(outputPath, output)
}