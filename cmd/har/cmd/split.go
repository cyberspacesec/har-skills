package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
	"github.com/spf13/cobra"
)

// splitCmd 拆分HAR文件
var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "拆分HAR文件",
	Long: `按各种条件拆分HAR文件为多个小文件。

支持的拆分方式:
  --by page     按页面引用（pageref）拆分
  --by domain   按请求域名拆分
  --by time     按时间间隔拆分（配合 --interval）
  --by size     按条目数量拆分（配合 --max-entries）
  --by status   按状态码范围拆分（2xx/3xx/4xx/5xx）
  --by method   按HTTP方法拆分

示例:
  har split -f capture.har --by domain                     # 按域名拆分
  har split -f capture.har --by time --interval 30m        # 每30分钟拆分
  har split -f capture.har --by size --max-entries 50      # 每50条拆分
  har split -f capture.har --by status -o result           # 输出前缀为result`,
	RunE: runSplit,
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().String("by", "", "拆分方式 (page/domain/time/size/status/method)")
	splitCmd.Flags().Duration("interval", 1*time.Hour, "时间间隔（配合 --by time）")
	splitCmd.Flags().Int("max-entries", 100, "每组最大条目数（配合 --by size）")
	splitCmd.Flags().StringP("output", "o", "split", "输出文件前缀")
}

// runSplit 执行拆分命令
func runSplit(cmd *cobra.Command, args []string) error {
	// 检查必需的 --by 标志
	by, _ := cmd.Flags().GetString("by")
	if by == "" {
		return fmt.Errorf("必须指定 --by 标志 (page/domain/time/size/status/method)")
	}

	// 加载HAR文件
	h := internal.LoadHar(cmd, args)

	// 获取输出前缀（优先使用本地 -o，否则使用全局 --output）
	prefix, _ := cmd.Flags().GetString("output")
	if prefix == "" {
		prefix = "split"
	}

	// 根据拆分方式执行拆分
	var fileCount int
	var err error

	switch by {
	case "page":
		fileCount, err = splitByPage(h, prefix)
	case "domain":
		fileCount, err = splitByDomain(h, prefix)
	case "time":
		interval, _ := cmd.Flags().GetDuration("interval")
		fileCount, err = splitByTime(h, prefix, interval)
	case "size":
		maxEntries, _ := cmd.Flags().GetInt("max-entries")
		fileCount, err = splitBySize(h, prefix, maxEntries)
	case "status":
		fileCount, err = splitByStatus(h, prefix)
	case "method":
		fileCount, err = splitByMethod(h, prefix)
	default:
		return fmt.Errorf("不支持的拆分方式: %s (可选: page/domain/time/size/status/method)", by)
	}

	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "已拆分为 %d 个文件（前缀: %s）\n", fileCount, prefix)
	return nil
}

// splitByPage 按页面引用拆分
func splitByPage(h *har.Har, prefix string) (int, error) {
	parts := h.SplitByPage()
	return writeSplitMap(parts, prefix, "page")
}

// splitByDomain 按域名拆分
func splitByDomain(h *har.Har, prefix string) (int, error) {
	parts := h.SplitByDomain()
	return writeSplitMap(parts, prefix, "domain")
}

// splitByTime 按时间间隔拆分
func splitByTime(h *har.Har, prefix string, interval time.Duration) (int, error) {
	parts := h.SplitByTimeRange(interval)
	return writeSplitSlice(parts, prefix, "time")
}

// splitBySize 按条目数量拆分
func splitBySize(h *har.Har, prefix string, maxEntries int) (int, error) {
	parts := h.SplitBySize(maxEntries)
	return writeSplitSlice(parts, prefix, "size")
}

// splitByStatus 按状态码范围拆分
func splitByStatus(h *har.Har, prefix string) (int, error) {
	parts := h.SplitByStatusCode()
	return writeSplitMap(parts, prefix, "status")
}

// splitByMethod 按HTTP方法拆分
func splitByMethod(h *har.Har, prefix string) (int, error) {
	parts := h.SplitByMethod()
	return writeSplitMap(parts, prefix, "method")
}

// writeSplitMap 将map形式的拆分结果写入文件
func writeSplitMap(parts map[string]*har.Har, prefix, kind string) (int, error) {
	for key, harData := range parts {
		// 清理key中的特殊字符
		safeKey := sanitizeFilename(key)
		if safeKey == "" {
			safeKey = "unnamed"
		}
		filename := fmt.Sprintf("%s_%s_%s.har", prefix, kind, safeKey)
		if err := writeHarToFile(harData, filename); err != nil {
			return 0, err
		}
	}
	return len(parts), nil
}

// writeSplitSlice 将切片形式的拆分结果写入文件
func writeSplitSlice(parts []*har.Har, prefix, kind string) (int, error) {
	for i, harData := range parts {
		filename := fmt.Sprintf("%s_%s_%03d.har", prefix, kind, i+1)
		if err := writeHarToFile(harData, filename); err != nil {
			return 0, err
		}
	}
	return len(parts), nil
}

// writeHarToFile 将HAR数据写入文件
func writeHarToFile(h *har.Har, filename string) error {
	output, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}
	output = append(output, '\n')

	// 确保目录存在
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("无法创建目录 '%s': %w", dir, err)
		}
	}

	if err := os.WriteFile(filename, output, 0644); err != nil {
		return fmt.Errorf("无法写入文件 '%s': %w", filename, err)
	}

	fmt.Fprintf(os.Stderr, "  写入: %s (%d 条目)\n", filename, len(h.Log.Entries))
	return nil
}

// sanitizeFilename 清理文件名中的特殊字符
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}
