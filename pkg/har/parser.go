package har

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// ParseHarWithOptions 解析HAR格式的字节数据，使用自定义解析选项
func ParseHarWithOptions(harFileBytes []byte, options ParseOptions) (*Har, error) {
	if len(harFileBytes) == 0 {
		return nil, NewInvalidFormatError("输入为空")
	}

	// 检查文件是否是JSON格式
	if !isJSONContent(harFileBytes) {
		return nil, NewInvalidFormatError("输入不是有效的JSON格式")
	}

	// 如果是严格模式，直接解析
	if !options.Lenient {
		har := new(Har)
		err := json.Unmarshal(harFileBytes, har)
		if err != nil {
			return nil, WrapJSONUnmarshalError(err)
		}

		// 如果需要验证
		if !options.SkipValidation {
			if err := validateHar(har); err != nil {
				return nil, err
			}
		}

		return har, nil
	}

	// 宽松模式（Lenient）：尝试解析尽可能多的内容
	return parseLenient(harFileBytes, options)
}

// ParseHarFileWithOptions 解析HAR格式的文件，使用自定义解析选项
func ParseHarFileWithOptions(harFilePath string, options ParseOptions) (*Har, error) {
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("无法读取文件 '%s'", harFilePath), err)
	}

	har, err := ParseHarWithOptions(harFileBytes, options)
	if err != nil {
		harErr, ok := err.(*HarError)
		if ok {
			_ = harErr.WithMetadata("filePath", harFilePath)
		}
		return nil, err
	}

	return har, nil
}

// ParseHarEnhanced 增强版HAR解析，提供详细错误信息
func ParseHarEnhanced(harFileBytes []byte) (*Har, *HarError) {
	har, err := ParseHarWithOptions(harFileBytes, DefaultParseOptions())
	if err != nil {
		if harErr, ok := err.(*HarError); ok {
			return nil, harErr
		}
		// 将普通错误包装为自定义错误
		return nil, NewHarError(ErrCodeUnknown, err.Error(), err)
	}
	return har, nil
}

// ParseHarFileEnhanced 增强版HAR文件解析，提供详细错误信息
func ParseHarFileEnhanced(harFilePath string) (*Har, *HarError) {
	har, err := ParseHarFileWithOptions(harFilePath, DefaultParseOptions())
	if err != nil {
		if harErr, ok := err.(*HarError); ok {
			return nil, harErr
		}
		// 将普通错误包装为自定义错误
		return nil, NewHarError(ErrCodeUnknown, err.Error(), err)
	}
	return har, nil
}

// ParseHarLenient 宽松模式解析HAR文件内容
func ParseHarLenient(harFileBytes []byte) (*Har, error) {
	options := DefaultParseOptions()
	options.Lenient = true
	options.CollectWarnings = true
	return ParseHarWithOptions(harFileBytes, options)
}

// ParseHarFileLenient 宽松模式解析HAR文件
func ParseHarFileLenient(harFilePath string) (*Har, error) {
	options := DefaultParseOptions()
	options.Lenient = true
	options.CollectWarnings = true
	return ParseHarFileWithOptions(harFilePath, options)
}

// isJSONContent 检查内容是否是JSON格式
func isJSONContent(content []byte) bool {
	trimmed := strings.TrimSpace(string(content))
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

// validateHar 验证HAR对象内容的有效性
// 该函数现在转发到validator.go中实现的ValidateHarFile
func validateHar(har *Har) error {
	if har == nil {
		return NewInvalidFormatError("HAR对象为空")
	}

	return ValidateHarFile(har)
}

// parseLenient 宽松模式解析，尝试解析尽可能多的内容
func parseLenient(harFileBytes []byte, options ParseOptions) (*Har, error) {
	// 创建一个空的HAR对象
	har := &Har{
		Log: Log{
			Entries: []Entries{},
			Pages:   []Pages{},
		},
	}

	// 使用map来进行初步解析，这样即使部分字段无效也能解析其他部分
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(harFileBytes, &rawData); err != nil {
		return nil, WrapJSONUnmarshalError(err)
	}

	// 跟踪所有错误
	rootError := &HarError{
		Code:    ErrCodeJSONParse,
		Message: "HAR解析过程中发生错误，但部分内容已成功解析",
	}

	// 解析log字段
	if logBytes, ok := rawData["log"]; ok {
		var logData map[string]json.RawMessage
		if err := json.Unmarshal(logBytes, &logData); err != nil {
			_ = rootError.AddPartialError(
				NewJSONParseError("无法解析log字段", err).WithField("log"))
		} else {
			// 解析version字段
			if versionBytes, ok := logData["version"]; ok {
				var version string
				if err := json.Unmarshal(versionBytes, &version); err == nil {
					har.Log.Version = version
				} else {
					_ = rootError.AddPartialError(
						NewJSONParseError("无法解析version字段", err).WithField("log.version"))
				}
			}

			// 解析creator字段
			if creatorBytes, ok := logData["creator"]; ok {
				var creator Creator
				if err := json.Unmarshal(creatorBytes, &creator); err == nil {
					har.Log.Creator = creator
				} else {
					_ = rootError.AddPartialError(
						NewJSONParseError("无法解析creator字段", err).WithField("log.creator"))
				}
			}

			// 解析pages字段
			if pagesBytes, ok := logData["pages"]; ok {
				var pages []json.RawMessage
				if err := json.Unmarshal(pagesBytes, &pages); err == nil {
					for i, pageBytes := range pages {
						var page Pages
						if err := json.Unmarshal(pageBytes, &page); err == nil {
							har.Log.Pages = append(har.Log.Pages, page)
						} else {
							_ = rootError.AddPartialError(
								NewJSONParseError(
									fmt.Sprintf("无法解析第%d个page", i+1), err).
									WithField(fmt.Sprintf("log.pages[%d]", i)))
						}
					}
				} else {
					_ = rootError.AddPartialError(
						NewJSONParseError("无法解析pages字段", err).WithField("log.pages"))
				}
			}

			// 解析entries字段，这是最重要的部分
			if entriesBytes, ok := logData["entries"]; ok {
				var entries []json.RawMessage
				if err := json.Unmarshal(entriesBytes, &entries); err == nil {
					for i, entryBytes := range entries {
						var entry Entries
						if err := json.Unmarshal(entryBytes, &entry); err == nil {
							har.Log.Entries = append(har.Log.Entries, entry)
						} else {
							_ = rootError.AddPartialError(
								NewJSONParseError(
									fmt.Sprintf("无法解析第%d个entry", i+1), err).
									WithField(fmt.Sprintf("log.entries[%d]", i)))
						}
					}
				} else {
					_ = rootError.AddPartialError(
						NewJSONParseError("无法解析entries字段", err).WithField("log.entries"))
				}
			}
		}
	} else {
		rootError.AddPartialError(NewMissingFieldError("log"))
	}

	// 如果有错误，并且选项指定收集警告
	if rootError.HasPartialErrors() && options.CollectWarnings {
		// 如果解析了部分内容，返回HAR对象和错误
		if har.Log.Version != "" || len(har.Log.Entries) > 0 || len(har.Log.Pages) > 0 {
			return har, rootError
		}
		// 否则认为解析完全失败
		return nil, rootError
	} else if rootError.HasPartialErrors() {
		// 如果有错误但不收集警告，只返回错误
		return nil, rootError
	}

	return har, nil
}

// Result 解析结果，包含HAR对象和可能的警告
type Result struct {
	Har      *Har
	Warnings []*HarError
}

// ParseHarWithWarnings 解析HAR文件同时返回警告信息
// 该函数使用宽松模式解析，并收集所有警告而不是直接失败
func ParseHarWithWarnings(harFileBytes []byte) (*Result, error) {
	// 使用宽松模式和警告收集
	options := DefaultParseOptions()
	options.Lenient = true
	options.CollectWarnings = true

	// 解析HAR数据
	har, err := ParseHarWithOptions(harFileBytes, options)

	// 初始化结果对象
	result := &Result{
		Har:      har,
		Warnings: []*HarError{},
	}

	// 处理解析阶段的警告
	if err != nil {
		if harErr, ok := err.(*HarError); ok && har != nil {
			// 在宽松模式下，将解析错误转换为警告
			result.Warnings = appendWarnings(result.Warnings, harErr.GetPartialErrors())
		} else {
			// 解析完全失败的情况
			return nil, err
		}
	}

	// 执行URL验证
	urlWarnings := validateURLs(har)
	if len(urlWarnings) > 0 {
		result.Warnings = appendWarnings(result.Warnings, urlWarnings)
	}

	// 如果仍未找到警告，尝试运行完整验证
	if len(result.Warnings) == 0 {
		validationWarnings := performFullValidation(har)
		result.Warnings = appendWarnings(result.Warnings, validationWarnings)
	}

	return result, nil
}

// validateURLs 验证所有条目中的URL字段
func validateURLs(har *Har) []*HarError {
	if har == nil || len(har.Log.Entries) == 0 {
		return nil
	}

	var warnings []*HarError
	for i, entry := range har.Log.Entries {
		if entry.Request.URL == "" {
			continue
		}

		// 严格URL验证
		if _, err := url.Parse(entry.Request.URL); err != nil {
			urlError := NewValidationError(
				fmt.Sprintf("无效的URL格式: %s", err.Error()),
				fmt.Sprintf("log.entries[%d].request.url", i),
			)
			warnings = append(warnings, urlError)
			continue
		}

		// 额外检查常见URL问题
		if strings.Contains(entry.Request.URL, " ") {
			urlError := NewValidationError(
				fmt.Sprintf("URL包含空格: %s", entry.Request.URL),
				fmt.Sprintf("log.entries[%d].request.url", i),
			)
			warnings = append(warnings, urlError)
		}

		if !strings.Contains(entry.Request.URL, "://") {
			urlError := NewValidationError(
				fmt.Sprintf("URL缺少协议: %s", entry.Request.URL),
				fmt.Sprintf("log.entries[%d].request.url", i),
			)
			warnings = append(warnings, urlError)
		}
	}

	return warnings
}

// performFullValidation 执行完整的HAR验证，并将错误转换为警告
func performFullValidation(har *Har) []*HarError {
	if har == nil {
		return nil
	}

	validationErr := ValidateHarFile(har)
	if validationErr == nil {
		return nil
	}

	if harErr, ok := validationErr.(*HarError); ok {
		return harErr.GetPartialErrors()
	}

	// 如果是其他类型错误，包装为HarError
	return []*HarError{
		NewValidationError(validationErr.Error(), ""),
	}
}

// appendWarnings 将新警告追加到现有警告列表，避免重复
func appendWarnings(existing []*HarError, newWarnings []*HarError) []*HarError {
	if len(newWarnings) == 0 {
		return existing
	}

	if existing == nil {
		return newWarnings
	}

	// 使用映射检测重复
	warningMap := make(map[string]bool)
	for _, warn := range existing {
		key := warn.Field + ":" + warn.Message
		warningMap[key] = true
	}

	// 添加非重复的警告
	for _, warn := range newWarnings {
		key := warn.Field + ":" + warn.Message
		if !warningMap[key] {
			existing = append(existing, warn)
			warningMap[key] = true
		}
	}

	return existing
}

// ParseHarFileWithWarnings 解析HAR文件同时返回警告信息
func ParseHarFileWithWarnings(harFilePath string) (*Result, error) {
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("无法读取文件 '%s'", harFilePath), err)
	}

	return ParseHarWithWarnings(harFileBytes)
}
