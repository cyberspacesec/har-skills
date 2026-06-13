package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

var contentCmd = &cobra.Command{
	Use:   "content",
	Short: "Analyze content types and sizes",
	Long: `Show content type breakdown with sizes, compression ratios,
and MIME category distribution. Useful for understanding what types of
content make up the HAR file.`,
	Example: `  har -f capture.har content
  har -f capture.har content --format json
  har -f capture.har content --by-mime`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)
		summary := h.ContentSummary()

		byMime, _ := cmd.Flags().GetBool("by-mime")

		return internal.WriteOutput(cmd, summary, func() string {
			var sb strings.Builder
			sb.WriteString("Content Summary\n")
			sb.WriteString("===============\n")
			sb.WriteString(fmt.Sprintf("Total size:      %s\n", internal.FormatBytes(summary.TotalSize)))
			sb.WriteString(fmt.Sprintf("Text size:       %s\n", internal.FormatBytes(summary.TextSize)))
			sb.WriteString(fmt.Sprintf("Binary size:     %s\n", internal.FormatBytes(summary.BinarySize)))
			if summary.CompressedSize > 0 {
				sb.WriteString(fmt.Sprintf("Compressed size: %s\n", internal.FormatBytes(summary.CompressedSize)))
			}

			sb.WriteString("\nBy Category\n-----------\n")
			cats := make([]string, 0, len(summary.ByCategory))
			for cat := range summary.ByCategory {
				cats = append(cats, string(cat))
			}
			sort.Strings(cats)
			for _, cat := range cats {
				size := summary.ByCategory[har.MIMECategory(cat)]
				pct := float64(0)
				if summary.TotalSize > 0 {
					pct = float64(size) / float64(summary.TotalSize) * 100
				}
				sb.WriteString(fmt.Sprintf("  %-15s %s (%.1f%%)\n", cat, internal.FormatBytes(size), pct))
			}

			if byMime {
				sb.WriteString("\nBy MIME Type\n------------\n")
				mimes := make([]string, 0, len(summary.ByMIMEType))
				for m := range summary.ByMIMEType {
					mimes = append(mimes, m)
				}
				sort.Slice(mimes, func(i, j int) bool {
					return summary.ByMIMEType[mimes[i]] > summary.ByMIMEType[mimes[j]]
				})
				for _, m := range mimes {
					size := summary.ByMIMEType[m]
					sb.WriteString(fmt.Sprintf("  %-40s %s\n", m, internal.FormatBytes(size)))
				}
			}

			return sb.String()
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(contentCmd)

	contentCmd.Flags().Bool("by-mime", false, "Show detailed breakdown by MIME type")
}
