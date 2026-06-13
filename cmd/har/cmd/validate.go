package cmd

import (
	"fmt"

	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
	"github.com/spf13/cobra"
)

// validateCmd 验证HAR文件是否符合规范
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证HAR文件",
	Long: `验证HAR文件是否符合HAR规范。

支持标准验证和严格验证模式：
  - 标准验证：检查基本结构和必填字段
  - 严格验证：额外检查交叉引用、HTTP方法、状态码范围等
  - 时间一致性验证：检查Time字段与Timings各字段之和的一致性

示例:
  har validate -f capture.har                      # 标准验证
  har validate -f capture.har --strict             # 严格验证
  har validate -f capture.har --timings-tolerance 5  # 时间一致性容差5ms
  har validate -f capture.har --strict --timings-tolerance 0`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().Bool("strict", false, "启用严格验证模式")
	validateCmd.Flags().Float64("timings-tolerance", 10, "时间一致性容差（毫秒），0表示严格一致")
}

// runValidate 执行验证命令
func runValidate(cmd *cobra.Command, args []string) error {
	// 加载HAR文件
	h := internal.LoadHar(cmd, args)

	strict, _ := cmd.Flags().GetBool("strict")
	timingsTolerance, _ := cmd.Flags().GetFloat64("timings-tolerance")

	// 收集所有验证错误
	var allErrors []*har.ValidationError

	// 执行标准或严格验证
	if strict {
		if err := har.ValidateStrict(h); err != nil {
			collectErrors(err, &allErrors)
		}
	} else {
		if err := har.ValidateHarFile(h); err != nil {
			collectErrors(err, &allErrors)
		}
	}

	// 执行时间一致性验证（容差>=0时启用）
	if timingsTolerance >= 0 {
		timingErrors := har.ValidateTimingsConsistency(h, timingsTolerance)
		allErrors = append(allErrors, timingErrors...)
	}

	// 根据输出格式输出结果
	return internal.WriteOutput(cmd, allErrors,
		func() string { return formatValidateText(allErrors) },
		nil,
	)
}

// collectErrors 从HarError中提取ValidationError列表
func collectErrors(err error, errors *[]*har.ValidationError) {
	if err == nil {
		return
	}

	// 尝试作为HarError处理
	if harErr, ok := err.(*har.HarError); ok {
		for _, pe := range harErr.GetPartialErrors() {
			// HarError包含Field和Message，转换为ValidationError
			ve := &har.ValidationError{
				Field:   pe.Field,
				Message: pe.Message,
			}
			*errors = append(*errors, ve)
		}
		return
	}

	// 其他错误类型
	*errors = append(*errors, &har.ValidationError{
		Field:   "",
		Message: err.Error(),
	})
}

// formatValidateText 格式化验证结果的文本输出
func formatValidateText(errors []*har.ValidationError) string {
	if len(errors) == 0 {
		return "✓ Valid\n"
	}

	result := fmt.Sprintf("✗ 发现 %d 个验证错误:\n\n", len(errors))
	for i, e := range errors {
		if e.Field != "" {
			result += fmt.Sprintf("  %d. [%s] %s: %s\n", i+1, e.Rule, e.Field, e.Message)
		} else {
			result += fmt.Sprintf("  %d. %s\n", i+1, e.Message)
		}
	}
	return result
}