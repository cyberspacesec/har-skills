package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

// cookieCmd 分析HAR文件中的Cookie
var cookieCmd = &cobra.Command{
	Use:   "cookie",
	Short: "分析HAR文件中的Cookie",
	Long: `对HAR文件中的Cookie进行安全审计和演变分析。

安全审计检查Cookie的Secure、HttpOnly、SameSite等安全属性，
演变分析追踪Cookie在不同请求中的变化情况。

示例:
  har -f capture.har cookie
  har -f capture.har cookie --audit=false --evolution
  har -f capture.har cookie --name "session_id"
  har -f capture.har cookie --severity medium --format json`,
	RunE: runCookie,
}

func init() {
	rootCmd.AddCommand(cookieCmd)

	cookieCmd.Flags().Bool("audit", true, "执行Cookie安全审计")
	cookieCmd.Flags().Bool("evolution", false, "显示Cookie演变时间线")
	cookieCmd.Flags().String("name", "", "仅显示指定名称的Cookie")
	cookieCmd.Flags().String("severity", "info", "最低严重性过滤 (info/low/medium/high)")
}

func runCookie(cmd *cobra.Command, args []string) error {
	h := internal.LoadHar(cmd, args)

	_, _ = cmd.Flags().GetBool("audit")
	doEvolution, _ := cmd.Flags().GetBool("evolution")
	cookieName, _ := cmd.Flags().GetString("name")
	severity, _ := cmd.Flags().GetString("severity")

	// 如果没有指定演变分析，默认只做审计
	if !doEvolution {
		report := h.CookieAudit()

		// 按严重性过滤
		filtered := filterCookieFindingsBySeverity(report.Findings, severity)
		report.Findings = filtered

		// 按Cookie名称过滤
		if cookieName != "" {
			var nameFiltered []har.CookieFinding
			for _, f := range report.Findings {
				if f.CookieName == cookieName {
					nameFiltered = append(nameFiltered, f)
				}
			}
			report.Findings = nameFiltered
		}

		return internal.WriteOutput(cmd, report, func() string {
			return formatCookieAuditReport(report)
		}, nil)
	}

	// Cookie演变分析
	evolution := h.CookieEvolution()

	// 按名称过滤
	if cookieName != "" {
		if entries, ok := evolution[cookieName]; ok {
			evolution = map[string][]har.CookieEvolutionEntry{
				cookieName: entries,
			}
		} else {
			evolution = map[string][]har.CookieEvolutionEntry{}
		}
	}

	return internal.WriteOutput(cmd, evolution, func() string {
		return formatCookieEvolution(evolution)
	}, nil)
}

// filterCookieFindingsBySeverity 根据严重性过滤Cookie发现
func filterCookieFindingsBySeverity(findings []har.CookieFinding, minSeverity string) []har.CookieFinding {
	severityOrder := map[string]int{
		"info":   1,
		"low":    2,
		"medium": 3,
		"high":   4,
	}

	minLevel, ok := severityOrder[strings.ToLower(minSeverity)]
	if !ok {
		return findings
	}

	var result []har.CookieFinding
	for _, f := range findings {
		level, exists := severityOrder[f.Severity]
		if exists && level >= minLevel {
			result = append(result, f)
		}
	}
	return result
}

// formatCookieAuditReport 格式化Cookie审计报告为文本
func formatCookieAuditReport(report *har.CookieAuditReport) string {
	var sb strings.Builder

	sb.WriteString("Cookie安全审计报告\n")
	sb.WriteString("==================\n")
	sb.WriteString(fmt.Sprintf("Cookie总数: %d (唯一: %d)\n", report.TotalCookies, report.UniqueCookies))
	sb.WriteString(fmt.Sprintf("安全属性: Secure=%d, HttpOnly=%d, SameSite=%d\n\n",
		report.SecureCount, report.HttpOnlyCount, report.SameSiteCount))

	if len(report.Findings) == 0 {
		sb.WriteString("未发现Cookie安全问题。\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("发现 %d 个问题:\n", len(report.Findings)))
	sb.WriteString(strings.Repeat("-", 80) + "\n")
	sb.WriteString(fmt.Sprintf("%-8s %-15s %-20s %s\n", "严重性", "类别", "Cookie名称", "描述"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	for _, f := range report.Findings {
		sb.WriteString(fmt.Sprintf("%-8s %-15s %-20s %s\n",
			strings.ToUpper(f.Severity), f.Category, f.CookieName, f.Description))
	}

	return sb.String()
}

// formatCookieEvolution 格式化Cookie演变时间线为文本
func formatCookieEvolution(evolution map[string][]har.CookieEvolutionEntry) string {
	var sb strings.Builder

	sb.WriteString("Cookie演变时间线\n")
	sb.WriteString("================\n\n")

	if len(evolution) == 0 {
		sb.WriteString("未发现Cookie变化。\n")
		return sb.String()
	}

	for name, entries := range evolution {
		sb.WriteString(fmt.Sprintf("Cookie: %s\n", name))
		sb.WriteString(strings.Repeat("-", 40) + "\n")

		for i, e := range entries {
			sb.WriteString(fmt.Sprintf("  #%d [Entry %d]\n", i+1, e.EntryIndex))
			sb.WriteString(fmt.Sprintf("    Value:   %s\n", truncateString(e.Value, 50)))
			sb.WriteString(fmt.Sprintf("    Secure:  %v  HttpOnly: %v  SameSite: %s\n",
				e.Secure, e.HttpOnly, e.SameSite))
			sb.WriteString(fmt.Sprintf("    Domain:  %s  Path: %s\n", e.Domain, e.Path))
			if i < len(entries)-1 {
				sb.WriteString("    |\n")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// truncateString 截断过长字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
