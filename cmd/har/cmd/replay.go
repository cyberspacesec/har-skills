package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// replayCmd 重放HAR文件中的HTTP请求
var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "重放HAR文件中的HTTP请求",
	Long: `重新执行HAR文件中记录的HTTP请求，并显示响应结果。

支持设置超时、重定向、SSL验证等选项，也支持
仅预览而不实际执行的干跑模式。

示例:
  har -f capture.har replay
  har -f capture.har replay --dry-run
  har -f capture.har replay --timeout 10s --skip-ssl
  har -f capture.har replay --index 0
  har -f capture.har replay --filter "api/users" --format json`,
	RunE: runReplay,
}

func init() {
	rootCmd.AddCommand(replayCmd)

	replayCmd.Flags().Duration("timeout", 30*time.Second, "请求超时时间")
	replayCmd.Flags().Bool("no-follow-redirects", false, "不跟随重定向")
	replayCmd.Flags().Int("max-redirects", 10, "最大重定向次数")
	replayCmd.Flags().Bool("skip-ssl", false, "跳过SSL证书验证")
	replayCmd.Flags().StringSlice("header", nil, "覆盖请求头 (格式: name:value)")
	replayCmd.Flags().Int("index", -1, "仅重放指定索引的条目")
	replayCmd.Flags().String("filter", "", "URL过滤模式 (仅重放匹配的条目)")
	replayCmd.Flags().Bool("dry-run", false, "仅预览将重放的请求，不实际执行")
}

func runReplay(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	// 解析选项
	timeout, _ := cmd.Flags().GetDuration("timeout")
	noFollowRedirects, _ := cmd.Flags().GetBool("no-follow-redirects")
	maxRedirects, _ := cmd.Flags().GetInt("max-redirects")
	skipSSL, _ := cmd.Flags().GetBool("skip-ssl")
	overrideHeadersSlice, _ := cmd.Flags().GetStringSlice("header")
	idx, _ := cmd.Flags().GetInt("index")
	filterPattern, _ := cmd.Flags().GetString("filter")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// 解析覆盖请求头
	overrideHeaders := make(map[string]string)
	for _, h := range overrideHeadersSlice {
		name, value, err := parseColonKeyValue(h)
		if err != nil {
			return fmt.Errorf("无效的header参数 '%s': %w", h, err)
		}
		overrideHeaders[name] = value
	}

	// 构建重放选项
	opts := har.ReplayOptions{
		Timeout:         timeout,
		FollowRedirects: !noFollowRedirects,
		MaxRedirects:    maxRedirects,
		SkipSSLVerify:   skipSSL,
		OverrideHeaders: overrideHeaders,
	}

	// 按索引或过滤筛选条目
	entries := selectReplayEntries(h, idx, filterPattern)

	// 干跑模式：仅显示将重放的请求
	if dryRun {
		return internal.WriteOutput(cmd, formatDryRunEntries(entries), func() string {
			return formatDryRunText(entries)
		}, nil)
	}

	// 实际执行重放
	results, err := replayEntries(entries, opts)
	if err != nil {
		return fmt.Errorf("重放请求失败: %w", err)
	}

	return internal.WriteOutput(cmd, results, func() string {
		return formatReplayResults(results)
	}, nil)
}

// replayEntryInfo 用于干跑模式和结果展示的条目信息
type replayEntryInfo struct {
	Index   int    `json:"index"`
	Method  string `json:"method"`
	URL     string `json:"url"`
	Headers int    `json:"headers"`
	HasBody bool   `json:"hasBody"`
}

// replayResultInfo 用于JSON输出的重放结果
type replayResultInfo struct {
	Index      int    `json:"index"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	StatusCode int    `json:"statusCode,omitempty"`
	Status     string `json:"status,omitempty"`
	Duration   string `json:"duration"`
	Error      string `json:"error,omitempty"`
}

// selectReplayEntries 根据索引或过滤模式选择要重放的条目
func selectReplayEntries(h *har.Har, idx int, filter string) []har.Entries {
	if idx >= 0 {
		if idx >= len(h.Log.Entries) {
			return nil
		}
		return h.Log.Entries[idx : idx+1]
	}

	if filter != "" {
		var selected []har.Entries
		for i, entry := range h.Log.Entries {
			if strings.Contains(entry.Request.URL, filter) {
				_ = i
				selected = append(selected, entry)
			}
		}
		return selected
	}

	return h.Log.Entries
}

// formatDryRunEntries 生成干跑模式条目信息
func formatDryRunEntries(entries []har.Entries) []replayEntryInfo {
	var infos []replayEntryInfo
	for i, entry := range entries {
		infos = append(infos, replayEntryInfo{
			Index:   i,
			Method:  entry.Request.Method,
			URL:     entry.Request.URL,
			Headers: len(entry.Request.Headers),
			HasBody: entry.Request.PostData != nil && entry.Request.PostData.Text != "",
		})
	}
	return infos
}

// formatDryRunText 格式化干跑模式为文本
func formatDryRunText(entries []har.Entries) string {
	var sb strings.Builder

	sb.WriteString("干跑模式 - 将重放以下请求:\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf("#%d %s %s\n", i, entry.Request.Method, entry.Request.URL))
		sb.WriteString(fmt.Sprintf("   请求头: %d个", len(entry.Request.Headers)))
		if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
			sb.WriteString(", 有请求体")
		}
		sb.WriteString("\n\n")
	}

	sb.WriteString(fmt.Sprintf("共 %d 个请求待重放\n", len(entries)))
	return sb.String()
}

// replayEntries 执行重放
func replayEntries(entries []har.Entries, opts har.ReplayOptions) ([]replayResultInfo, error) {
	results := make([]replayResultInfo, len(entries))
	var firstErr error

	for i, entry := range entries {
		result, err := entry.Replay(opts)
		info := replayResultInfo{
			Index:   i,
			Method:  entry.Request.Method,
			URL:     entry.Request.URL,
			Duration: result.Duration.String(),
		}

		if err != nil {
			info.Error = err.Error()
			if firstErr == nil {
				firstErr = err
			}
		} else if result.Response != nil {
			info.StatusCode = result.Response.StatusCode
			info.Status = result.Response.Status
			// 读取并丢弃响应体以释放连接
			_, _ = io.Copy(io.Discard, result.Response.Body)
			result.Response.Body.Close()
		}

		results[i] = info
	}

	return results, firstErr
}

// formatReplayResults 格式化重放结果为文本
func formatReplayResults(results []replayResultInfo) string {
	var sb strings.Builder

	sb.WriteString("重放结果\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	successCount := 0
	failCount := 0

	for _, r := range results {
		status := "OK"
		if r.Error != "" {
			status = "FAIL"
			failCount++
		} else {
			successCount++
		}

		sb.WriteString(fmt.Sprintf("#%d %s %s\n", r.Index, r.Method, r.URL))

		if r.StatusCode > 0 {
			sb.WriteString(fmt.Sprintf("   状态: %d %s  耗时: %s\n", r.StatusCode, r.Status, r.Duration))
		} else {
			sb.WriteString(fmt.Sprintf("   状态: %s  耗时: %s\n", status, r.Duration))
		}

		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("   错误: %s\n", r.Error))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("总计: %d 成功, %d 失败\n", successCount, failCount))
	return sb.String()
}
