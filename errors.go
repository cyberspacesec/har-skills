package har

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ErrorCode 定义错误类型代码
type ErrorCode int

const (
	// ErrCodeUnknown 未知错误
	ErrCodeUnknown ErrorCode = iota
	// ErrCodeFileSystem 文件系统错误
	ErrCodeFileSystem
	// ErrCodeJSONParse JSON解析错误
	ErrCodeJSONParse
	// ErrCodeInvalidFormat 格式错误
	ErrCodeInvalidFormat
	// ErrCodeValidation 验证错误
	ErrCodeValidation
	// ErrCodeMissingField 缺少必要字段
	ErrCodeMissingField
	// ErrCodeInvalidValue 字段值无效
	ErrCodeInvalidValue
	// ErrCodeUnsupported 不支持的操作
	ErrCodeUnsupported
)

// HarError 自定义HAR错误类型
type HarError struct {
	// 错误代码
	Code ErrorCode
	// 错误信息
	Message string
	// 原始错误（可选）
	Err error
	// 字段路径，用点号分隔，如 "log.entries[0].request.url"
	Field string
	// 包含更多上下文信息的元数据
	Metadata map[string]interface{}
	// 如果是部分解析错误，包含的其他错误
	PartialErrors []*HarError
}

// 实现error接口
func (e *HarError) Error() string {
	if e == nil {
		return "<nil>"
	}
	msg := e.Message
	if e.Field != "" {
		msg = fmt.Sprintf("字段 '%s': %s", e.Field, msg)
	}

	if e.Err != nil {
		msg = fmt.Sprintf("%s - %v", msg, e.Err)
	}

	if len(e.PartialErrors) > 0 {
		partialMsgs := make([]string, 0, len(e.PartialErrors))
		for _, pe := range e.PartialErrors {
			if pe == nil {
				continue
			}
			partialMsgs = append(partialMsgs, pe.Error())
		}
		if len(partialMsgs) > 0 {
			msg = fmt.Sprintf("%s (部分错误: %s)", msg, strings.Join(partialMsgs, "; "))
		}
	}

	return msg
}

// Unwrap returns the underlying error for errors.Is/errors.As support.
func (e *HarError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// WithField 添加字段路径到错误
func (e *HarError) WithField(field string) *HarError {
	if e == nil {
		return nil
	}
	if e.Field == "" {
		e.Field = field
	} else {
		e.Field = field + "." + e.Field
	}
	return e
}

// WithMetadata 添加元数据到错误
func (e *HarError) WithMetadata(key string, value interface{}) *HarError {
	if e == nil {
		return nil
	}
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// AddPartialError 添加部分解析错误
func (e *HarError) AddPartialError(err *HarError) *HarError {
	if e == nil {
		return nil
	}
	if err == nil {
		return e
	}
	e.PartialErrors = append(e.PartialErrors, err)
	return e
}

// HasPartialErrors 检查是否包含部分错误
func (e *HarError) HasPartialErrors() bool {
	if e == nil {
		return false
	}
	for _, pe := range e.PartialErrors {
		if pe != nil {
			return true
		}
	}
	return false
}

// GetPartialErrors 获取所有部分错误
func (e *HarError) GetPartialErrors() []*HarError {
	if e == nil {
		return nil
	}
	return e.PartialErrors
}

// GetCode 获取错误代码
func (e *HarError) GetCode() ErrorCode {
	if e == nil {
		return ErrCodeUnknown
	}
	return e.Code
}

// IsFileSystemError 是否为文件系统错误
func (e *HarError) IsFileSystemError() bool {
	if e == nil {
		return false
	}
	return e.Code == ErrCodeFileSystem
}

// IsJSONParseError 是否为JSON解析错误
func (e *HarError) IsJSONParseError() bool {
	if e == nil {
		return false
	}
	return e.Code == ErrCodeJSONParse
}

// IsFormatError 是否为格式错误
func (e *HarError) IsFormatError() bool {
	if e == nil {
		return false
	}
	return e.Code == ErrCodeInvalidFormat
}

// IsValidationError 是否为验证错误
func (e *HarError) IsValidationError() bool {
	if e == nil {
		return false
	}
	return e.Code == ErrCodeValidation
}

// NewHarError 创建新的HAR错误
func NewHarError(code ErrorCode, message string, err error) *HarError {
	return &HarError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewFileSystemError 创建文件系统错误
func NewFileSystemError(message string, err error) *HarError {
	return NewHarError(ErrCodeFileSystem, message, err)
}

// NewJSONParseError 创建JSON解析错误
func NewJSONParseError(message string, err error) *HarError {
	return NewHarError(ErrCodeJSONParse, message, err)
}

// WrapJSONUnmarshalError 封装JSON解析错误以提供更详细信息
func WrapJSONUnmarshalError(err error) *HarError {
	if err == nil {
		return nil
	}

	// 尝试从JSON错误中提取详细信息
	var jsonErr *json.UnmarshalTypeError
	var syntaxErr *json.SyntaxError

	if e, ok := err.(*json.UnmarshalTypeError); ok {
		jsonErr = e
		return NewJSONParseError(
			fmt.Sprintf("类型不匹配: 预期 %s 类型，但得到 %s",
				jsonErr.Type.String(), jsonErr.Value),
			err,
		).WithField(jsonErr.Field).WithMetadata("offset", jsonErr.Offset)
	} else if e, ok := err.(*json.SyntaxError); ok {
		syntaxErr = e
		return NewJSONParseError(
			fmt.Sprintf("JSON语法错误: %s", syntaxErr.Error()),
			err,
		).WithMetadata("offset", syntaxErr.Offset)
	} else if strings.Contains(err.Error(), "cannot unmarshal") {
		// 处理其他无法精确识别类型的JSON解析错误
		parts := strings.Split(err.Error(), ":")
		if len(parts) >= 2 {
			return NewJSONParseError(
				fmt.Sprintf("JSON解析错误: %s", strings.TrimSpace(parts[1])),
				err,
			)
		}
	}

	// 默认JSON错误处理
	return NewJSONParseError("JSON解析错误", err)
}

// NewValidationError 创建验证错误
func NewValidationError(message string, field string) *HarError {
	return NewHarError(ErrCodeValidation, message, nil).WithField(field)
}

// NewInvalidFormatError 创建格式错误
func NewInvalidFormatError(message string) *HarError {
	return NewHarError(ErrCodeInvalidFormat, message, nil)
}

// NewMissingFieldError 创建缺少字段错误
func NewMissingFieldError(field string) *HarError {
	return NewHarError(ErrCodeMissingField, "必需字段缺失", nil).WithField(field)
}

// NewInvalidValueError 创建字段值无效错误
func NewInvalidValueError(field string, value interface{}, reason string) *HarError {
	msg := "字段值无效"
	if reason != "" {
		msg = msg + ": " + reason
	}
	return NewHarError(ErrCodeInvalidValue, msg, nil).
		WithField(field).
		WithMetadata("value", value)
}

// NewUnsupportedError 创建不支持的操作错误
func NewUnsupportedError(message string) *HarError {
	return NewHarError(ErrCodeUnsupported, message, nil)
}

// ParseOptions 解析选项
type ParseOptions struct {
	// 是否开启宽松解析模式，会尽量解析有效部分
	Lenient bool
	// 是否跳过验证
	SkipValidation bool
	// 是否记录所有解析警告（不会导致解析失败）
	CollectWarnings bool
	// 最大允许的警告数量，超过则停止解析
	MaxWarnings int
}

// DefaultParseOptions 默认解析选项
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		Lenient:         false,
		SkipValidation:  false,
		CollectWarnings: false,
		MaxWarnings:     100,
	}
}
