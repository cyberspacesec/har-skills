package har

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// YAMLFormat 常量
const FormatYAML ConvertFormat = "yaml"

// ToYAML 将HAR对象转换为YAML格式字符串
//
// 该方法将HAR数据转为YAML格式，便于阅读和编辑。
// 注意：该实现不依赖外部YAML库，使用内置的JSON转YAML转换。
func (h *Har) ToYAML() (string, error) {
	if h == nil {
		return "", nil
	}

	// 先转为JSON
	jsonData, err := h.ToJSON(true)
	if err != nil {
		return "", err
	}

	// 将JSON转为YAML格式
	return jsonToYAML(jsonData), nil
}

// SaveAsYAML 将HAR对象保存为YAML文件
func (h *Har) SaveAsYAML(filePath string) error {
	yamlData, err := h.ToYAML()
	if err != nil {
		return err
	}
	return writeToFile(filePath, []byte(yamlData))
}

// jsonToYAML 简单的JSON到YAML转换器
// 不依赖外部库，提供基本的YAML输出
func jsonToYAML(jsonData []byte) string {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return string(jsonData)
	}
	return valueToYAML(data, 0)
}

// valueToYAML 递归地将值转为YAML格式
func valueToYAML(v interface{}, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	switch val := v.(type) {
	case map[string]interface{}:
		first := true
		for k, v := range val {
			if !first {
				sb.WriteString("\n")
			}
			first = false

			switch child := v.(type) {
			case map[string]interface{}:
				sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
				sb.WriteString(valueToYAML(child, indent+1))
			case []interface{}:
				sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
				sb.WriteString(arrayToYAML(child, indent+1))
			case nil:
				sb.WriteString(fmt.Sprintf("%s%s: null\n", prefix, k))
			case string:
				if strings.ContainsAny(child, ":{}[]&*?|>-!%@`\"'\n") || child == "" {
					sb.WriteString(fmt.Sprintf("%s%s: \"%s\"\n", prefix, k, escapeYAMLString(child)))
				} else {
					sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, k, child))
				}
			case float64:
				if child == float64(int64(child)) {
					sb.WriteString(fmt.Sprintf("%s%s: %d\n", prefix, k, int64(child)))
				} else {
					sb.WriteString(fmt.Sprintf("%s%s: %g\n", prefix, k, child))
				}
			case bool:
				sb.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, k, child))
			default:
				sb.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, k, child))
			}
		}
	case []interface{}:
		sb.WriteString(arrayToYAML(val, indent))
	case nil:
		sb.WriteString(fmt.Sprintf("%snull\n", prefix))
	case string:
		sb.WriteString(fmt.Sprintf("%s%s\n", prefix, val))
	case float64:
		if val == float64(int64(val)) {
			sb.WriteString(fmt.Sprintf("%s%d\n", prefix, int64(val)))
		} else {
			sb.WriteString(fmt.Sprintf("%s%g\n", prefix, val))
		}
	case bool:
		sb.WriteString(fmt.Sprintf("%s%v\n", prefix, val))
	default:
		sb.WriteString(fmt.Sprintf("%s%v\n", prefix, val))
	}

	return sb.String()
}

// arrayToYAML 将数组转为YAML格式
func arrayToYAML(arr []interface{}, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	for _, item := range arr {
		switch val := item.(type) {
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("%s-\n", prefix))
			sb.WriteString(valueToYAML(val, indent+1))
		case string:
			if strings.ContainsAny(val, ":{}[]&*?|>-!%@`\"'\n") || val == "" {
				sb.WriteString(fmt.Sprintf("%s- \"%s\"\n", prefix, escapeYAMLString(val)))
			} else {
				sb.WriteString(fmt.Sprintf("%s- %s\n", prefix, val))
			}
		case float64:
			if val == float64(int64(val)) {
				sb.WriteString(fmt.Sprintf("%s- %d\n", prefix, int64(val)))
			} else {
				sb.WriteString(fmt.Sprintf("%s- %g\n", prefix, val))
			}
		case bool:
			sb.WriteString(fmt.Sprintf("%s- %v\n", prefix, val))
		case nil:
			sb.WriteString(fmt.Sprintf("%s- null\n", prefix))
		default:
			sb.WriteString(fmt.Sprintf("%s- %v\n", prefix, val))
		}
	}

	return sb.String()
}

// writeToFile 写入数据到文件
func writeToFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0644)
}

// escapeYAMLString 转义YAML字符串中的特殊字符
func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// WriteToWriter 将HAR写入指定的Writer
// （已定义在builder.go中，这里提供ConvertFormat参数版本）

// ConvertTo 将HAR转换为指定格式并写入Writer
func (h *Har) ConvertTo(format ConvertFormat, w io.Writer, options ConvertOptions) error {
	var content string
	var err error

	switch format {
	case FormatYAML:
		content, err = h.ToYAML()
	case FormatCSV, FormatMarkdown, FormatHTML, FormatText:
		content, err = h.Convert(format, options)
	default:
		// 默认输出JSON
		var data []byte
		data, err = h.ToJSON(true)
		if err == nil {
			_, err = w.Write(data)
			return err
		}
	}

	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, content)
	return err
}
