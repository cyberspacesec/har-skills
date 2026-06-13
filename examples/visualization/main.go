package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cyberspacesec/har-skills"
)

// 简易可视化类型
type Visualization struct {
	TotalRequests    int
	TotalSize        float64 // MB
	TotalDuration    float64 // ms
	Timeline         []TimelineEntry
	ContentTypeChart map[string]int
	DomainChart      map[string]int
	WaterfallData    []WaterfallEntry
}

// 时间线条目
type TimelineEntry struct {
	Timestamp time.Time
	EventType string
	URL       string
}

// 瀑布图条目
type WaterfallEntry struct {
	URL      string
	Method   string
	Status   int
	Start    float64 // 相对时间（毫秒）
	Duration float64 // 时长（毫秒）
	Size     float64 // 大小（KB）
	Type     string  // 内容类型
	Blocked  float64 // 阻塞时间
	DNS      float64 // DNS查询时间
	Connect  float64 // 连接时间
	Send     float64 // 发送时间
	Wait     float64 // 等待时间
	Receive  float64 // 接收时间
}

func main() {
	// 解析命令行参数
	harPath := flag.String("file", "", "HAR文件路径")
	outputDir := flag.String("output", "har_viz", "输出目录")
	flag.Parse()

	// 验证HAR文件路径
	if *harPath == "" {
		fmt.Println("请提供HAR文件路径。使用 -file 参数。")
		flag.Usage()
		os.Exit(1)
	}

	// 创建输出目录
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("无法创建输出目录: %v", err)
	}

	// 加载HAR文件
	fmt.Printf("正在分析HAR文件: %s\n", *harPath)
	harFile, err := har.ParseFile(*harPath)
	if err != nil {
		log.Fatalf("无法解析HAR文件: %v", err)
	}

	// 生成可视化数据
	vizData := generateVisualization(harFile)

	// 输出可视化HTML
	htmlPath := fmt.Sprintf("%s/visualization.html", *outputDir)
	if err := generateHTML(vizData, htmlPath); err != nil {
		log.Fatalf("生成HTML失败: %v", err)
	}

	fmt.Printf("可视化HTML已生成: %s\n", htmlPath)
}

// 生成可视化数据
func generateVisualization(harFile har.HARProvider) Visualization {
	viz := Visualization{
		ContentTypeChart: make(map[string]int),
		DomainChart:      make(map[string]int),
	}

	// 获取所有条目
	entries := harFile.GetEntries()
	if len(entries) == 0 {
		return viz
	}

	viz.TotalRequests = len(entries)

	// 寻找最早的请求时间
	var firstRequestTime time.Time
	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()
		if firstRequestTime.IsZero() || entry.StartedDateTime.Before(firstRequestTime) {
			firstRequestTime = entry.StartedDateTime
		}
	}

	// 处理每个条目
	for _, entryProvider := range entries {
		entry := entryProvider.ToStandard()

		// 统计条目信息
		domain := extractDomain(entry.Request.URL)
		contentType := extractContentType(entry.Response.Content.MimeType)

		// 更新总大小
		viz.TotalSize += float64(entry.Response.Content.Size) / (1024 * 1024) // 转换为MB

		// 更新总时长
		viz.TotalDuration += entry.Time

		// 更新内容类型统计
		viz.ContentTypeChart[contentType]++

		// 更新域名统计
		viz.DomainChart[domain]++

		// 添加到时间线
		relativeStart := entry.StartedDateTime.Sub(firstRequestTime).Milliseconds()

		// 创建瀑布图条目
		waterfall := WaterfallEntry{
			URL:      entry.Request.URL,
			Method:   entry.Request.Method,
			Status:   entry.Response.Status,
			Start:    float64(relativeStart),
			Duration: entry.Time,
			Size:     float64(entry.Response.Content.Size) / 1024, // KB
			Type:     contentType,
		}

		// 添加计时信息
		waterfall.Blocked = entry.Timings.Blocked
		waterfall.DNS = entry.Timings.DNS
		waterfall.Connect = entry.Timings.Connect
		waterfall.Send = entry.Timings.Send
		waterfall.Wait = entry.Timings.Wait
		waterfall.Receive = entry.Timings.Receive

		viz.WaterfallData = append(viz.WaterfallData, waterfall)
	}

	// 按照开始时间排序瀑布图数据
	sort.Slice(viz.WaterfallData, func(i, j int) bool {
		return viz.WaterfallData[i].Start < viz.WaterfallData[j].Start
	})

	return viz
}

// 从URL中提取域名
func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, "/")
	return parts[0]
}

// 提取内容类型
func extractContentType(mimeType string) string {
	// 简化MIME类型
	if strings.Contains(mimeType, "javascript") || strings.Contains(mimeType, "json") {
		return "JS"
	} else if strings.Contains(mimeType, "css") {
		return "CSS"
	} else if strings.Contains(mimeType, "html") {
		return "HTML"
	} else if strings.Contains(mimeType, "image") {
		return "图片"
	} else if strings.Contains(mimeType, "font") {
		return "字体"
	} else if strings.Contains(mimeType, "audio") || strings.Contains(mimeType, "video") {
		return "媒体"
	} else if strings.Contains(mimeType, "text") {
		return "文本"
	}
	return "其他"
}

// 生成HTML文件
func generateHTML(viz Visualization, outputPath string) error {
	// 创建HTML文件
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入HTML头部
	file.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>HAR文件可视化</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .section { margin-bottom: 30px; }
        .summary { display: flex; justify-content: space-between; flex-wrap: wrap; }
        .summary-item { 
            background: #f5f5f5; padding: 15px; border-radius: 5px; 
            flex: 1; min-width: 200px; margin: 5px; text-align: center;
        }
        .chart-container { display: flex; justify-content: space-between; flex-wrap: wrap; }
        .chart { flex: 1; min-width: 400px; height: 300px; margin: 10px; }
        .waterfall { 
            width: 100%; overflow-x: auto; margin-top: 20px;
            font-size: 12px; border-collapse: collapse;
        }
        .waterfall th, .waterfall td { padding: 5px; text-align: left; border-bottom: 1px solid #ddd; }
        .waterfall tr:hover { background-color: #f5f5f5; }
        .bar { height: 20px; position: relative; margin: 5px 0; }
        .bar-segment { position: absolute; height: 100%; }
        .blocked { background-color: #ccc; }
        .dns { background-color: #9C27B0; }
        .connect { background-color: #2196F3; }
        .send { background-color: #4CAF50; }
        .wait { background-color: #FF9800; }
        .receive { background-color: #F44336; }
        .legend { display: flex; margin: 10px 0; }
        .legend-item { display: flex; align-items: center; margin-right: 15px; }
        .legend-color { width: 15px; height: 15px; margin-right: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>HAR文件可视化</h1>
`)

	// 写入摘要信息
	file.WriteString(fmt.Sprintf(`
        <div class="section">
            <h2>摘要信息</h2>
            <div class="summary">
                <div class="summary-item">
                    <h3>总请求数</h3>
                    <div>%d</div>
                </div>
                <div class="summary-item">
                    <h3>总传输大小</h3>
                    <div>%.2f MB</div>
                </div>
                <div class="summary-item">
                    <h3>总加载时间</h3>
                    <div>%.2f 秒</div>
                </div>
            </div>
        </div>
`, viz.TotalRequests, viz.TotalSize, viz.TotalDuration/1000))

	// 写入图表部分
	file.WriteString(`
        <div class="section">
            <h2>统计图表</h2>
            <div class="chart-container">
                <div class="chart">
                    <canvas id="contentTypeChart"></canvas>
                </div>
                <div class="chart">
                    <canvas id="domainChart"></canvas>
                </div>
            </div>
        </div>
`)

	// 写入瀑布图部分
	file.WriteString(`
        <div class="section">
            <h2>请求瀑布图</h2>
            <div class="legend">
                <div class="legend-item"><div class="legend-color blocked"></div> 阻塞</div>
                <div class="legend-item"><div class="legend-color dns"></div> DNS</div>
                <div class="legend-item"><div class="legend-color connect"></div> 连接</div>
                <div class="legend-item"><div class="legend-color send"></div> 发送</div>
                <div class="legend-item"><div class="legend-color wait"></div> 等待</div>
                <div class="legend-item"><div class="legend-color receive"></div> 接收</div>
            </div>
            <table class="waterfall">
                <thead>
                    <tr>
                        <th>URL</th>
                        <th>方法</th>
                        <th>状态</th>
                        <th>类型</th>
                        <th>大小</th>
                        <th>时间</th>
                        <th>时间线</th>
                    </tr>
                </thead>
                <tbody>
`)

	// 写入瀑布图行
	maxTime := viz.WaterfallData[len(viz.WaterfallData)-1].Start + viz.WaterfallData[len(viz.WaterfallData)-1].Duration
	for _, entry := range viz.WaterfallData {
		// 计算开始比例
		startPercent := entry.Start / maxTime * 100

		// 计算时间段的宽度比例
		blockedWidth := entry.Blocked / entry.Duration * 100
		dnsWidth := entry.DNS / entry.Duration * 100
		connectWidth := entry.Connect / entry.Duration * 100
		sendWidth := entry.Send / entry.Duration * 100
		waitWidth := entry.Wait / entry.Duration * 100
		receiveWidth := entry.Receive / entry.Duration * 100

		// URL显示简化
		displayURL := entry.URL
		if len(displayURL) > 50 {
			displayURL = displayURL[:47] + "..."
		}

		file.WriteString(fmt.Sprintf(`
                <tr>
                    <td title="%s">%s</td>
                    <td>%s</td>
                    <td>%d</td>
                    <td>%s</td>
                    <td>%.1f KB</td>
                    <td>%.0f ms</td>
                    <td style="width: 400px;">
                        <div class="bar">
                            <div class="bar-segment blocked" style="left: %.1f%%; width: %.1f%%;"></div>
                            <div class="bar-segment dns" style="left: %.1f%%; width: %.1f%%;"></div>
                            <div class="bar-segment connect" style="left: %.1f%%; width: %.1f%%;"></div>
                            <div class="bar-segment send" style="left: %.1f%%; width: %.1f%%;"></div>
                            <div class="bar-segment wait" style="left: %.1f%%; width: %.1f%%;"></div>
                            <div class="bar-segment receive" style="left: %.1f%%; width: %.1f%%;"></div>
                        </div>
                    </td>
                </tr>
`,
			entry.URL, displayURL, entry.Method, entry.Status, entry.Type, entry.Size, entry.Duration,
			startPercent, blockedWidth,
			startPercent+blockedWidth, dnsWidth,
			startPercent+blockedWidth+dnsWidth, connectWidth,
			startPercent+blockedWidth+dnsWidth+connectWidth, sendWidth,
			startPercent+blockedWidth+dnsWidth+connectWidth+sendWidth, waitWidth,
			startPercent+blockedWidth+dnsWidth+connectWidth+sendWidth+waitWidth, receiveWidth,
		))
	}

	// 关闭表格和容器
	file.WriteString(`
                </tbody>
            </table>
        </div>
    </div>
`)

	// 写入JavaScript图表代码
	file.WriteString(`
    <script>
        // 内容类型图表
        const contentTypeData = {
            labels: [`)

	// 写入内容类型标签
	var contentTypeLabels []string
	var contentTypeValues []int
	for label, value := range viz.ContentTypeChart {
		contentTypeLabels = append(contentTypeLabels, label)
		contentTypeValues = append(contentTypeValues, value)
	}
	for i, label := range contentTypeLabels {
		if i > 0 {
			file.WriteString(", ")
		}
		file.WriteString(fmt.Sprintf(`"%s"`, label))
	}

	file.WriteString(`],
            datasets: [{
                label: '请求数',
                data: [`)

	// 写入内容类型值
	for i, value := range contentTypeValues {
		if i > 0 {
			file.WriteString(", ")
		}
		file.WriteString(fmt.Sprintf("%d", value))
	}

	file.WriteString(`],
                backgroundColor: [
                    '#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF', '#FF9F40', '#C9CBCF'
                ]
            }]
        };

        // 域名图表
        const domainData = {
            labels: [`)

	// 写入域名标签
	var domainLabels []string
	var domainValues []int
	for label, value := range viz.DomainChart {
		domainLabels = append(domainLabels, label)
		domainValues = append(domainValues, value)
	}
	for i, label := range domainLabels {
		if i > 0 {
			file.WriteString(", ")
		}
		file.WriteString(fmt.Sprintf(`"%s"`, label))
	}

	file.WriteString(`],
            datasets: [{
                label: '请求数',
                data: [`)

	// 写入域名值
	for i, value := range domainValues {
		if i > 0 {
			file.WriteString(", ")
		}
		file.WriteString(fmt.Sprintf("%d", value))
	}

	file.WriteString(`],
                backgroundColor: [
                    '#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF', '#FF9F40', '#C9CBCF',
                    '#2E2EFE', '#088A08', '#FF0040', '#FF8000', '#01A9DB', '#D7DF01', '#6A0888'
                ]
            }]
        };

        // 渲染图表
        window.onload = function() {
            // 内容类型图表
            new Chart(document.getElementById('contentTypeChart').getContext('2d'), {
                type: 'pie',
                data: contentTypeData,
                options: {
                    responsive: true,
                    plugins: {
                        title: {
                            display: true,
                            text: '按内容类型分布'
                        }
                    }
                }
            });

            // 域名图表
            new Chart(document.getElementById('domainChart').getContext('2d'), {
                type: 'pie',
                data: domainData,
                options: {
                    responsive: true,
                    plugins: {
                        title: {
                            display: true,
                            text: '按域名分布'
                        }
                    }
                }
            });
        };
    </script>
</body>
</html>`)

	return nil
}
