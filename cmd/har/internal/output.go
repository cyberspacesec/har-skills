package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// OutputFormat 输出格式类型
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatCSV  OutputFormat = "csv"
	FormatYAML OutputFormat = "yaml"
)

// GetFormat 从命令行标志获取输出格式
func GetFormat(cmd *cobra.Command) OutputFormat {
	f, _ := cmd.Flags().GetString("format")
	switch f {
	case "json":
		return FormatJSON
	case "csv":
		return FormatCSV
	case "yaml":
		return FormatYAML
	default:
		return FormatText
	}
}

// GetOutputPath 从命令行标志获取输出文件路径
func GetOutputPath(cmd *cobra.Command) string {
	path, _ := cmd.Flags().GetString("output")
	return path
}

// NoHeader 从命令行标志获取是否隐藏表头
func NoHeader(cmd *cobra.Command) bool {
	nh, _ := cmd.Flags().GetBool("no-header")
	return nh
}

// WriteOutput 根据格式输出数据
// data: 用于 json/yaml 序列化的数据结构
// textFunc: text 格式的输出字符串
// csvFunc: csv 格式的输出字符串
func WriteOutput(cmd *cobra.Command, data interface{}, textFunc func() string, csvFunc func() string) error {
	format := GetFormat(cmd)
	outputPath := GetOutputPath(cmd)

	var output []byte
	var err error

	switch format {
	case FormatJSON:
		output, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("JSON序列化失败: %w", err)
		}
		output = append(output, '\n')
	case FormatCSV:
		if csvFunc != nil {
			output = []byte(csvFunc())
		} else {
			output, err = json.Marshal(data)
			if err != nil {
				return fmt.Errorf("CSV序列化失败: %w", err)
			}
		}
	case FormatYAML:
		// 使用SDK的YAML功能
		if yamlMarshaler, ok := data.(interface{ ToYAML() (string, error) }); ok {
			yamlStr, yamlErr := yamlMarshaler.ToYAML()
			if yamlErr != nil {
				return fmt.Errorf("YAML序列化失败: %w", yamlErr)
			}
			output = []byte(yamlStr)
		} else {
			// 简单回退：使用JSON缩进格式
			output, err = json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("序列化失败: %w", err)
			}
			output = append(output, '\n')
		}
	default: // text
		if textFunc != nil {
			output = []byte(textFunc())
		} else {
			output, err = json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("序列化失败: %w", err)
			}
			output = append(output, '\n')
		}
	}

	return WriteToFileOrStdout(outputPath, output)
}

// WriteStringOutput 输出字符串到文件或stdout
func WriteStringOutput(cmd *cobra.Command, content string) error {
	outputPath := GetOutputPath(cmd)
	return WriteToFileOrStdout(outputPath, []byte(content))
}

// WriteToFileOrStdout 将字节数据写入文件或标准输出
func WriteToFileOrStdout(path string, data []byte) error {
	if path != "" {
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("无法写入文件 '%s': %w", path, err)
		}
		fmt.Fprintf(os.Stderr, "已写入 %d 字节到 %s\n", len(data), path)
		return nil
	}
	_, err := os.Stdout.Write(data)
	return err
}

// FormatBytes 格式化字节数为人类可读字符串
func FormatBytes(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatDuration 格式化毫秒为人类可读字符串
func FormatDuration(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.1f ms", ms)
	}
	return fmt.Sprintf("%.2f s", ms/1000)
}
