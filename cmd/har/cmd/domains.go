package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Show per-domain statistics",
	Long: `Show detailed statistics broken down by domain: request count,
total time, average time, transfer size, and error count.`,
	Example: `  har -f capture.har domains
  har -f capture.har domains --sort time
  har -f capture.har domains --sort size --limit 10
  har -f capture.har domains --format json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)
		summary := h.DomainSummary()

		sortBy, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")

		// Sort domains
		type domainRow struct {
			domain string
			stats  *har.DomainStats
		}
		rows := make([]domainRow, 0, len(summary))
		for d, s := range summary {
			rows = append(rows, domainRow{d, s})
		}

		switch sortBy {
		case "time":
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].stats.TotalTime > rows[j].stats.TotalTime
			})
		case "size":
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].stats.TotalTransferred > rows[j].stats.TotalTransferred
			})
		case "errors":
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].stats.ErrorCount > rows[j].stats.ErrorCount
			})
		default: // "count"
			sort.Slice(rows, func(i, j int) bool {
				return rows[i].stats.RequestCount > rows[j].stats.RequestCount
			})
		}

		if limit > 0 && limit < len(rows) {
			rows = rows[:limit]
		}

		return internal.WriteOutput(cmd, summary, func() string {
			var sb tabWriterBuf
			w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "DOMAIN\tREQUESTS\tAVG TIME\tTOTAL TIME\tSIZE\tERRORS\n")
			for _, r := range rows {
				fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%d\n",
					r.domain,
					r.stats.RequestCount,
					internal.FormatDuration(r.stats.AvgTime),
					internal.FormatDuration(r.stats.TotalTime),
					internal.FormatBytes(int(r.stats.TotalTransferred)),
					r.stats.ErrorCount,
				)
			}
			w.Flush()
			return sb.String()
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(domainsCmd)

	domainsCmd.Flags().String("sort", "count", "Sort by: count, time, size, errors")
	domainsCmd.Flags().IntP("limit", "n", 0, "Limit to top N domains (0=all)")
}
