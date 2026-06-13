package har

import (
	"net/url"
	"strings"
	"time"
)

// CookieFinding represents a single cookie security/privacy finding.
type CookieFinding struct {
	Severity    string // "high", "medium", "low", "info"
	Category    string // "expired", "session", "insecure", "samesite", "third-party", "oversized", "duplicate"
	CookieName  string
	EntryIndex  int
	EntryURL    string
	Description string
	Remedy      string
}

// CookieAuditReport contains the results of a comprehensive cookie audit.
type CookieAuditReport struct {
	Findings      []CookieFinding
	TotalCookies  int
	UniqueCookies int
	SecureCount   int
	HttpOnlyCount int
	SameSiteCount int
}

// HasHighSeverity returns true if any finding has high severity.
func (r *CookieAuditReport) HasHighSeverity() bool {
	for _, f := range r.Findings {
		if f.Severity == "high" {
			return true
		}
	}
	return false
}

// FindByCategory returns all findings matching the given category.
func (r *CookieAuditReport) FindByCategory(cat string) []CookieFinding {
	var result []CookieFinding
	for _, f := range r.Findings {
		if f.Category == cat {
			result = append(result, f)
		}
	}
	return result
}

// CookieEvolutionEntry tracks a single observation of a cookie's value/attributes.
type CookieEvolutionEntry struct {
	EntryIndex int
	Value      string
	Secure     bool
	HttpOnly   bool
	SameSite   string
	Domain     string
	Path       string
}

// CookieAudit performs a comprehensive cookie security analysis on all entries.
func (h *Har) CookieAudit() *CookieAuditReport {
	if h == nil || len(h.Log.Entries) == 0 {
		return &CookieAuditReport{}
	}

	report := &CookieAuditReport{}
	seen := make(map[string]bool) // track unique cookie names
	cookieEntries := make(map[string][]int) // cookie name -> list of entry indices where it appears

	now := time.Now()

	for i, entry := range h.Log.Entries {
		allCookies := append(append([]Cookie{}, entry.Request.Cookies...), entry.Response.Cookies...)

		for _, c := range allCookies {
			report.TotalCookies++
			seen[c.Name] = true

			// Track duplicates
			cookieEntries[c.Name] = append(cookieEntries[c.Name], i)

			// 1. Expired cookies
			if !c.Expires.IsZero() && c.Expires.Before(now) {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "medium",
					Category:    "expired",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie is expired (expires: " + c.Expires.Format(time.RFC3339) + ")",
					Remedy:      "Remove expired cookies or update expiration",
				})
			}

			// 2. Session cookies (no Expires and no Max-Age) — INFO
			if c.Expires.IsZero() {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "info",
					Category:    "session",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie is a session cookie (no Expires set)",
					Remedy:      "Consider setting an explicit expiration if persistence is needed",
				})
			}

			// 3. Cookies without Secure flag on HTTPS pages
			if !c.Secure && isHTTPS(entry.Request.URL) {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "high",
					Category:    "insecure",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie lacks Secure flag on HTTPS page",
					Remedy:      "Set the Secure flag to prevent transmission over HTTP",
				})
			}

			// 4. Cookies without HttpOnly flag for session-like names
			if !c.HTTPOnly && isSessionLikeName(c.Name) {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "medium",
					Category:    "insecure",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Session-like cookie lacks HttpOnly flag",
					Remedy:      "Set HttpOnly flag to prevent JavaScript access",
				})
			}

			// 5. SameSite=None without Secure
			if strings.EqualFold(c.SameSite, "None") && !c.Secure {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "high",
					Category:    "samesite",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie has SameSite=None without Secure flag",
					Remedy:      "Set Secure flag when using SameSite=None",
				})
			}

			// 6. SameSite attribute missing
			if c.SameSite == "" {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "low",
					Category:    "samesite",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie is missing SameSite attribute",
					Remedy:      "Set SameSite to Lax or Strict to prevent CSRF",
				})
			}

			// 7. Third-party cookies
			if c.Domain != "" && isThirdParty(c.Domain, entry.Request.URL) {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "info",
					Category:    "third-party",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie domain differs from page domain (third-party cookie)",
					Remedy:      "Review third-party cookie usage for privacy compliance",
				})
			}

			// 8. Oversized cookies (> 4KB)
			if len(c.Value) > 4096 {
				report.Findings = append(report.Findings, CookieFinding{
					Severity:    "low",
					Category:    "oversized",
					CookieName:  c.Name,
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Description: "Cookie value exceeds 4KB",
					Remedy:      "Reduce cookie size or use alternative storage",
				})
			}

			// Counts
			if c.Secure {
				report.SecureCount++
			}
			if c.HTTPOnly {
				report.HttpOnlyCount++
			}
			if c.SameSite != "" {
				report.SameSiteCount++
			}
		}
	}

	report.UniqueCookies = len(seen)

	// 9. Duplicate cookie names across entries
	for name, indices := range cookieEntries {
		if len(indices) > 1 {
			report.Findings = append(report.Findings, CookieFinding{
				Severity:    "info",
				Category:    "duplicate",
				CookieName:  name,
				EntryIndex:  indices[0],
				EntryURL:    h.Log.Entries[indices[0]].Request.URL,
				Description: "Cookie appears in multiple entries",
				Remedy:      "Review duplicate cookie usage across requests",
			})
		}
	}

	return report
}

// CookieEvolution tracks how the same cookie changes across entries.
func (h *Har) CookieEvolution() map[string][]CookieEvolutionEntry {
	result := make(map[string][]CookieEvolutionEntry)

	if h == nil || len(h.Log.Entries) == 0 {
		return result
	}

	for i, entry := range h.Log.Entries {
		allCookies := append(append([]Cookie{}, entry.Request.Cookies...), entry.Response.Cookies...)

		for _, c := range allCookies {
			result[c.Name] = append(result[c.Name], CookieEvolutionEntry{
				EntryIndex: i,
				Value:      c.Value,
				Secure:     c.Secure,
				HttpOnly:   c.HTTPOnly,
				SameSite:   c.SameSite,
				Domain:     c.Domain,
				Path:       c.Path,
			})
		}
	}

	return result
}

// isHTTPS checks if a URL uses HTTPS.
func isHTTPS(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "https"
}

// isThirdParty checks if a cookie domain differs from the page domain.
func isThirdParty(cookieDomain, pageURL string) bool {
	u, err := url.Parse(pageURL)
	if err != nil {
		return false
	}
	pageHost := u.Hostname()
	domain := strings.TrimPrefix(cookieDomain, ".")

	// Exact match or subdomain match means first-party
	if pageHost == domain || strings.HasSuffix(pageHost, "."+domain) {
		return false
	}
	return true
}

// isSessionLikeName checks if a cookie name looks like a session identifier.
func isSessionLikeName(name string) bool {
	lower := strings.ToLower(name)
	sessionPatterns := []string{"session", "sess", "sid", "jsession", "phpsessid", "asp.net_sessionid", "token", "auth"}
	for _, pattern := range sessionPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
