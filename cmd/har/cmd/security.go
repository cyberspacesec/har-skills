package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// securityCmd 运行安全审计
var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "对HAR文件运行安全审计",
	Long: `对HAR文件中的请求和响应进行安全审计，检查以下方面：

  - 安全头部缺失 (Strict-Transport-Security, X-Content-Type-Options 等)
  - Cookie安全性 (Secure, HttpOnly, SameSite 属性)
  - 混合内容 (HTTPS页面中的HTTP资源)
  - 敏感数据泄露 (API密钥、令牌等)
  - CORS配置问题
  - 信息泄露 (Server头部、错误消息等)

示例:
  har -f capture.har security
  har -f capture.har security --severity high
  har -f capture.har security --check-cookies=false --format json`,
	RunE: runSecurity,
}

func init() {
	rootCmd.AddCommand(securityCmd)

	securityCmd.Flags().Bool("check-headers", true, "检查安全头部")
	securityCmd.Flags().Bool("check-cookies", true, "检查Cookie安全性")
	securityCmd.Flags().Bool("check-mixed-content", true, "检查混合内容")
	securityCmd.Flags().Bool("check-sensitive-data", true, "检查敏感数据泄露")
	securityCmd.Flags().Bool("check-cors", true, "检查CORS配置")
	securityCmd.Flags().Bool("check-info-disclosure", true, "检查信息泄露")
	securityCmd.Flags().String("severity", "low", "最低严重性过滤 (all/info/low/medium/high)")
}

func runSecurity(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	// 构建审计选项
	opts := har.SecurityAuditOptions{
		CheckSecurityHeaders: mustGetBool(cmd, "check-headers"),
		CheckCookies:         mustGetBool(cmd, "check-cookies"),
		CheckMixedContent:    mustGetBool(cmd, "check-mixed-content"),
		CheckSensitiveData:   mustGetBool(cmd, "check-sensitive-data"),
		CheckCORS:            mustGetBool(cmd, "check-cors"),
		CheckInfoDisclosure:  mustGetBool(cmd, "check-info-disclosure"),
	}

	report := h.SecurityAuditWithOptions(opts)

	// 根据严重性过滤
	severity, _ := cmd.Flags().GetString("severity")
	filtered := filterFindingsBySeverity(report.Findings, severity)
	report.Findings = filtered

	return internal.WriteOutput(cmd, report, func() string {
		return formatSecurityReport(report, severity)
	}, nil)
}

// filterFindingsBySeverity 根据最低严重性级别过滤发现
func filterFindingsBySeverity(findings []har.SecurityFinding, minSeverity string) []har.SecurityFinding {
	severityOrder := map[string]int{
		"all":    0,
		"info":   1,
		"low":    2,
		"medium": 3,
		"high":   4,
	}

	minLevel, ok := severityOrder[strings.ToLower(minSeverity)]
	if !ok || minSeverity == "all" {
		return findings
	}

	var result []har.SecurityFinding
	for _, f := range findings {
		level, exists := severityOrder[f.Severity]
		if exists && level >= minLevel {
			result = append(result, f)
		}
	}
	return result
}

// formatSecurityReport 格式化安全审计报告为文本
func formatSecurityReport(report *har.SecurityReport, severity string) string {
	var sb strings.Builder

	sb.WriteString("安全审计报告\n")
	sb.WriteString("============\n")
	sb.WriteString(fmt.Sprintf("评分: %d/100\n", report.Score))
	sb.WriteString(fmt.Sprintf("发现: %d 个问题\n\n", len(report.Findings)))

	if len(report.Findings) == 0 {
		sb.WriteString("未发现安全问题。\n")
		return sb.String()
	}

	// 按严重性分组
	groups := map[string][]har.SecurityFinding{
		"high":   {},
		"medium": {},
		"low":    {},
		"info":   {},
	}
	for _, f := range report.Findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}

	severityLabels := []string{"high", "medium", "low", "info"}
	severityNames := map[string]string{
		"high":   "高危",
		"medium": "中危",
		"low":    "低危",
		"info":   "信息",
	}

	for _, sev := range severityLabels {
		findings := groups[sev]
		if len(findings) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("[%s] %s (%d)\n", strings.ToUpper(sev), severityNames[sev], len(findings)))
		sb.WriteString(strings.Repeat("-", 60) + "\n")

		for i, f := range findings {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, f.Title))
			if f.EntryURL != "" {
				sb.WriteString(fmt.Sprintf("     URL: %s\n", f.EntryURL))
			}
			sb.WriteString(fmt.Sprintf("     类别: %s\n", f.Category))
			sb.WriteString(fmt.Sprintf("     描述: %s\n", f.Description))
			if f.Remedy != "" {
				sb.WriteString(fmt.Sprintf("     修复: %s\n", f.Remedy))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// mustGetBool 从命令行标志获取布尔值
func mustGetBool(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Flags().GetBool(name)
	return v
}
