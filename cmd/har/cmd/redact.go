package cmd

import (
	"encoding/json"
	"fmt"

	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
	"github.com/spf13/cobra"
)

// redactCmd 脱敏HAR文件中的敏感数据
var redactCmd = &cobra.Command{
	Use:   "redact",
	Short: "脱敏敏感数据",
	Long: `将HAR文件中的敏感数据（如密码、令牌、API密钥等）替换为占位符。

支持自定义脱敏字段：
  - 指定头部字段名（如 Authorization、X-Api-Key）
  - 指定Cookie名称（如 session、token）
  - 指定查询参数名（如 password、api_key）
  - 指定POST字段名（如 password、secret）
  - IP地址匿名化（将最后一段替换为0）

示例:
  har redact -f capture.har                             # 使用默认脱敏规则
  har redact -f capture.har --defaults=false            # 不使用默认规则
  har redact -f capture.har --header=X-Custom-Key      # 添加自定义头部
  har redact -f capture.har --redact-ips                # 匿名化IP地址
  har redact -f capture.har --replacement="***"          # 自定义替换文本
  har redact -f capture.har --in-place                  # 原地修改文件`,
	RunE: runRedact,
}

func init() {
	rootCmd.AddCommand(redactCmd)

	redactCmd.Flags().Bool("defaults", true, "使用默认脱敏规则")
	redactCmd.Flags().StringSlice("header", nil, "额外脱敏的头部字段名")
	redactCmd.Flags().StringSlice("cookie", nil, "额外脱敏的Cookie名称")
	redactCmd.Flags().StringSlice("query-param", nil, "额外脱敏的查询参数名")
	redactCmd.Flags().StringSlice("post-field", nil, "额外脱敏的POST字段名")
	redactCmd.Flags().String("replacement", "[REDACTED]", "替换文本")
	redactCmd.Flags().Bool("redact-ips", false, "匿名化IP地址")
	redactCmd.Flags().Bool("in-place", false, "原地修改文件")
}

// runRedact 执行脱敏命令
func runRedact(cmd *cobra.Command, args []string) error {
	// 加载HAR文件
	h := internal.LoadHar(cmd, args)

	// 构建脱敏选项
	var opts har.RedactOptions

	useDefaults, _ := cmd.Flags().GetBool("defaults")
	if useDefaults {
		opts = har.DefaultRedactOptions()
	} else {
		opts = har.RedactOptions{
			Replacement: "[REDACTED]",
		}
	}

	// 读取命令行标志
	extraHeaders, _ := cmd.Flags().GetStringSlice("header")
	extraCookies, _ := cmd.Flags().GetStringSlice("cookie")
	extraQueryParams, _ := cmd.Flags().GetStringSlice("query-param")
	extraPostFields, _ := cmd.Flags().GetStringSlice("post-field")
	replacement, _ := cmd.Flags().GetString("replacement")
	redactIPs, _ := cmd.Flags().GetBool("redact-ips")
	inPlace, _ := cmd.Flags().GetBool("in-place")

	// 合并额外的脱敏字段
	opts.Headers = append(opts.Headers, extraHeaders...)
	opts.Cookies = append(opts.Cookies, extraCookies...)
	opts.QueryParams = append(opts.QueryParams, extraQueryParams...)
	opts.PostDataFields = append(opts.PostDataFields, extraPostFields...)
	opts.Replacement = replacement
	opts.RedactIPs = redactIPs

	// 执行脱敏
	var result *har.Har
	if inPlace {
		h.RedactInPlace(opts)
		result = h
	} else {
		result = h.Redact(opts)
	}

	// 序列化为JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}
	output = append(output, '\n')

	// 写入输出
	outputPath := internal.GetOutputPath(cmd)
	if inPlace && outputPath == "" {
		// 原地修改模式：需要从 --file 获取路径并写回
		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" && filePath != "-" {
			outputPath = filePath
		}
	}

	return internal.WriteToFileOrStdout(outputPath, output)
}
