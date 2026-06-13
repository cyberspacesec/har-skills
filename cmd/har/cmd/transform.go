package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// transformCmd 转换HAR文件中的请求
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "转换HAR文件中的请求",
	Long: `对HAR文件中的请求执行各种转换操作，包括URL重写、
请求头增删、协议变更、查询参数移除等。

转换规则按顺序依次应用，结果输出为新的HAR JSON。

示例:
  har -f capture.har transform --rewrite-url "http://localhost->https://prod.example.com"
  har -f capture.har transform --remove-header "X-Debug" --add-header "X-Env:production"
  har -f capture.har transform --change-scheme "http->https" --remove-query-param "_"`,
	RunE: runTransform,
}

func init() {
	rootCmd.AddCommand(transformCmd)

	transformCmd.Flags().StringSlice("rewrite-url", nil, "URL重写规则 (格式: from->to)")
	transformCmd.Flags().StringSlice("remove-header", nil, "移除指定请求头")
	transformCmd.Flags().StringSlice("add-header", nil, "添加请求头 (格式: name:value)")
	transformCmd.Flags().String("add-header-target", "both", "添加请求头目标 (request/response/both)")
	transformCmd.Flags().StringSlice("change-scheme", nil, "协议变更规则 (格式: from->to)")
	transformCmd.Flags().StringSlice("remove-query-param", nil, "移除指定查询参数")
}

func runTransform(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	var rules []har.TransformRule

	// 解析 --rewrite-url 规则
	rewriteURLs, _ := cmd.Flags().GetStringSlice("rewrite-url")
	for _, rw := range rewriteURLs {
		from, to, err := parseArrowRule(rw)
		if err != nil {
			return fmt.Errorf("无效的rewrite-url规则 '%s': %w", rw, err)
		}
		rules = append(rules, har.TransformRule{
			Type:        har.TransformURLRewrite,
			Pattern:     from,
			Replacement: to,
		})
	}

	// 解析 --remove-header
	removeHeaders, _ := cmd.Flags().GetStringSlice("remove-header")
	for _, name := range removeHeaders {
		rules = append(rules, har.TransformRule{
			Type:       har.TransformHeaderRemove,
			HeaderName: name,
		})
	}

	// 解析 --add-header
	addHeaders, _ := cmd.Flags().GetStringSlice("add-header")
	addHeaderTarget, _ := cmd.Flags().GetString("add-header-target")
	if len(addHeaders) > 0 {
		headersMap := make(map[string]string)
		for _, h := range addHeaders {
			name, value, err := parseColonKeyValue(h)
			if err != nil {
				return fmt.Errorf("无效的add-header规则 '%s': %w", h, err)
			}
			headersMap[name] = value
		}
		// AddHeaders 有独立逻辑，直接调用
		result := h.AddHeaders(headersMap, addHeaderTarget)
		h = result
	}

	// 解析 --change-scheme
	changeSchemes, _ := cmd.Flags().GetStringSlice("change-scheme")
	for _, cs := range changeSchemes {
		from, to, err := parseArrowRule(cs)
		if err != nil {
			return fmt.Errorf("无效的change-scheme规则 '%s': %w", cs, err)
		}
		rules = append(rules, har.TransformRule{
			Type:        har.TransformSchemeChange,
			Pattern:     from,
			Replacement: to,
		})
	}

	// 解析 --remove-query-param
	removeQueryParams, _ := cmd.Flags().GetStringSlice("remove-query-param")
	for _, param := range removeQueryParams {
		rules = append(rules, har.TransformRule{
			Type:    har.TransformQueryParamRemove,
			Pattern: param,
		})
	}

	// 应用通用转换规则
	if len(rules) > 0 {
		h = h.Transform(rules)
	}

	// 输出转换后的HAR JSON
	data, err := h.ToJSON(true)
	if err != nil {
		return fmt.Errorf("序列化HAR失败: %w", err)
	}

	return internal.WriteStringOutput(cmd, string(data))
}

// parseArrowRule 解析 "from->to" 格式的规则
func parseArrowRule(s string) (string, string, error) {
	parts := strings.SplitN(s, "->", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("格式应为 from->to")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

// parseColonKeyValue 解析 "name:value" 格式的键值对
func parseColonKeyValue(s string) (string, string, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("格式应为 name:value")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}
