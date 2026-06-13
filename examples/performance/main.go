package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/cyberspacesec/har-skills"
)

func main() {
	// 定义命令行参数
	filePathPtr := flag.String("file", "", "HAR文件路径")
	modePtr := flag.String("mode", "all", "性能测试模式 (standard, optimized, lazy, streaming, all)")
	flag.Parse()

	// 如果没有指定文件路径，检查是否存在默认示例文件
	filePath := *filePathPtr
	if filePath == "" {
		defaultPath := "../../data/example.har"
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			fmt.Println("错误: 请提供HAR文件路径或放置示例文件在 data/example.har")
			flag.Usage()
			return
		}
		filePath = defaultPath
		fmt.Printf("使用默认HAR文件: %s\n", filePath)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("无法获取文件信息: %v", err)
	}
	fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
	fmt.Printf("文件大小: %.2f MB\n\n", fileSizeMB)

	mode := *modePtr
	switch mode {
	case "standard":
		benchmarkStandard(filePath)
	case "optimized":
		benchmarkOptimized(filePath)
	case "lazy":
		benchmarkLazy(filePath)
	case "streaming":
		benchmarkStreaming(filePath)
	case "all":
		benchmarkAll(filePath)
	default:
		fmt.Printf("未知模式: %s\n", mode)
		flag.Usage()
	}

	// 演示新API的使用
	demonstrateNewApi()
}

// benchmarkAll 运行所有性能测试
func benchmarkAll(filePath string) {
	fmt.Println("=== 性能比较 ===")
	fmt.Println("测试所有解析方法的内存使用和性能...")

	benchmarkStandard(filePath)
	benchmarkOptimized(filePath)
	benchmarkLazy(filePath)
	benchmarkStreaming(filePath)
}

// benchmarkStandard 测试标准解析性能
func benchmarkStandard(filePath string) {
	fmt.Println("\n=== 标准解析 ===")
	start := time.Now()
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	// 标准模式
	har, err := har.ParseFile(filePath)
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}

	// 获取内存使用
	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc
	memUsed := float64(memAfter-memBefore) / (1024 * 1024)
	elapsed := time.Since(start)

	// 计算和显示结果
	entriesCount := len(har.GetEntries())
	fmt.Printf("解析耗时: %v\n", elapsed)
	fmt.Printf("内存用量: %.2f MB\n", memUsed)
	fmt.Printf("条目数量: %d\n", entriesCount)
}

// benchmarkOptimized 测试内存优化解析性能
func benchmarkOptimized(filePath string) {
	fmt.Println("\n=== 内存优化解析 ===")
	start := time.Now()
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	// 内存优化模式
	har, err := har.ParseFile(filePath, har.WithMemoryOptimized())
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}

	// 获取内存使用
	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc
	memUsed := float64(memAfter-memBefore) / (1024 * 1024)
	elapsed := time.Since(start)

	// 计算和显示结果
	entriesCount := len(har.GetEntries())
	fmt.Printf("解析耗时: %v\n", elapsed)
	fmt.Printf("内存用量: %.2f MB\n", memUsed)
	fmt.Printf("条目数量: %d\n", entriesCount)
}

// benchmarkLazy 测试懒加载解析性能
func benchmarkLazy(filePath string) {
	fmt.Println("\n=== 懒加载解析 ===")
	start := time.Now()
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	// 懒加载模式
	har, err := har.ParseFile(filePath, har.WithLazyLoading())
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}

	// 获取内存使用（仅初始化内存）
	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc
	memUsed := float64(memAfter-memBefore) / (1024 * 1024)
	elapsed := time.Since(start)

	// 计算和显示结果
	entriesCount := len(har.GetEntries())
	fmt.Printf("解析耗时: %v\n", elapsed)
	fmt.Printf("初始内存用量: %.2f MB\n", memUsed)
	fmt.Printf("条目数量: %d\n", entriesCount)

	// 访问响应内容，触发完整加载
	if entriesCount > 0 {
		fmt.Println("\n访问第一个条目的内容，触发懒加载...")
		runtime.GC()
		runtime.ReadMemStats(&memStats)
		memBefore = memStats.Alloc
		startAccess := time.Now()

		// 获取第一个条目的内容
		entry := har.GetEntries()[0]
		response := entry.GetResponse()
		content := response.GetContent()
		fmt.Printf("内容类型: %s, 大小: %d bytes\n", content.GetMimeType(), content.GetSize())

		// 计算访问耗时和额外内存
		runtime.ReadMemStats(&memStats)
		memAfter = memStats.Alloc
		additionalMem := float64(memAfter-memBefore) / (1024 * 1024)
		accessTime := time.Since(startAccess)
		fmt.Printf("访问耗时: %v\n", accessTime)
		fmt.Printf("额外内存用量: %.2f MB\n", additionalMem)
	}
}

// benchmarkStreaming 测试流式解析性能
func benchmarkStreaming(filePath string) {
	fmt.Println("\n=== 流式解析 ===")
	start := time.Now()
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	memBefore := memStats.Alloc

	// 流式模式
	iterator, err := har.NewStreamingParserFromFile(filePath)
	if err != nil {
		log.Fatalf("创建流式解析器失败: %v", err)
	}

	// 处理所有条目
	count := 0
	for iterator.Next() {
		count++
		// 只获取但不做任何处理
		_ = iterator.Entry()
	}

	if err := iterator.Err(); err != nil {
		log.Fatalf("流式解析过程中发生错误: %v", err)
	}

	// 获取内存使用
	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc
	memUsed := float64(memAfter-memBefore) / (1024 * 1024)
	elapsed := time.Since(start)

	// 显示结果
	fmt.Printf("解析耗时: %v\n", elapsed)
	fmt.Printf("内存用量: %.2f MB\n", memUsed)
	fmt.Printf("条目数量: %d\n", count)
}
