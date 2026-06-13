package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	har "github.com/cyberspacesec/har-skills"
	"github.com/cyberspacesec/har-skills/cmd/har/internal"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Build and query entry index",
	Long: `Build an in-memory index of HAR entries for fast lookup.
Supports querying by URL, method, status code, domain, MIME type, URL pattern, and time range.
Also shows index statistics.`,
	Example: `  har -f capture.har index --stats
  har -f capture.har index --url "https://api.example.com/users"
  har -f capture.har index --method POST
  har -f capture.har index --status 200
  har -f capture.har index --domain "api.example.com"
  har -f capture.har index --mime "application/json"
  har -f capture.har index --pattern "api/v[0-9]+"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		h := internal.LoadHar(cmd, args)
		idx := h.BuildIndex()

		showStats, _ := cmd.Flags().GetBool("stats")
		urlQuery, _ := cmd.Flags().GetString("url")
		methodQuery, _ := cmd.Flags().GetString("method")
		statusQuery, _ := cmd.Flags().GetInt("status")
		domainQuery, _ := cmd.Flags().GetString("domain")
		mimeQuery, _ := cmd.Flags().GetString("mime")
		patternQuery, _ := cmd.Flags().GetString("pattern")

		// If no specific query, show stats
		if !showStats && urlQuery == "" && methodQuery == "" &&
			statusQuery == 0 && domainQuery == "" && mimeQuery == "" && patternQuery == "" {
			showStats = true
		}

		// Show stats
		if showStats {
			stats := idx.Stats()
			return internal.WriteOutput(cmd, stats, func() string {
				var sb strings.Builder
				sb.WriteString("Index Statistics\n")
				sb.WriteString("================\n")
				sb.WriteString(fmt.Sprintf("Total entries:  %d\n", idx.Size()))
				sb.WriteString(fmt.Sprintf("Unique URLs:    %d\n", stats.UniqueURLs))
				sb.WriteString(fmt.Sprintf("Unique domains: %d\n", stats.UniqueDomains))
				sort.Ints(stats.StatusCodes)
				sb.WriteString(fmt.Sprintf("Status codes:   %v\n", stats.StatusCodes))
				sort.Strings(stats.Methods)
				sb.WriteString(fmt.Sprintf("Methods:        %v\n", stats.Methods))
				return sb.String()
			}, nil)
		}

		// Query by URL
		if urlQuery != "" {
			entries := idx.ByURL(urlQuery)
			return outputIndexResult(cmd, entries)
		}

		// Query by method
		if methodQuery != "" {
			entries := idx.ByMethod(methodQuery)
			return outputIndexResult(cmd, entries)
		}

		// Query by status
		if statusQuery > 0 {
			entries := idx.ByStatus(statusQuery)
			return outputIndexResult(cmd, entries)
		}

		// Query by domain
		if domainQuery != "" {
			entries := idx.ByDomain(domainQuery)
			return outputIndexResult(cmd, entries)
		}

		// Query by MIME type
		if mimeQuery != "" {
			entries := idx.ByMimeType(mimeQuery)
			return outputIndexResult(cmd, entries)
		}

		// Query by URL pattern (regex)
		if patternQuery != "" {
			entries := idx.ByURLPattern(patternQuery)
			return outputIndexResult(cmd, entries)
		}

		return nil
	},
}

func outputIndexResult(cmd *cobra.Command, entries []*har.Entries) error {
	// Convert []*Entries to []Entries for FilterResult
	entriesSlice := make([]har.Entries, len(entries))
	for i, e := range entries {
		entriesSlice[i] = *e
	}
	result := &har.FilterResult{Entries: entriesSlice}
	return internal.WriteOutput(cmd, buildListJSON(result), func() string {
		return formatFindTable(result)
	}, nil)
}

func init() {
	rootCmd.AddCommand(indexCmd)

	indexCmd.Flags().Bool("stats", false, "Show index statistics")
	indexCmd.Flags().String("url", "", "Look up entries by exact URL")
	indexCmd.Flags().String("method", "", "Look up entries by HTTP method")
	indexCmd.Flags().Int("status", 0, "Look up entries by status code")
	indexCmd.Flags().String("domain", "", "Look up entries by domain")
	indexCmd.Flags().String("mime", "", "Look up entries by MIME type")
	indexCmd.Flags().String("pattern", "", "Look up entries by URL pattern (regex)")
}
