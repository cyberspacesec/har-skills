package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cyberspacesec/har-skills"
)

// 演示使用新的函数选项模式API
func demonstrateNewApi() {
	exampleHarPath := "../../data/example.har"
	if _, err := os.Stat(exampleHarPath); os.IsNotExist(err) {
		log.Println("示例HAR文件不存在，跳过新API演示")
		return
	}

	fmt.Println("\n=== 新函数选项模式API示例 ===")

	// 使用标准解析
	harData, err := har.ParseFile(exampleHarPath)
	if err != nil {
		log.Printf("标准解析失败: %v", err)
	} else {
		fmt.Printf("标准解析: 加载了 %d 个entries\n", len(harData.GetEntries()))
	}

	// 使用内存优化解析
	harData, err = har.ParseFile(exampleHarPath, har.WithMemoryOptimized())
	if err != nil {
		log.Printf("内存优化解析失败: %v", err)
	} else {
		fmt.Printf("内存优化解析: 加载了 %d 个entries\n", len(harData.GetEntries()))
	}

	// 使用懒加载解析
	harData, err = har.ParseFile(exampleHarPath, har.WithLazyLoading())
	if err != nil {
		log.Printf("懒加载解析失败: %v", err)
	} else {
		fmt.Printf("懒加载解析: 加载了 %d 个entries\n", len(harData.GetEntries()))
	}

	// 使用组合选项
	harData, err = har.ParseFile(exampleHarPath, har.WithMemoryOptimized(), har.WithSkipValidation(), har.WithLenient())
	if err != nil {
		log.Printf("组合选项解析失败: %v", err)
	} else {
		fmt.Printf("组合选项解析: 加载了 %d 个entries\n", len(harData.GetEntries()))
	}

	// 使用预定义选项组
	harData, err = har.ParseFile(exampleHarPath, har.OptMemoryEfficient...)
	if err != nil {
		log.Printf("预定义选项组解析失败: %v", err)
	} else {
		fmt.Printf("预定义选项组解析: 加载了 %d 个entries\n", len(harData.GetEntries()))
	}

	// 使用流式解析
	iterator, err := har.NewStreamingParserFromFile(exampleHarPath)
	if err != nil {
		log.Printf("流式解析器创建失败: %v", err)
	} else {
		count := 0
		for iterator.Next() {
			count++
		}
		if err := iterator.Err(); err != nil {
			log.Printf("流式解析过程中发生错误: %v", err)
		} else {
			fmt.Printf("流式解析: 处理了 %d 个entries\n", count)
		}
	}

	// 使用接口进行通用处理
	harData, err = har.ParseFile(exampleHarPath)
	if err != nil {
		log.Printf("接口解析失败: %v", err)
	} else {
		fmt.Println("\n使用接口处理HAR数据:")
		processAnyHar(harData)
	}
}

// 通用处理函数，可以处理任何实现了HARProvider接口的类型
func processAnyHar(har har.HARProvider) {
	fmt.Printf("HAR 版本: %s\n", har.GetVersion())
	fmt.Printf("创建者: %s %s\n", har.GetCreator().Name, har.GetCreator().Version)
	fmt.Printf("条目数量: %d\n", len(har.GetEntries()))

	// 处理所有条目
	if len(har.GetEntries()) > 0 {
		fmt.Println("\n第一个条目信息:")
		entry := har.GetEntries()[0]
		request := entry.GetRequest()
		response := entry.GetResponse()

		fmt.Printf("  请求: %s %s\n", request.GetMethod(), request.GetURL())
		fmt.Printf("  响应: %d %s\n", response.GetStatus(), response.GetStatusText())
		fmt.Printf("  内容类型: %s\n", response.GetContent().GetMimeType())
		fmt.Printf("  内容大小: %d 字节\n", response.GetContent().GetSize())
	}
}
