package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/cyberspacesec/har-skills"
)

const (
	VERSION = "0.1.0"
)

// 命令行参数
type CommandArgs struct {
	HarFile   string
	Command   string
	Filter    string
	Format    string
	Limit     int
	SortField string
	SortOrder string
	Output    string
}

// 主函数
func main() {
	// 解析命令行参数
	args := parseArgs()

	// 验证HAR文件路径
	if args.HarFile == "" {
		fmt.Println("错误: 未提供HAR文件路径")
		printUsage()
		os.Exit(1)
	}

	// 加载HAR文件
	harFile, err := har.ParseFile(args.HarFile, har.WithMemoryOptimized())
	if err != nil {
		log.Fatalf("无法解析HAR文件: %v", err)
	}

	// 执行命令
	switch args.Command {
	case "info":
		showInfo(harFile)
	case "list":
		listEntries(harFile, args)
	case "find":
		findEntries(harFile, args)
	case "headers":
		showHeaders(harFile, args)
	case "timing":
		showTiming(harFile, args)
	case "extract":
		extractContent(harFile, args)
	default:
		fmt.Printf("未知命令: %s\n", args.Command)
		printUsage()
		os.Exit(1)
	}
}

// 解析命令行参数
func parseArgs() CommandArgs {
	// 定义命令行参数
	harFilePtr := flag.String("file", "", "HAR文件路径")
	commandPtr := flag.String("cmd", "info", "要执行的命令 (info, list, find, headers, timing, extract)")
	filterPtr := flag.String("filter", "", "筛选条件 (URL正则表达式、状态码、类型等)")
	formatPtr := flag.String("format", "text", "输出格式 (text, json, csv)")
	limitPtr := flag.Int("limit", 10, "结果数量限制")
	sortFieldPtr := flag.String("sort", "time", "排序字段 (time, size, url, status)")
	sortOrderPtr := flag.String("order", "desc", "排序顺序 (asc, desc)")
	outputPtr := flag.String("output", "", "输出文件路径")

	// 自定义使用说明
	flag.Usage = printUsage

	// 解析参数
	flag.Parse()

	return CommandArgs{
		HarFile:   *harFilePtr,
		Command:   *commandPtr,
		Filter:    *filterPtr,
		Format:    *formatPtr,
		Limit:     *limitPtr,
		SortField: *sortFieldPtr,
		SortOrder: *sortOrderPtr,
		Output:    *outputPtr,
	}
}

// 打印使用说明
func printUsage() {
	fmt.Printf("HAR CLI 工具 v%s - HTTP Archive 文件命令行分析工具\n\n", VERSION)
	fmt.Println("用法: har-cli -file <har文件路径> -cmd <命令> [选项]")
	fmt.Println("\n可用命令:")
	fmt.Println("  info      - 显示HAR文件基本信息")
	fmt.Println("  list      - 列出HAR文件中的请求")
	fmt.Println("  find      - 查找匹配条件的请求")
	fmt.Println("  headers   - 显示请求或响应头")
	fmt.Println("  timing    - 显示请求时间分析")
	fmt.Println("  extract   - 提取响应内容")
	fmt.Println("\n选项:")
	flag.PrintDefaults()
	fmt.Println("\n示例:")
	fmt.Println("  har-cli -file example.har -cmd info")
	fmt.Println("  har-cli -file example.har -cmd list -limit 20")
	fmt.Println("  har-cli -file example.har -cmd find -filter \"api/users\"")
	fmt.Println("  har-cli -file example.har -cmd timing -sort time -order desc")
}

// 显示HAR文件基本信息
func showInfo(harFile har.HARProvider) {
	entries := harFile.GetEntries()
	pages := harFile.GetPages()
	creator := harFile.GetCreator()

	fmt.Println("=== HAR文件信息 ===")
	fmt.Printf("版本: %s\n", harFile.GetVersion())
	fmt.Printf("创建者: %s %s\n", creator.Name, creator.Version)
	fmt.Printf("页面数: %d\n", len(pages))
	fmt.Printf("请求数: %d\n", len(entries))

	// 计算总请求大小和总响应时间
	var totalSize int64
	var totalTime float64
	statusCodes := make(map[int]int)
	methods := make(map[string]int)
	contentTypes := make(map[string]int)
	domains := make(map[string]int)

	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()
		totalSize += int64(entry.Response.Content.Size)
		totalTime += entry.Time

		// 状态码统计
		statusCodes[entry.Response.Status]++

		// 请求方法统计
		methods[entry.Request.Method]++

		// 内容类型统计
		contentType := entry.Response.Content.MimeType
		if contentType != "" {
			// 简化内容类型
			contentType = strings.Split(contentType, ";")[0]
			contentTypes[contentType]++
		}

		// 域名统计
		domain := extractDomain(entry.Request.URL)
		domains[domain]++
	}

	fmt.Printf("\n总传输大小: %.2f MB\n", float64(totalSize)/(1024*1024))
	fmt.Printf("总响应时间: %.2f 秒\n", totalTime/1000)

	if len(entries) > 0 {
		fmt.Printf("平均响应大小: %.2f KB\n", float64(totalSize)/float64(len(entries))/1024)
		fmt.Printf("平均响应时间: %.2f ms\n", totalTime/float64(len(entries)))
	}

	// 显示状态码分布
	fmt.Println("\n状态码分布:")
	for status, count := range statusCodes {
		fmt.Printf("  %d: %d (%.1f%%)\n", status, count, float64(count)/float64(len(entries))*100)
	}

	// 显示请求方法分布
	fmt.Println("\n请求方法分布:")
	for method, count := range methods {
		fmt.Printf("  %s: %d (%.1f%%)\n", method, count, float64(count)/float64(len(entries))*100)
	}

	// 显示前几个域名
	fmt.Println("\n域名分布 (Top 5):")
	domainList := sortMapByValue(domains)
	for i, item := range domainList {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s: %d (%.1f%%)\n", item.Key, item.Value, float64(item.Value)/float64(len(entries))*100)
	}

	// 显示前几个内容类型
	fmt.Println("\n内容类型分布 (Top 5):")
	contentTypeList := sortMapByValue(contentTypes)
	for i, item := range contentTypeList {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s: %d (%.1f%%)\n", item.Key, item.Value, float64(item.Value)/float64(len(entries))*100)
	}
}

// 列出HAR文件中的请求
func listEntries(harFile har.HARProvider, args CommandArgs) {
	entries := harFile.GetEntries()

	// 转换为标准格式
	var standardEntries []har.Entries
	for _, entryProvider := range entries {
		standardEntries = append(standardEntries, entryProvider.ToStandard())
	}

	// 排序
	sortEntries(standardEntries, args.SortField, args.SortOrder)

	// 限制结果数量
	if args.Limit > 0 && args.Limit < len(standardEntries) {
		standardEntries = standardEntries[:args.Limit]
	}

	// 打印结果
	printEntries(standardEntries, args.Format, args.Output)
}

// 按条件查找请求
func findEntries(harFile har.HARProvider, args CommandArgs) {
	if args.Filter == "" {
		fmt.Println("错误: 查找命令需要指定 -filter 参数")
		return
	}

	entries := harFile.GetEntries()
	var filtered []har.Entries

	// 检查是否是状态码筛选
	statusCode, err := strconv.Atoi(args.Filter)
	isStatusFilter := (err == nil)

	// 创建正则表达式
	var re *regexp.Regexp
	if !isStatusFilter {
		re, err = regexp.Compile(args.Filter)
		if err != nil {
			fmt.Printf("警告: 无效的正则表达式 '%s', 将使用简单字符串匹配\n", args.Filter)
			re = nil
		}
	}

	// 筛选条目
	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()

		// 状态码筛选
		if isStatusFilter && entry.Response.Status == statusCode {
			filtered = append(filtered, entry)
			continue
		}

		// URL正则筛选
		if !isStatusFilter {
			if re != nil && re.MatchString(entry.Request.URL) {
				filtered = append(filtered, entry)
			} else if !isStatusFilter && strings.Contains(entry.Request.URL, args.Filter) {
				filtered = append(filtered, entry)
			}
		}
	}

	fmt.Printf("找到 %d 个匹配的请求\n", len(filtered))

	// 排序
	sortEntries(filtered, args.SortField, args.SortOrder)

	// 限制结果数量
	if args.Limit > 0 && args.Limit < len(filtered) {
		filtered = filtered[:args.Limit]
	}

	// 打印结果
	printEntries(filtered, args.Format, args.Output)
}

// 排序用的键值对
type KeyValue struct {
	Key   string
	Value int
}

// 从URL中提取域名
func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, "/")
	return parts[0]
}

// 按值排序map
func sortMapByValue(m map[string]int) []KeyValue {
	var ss []KeyValue
	for k, v := range m {
		ss = append(ss, KeyValue{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss
}

// 排序条目列表
func sortEntries(entries []har.Entries, field, order string) {
	sort.Slice(entries, func(i, j int) bool {
		var less bool
		switch strings.ToLower(field) {
		case "time":
			less = entries[i].Time < entries[j].Time
		case "size":
			less = entries[i].Response.Content.Size < entries[j].Response.Content.Size
		case "url":
			less = entries[i].Request.URL < entries[j].Request.URL
		case "status":
			less = entries[i].Response.Status < entries[j].Response.Status
		default:
			less = entries[i].Time < entries[j].Time
		}

		if strings.ToLower(order) == "desc" {
			return !less
		}
		return less
	})
}

// 打印条目列表
func printEntries(entries []har.Entries, format, outputPath string) {
	if len(entries) == 0 {
		fmt.Println("无结果")
		return
	}

	switch format {
	case "json":
		printEntriesJSON(entries, outputPath)
	case "csv":
		printEntriesCSV(entries, outputPath)
	default: // text
		printEntriesText(entries, outputPath)
	}
}

// 打印条目(文本格式)
func printEntriesText(entries []har.Entries, outputPath string) {
	// 创建输出流
	var output *os.File
	var err error
	if outputPath != "" {
		output, err = os.Create(outputPath)
		if err != nil {
			fmt.Printf("无法创建输出文件: %v\n", err)
			output = os.Stdout
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// 使用表格格式化
	w := tabwriter.NewWriter(output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "索引\t方法\t状态\t大小(KB)\t时间(ms)\tURL")
	fmt.Fprintln(w, "----\t----\t----\t--------\t--------\t---")

	for i, entry := range entries {
		fmt.Fprintf(w, "%d\t%s\t%d\t%.1f\t%.1f\t%s\n",
			i+1,
			entry.Request.Method,
			entry.Response.Status,
			float64(entry.Response.Content.Size)/1024,
			entry.Time,
			entry.Request.URL,
		)
	}
	w.Flush()
}

// 打印条目(JSON格式)
func printEntriesJSON(entries []har.Entries, outputPath string) {
	// 创建简化的输出结构
	type SimpleEntry struct {
		Method     string  `json:"method"`
		URL        string  `json:"url"`
		Status     int     `json:"status"`
		StatusText string  `json:"statusText"`
		MimeType   string  `json:"mimeType"`
		Size       int     `json:"size"`
		Time       float64 `json:"time"`
	}

	var simpleEntries []SimpleEntry
	for _, entry := range entries {
		simpleEntries = append(simpleEntries, SimpleEntry{
			Method:     entry.Request.Method,
			URL:        entry.Request.URL,
			Status:     entry.Response.Status,
			StatusText: entry.Response.StatusText,
			MimeType:   entry.Response.Content.MimeType,
			Size:       entry.Response.Content.Size,
			Time:       entry.Time,
		})
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(simpleEntries, "", "  ")
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)
		return
	}

	// 输出
	if outputPath != "" {
		err := os.WriteFile(outputPath, jsonData, 0644)
		if err != nil {
			fmt.Printf("写入文件失败: %v\n", err)
			fmt.Println(string(jsonData))
		} else {
			fmt.Printf("已写入 %d 条记录到 %s\n", len(entries), outputPath)
		}
	} else {
		fmt.Println(string(jsonData))
	}
}

// 打印条目(CSV格式)
func printEntriesCSV(entries []har.Entries, outputPath string) {
	// 创建输出流
	var output *os.File
	var err error
	if outputPath != "" {
		output, err = os.Create(outputPath)
		if err != nil {
			fmt.Printf("无法创建输出文件: %v\n", err)
			output = os.Stdout
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// 写入CSV头
	fmt.Fprintln(output, "method,url,status,statusText,mimeType,size,time")

	// 写入数据行
	for _, entry := range entries {
		// 处理URL中的逗号，确保CSV格式正确
		url := strings.ReplaceAll(entry.Request.URL, ",", "%2C")
		statusText := strings.ReplaceAll(entry.Response.StatusText, ",", " ")
		mimeType := strings.ReplaceAll(entry.Response.Content.MimeType, ",", " ")

		fmt.Fprintf(output, "%s,%s,%d,%s,%s,%d,%.1f\n",
			entry.Request.Method,
			url,
			entry.Response.Status,
			statusText,
			mimeType,
			entry.Response.Content.Size,
			entry.Time,
		)
	}

	if outputPath != "" {
		fmt.Printf("已写入 %d 条记录到 %s\n", len(entries), outputPath)
	}
}

// 显示请求或响应头
func showHeaders(harFile har.HARProvider, args CommandArgs) {
	if args.Filter == "" {
		fmt.Println("错误: 请指定 -filter 参数来匹配URL")
		return
	}

	entries := harFile.GetEntries()
	found := false

	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()
		if strings.Contains(entry.Request.URL, args.Filter) {
			found = true
			fmt.Printf("=== %s %s ===\n", entry.Request.Method, entry.Request.URL)

			fmt.Println("\n请求头:")
			for _, header := range entry.Request.Headers {
				fmt.Printf("  %s: %s\n", header.Name, header.Value)
			}

			fmt.Println("\n响应头:")
			for _, header := range entry.Response.Headers {
				fmt.Printf("  %s: %s\n", header.Name, header.Value)
			}

			fmt.Println("\n响应状态:", entry.Response.Status, entry.Response.StatusText)
			fmt.Println("内容类型:", entry.Response.Content.MimeType)
			fmt.Println("内容大小:", entry.Response.Content.Size, "字节")

			// 如果找到多个匹配，只处理第一个
			if args.Limit <= 1 {
				break
			}
		}
	}

	if !found {
		fmt.Println("未找到匹配的请求")
	}
}

// 显示请求时间分析
func showTiming(harFile har.HARProvider, args CommandArgs) {
	entries := harFile.GetEntries()

	// 转换为标准格式并预处理
	var timingEntries []struct {
		URL     string
		Method  string
		Status  int
		Time    float64
		Blocked float64
		DNS     float64
		Connect float64
		Send    float64
		Wait    float64
		Receive float64
		SSL     float64
	}

	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()

		// 如果有过滤条件，跳过不匹配的
		if args.Filter != "" && !strings.Contains(entry.Request.URL, args.Filter) {
			continue
		}

		timing := struct {
			URL     string
			Method  string
			Status  int
			Time    float64
			Blocked float64
			DNS     float64
			Connect float64
			Send    float64
			Wait    float64
			Receive float64
			SSL     float64
		}{
			URL:     entry.Request.URL,
			Method:  entry.Request.Method,
			Status:  entry.Response.Status,
			Time:    entry.Time,
			Blocked: entry.Timings.Blocked,
			DNS:     entry.Timings.DNS,
			Connect: entry.Timings.Connect,
			Send:    entry.Timings.Send,
			Wait:    entry.Timings.Wait,
			Receive: entry.Timings.Receive,
			SSL:     entry.Timings.Ssl,
		}

		timingEntries = append(timingEntries, timing)
	}

	// 排序
	sort.Slice(timingEntries, func(i, j int) bool {
		if args.SortField == "time" {
			return timingEntries[i].Time > timingEntries[j].Time
		}
		return timingEntries[i].Wait > timingEntries[j].Wait
	})

	// 限制结果数
	if args.Limit > 0 && args.Limit < len(timingEntries) {
		timingEntries = timingEntries[:args.Limit]
	}

	// 打印结果
	fmt.Println("=== 请求时间分析 ===")
	fmt.Println("(时间单位: 毫秒)")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "URL\t总时间\t阻塞\tDNS\t连接\tSSL\t发送\t等待\t接收")

	for _, t := range timingEntries {
		// 简化URL显示
		url := t.URL
		if len(url) > 50 {
			url = url[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\n",
			url, t.Time, t.Blocked, t.DNS, t.Connect, t.SSL, t.Send, t.Wait, t.Receive)
	}
	w.Flush()
}

// 提取响应内容
func extractContent(harFile har.HARProvider, args CommandArgs) {
	if args.Filter == "" {
		fmt.Println("错误: 请指定 -filter 参数来匹配URL")
		return
	}

	if args.Output == "" {
		fmt.Println("错误: 请指定 -output 参数来设置输出文件")
		return
	}

	entries := harFile.GetEntries()
	found := false

	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()
		if strings.Contains(entry.Request.URL, args.Filter) {
			found = true

			// 确认输出文件
			fmt.Printf("找到匹配请求: %s\n", entry.Request.URL)
			fmt.Printf("内容类型: %s, 大小: %d 字节\n",
				entry.Response.Content.MimeType, entry.Response.Content.Size)

			// 实际代码中应该从HAR中提取内容，这里为演示简化处理
			fmt.Println("注意: 当前版本不支持从HAR中提取实际内容")
			fmt.Printf("将在未来版本添加此功能，会写入到: %s\n", args.Output)

			// 如果找到多个匹配，只处理第一个
			if args.Limit <= 1 {
				break
			}
		}
	}

	if !found {
		fmt.Println("未找到匹配的请求")
	}
}
