package cmd

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

var connectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "Analyze connection reuse",
	Long: `Show which HTTP entries share connections (as indicated by the Connection field).
Useful for understanding connection pooling and reuse patterns.`,
	Example: `  har -f capture.har connections
  har -f capture.har connections --format json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)
		reuse := h.ConnectionReuse()

		type connRow struct {
			connection string
			indices    []int
			count      int
		}
		rows := make([]connRow, 0, len(reuse))
		for conn, indices := range reuse {
			rows = append(rows, connRow{conn, indices, len(indices)})
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].count > rows[j].count
		})

		return internal.WriteOutput(cmd, reuse, func() string {
			var sb tabWriterBuf
			w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "CONNECTION\tENTRIES\tINDICES\n")
			for _, r := range rows {
				indices := make([]string, len(r.indices))
				for i, idx := range r.indices {
					indices[i] = fmt.Sprintf("%d", idx)
				}
				fmt.Fprintf(w, "%s\t%d\t%s\n", r.connection, r.count, strings.Join(indices, ","))
			}
			w.Flush()

			if len(rows) == 0 {
				return "No connection reuse data found.\n"
			}
			return sb.String()
		}, nil)
	},
}

func init() {
	rootCmd.AddCommand(connectionsCmd)
}
