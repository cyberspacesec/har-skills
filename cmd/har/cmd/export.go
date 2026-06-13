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
	Short: "Export HAR file to another format",
	Long: `Export HAR file to a specified format. Supported formats:

  curl     - cURL commands
  wget     - Wget commands
  python   - Python requests code
  postman  - Postman Collection JSON
  xml      - XML format
  yaml     - YAML format
  json     - Standard HAR JSON format
  jsonl    - JSON Lines format (one entry per line)
  csv      - CSV table format
  markdown - Markdown table format
  html     - HTML table format
  text     - Plain text table format

Examples:
  har -f capture.har export curl
  har -f capture.har export python -o replay.py
  har -f capture.har export postman --index 0
  har -f capture.har export csv --filter "api/users"
  har -f capture.har export markdown -o report.md
  har -f capture.har export jsonl -o entries.jsonl`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().Int("index", -1, "Export only the entry at this index")
	exportCmd.Flags().String("filter", "", "URL filter pattern (export only matching entries)")
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
			return fmt.Errorf("index %d out of range (total %d entries)", idx, len(h.Log.Entries))
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
			return fmt.Errorf("failed to export Postman Collection: %w", jsonErr)
		}
		output = string(data)
	case "xml":
		xmlStr, xmlErr := h.ToXML()
		if xmlErr != nil {
			return fmt.Errorf("failed to export XML: %w", xmlErr)
		}
		output = xmlStr
	case "yaml":
		yamlStr, yamlErr := h.ToYAML()
		if yamlErr != nil {
			return fmt.Errorf("failed to export YAML: %w", yamlErr)
		}
		output = yamlStr
	case "json":
		data, jsonErr := h.ToJSON(true)
		if jsonErr != nil {
			return fmt.Errorf("failed to export JSON: %w", jsonErr)
		}
		output = string(data)
	case "jsonl":
		jsonlStr, jsonlErr := h.ToJSONLines()
		if jsonlErr != nil {
			return fmt.Errorf("failed to export JSONL: %w", jsonlErr)
		}
		output = jsonlStr
	case "csv":
		result, err := h.Convert(har.FormatCSV, har.DefaultConvertOptions())
		if err != nil {
			return fmt.Errorf("failed to export CSV: %w", err)
		}
		output = result
	case "markdown", "md":
		result, err := h.Convert(har.FormatMarkdown, har.DefaultConvertOptions())
		if err != nil {
			return fmt.Errorf("failed to export Markdown: %w", err)
		}
		output = result
	case "html":
		result, err := h.Convert(har.FormatHTML, har.DefaultConvertOptions())
		if err != nil {
			return fmt.Errorf("failed to export HTML: %w", err)
		}
		output = result
	case "text":
		result, err := h.Convert(har.FormatText, har.DefaultConvertOptions())
		if err != nil {
			return fmt.Errorf("failed to export text: %w", err)
		}
		output = result
	default:
		return fmt.Errorf("unsupported export format: %s (supported: curl, wget, python, postman, xml, yaml, json, jsonl, csv, markdown, html, text)", format)
	}

	return internal.WriteStringOutput(cmd, output)
}
