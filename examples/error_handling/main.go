package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cyberspacesec/har-skills"
)

func main() {
	fmt.Println("========= HAR 错误处理增强示例 =========")

	// 示例1: 使用增强的解析API，获取详细错误信息
	fmt.Println("\n【示例1】详细错误信息")
	harFilePath := "../../data/www.google.com.har"
	harFile, harErr := har.ParseHarFileEnhanced(harFilePath)
	if harErr != nil {
		fmt.Println("解析失败，详细错误信息:")
		printHarError(harErr, 0)
	} else {
		fmt.Printf("成功解析HAR文件，包含 %d 个请求条目\n", len(harFile.Log.Entries))
	}

	// 示例2: 处理不存在的文件 - 文件系统错误
	fmt.Println("\n【示例2】处理文件系统错误")
	nonExistentFile := "non-existent.har"
	_, harErr = har.ParseHarFileEnhanced(nonExistentFile)
	if harErr != nil {
		fmt.Printf("错误类型: %v\n", harErr.GetCode())
		fmt.Printf("是否为文件系统错误: %v\n", harErr.IsFileSystemError())
		fmt.Printf("错误信息: %s\n", harErr.Error())

		if harErr.Metadata != nil {
			fmt.Println("元数据:")
			for k, v := range harErr.Metadata {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}

	// 示例3: 创建包含无效字段的JSON，测试部分解析
	fmt.Println("\n【示例3】部分解析 - 处理无效字段")
	invalidJSON := createInvalidJSON()
	tempFile := "temp_invalid.har"

	// 写入临时文件
	err := os.WriteFile(tempFile, []byte(invalidJSON), 0644)
	if err != nil {
		fmt.Println("创建临时文件失败:", err)
		return
	}
	defer os.Remove(tempFile) // 清理临时文件

	// 使用宽松模式解析
	fmt.Println("使用宽松解析模式:")
	harFile, err = har.ParseHarFileLenient(tempFile)
	if err != nil {
		if harErr, ok := err.(*har.HarError); ok {
			fmt.Println("解析过程中发生以下警告，但仍然返回了部分解析结果:")
			printHarError(harErr, 0)

			// 输出成功解析的部分
			if harFile != nil {
				fmt.Println("\n成功解析的部分:")
				fmt.Printf("  版本: %s\n", harFile.Log.Version)
				fmt.Printf("  Creator: %s %s\n", harFile.Log.Creator.Name, harFile.Log.Creator.Version)
				fmt.Printf("  页面数: %d\n", len(harFile.Log.Pages))
				fmt.Printf("  条目数: %d\n", len(harFile.Log.Entries))
			}
		} else {
			fmt.Println("解析失败:", err)
		}
	} else {
		fmt.Println("完全成功解析，没有警告")
	}

	// 示例4: 使用带警告的解析
	fmt.Println("\n【示例4】收集警告信息")
	result, err := har.ParseHarFileWithWarnings(tempFile)
	if err != nil {
		fmt.Println("解析完全失败:", err)
	} else {
		fmt.Printf("解析成功，收集到 %d 个警告\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Printf("警告 %d: %s\n", i+1, warning.Error())
		}

		// 输出成功解析的部分
		fmt.Println("\n成功解析的部分:")
		fmt.Printf("  版本: %s\n", result.Har.Log.Version)
		fmt.Printf("  Creator: %s %s\n", result.Har.Log.Creator.Name, result.Har.Log.Creator.Version)
		fmt.Printf("  页面数: %d\n", len(result.Har.Log.Pages))
		fmt.Printf("  条目数: %d\n", len(result.Har.Log.Entries))
	}
}

// 打印HAR错误信息，包括嵌套的部分错误
func printHarError(harErr *har.HarError, level int) {
	prefix := strings.Repeat("  ", level)
	fmt.Printf("%s错误: %s\n", prefix, harErr.Message)

	if harErr.Field != "" {
		fmt.Printf("%s字段: %s\n", prefix, harErr.Field)
	}

	if harErr.Metadata != nil && len(harErr.Metadata) > 0 {
		fmt.Printf("%s元数据:\n", prefix)
		for k, v := range harErr.Metadata {
			fmt.Printf("%s  %s: %v\n", prefix, k, v)
		}
	}

	if harErr.HasPartialErrors() {
		fmt.Printf("%s包含 %d 个部分错误:\n", prefix, len(harErr.GetPartialErrors()))
		for i, pe := range harErr.GetPartialErrors() {
			fmt.Printf("%s部分错误 %d:\n", prefix, i+1)
			printHarError(pe, level+1)
		}
	}
}

// 创建一个包含无效字段的JSON
func createInvalidJSON() string {
	return `{
		"log": {
			"version": "1.2",
			"creator": {
				"name": "测试用例",
				"version": "1.0"
			},
			"pages": [
				{
					"startedDateTime": "invalid-date",
					"id": "page_1",
					"title": "测试页面",
					"pageTimings": {
						"onContentLoad": "非数字",
						"onLoad": 500
					}
				}
			],
			"entries": [
				{
					"startedDateTime": "2023-01-01T00:00:00.000Z",
					"time": 100,
					"request": {
						"method": "GET",
						"url": "https://example.com",
						"httpVersion": "HTTP/1.1",
						"headers": [
							{
								"name": "Accept",
								"value": "text/html"
							}
						]
					},
					"response": {
						"status": "非数字",
						"statusText": "OK",
						"content": {
							"size": 1024,
							"mimeType": "text/html"
						}
					}
				},
				{
					"无效字段": "这会导致解析错误",
					"request": {
						"method": "POST",
						"url": "https://example.com/api"
					}
				}
			]
		}
	}`
}
