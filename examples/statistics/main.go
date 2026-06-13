package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cyberspacesec/har-skills"
)

// 统计类型
type Stats struct {
	TotalRequests       int
	TotalSize           int64
	TotalDuration       float64
	AvgRequestSize      float64
	AvgResponseTime     float64
	MedianResponseTime  float64
	MinResponseTime     float64
	MaxResponseTime     float64
	P95ResponseTime     float64
	StatusCodes         map[int]int
	ContentTypes        map[string]int
	DomainsCount        map[string]int
	SlowRequests        []RequestInfo
	LargeResponses      []RequestInfo
	TimeByContentType   map[string]float64
	SizeByContentType   map[string]int64
	RequestsPerSecond   map[int]int
	SuccessRate         float64
	CacheHitRate        float64
	PerformanceByDomain map[string]DomainPerformance
}

// 请求信息
type RequestInfo struct {
	URL          string
	Method       string
	StatusCode   int
	Size         int
	Duration     float64
	StartTime    time.Time
	ContentType  string
	Domain       string
	CacheStatus  string
	ResponseSize int
}

// 域名性能统计
type DomainPerformance struct {
	Requests     int
	TotalTime    float64
	TotalSize    int64
	AvgTime      float64
	AvgSize      float64
	MaxTime      float64
	MinTime      float64
	SuccessCount int
	ErrorCount   int
}

func main() {
	// 解析命令行参数
	harPath := flag.String("file", "", "HAR文件路径")
	outputFormat := flag.String("format", "text", "输出格式：text, json")
	slowThreshold := flag.Float64("slow", 500, "慢请求阈值（毫秒）")
	largeThreshold := flag.Int("large", 1000000, "大响应阈值（字节）")
	flag.Parse()

	// 验证HAR文件路径
	if *harPath == "" {
		fmt.Println("请提供HAR文件路径。使用 -file 参数。")
		flag.Usage()
		os.Exit(1)
	}

	// 加载HAR文件 - 使用内存优化模式
	fmt.Printf("正在分析HAR文件: %s\n", *harPath)
	harFile, err := har.ParseFile(*harPath, har.WithMemoryOptimized())
	if err != nil {
		log.Fatalf("无法解析HAR文件: %v", err)
	}

	// 计算统计数据
	stats := calculateStats(harFile, *slowThreshold, *largeThreshold)

	// 输出统计信息
	switch *outputFormat {
	case "text":
		printTextStats(stats)
	case "json":
		printJSONStats(stats)
	default:
		fmt.Printf("不支持的输出格式: %s\n", *outputFormat)
	}
}

// 计算统计数据
func calculateStats(harFile har.HARProvider, slowThreshold float64, largeThreshold int) Stats {
	stats := Stats{
		StatusCodes:         make(map[int]int),
		ContentTypes:        make(map[string]int),
		DomainsCount:        make(map[string]int),
		TimeByContentType:   make(map[string]float64),
		SizeByContentType:   make(map[string]int64),
		RequestsPerSecond:   make(map[int]int),
		PerformanceByDomain: make(map[string]DomainPerformance),
	}

	entries := harFile.GetEntries()
	if len(entries) == 0 {
		return stats
	}

	// 收集响应时间以计算中位数和百分位数
	var responseTimes []float64
	var totalSuccessful, totalCached int

	// 最早的请求时间
	var firstRequestTime time.Time

	// 处理每个条目
	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()
		startTime := entry.StartedDateTime

		// 初始化首次请求时间
		if firstRequestTime.IsZero() || startTime.Before(firstRequestTime) {
			firstRequestTime = startTime
		}

		// 提取请求信息
		reqInfo := extractRequestInfo(entry)

		// 更新基本计数统计
		stats.TotalRequests++
		stats.TotalSize += int64(reqInfo.Size)
		stats.TotalDuration += reqInfo.Duration
		stats.StatusCodes[reqInfo.StatusCode]++
		stats.ContentTypes[reqInfo.ContentType]++
		stats.DomainsCount[reqInfo.Domain]++

		// 按内容类型统计
		baseContentType := getBaseContentType(reqInfo.ContentType)
		stats.TimeByContentType[baseContentType] += reqInfo.Duration
		stats.SizeByContentType[baseContentType] += int64(reqInfo.Size)

		// 按秒统计请求数
		secondsSinceFirst := int(startTime.Sub(firstRequestTime).Seconds())
		stats.RequestsPerSecond[secondsSinceFirst]++

		// 收集响应时间
		responseTimes = append(responseTimes, reqInfo.Duration)

		// 慢请求和大响应
		if reqInfo.Duration > slowThreshold {
			stats.SlowRequests = append(stats.SlowRequests, reqInfo)
		}
		if reqInfo.Size > int(largeThreshold) {
			stats.LargeResponses = append(stats.LargeResponses, reqInfo)
		}

		// 统计成功请求和缓存命中
		if reqInfo.StatusCode >= 200 && reqInfo.StatusCode < 400 {
			totalSuccessful++
		}
		if reqInfo.CacheStatus == "hit" {
			totalCached++
		}

		// 按域名统计性能
		domainPerf, exists := stats.PerformanceByDomain[reqInfo.Domain]
		if !exists {
			domainPerf = DomainPerformance{
				MinTime: math.MaxFloat64,
			}
		}
		domainPerf.Requests++
		domainPerf.TotalTime += reqInfo.Duration
		domainPerf.TotalSize += int64(reqInfo.Size)
		if reqInfo.Duration > domainPerf.MaxTime {
			domainPerf.MaxTime = reqInfo.Duration
		}
		if reqInfo.Duration < domainPerf.MinTime {
			domainPerf.MinTime = reqInfo.Duration
		}
		if reqInfo.StatusCode >= 200 && reqInfo.StatusCode < 400 {
			domainPerf.SuccessCount++
		} else {
			domainPerf.ErrorCount++
		}
		stats.PerformanceByDomain[reqInfo.Domain] = domainPerf
	}

	// 计算平均值
	stats.AvgRequestSize = float64(stats.TotalSize) / float64(stats.TotalRequests)
	stats.AvgResponseTime = stats.TotalDuration / float64(stats.TotalRequests)

	// 计算成功率和缓存命中率
	stats.SuccessRate = float64(totalSuccessful) / float64(stats.TotalRequests) * 100
	if totalCached > 0 {
		stats.CacheHitRate = float64(totalCached) / float64(stats.TotalRequests) * 100
	}

	// 计算中位数和百分位数
	if len(responseTimes) > 0 {
		sort.Float64s(responseTimes)
		stats.MinResponseTime = responseTimes[0]
		stats.MaxResponseTime = responseTimes[len(responseTimes)-1]
		stats.MedianResponseTime = percentile(responseTimes, 50)
		stats.P95ResponseTime = percentile(responseTimes, 95)
	}

	// 计算每个域名的平均性能
	for domain, perf := range stats.PerformanceByDomain {
		if perf.Requests > 0 {
			perf.AvgTime = perf.TotalTime / float64(perf.Requests)
			perf.AvgSize = float64(perf.TotalSize) / float64(perf.Requests)
			stats.PerformanceByDomain[domain] = perf
		}
	}

	return stats
}

// 提取请求信息
func extractRequestInfo(entry har.Entries) RequestInfo {
	info := RequestInfo{
		URL:         entry.Request.URL,
		Method:      entry.Request.Method,
		StatusCode:  entry.Response.Status,
		StartTime:   entry.StartedDateTime,
		Duration:    entry.Time,
		Size:        entry.Response.Content.Size,
		ContentType: entry.Response.Content.MimeType,
		CacheStatus: "miss", // 默认为miss
	}

	// 提取域名
	info.Domain = extractDomain(entry.Request.URL)

	// 识别缓存状态 - 简化版，实际使用时可能需要调整
	// 由于实际的HAR结构复杂性，简化处理方式
	if entry.Cache.AfterRequest != nil {
		info.CacheStatus = "hit"
	}

	return info
}

// 从URL中提取域名
func extractDomain(url string) string {
	// 简单的域名提取，可以根据需要改进
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, "/")
	return parts[0]
}

// 获取内容类型的基本部分
func getBaseContentType(contentType string) string {
	parts := strings.Split(contentType, ";")
	baseType := strings.TrimSpace(parts[0])

	// 进一步归类为大类
	if strings.Contains(baseType, "javascript") || strings.Contains(baseType, "json") {
		return "javascript"
	} else if strings.Contains(baseType, "css") {
		return "css"
	} else if strings.Contains(baseType, "html") {
		return "html"
	} else if strings.Contains(baseType, "image") {
		return "image"
	} else if strings.Contains(baseType, "font") {
		return "font"
	} else if strings.Contains(baseType, "video") || strings.Contains(baseType, "audio") {
		return "media"
	}

	return baseType
}

// 计算百分位数
func percentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	index := int(math.Ceil(float64(len(values))*percentile/100)) - 1
	if index < 0 {
		index = 0
	} else if index >= len(values) {
		index = len(values) - 1
	}

	return values[index]
}

// 以文本格式打印统计信息
func printTextStats(stats Stats) {
	fmt.Println("\n========== HAR 文件统计分析 ==========")
	fmt.Printf("总请求数: %d\n", stats.TotalRequests)
	fmt.Printf("总传输大小: %.2f MB\n", float64(stats.TotalSize)/(1024*1024))
	fmt.Printf("总加载时间: %.2f 秒\n", stats.TotalDuration/1000)
	fmt.Printf("平均请求大小: %.2f KB\n", stats.AvgRequestSize/1024)
	fmt.Printf("平均响应时间: %.2f ms\n", stats.AvgResponseTime)
	fmt.Printf("响应时间中位数: %.2f ms\n", stats.MedianResponseTime)
	fmt.Printf("最小响应时间: %.2f ms\n", stats.MinResponseTime)
	fmt.Printf("最大响应时间: %.2f ms\n", stats.MaxResponseTime)
	fmt.Printf("95%%响应时间: %.2f ms\n", stats.P95ResponseTime)
	fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate)
	fmt.Printf("缓存命中率: %.2f%%\n", stats.CacheHitRate)

	fmt.Println("\n---------- 状态码分布 ----------")
	for code, count := range stats.StatusCodes {
		fmt.Printf("%d: %d (%.2f%%)\n", code, count, float64(count)/float64(stats.TotalRequests)*100)
	}

	fmt.Println("\n---------- 内容类型分布 ----------")
	for contentType, count := range stats.ContentTypes {
		if count > 0 {
			fmt.Printf("%s: %d (%.2f%%)\n", contentType, count, float64(count)/float64(stats.TotalRequests)*100)
		}
	}

	fmt.Println("\n---------- 按内容类型统计加载时间 ----------")
	for contentType, time := range stats.TimeByContentType {
		if time > 0 {
			fmt.Printf("%s: %.2f ms (%.2f%%)\n", contentType, time, time/stats.TotalDuration*100)
		}
	}

	fmt.Println("\n---------- 按内容类型统计传输大小 ----------")
	for contentType, size := range stats.SizeByContentType {
		if size > 0 {
			fmt.Printf("%s: %.2f KB (%.2f%%)\n", contentType, float64(size)/1024, float64(size)/float64(stats.TotalSize)*100)
		}
	}

	fmt.Println("\n---------- 域名性能 ----------")
	// 按请求数排序
	type DomainStat struct {
		Domain string
		Perf   DomainPerformance
	}
	var domainStats []DomainStat
	for domain, perf := range stats.PerformanceByDomain {
		domainStats = append(domainStats, DomainStat{domain, perf})
	}

	sort.Slice(domainStats, func(i, j int) bool {
		return domainStats[i].Perf.Requests > domainStats[j].Perf.Requests
	})

	for _, ds := range domainStats {
		fmt.Printf("%s:\n", ds.Domain)
		fmt.Printf("  请求数: %d (%.2f%%)\n", ds.Perf.Requests, float64(ds.Perf.Requests)/float64(stats.TotalRequests)*100)
		fmt.Printf("  平均响应时间: %.2f ms\n", ds.Perf.AvgTime)
		fmt.Printf("  平均大小: %.2f KB\n", ds.Perf.AvgSize/1024)
		fmt.Printf("  成功请求: %d, 失败请求: %d\n", ds.Perf.SuccessCount, ds.Perf.ErrorCount)
	}

	fmt.Println("\n---------- 最慢的请求 (Top 5) ----------")
	sort.Slice(stats.SlowRequests, func(i, j int) bool {
		return stats.SlowRequests[i].Duration > stats.SlowRequests[j].Duration
	})

	for i, req := range stats.SlowRequests {
		if i >= 5 {
			break
		}
		fmt.Printf("%d. %s %s\n", i+1, req.Method, req.URL)
		fmt.Printf("   响应时间: %.2f ms, 状态码: %d, 大小: %.2f KB\n",
			req.Duration, req.StatusCode, float64(req.Size)/1024)
	}

	fmt.Println("\n---------- 最大的响应 (Top 5) ----------")
	sort.Slice(stats.LargeResponses, func(i, j int) bool {
		return stats.LargeResponses[i].Size > stats.LargeResponses[j].Size
	})

	for i, req := range stats.LargeResponses {
		if i >= 5 {
			break
		}
		fmt.Printf("%d. %s %s\n", i+1, req.Method, req.URL)
		fmt.Printf("   大小: %.2f KB, 状态码: %d, 响应时间: %.2f ms\n",
			float64(req.Size)/1024, req.StatusCode, req.Duration)
	}
}

// 以JSON格式打印统计信息
func printJSONStats(stats Stats) {
	// 此处应实现JSON格式输出
	// 为简化示例，这里仅打印文本提示
	fmt.Println("JSON输出功能正在开发中...")
}
