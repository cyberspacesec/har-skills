package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// extractCmd 提取响应内容
var extractCmd = &cobra.Command{
	Use:   "extract [url-pattern]",
	Short: "提取响应内容",
	Long: `提取匹配条目的响应内容。支持按URL模式匹配或按条目索引提取，
支持自动解码base64编码和gzip/deflate压缩的内容。
提取的内容默认输出到stdout，也可使用--output写入文件。`,
	Example: `  har -f capture.har extract
  har -f capture.har extract "api/users"
  har -f capture.har extract --index 0
  har -f capture.har extract --all --decode
  har -f capture.har extract --index 3 -o response.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)

		// 获取参数
		urlPattern := ""
		if len(args) > 0 {
			urlPattern = args[0]
		}
		entryIndex, _ := cmd.Flags().GetInt("index")
		decode, _ := cmd.Flags().GetBool("decode")
		extractAll, _ := cmd.Flags().GetBool("all")

		// 按索引提取
		if entryIndex >= 0 && entryIndex < len(h.Log.Entries) {
			return extractSingleEntry(cmd, &h.Log.Entries[entryIndex], decode)
		}

		// 过滤匹配条目
		var entries []har.Entries
		for _, entry := range h.Log.Entries {
			if urlPattern != "" && !strings.Contains(entry.Request.URL, urlPattern) {
				continue
			}
			entries = append(entries, entry)
		}

		// 提取所有匹配或仅第一个
		if extractAll {
			return extractMultipleEntries(cmd, entries, decode)
		}

		if len(entries) == 0 {
			fmt.Fprintln(os.Stderr, "未找到匹配的条目")
			return nil
		}

		return extractSingleEntry(cmd, &entries[0], decode)
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	extractCmd.Flags().Int("index", -1, "按索引提取指定条目")
	extractCmd.Flags().Bool("decode", true, "自动解码base64/压缩内容")
	extractCmd.Flags().Bool("all", false, "提取所有匹配条目")
}

// extractSingleEntry 提取单个条目的响应内容
func extractSingleEntry(cmd *cobra.Command, entry *har.Entries, decode bool) error {
	if decode {
		data, err := entry.DecodeContent()
		if err != nil {
			return fmt.Errorf("解码内容失败: %w", err)
		}
		if data == nil {
			fmt.Fprintln(os.Stderr, "该条目无响应内容")
			return nil
		}
		return internal.WriteStringOutput(cmd, string(data))
	}

	// 不解码，直接输出原始文本
	if entry.Response.Content.Text == "" {
		fmt.Fprintln(os.Stderr, "该条目无响应内容")
		return nil
	}
	return internal.WriteStringOutput(cmd, entry.Response.Content.Text)
}

// extractMultipleEntries 提取多个条目的响应内容
func extractMultipleEntries(cmd *cobra.Command, entries []har.Entries, decode bool) error {
	var sb strings.Builder

	for i, entry := range entries {
		if i > 0 {
			sb.WriteString("\n--- 分隔线 ---\n\n")
		}
		sb.WriteString(fmt.Sprintf("# 条目 #%d: %s %s\n", i, entry.Request.Method, entry.Request.URL))
		sb.WriteString(fmt.Sprintf("# 状态: %d %s\n", entry.Response.Status, entry.Response.StatusText))
		sb.WriteString(fmt.Sprintf("# MIME类型: %s\n\n", entry.Response.Content.MimeType))

		if decode {
			data, err := entry.DecodeContent()
			if err != nil {
				sb.WriteString(fmt.Sprintf("# 解码失败: %v\n", err))
				continue
			}
			if data != nil {
				sb.WriteString(string(data))
			} else {
				sb.WriteString("# 无内容\n")
			}
		} else {
			if entry.Response.Content.Text != "" {
				sb.WriteString(entry.Response.Content.Text)
			} else {
				sb.WriteString("# 无内容\n")
			}
		}
	}

	return internal.WriteStringOutput(cmd, sb.String())
}