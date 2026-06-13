package har

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// HarStatistics 表示HAR文件的统计信息
type HarStatistics struct {
	TotalRequests     int              // 总请求数
	TotalTransferred int64            // 总传输字节数
	TotalUncompressed int64           // 总未压缩字节数
	TotalTime         float64         // 从第一个请求到最后一个响应的总时间(ms)
	AvgTime           float64         // 平均请求时间(ms)
	MaxTime           float64         // 最大请求时间(ms)
	MinTime           float64         // 最小请求时间(ms)
	MedianTime        float64         // 中位数请求时间(ms)
	P95Time           float64         // 95百分位请求时间(ms)
	P99Time           float64         // 99百分位请求时间(ms)
	Methods           map[string]int  // HTTP方法分布
	StatusCodes       map[int]int     // 状态码分布
	ContentTypes      map[string]int  // 内容类型分布
	Domains           map[string]int  // 域名分布
	ErrorCount        int             // 错误请求数(4xx+5xx)
	RedirectCount     int             // 重定向数(3xx)
	TimingsSummary    TimingsSummary  // 时间指标汇总
	StartTime         time.Time       // 最早请求时间
	EndTime           time.Time       // 最晚请求时间
}

// TimingsSummary 表示时间指标的统计汇总
type TimingsSummary struct {
	AvgBlocked float64 // 平均阻塞时间(ms)
	AvgDNS     float64 // 平均DNS解析时间(ms)
	AvgConnect float64 // 平均TCP连接时间(ms)
	AvgSend    float64 // 平均发送时间(ms)
	AvgWait    float64 // 平均等待时间(ms)
	AvgReceive float64 // 平均接收时间(ms)
	AvgSSL     float64 // 平均SSL握手时间(ms)
	MaxBlocked float64 // 最大阻塞时间(ms)
	MaxDNS     float64 // 最大DNS解析时间(ms)
	MaxConnect float64 // 最大TCP连接时间(ms)
	MaxSend    float64 // 最大发送时间(ms)
	MaxWait    float64 // 最大等待时间(ms)
	MaxReceive float64 // 最大接收时间(ms)
	MaxSSL     float64 // 最大SSL握手时间(ms)
	MinBlocked float64 // 最小阻塞时间(ms)
	MinDNS     float64 // 最小DNS解析时间(ms)
	MinConnect float64 // 最小TCP连接时间(ms)
	MinSend    float64 // 最小发送时间(ms)
	MinWait    float64 // 最小等待时间(ms)
	MinReceive float64 // 最小接收时间(ms)
	MinSSL     float64 // 最小SSL握手时间(ms)
}

// DomainStats 表示按域名的统计信息
type DomainStats struct {
	RequestCount     int     // 请求数
	TotalTime        float64 // 总耗时(ms)
	AvgTime          float64 // 平均耗时(ms)
	TotalTransferred int64   // 总传输字节数
	ErrorCount       int     // 错误数
}

// Statistics 计算HAR文件的完整统计信息
func (h *Har) Statistics() *HarStatistics {
	if h == nil || len(h.Log.Entries) == 0 {
		return &HarStatistics{
			Methods:      make(map[string]int),
			StatusCodes:  make(map[int]int),
			ContentTypes: make(map[string]int),
			Domains:      make(map[string]int),
		}
	}

	stats := &HarStatistics{
		TotalRequests: len(h.Log.Entries),
		Methods:       make(map[string]int),
		StatusCodes:   make(map[int]int),
		ContentTypes:  make(map[string]int),
		Domains:       make(map[string]int),
	}

	var totalTime float64
	var times []float64
	var startTime, endTime time.Time
	var validTimings int
	var sumBlocked, sumDNS, sumConnect, sumSend, sumWait, sumReceive, sumSSL float64
	var countBlocked, countDNS, countConnect, countSend, countWait, countReceive, countSSL int

	// 初始化最小值为一个很大的数
	minTime := float64(1<<62)
	var minBlocked, minDNS, minConnect, minSend, minWait, minReceive, minSSL float64 = 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62, 1 << 62

	for i, entry := range h.Log.Entries {
		// 累计总时间
		totalTime += entry.Time
		times = append(times, entry.Time)

		// 最小/最大请求时间
		if entry.Time > stats.MaxTime {
			stats.MaxTime = entry.Time
		}
		if entry.Time < minTime {
			minTime = entry.Time
		}

		// 传输字节数
		if entry.Response.BodySize > 0 {
			stats.TotalTransferred += int64(entry.Response.BodySize)
		}
		if entry.Response.Content.Size > 0 {
			stats.TotalUncompressed += int64(entry.Response.Content.Size)
		}

		// HTTP方法分布
		stats.Methods[entry.Request.Method]++

		// 状态码分布
		stats.StatusCodes[entry.Response.Status]++

		// 错误计数
		if entry.Response.Status >= 400 {
			stats.ErrorCount++
		}
		if entry.Response.Status >= 300 && entry.Response.Status < 400 {
			stats.RedirectCount++
		}

		// 内容类型分布
		contentType := entry.Response.Content.MimeType
		if contentType != "" {
			// 去掉参数部分（如 charset=utf-8）
			if idx := strings.Index(contentType, ";"); idx != -1 {
				contentType = strings.TrimSpace(contentType[:idx])
			}
			stats.ContentTypes[contentType]++
		}

		// 域名分布
		if domain := extractDomain(entry.Request.URL); domain != "" {
			stats.Domains[domain]++
		}

		// 时间范围
		if i == 0 || entry.StartedDateTime.Before(startTime) {
			startTime = entry.StartedDateTime
		}
		if endTime.IsZero() || entry.StartedDateTime.Add(time.Duration(entry.Time)*time.Millisecond).After(endTime) {
			endTime = entry.StartedDateTime.Add(time.Duration(entry.Time) * time.Millisecond)
		}

		// 时间指标汇总
		if entry.Timings.Blocked > 0 {
			sumBlocked += entry.Timings.Blocked
				countBlocked++
			if entry.Timings.Blocked > stats.TimingsSummary.MaxBlocked {
				stats.TimingsSummary.MaxBlocked = entry.Timings.Blocked
			}
			if entry.Timings.Blocked < minBlocked {
				minBlocked = entry.Timings.Blocked
			}
		}
		if entry.Timings.DNS > 0 {
			sumDNS += entry.Timings.DNS
				countDNS++
			if entry.Timings.DNS > stats.TimingsSummary.MaxDNS {
				stats.TimingsSummary.MaxDNS = entry.Timings.DNS
			}
			if entry.Timings.DNS < minDNS {
				minDNS = entry.Timings.DNS
			}
		}
		if entry.Timings.Connect > 0 {
			sumConnect += entry.Timings.Connect
				countConnect++
			if entry.Timings.Connect > stats.TimingsSummary.MaxConnect {
				stats.TimingsSummary.MaxConnect = entry.Timings.Connect
			}
			if entry.Timings.Connect < minConnect {
				minConnect = entry.Timings.Connect
			}
		}
		if entry.Timings.Send > 0 {
			sumSend += entry.Timings.Send
				countSend++
			if entry.Timings.Send > stats.TimingsSummary.MaxSend {
				stats.TimingsSummary.MaxSend = entry.Timings.Send
			}
			if entry.Timings.Send < minSend {
				minSend = entry.Timings.Send
			}
		}
		if entry.Timings.Wait > 0 {
			sumWait += entry.Timings.Wait
				countWait++
			if entry.Timings.Wait > stats.TimingsSummary.MaxWait {
				stats.TimingsSummary.MaxWait = entry.Timings.Wait
			}
			if entry.Timings.Wait < minWait {
				minWait = entry.Timings.Wait
			}
		}
		if entry.Timings.Receive > 0 {
			sumReceive += entry.Timings.Receive
				countReceive++
			if entry.Timings.Receive > stats.TimingsSummary.MaxReceive {
				stats.TimingsSummary.MaxReceive = entry.Timings.Receive
			}
			if entry.Timings.Receive < minReceive {
				minReceive = entry.Timings.Receive
			}
		}
		if entry.Timings.Ssl > 0 {
			sumSSL += entry.Timings.Ssl
				countSSL++
			if entry.Timings.Ssl > stats.TimingsSummary.MaxSSL {
				stats.TimingsSummary.MaxSSL = entry.Timings.Ssl
			}
			if entry.Timings.Ssl < minSSL {
				minSSL = entry.Timings.Ssl
			}
		}
		validTimings++
	}

	// 计算平均值
	stats.AvgTime = totalTime / float64(stats.TotalRequests)
	stats.MinTime = minTime

	// 计算百分位数
	stats.MedianTime = percentile(times, 50)
	stats.P95Time = percentile(times, 95)
	stats.P99Time = percentile(times, 99)

	// 总时间范围
	if !startTime.IsZero() && !endTime.IsZero() {
		stats.TotalTime = float64(endTime.Sub(startTime).Milliseconds())
		stats.StartTime = startTime
		stats.EndTime = endTime
	}

	// 计算时间指标平均值
	if validTimings > 0 {
		n := float64(validTimings)
		stats.TimingsSummary.AvgBlocked = sumBlocked / n
		stats.TimingsSummary.AvgDNS = sumDNS / n
		stats.TimingsSummary.AvgConnect = sumConnect / n
		stats.TimingsSummary.AvgSend = sumSend / n
		stats.TimingsSummary.AvgWait = sumWait / n
		stats.TimingsSummary.AvgReceive = sumReceive / n
		stats.TimingsSummary.AvgSSL = sumSSL / n
	}

	// 设置最小值（如果有有效值）
	if minBlocked < float64(1<<62) {
		stats.TimingsSummary.MinBlocked = minBlocked
	}
	if minDNS < float64(1<<62) {
		stats.TimingsSummary.MinDNS = minDNS
	}
	if minConnect < float64(1<<62) {
		stats.TimingsSummary.MinConnect = minConnect
	}
	if minSend < float64(1<<62) {
		stats.TimingsSummary.MinSend = minSend
	}
	if minWait < float64(1<<62) {
		stats.TimingsSummary.MinWait = minWait
	}
	if minReceive < float64(1<<62) {
		stats.TimingsSummary.MinReceive = minReceive
	}
	if minSSL < float64(1<<62) {
		stats.TimingsSummary.MinSSL = minSSL
	}

	return stats
}

// TimingStatistics 仅计算时间指标的统计信息
func (h *Har) TimingStatistics() *TimingsSummary {
	stats := h.Statistics()
	return &stats.TimingsSummary
}

// DomainSummary 按域名汇总统计信息
func (h *Har) DomainSummary() map[string]*DomainStats {
	result := make(map[string]*DomainStats)

	if h == nil {
		return result
	}

	for _, entry := range h.Log.Entries {
		domain := extractDomain(entry.Request.URL)
		if domain == "" {
			continue
		}

		if _, ok := result[domain]; !ok {
			result[domain] = &DomainStats{}
		}

		ds := result[domain]
		ds.RequestCount++
		ds.TotalTime += entry.Time

		if entry.Response.BodySize > 0 {
			ds.TotalTransferred += int64(entry.Response.BodySize)
		}

		if entry.Response.Status >= 400 {
			ds.ErrorCount++
		}
	}

	// 计算平均时间
	for _, ds := range result {
		if ds.RequestCount > 0 {
			ds.AvgTime = ds.TotalTime / float64(ds.RequestCount)
		}
	}

	return result
}

// StatusCodeDistribution 获取状态码分布
func (h *Har) StatusCodeDistribution() map[int]int {
	if h == nil {
		return make(map[int]int)
	}

	result := make(map[int]int)
	for _, entry := range h.Log.Entries {
		result[entry.Response.Status]++
	}
	return result
}

// MethodDistribution 获取HTTP方法分布
func (h *Har) MethodDistribution() map[string]int {
	if h == nil {
		return make(map[string]int)
	}

	result := make(map[string]int)
	for _, entry := range h.Log.Entries {
		result[entry.Request.Method]++
	}
	return result
}

// ContentTypeDistribution 获取内容类型分布
func (h *Har) ContentTypeDistribution() map[string]int {
	if h == nil {
		return make(map[string]int)
	}

	result := make(map[string]int)
	for _, entry := range h.Log.Entries {
		contentType := entry.Response.Content.MimeType
		if contentType != "" {
			if idx := strings.Index(contentType, ";"); idx != -1 {
				contentType = strings.TrimSpace(contentType[:idx])
			}
			result[contentType]++
		}
	}
	return result
}

// SlowestRequests 获取最慢的N个请求
func (h *Har) SlowestRequests(n int) []Entries {
	if h == nil || n <= 0 {
		return nil
	}

	entries := make([]Entries, len(h.Log.Entries))
	copy(entries, h.Log.Entries)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})

	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// FastestRequests 获取最快的N个请求
func (h *Har) FastestRequests(n int) []Entries {
	if h == nil || n <= 0 {
		return nil
	}

	entries := make([]Entries, len(h.Log.Entries))
	copy(entries, h.Log.Entries)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time < entries[j].Time
	})

	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// LargestResponses 获取响应体最大的N个请求
func (h *Har) LargestResponses(n int) []Entries {
	if h == nil || n <= 0 {
		return nil
	}

	entries := make([]Entries, len(h.Log.Entries))
	copy(entries, h.Log.Entries)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Response.Content.Size > entries[j].Response.Content.Size
	})

	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// Summary 获取HAR文件的文本摘要
func (h *Har) Summary() string {
	stats := h.Statistics()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("HAR 文件摘要\n"))
	sb.WriteString(fmt.Sprintf("=============\n"))
	sb.WriteString(fmt.Sprintf("版本: %s\n", h.Log.Version))
	if h.Log.Creator.Name != "" {
		sb.WriteString(fmt.Sprintf("创建者: %s %s\n", h.Log.Creator.Name, h.Log.Creator.Version))
	}
	if h.Log.Browser.Name != "" {
		sb.WriteString(fmt.Sprintf("浏览器: %s %s\n", h.Log.Browser.Name, h.Log.Browser.Version))
	}
	sb.WriteString(fmt.Sprintf("总请求数: %d\n", stats.TotalRequests))
	sb.WriteString(fmt.Sprintf("错误请求数: %d\n", stats.ErrorCount))
	sb.WriteString(fmt.Sprintf("重定向数: %d\n", stats.RedirectCount))
	sb.WriteString(fmt.Sprintf("总传输量: %s\n", formatBytes(stats.TotalTransferred)))
	sb.WriteString(fmt.Sprintf("总未压缩量: %s\n", formatBytes(stats.TotalUncompressed)))
	sb.WriteString(fmt.Sprintf("总时间: %.2f ms\n", stats.TotalTime))
	sb.WriteString(fmt.Sprintf("平均请求时间: %.2f ms\n", stats.AvgTime))
	sb.WriteString(fmt.Sprintf("中位数请求时间: %.2f ms\n", stats.MedianTime))
	sb.WriteString(fmt.Sprintf("P95请求时间: %.2f ms\n", stats.P95Time))
	sb.WriteString(fmt.Sprintf("P99请求时间: %.2f ms\n", stats.P99Time))
	sb.WriteString(fmt.Sprintf("最慢请求: %.2f ms\n", stats.MaxTime))
	sb.WriteString(fmt.Sprintf("最快请求: %.2f ms\n", stats.MinTime))

	if len(stats.Domains) > 0 {
		sb.WriteString(fmt.Sprintf("\n域名数: %d\n", len(stats.Domains)))
	}

	return sb.String()
}

// extractDomain 从URL中提取域名
func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Host
}

// percentile 计算百分位数
func percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	index := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

// formatBytes 格式化字节数为人类可读格式（内部使用FormatBytes）
func formatBytes(bytes int64) string {
	return FormatBytes(int(bytes))
}
