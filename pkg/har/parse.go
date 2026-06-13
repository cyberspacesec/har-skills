package har

import (
	"fmt"
	"os"
)

// Parse 使用函数选项模式解析HAR字节数据
//
// Parse函数是解析HAR数据的主要入口点，支持多种解析策略和选项。
// 该函数使用函数选项模式，允许灵活配置解析行为。
//
// 示例:
//
//	// 标准解析
//	har, err := Parse(harBytes)
//
//	// 使用内存优化
//	har, err := Parse(harBytes, WithMemoryOptimized())
//
//	// 组合多个选项
//	har, err := Parse(harBytes, WithMemoryOptimized(), WithSkipValidation())
//
// 返回实现了HARProvider接口的对象，可以统一访问不同实现的HAR结构。
func Parse(harFileBytes []byte, opts ...Option) (HARProvider, error) {
	// 应用选项
	options := applyOptions(opts...)

	// 验证输入
	if err := validateInput(harFileBytes); err != nil {
		return nil, err
	}

	// 根据选项选择相应的解析方法
	return parseWithStrategy(harFileBytes, options)
}

// validateInput 验证输入数据是否有效
func validateInput(harFileBytes []byte) error {
	// 检查输入是否为空
	if len(harFileBytes) == 0 {
		return NewInvalidFormatError("输入为空")
	}

	// 检查文件是否是JSON格式
	if !isJSONContent(harFileBytes) {
		return NewInvalidFormatError("输入不是有效的JSON格式")
	}

	return nil
}

// parseWithStrategy 根据选项选择合适的解析策略
func parseWithStrategy(harFileBytes []byte, options options) (HARProvider, error) {
	// 流式解析需要特殊处理
	if options.useStreaming {
		return nil, NewUnsupportedError("流式解析不支持直接返回完整HAR对象，请使用NewStreamingParser")
	}

	// 根据选项选择解析策略
	if options.useMemoryOptimized {
		// 内存优化解析
		return ParseHarOptimized(harFileBytes)
	} else if options.useLazyLoading {
		// 懒加载解析
		return ParseHarWithLazyLoading(harFileBytes)
	} else {
		// 标准解析
		parseOptions := options.toParseOptions()
		return ParseHarWithOptions(harFileBytes, parseOptions)
	}
}

// ParseFile 使用函数选项模式解析HAR文件
//
// ParseFile是解析HAR文件的便捷方法，支持与Parse函数相同的选项。
// 该函数负责文件读取，然后将内容传递给Parse函数进行解析。
//
// 示例:
//
//	// 标准解析
//	har, err := ParseFile("example.har")
//
//	// 使用预定义选项组合
//	har, err := ParseFile("large.har", OptMemoryEfficient...)
func ParseFile(harFilePath string, opts ...Option) (HARProvider, error) {
	// 读取文件
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("无法读取文件 '%s'", harFilePath), err)
	}

	// 解析文件内容
	har, err := Parse(harFileBytes, opts...)
	if err != nil {
		// 添加文件路径到错误上下文
		if harErr, ok := err.(*HarError); ok {
			_ = harErr.WithMetadata("filePath", harFilePath)
		}
		return nil, err
	}

	return har, nil
}

// NewStreamingParser 创建一个新的流式解析器
//
// 流式解析器允许逐个处理HAR条目，适用于大型HAR文件，避免一次性加载全部内容。
//
// 示例:
//
//	iterator, err := NewStreamingParser(harBytes)
//	if err != nil {
//	    return err
//	}
//	for iterator.Next() {
//	    entry := iterator.Entry()
//	    // 处理单个条目
//	}
func NewStreamingParser(harFileBytes []byte, opts ...Option) (EntryIterator, error) {
	// 验证输入
	if err := validateInput(harFileBytes); err != nil {
		return nil, err
	}

	// 创建流式解析器
	streamingHar, err := NewStreamingHarFromBytes(harFileBytes)
	if err != nil {
		return nil, err
	}
	return streamingHar.Entries(), nil
}

// NewStreamingParserFromFile 从文件创建一个新的流式解析器
//
// 这是一个便捷方法，用于从文件路径创建流式解析器，避免手动读取文件。
func NewStreamingParserFromFile(harFilePath string, opts ...Option) (EntryIterator, error) {
	harFileBytes, err := os.ReadFile(harFilePath)
	if err != nil {
		return nil, NewFileSystemError(fmt.Sprintf("无法读取文件 '%s'", harFilePath), err)
	}

	return NewStreamingParser(harFileBytes, opts...)
}
