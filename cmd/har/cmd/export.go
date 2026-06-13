package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// exportCmd 导出HAR文件为其他格式
var exportCmd = &cobra.Command{
	Use:   "export [format]",
	Short: "导出HAR文件为其他格式",
	Long: `将HAR文件导出为指定的格式。支持以下格式：

  curl    - cURL命令
  wget    - Wget命令
  python  - Python requests代码
  postman - Postman Collection JSON
  xml     - XML格式
  yaml    - YAML格式
  json    - 标准HAR JSON格式

示例:
  har -f capture.har export curl
  har -f capture.har export python -o replay.py
  har -f capture.har export postman --index 0
  har -f capture.har export yaml --filter "api/users"`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().Int("index", -1, "仅导出指定索引的条目")
	exportCmd.Flags().String("filter", "", "URL过滤模式 (仅导出匹配的条目)")
}

func runExport(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)
	format := args[0]

	// 如果指定了过滤，先过滤
	filterPattern, _ := cmd.Flags().GetString("filter")
	if filterPattern != "" {
		opts := har.NewFilterOptions(har.WithFilterURL(filterPattern))
		filtered := h.Filter(opts)
		harResult := filtered.ToHar()
		h = harResult
	}

	// 如果指定了索引，截取单个条目
	idx, _ := cmd.Flags().GetInt("index")
	if idx >= 0 {
		if idx >= len(h.Log.Entries) {
			return fmt.Errorf("索引 %d 超出范围 (共 %d 个条目)", idx, len(h.Log.Entries))
		}
		h = h.Clone()
		h.Log.Entries = h.Log.Entries[idx : idx+1]
	}

	var output string

	switch format {
	case "curl":
		output = h.ToCurl()
	case "wget":
		output = h.ToWget()
	case "python":
		output = h.ToPythonRequests()
	case "postman":
		data, jsonErr := h.ToPostmanCollection()
		if jsonErr != nil {
			return fmt.Errorf("导出Postman Collection失败: %w", jsonErr)
		}
		output = string(data)
	case "xml":
		xmlStr, xmlErr := h.ToXML()
		if xmlErr != nil {
			return fmt.Errorf("导出XML失败: %w", xmlErr)
		}
		output = xmlStr
	case "yaml":
		yamlStr, yamlErr := h.ToYAML()
		if yamlErr != nil {
			return fmt.Errorf("导出YAML失败: %w", yamlErr)
		}
		output = yamlStr
	case "json":
		data, jsonErr := h.ToJSON(true)
		if jsonErr != nil {
			return fmt.Errorf("导出JSON失败: %w", jsonErr)
		}
		output = string(data)
	default:
		return fmt.Errorf("不支持的导出格式: %s (支持: curl, wget, python, postman, xml, yaml, json)", format)
	}

	return internal.WriteStringOutput(cmd, output)
}
