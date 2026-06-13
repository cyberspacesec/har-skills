package har

import (
	"strings"
	"time"
)

// SecurityFinding represents a single security issue discovered during audit.
type SecurityFinding struct {
	Severity    string // "high", "medium", "low", "info"
	Category    string // "headers", "cookies", "mixed-content", "sensitive-data", "cors", "info-disclosure"
	Title       string
	Description string
	EntryIndex  int    // -1 if not entry-specific
	EntryURL    string
	Remedy      string
}

// SecurityReport contains the results of a security audit.
type SecurityReport struct {
	Findings  []SecurityFinding
	Score     int // 0-100, higher is better
	CheckedAt time.Time
}

// HasHighSeverity returns true if any finding has severity "high".
func (r *SecurityReport) HasHighSeverity() bool {
	for _, f := range r.Findings {
		if f.Severity == "high" {
			return true
		}
	}
	return false
}

// HasMediumSeverity returns true if any finding has severity "medium".
func (r *SecurityReport) HasMediumSeverity() bool {
	for _, f := range r.Findings {
		if f.Severity == "medium" {
			return true
		}
	}
	return false
}

// FindByCategory returns all findings matching the given category.
func (r *SecurityReport) FindByCategory(cat string) []SecurityFinding {
	var result []SecurityFinding
	for _, f := range r.Findings {
		if f.Category == cat {
			result = append(result, f)
		}
	}
	return result
}

// FindBySeverity returns all findings matching the given severity.
func (r *SecurityReport) FindBySeverity(sev string) []SecurityFinding {
	var result []SecurityFinding
	for _, f := range r.Findings {
		if f.Severity == sev {
			result = append(result, f)
		}
	}
	return result
}

// Summary returns a human-readable summary of the security report.
func (r *SecurityReport) Summary() string {
	high := len(r.FindBySeverity("high"))
	medium := len(r.FindBySeverity("medium"))
	low := len(r.FindBySeverity("low"))
	info := len(r.FindBySeverity("info"))

	var sb strings.Builder
	sb.WriteString("Security Audit Report\n")
	sb.WriteString("=====================\n")
	sb.WriteString("Score: ")
	sb.WriteString(intToStr(r.Score))
	sb.WriteString("/100\n")
	sb.WriteString("Findings: ")
	sb.WriteString(intToStr(len(r.Findings)))
	sb.WriteString(" total (")
	sb.WriteString(intToStr(high))
	sb.WriteString(" high, ")
	sb.WriteString(intToStr(medium))
	sb.WriteString(" medium, ")
	sb.WriteString(intToStr(low))
	sb.WriteString(" low, ")
	sb.WriteString(intToStr(info))
	sb.WriteString(" info)\n")

	if high > 0 {
		sb.WriteString("\nHIGH severity findings:\n")
		for _, f := range r.FindBySeverity("high") {
			sb.WriteString("  - ")
			sb.WriteString(f.Title)
			if f.EntryURL != "" {
				sb.WriteString(" [")
				sb.WriteString(f.EntryURL)
				sb.WriteString("]")
			}
			sb.WriteString("\n")
		}
	}
	if medium > 0 {
		sb.WriteString("\nMEDIUM severity findings:\n")
		for _, f := range r.FindBySeverity("medium") {
			sb.WriteString("  - ")
			sb.WriteString(f.Title)
			if f.EntryURL != "" {
				sb.WriteString(" [")
				sb.WriteString(f.EntryURL)
				sb.WriteString("]")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// SecurityCheckFunc is a function that performs custom security checks on a HAR.
type SecurityCheckFunc func(har *Har) []SecurityFinding

// SecurityAuditOptions configures which security checks to perform.
type SecurityAuditOptions struct {
	CheckSecurityHeaders bool
	CheckCookies         bool
	CheckMixedContent    bool
	CheckSensitiveData   bool
	CheckCORS            bool
	CheckInfoDisclosure  bool
	CustomChecks         []SecurityCheckFunc
}

// DefaultSecurityAuditOptions returns options with all built-in checks enabled.
func DefaultSecurityAuditOptions() SecurityAuditOptions {
	return SecurityAuditOptions{
		CheckSecurityHeaders: true,
		CheckCookies:         true,
		CheckMixedContent:    true,
		CheckSensitiveData:   true,
		CheckCORS:            true,
		CheckInfoDisclosure:  true,
	}
}

// SecurityAudit runs a security audit on the HAR with default options.
func (h *Har) SecurityAudit() *SecurityReport {
	return h.SecurityAuditWithOptions(DefaultSecurityAuditOptions())
}

// SecurityAuditWithOptions runs a security audit on the HAR with custom options.
func (h *Har) SecurityAuditWithOptions(opts SecurityAuditOptions) *SecurityReport {
	var findings []SecurityFinding

	if opts.CheckSecurityHeaders {
		findings = append(findings, checkSecurityHeaders(h)...)
	}
	if opts.CheckCookies {
		findings = append(findings, checkCookies(h)...)
	}
	if opts.CheckMixedContent {
		findings = append(findings, checkMixedContent(h)...)
	}
	if opts.CheckSensitiveData {
		findings = append(findings, checkSensitiveData(h)...)
	}
	if opts.CheckCORS {
		findings = append(findings, checkCORS(h)...)
	}
	if opts.CheckInfoDisclosure {
		findings = append(findings, checkInfoDisclosure(h)...)
	}

	for _, customCheck := range opts.CustomChecks {
		findings = append(findings, customCheck(h)...)
	}

	score := computeScore(findings)

	return &SecurityReport{
		Findings:  findings,
		Score:     score,
		CheckedAt: time.Now(),
	}
}

// computeScore calculates a security score from 0-100 based on findings.
func computeScore(findings []SecurityFinding) int {
	score := 100
	for _, f := range findings {
		switch f.Severity {
		case "high":
			score -= 10
		case "medium":
			score -= 5
		case "low":
			score -= 2
		// "info" does not reduce score
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}

// --- Built-in check functions ---

// checkSecurityHeaders checks for missing security headers in the page response.
func checkSecurityHeaders(h *Har) []SecurityFinding {
	var findings []SecurityFinding

	if len(h.Log.Entries) == 0 {
		return findings
	}

	// Use the first entry as the page document response
	entry := h.Log.Entries[0]
	headers := headerSet(entry.Response.Headers)
	isHTTPS := strings.HasPrefix(entry.Request.URL, "https://")

	// Missing Strict-Transport-Security
	if _, ok := headers["strict-transport-security"]; !ok {
		sev := "high"
		desc := "Strict-Transport-Security header is missing"
		remedy := "Add Strict-Transport-Security header (e.g., 'max-age=31536000; includeSubDomains')"
		if !isHTTPS {
			sev = "info"
			desc = "Strict-Transport-Security header is missing (page is served over HTTP)"
			remedy = "Enable HTTPS and add Strict-Transport-Security header"
		}
		findings = append(findings, SecurityFinding{
			Severity:    sev,
			Category:    "headers",
			Title:       "Missing Strict-Transport-Security header",
			Description: desc,
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      remedy,
		})
	}

	// Missing Content-Security-Policy
	if _, ok := headers["content-security-policy"]; !ok {
		findings = append(findings, SecurityFinding{
			Severity:    "medium",
			Category:    "headers",
			Title:       "Missing Content-Security-Policy header",
			Description: "Content-Security-Policy header is missing, which may allow XSS and content injection attacks",
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      "Add Content-Security-Policy header to restrict resource loading (e.g., \"default-src 'self'\")",
		})
	}

	// Missing X-Content-Type-Options
	if _, ok := headers["x-content-type-options"]; !ok {
		findings = append(findings, SecurityFinding{
			Severity:    "low",
			Category:    "headers",
			Title:       "Missing X-Content-Type-Options header",
			Description: "X-Content-Type-Options header is missing, browsers may MIME-sniff responses",
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      "Add 'X-Content-Type-Options: nosniff' header",
		})
	}

	// Missing X-Frame-Options
	if _, ok := headers["x-frame-options"]; !ok {
		findings = append(findings, SecurityFinding{
			Severity:    "medium",
			Category:    "headers",
			Title:       "Missing X-Frame-Options header",
			Description: "X-Frame-Options header is missing, page may be vulnerable to clickjacking",
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      "Add X-Frame-Options header (e.g., 'DENY' or 'SAMEORIGIN')",
		})
	}

	// Missing X-XSS-Protection
	if _, ok := headers["x-xss-protection"]; !ok {
		findings = append(findings, SecurityFinding{
			Severity:    "low",
			Category:    "headers",
			Title:       "Missing X-XSS-Protection header",
			Description: "X-XSS-Protection header is missing (deprecated but still useful for older browsers)",
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      "Add 'X-XSS-Protection: 1; mode=block' header for legacy browser support",
		})
	}

	// Missing Referrer-Policy
	if _, ok := headers["referrer-policy"]; !ok {
		findings = append(findings, SecurityFinding{
			Severity:    "low",
			Category:    "headers",
			Title:       "Missing Referrer-Policy header",
			Description: "Referrer-Policy header is missing, referrer information may be leaked to third parties",
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      "Add Referrer-Policy header (e.g., 'strict-origin-when-cross-origin')",
		})
	}

	// Missing Permissions-Policy
	if _, ok := headers["permissions-policy"]; !ok {
		findings = append(findings, SecurityFinding{
			Severity:    "info",
			Category:    "headers",
			Title:       "Missing Permissions-Policy header",
			Description: "Permissions-Policy header is missing, browser features are not explicitly restricted",
			EntryIndex:  0,
			EntryURL:    entry.Request.URL,
			Remedy:      "Add Permissions-Policy header to restrict browser features (e.g., 'camera=(), microphone=()')",
		})
	}

	return findings
}

// checkCookies checks for insecure cookie configurations.
func checkCookies(h *Har) []SecurityFinding {
	var findings []SecurityFinding

	for i, entry := range h.Log.Entries {
		isHTTPS := strings.HasPrefix(entry.Request.URL, "https://")

		// Check request cookies
		for _, cookie := range entry.Request.Cookies {
			findings = append(findings, checkSingleCookie(cookie, i, entry.Request.URL, isHTTPS)...)
		}

		// Check response cookies
		for _, cookie := range entry.Response.Cookies {
			findings = append(findings, checkSingleCookie(cookie, i, entry.Request.URL, isHTTPS)...)
		}
	}

	return findings
}

// checkSingleCookie checks a single cookie for security issues.
func checkSingleCookie(cookie Cookie, entryIndex int, entryURL string, isHTTPS bool) []SecurityFinding {
	var findings []SecurityFinding
	nameLower := strings.ToLower(cookie.Name)

	// Cookie without Secure flag on HTTPS
	if isHTTPS && !cookie.Secure {
		findings = append(findings, SecurityFinding{
			Severity:    "high",
			Category:    "cookies",
			Title:       "Cookie without Secure flag",
			Description: "Cookie '" + cookie.Name + "' is set without the Secure flag on an HTTPS page",
			EntryIndex:  entryIndex,
			EntryURL:    entryURL,
			Remedy:      "Set the Secure flag on cookie '" + cookie.Name + "'",
		})
	}

	// Cookie without HttpOnly if name suggests session/auth
	if isSessionCookieName(nameLower) && !cookie.HTTPOnly {
		findings = append(findings, SecurityFinding{
			Severity:    "medium",
			Category:    "cookies",
			Title:       "Session/auth cookie without HttpOnly flag",
			Description: "Cookie '" + cookie.Name + "' appears to be a session/auth cookie but lacks the HttpOnly flag, making it accessible to JavaScript",
			EntryIndex:  entryIndex,
			EntryURL:    entryURL,
			Remedy:      "Set the HttpOnly flag on cookie '" + cookie.Name + "'",
		})
	}

	// SameSite=None without Secure flag
	if strings.EqualFold(cookie.SameSite, "none") && !cookie.Secure {
		findings = append(findings, SecurityFinding{
			Severity:    "high",
			Category:    "cookies",
			Title:       "SameSite=None cookie without Secure flag",
			Description: "Cookie '" + cookie.Name + "' has SameSite=None but is missing the Secure flag, which is invalid per browser requirements",
			EntryIndex:  entryIndex,
			EntryURL:    entryURL,
			Remedy:      "Set the Secure flag on cookie '" + cookie.Name + "' when using SameSite=None",
		})
	}

	// Session cookie (no Expires) - INFO
	if cookie.Expires.IsZero() {
		findings = append(findings, SecurityFinding{
			Severity:    "info",
			Category:    "cookies",
			Title:       "Session cookie without expiration",
			Description: "Cookie '" + cookie.Name + "' has no expiration set and will be deleted when the browser session ends",
			EntryIndex:  entryIndex,
			EntryURL:    entryURL,
			Remedy:      "Consider setting an explicit expiration on cookie '" + cookie.Name + "' if persistent storage is intended",
		})
	}

	return findings
}

// isSessionCookieName checks if a cookie name suggests it's a session/auth cookie.
func isSessionCookieName(nameLower string) bool {
	sessionKeywords := []string{"session", "sess", "token", "auth", "login", "sid"}
	for _, kw := range sessionKeywords {
		if strings.Contains(nameLower, kw) {
			return true
		}
	}
	return false
}

// checkMixedContent checks for HTTPS pages loading HTTP resources.
func checkMixedContent(h *Har) []SecurityFinding {
	var findings []SecurityFinding

	if len(h.Log.Entries) == 0 {
		return findings
	}

	// Check if the first entry is HTTPS
	firstURL := h.Log.Entries[0].Request.URL
	if !strings.HasPrefix(firstURL, "https://") {
		return findings
	}

	// Find any HTTP entries
	for i, entry := range h.Log.Entries {
		if strings.HasPrefix(entry.Request.URL, "http://") {
			findings = append(findings, SecurityFinding{
				Severity:    "high",
				Category:    "mixed-content",
				Title:       "Mixed content: HTTPS page loading HTTP resource",
				Description: "The page is served over HTTPS but loads a resource over HTTP: " + entry.Request.URL,
				EntryIndex:  i,
				EntryURL:    entry.Request.URL,
				Remedy:      "Change the HTTP resource URL to HTTPS or remove the resource",
			})
		}
	}

	return findings
}

// sensitiveParamNames lists query parameter names that may contain sensitive data.
var sensitiveParamNames = []string{
	"password", "secret", "token", "api_key", "apikey",
	"access_token", "private_key", "client_secret",
}

// checkSensitiveData checks for sensitive data exposed in URL query parameters.
func checkSensitiveData(h *Har) []SecurityFinding {
	var findings []SecurityFinding

	for i, entry := range h.Log.Entries {
		for _, param := range entry.Request.QueryString {
			paramLower := strings.ToLower(param.Name)
			for _, sensitive := range sensitiveParamNames {
				if paramLower == sensitive {
					findings = append(findings, SecurityFinding{
						Severity:    "high",
						Category:    "sensitive-data",
						Title:       "Sensitive data in URL query parameter",
						Description: "Query parameter '" + param.Name + "' may contain sensitive data and is exposed in the URL",
						EntryIndex:  i,
						EntryURL:    entry.Request.URL,
						Remedy:      "Move sensitive data from URL parameters to request headers or body",
					})
					break
				}
			}
		}
	}

	return findings
}

// checkCORS checks for CORS misconfiguration.
func checkCORS(h *Har) []SecurityFinding {
	var findings []SecurityFinding

	for i, entry := range h.Log.Entries {
		headers := headerSet(entry.Response.Headers)

		allowOrigin, hasOrigin := headers["access-control-allow-origin"]
		allowCreds := headers["access-control-allow-credentials"]

		// Access-Control-Allow-Origin: * combined with Access-Control-Allow-Credentials: true
		if hasOrigin && allowOrigin == "*" && strings.EqualFold(allowCreds, "true") {
			findings = append(findings, SecurityFinding{
				Severity:    "high",
				Category:    "cors",
				Title:       "CORS allows credentials with wildcard origin",
				Description: "Access-Control-Allow-Origin is '*' combined with Access-Control-Allow-Credentials: true, which is a security misconfiguration",
				EntryIndex:  i,
				EntryURL:    entry.Request.URL,
				Remedy:      "Specify explicit origins instead of '*' when using Access-Control-Allow-Credentials: true",
			})
			continue
		}

		// Overly permissive CORS on API endpoints with auth headers
		if hasOrigin && allowOrigin == "*" {
			hasAuthHeader := false
			for _, h := range entry.Request.Headers {
				hLower := strings.ToLower(h.Name)
				if hLower == "authorization" || hLower == "cookie" || strings.Contains(hLower, "auth") {
					hasAuthHeader = true
					break
				}
			}
			if hasAuthHeader {
				findings = append(findings, SecurityFinding{
					Severity:    "medium",
					Category:    "cors",
					Title:       "Overly permissive CORS with authentication",
					Description: "Access-Control-Allow-Origin is '*' on an endpoint that receives authentication headers",
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Remedy:      "Restrict Access-Control-Allow-Origin to specific trusted origins",
				})
			}
		}
	}

	return findings
}

// checkInfoDisclosure checks for information disclosure in response headers.
func checkInfoDisclosure(h *Har) []SecurityFinding {
	var findings []SecurityFinding

	for i, entry := range h.Log.Entries {
		for _, h := range entry.Response.Headers {
			nameLower := strings.ToLower(h.Name)

			switch nameLower {
			case "server":
				// Server header with version info
				if containsVersionInfo(h.Value) {
					findings = append(findings, SecurityFinding{
						Severity:    "low",
						Category:    "info-disclosure",
						Title:       "Server header reveals version information",
						Description: "Server header '" + h.Value + "' reveals server version information",
						EntryIndex:  i,
						EntryURL:    entry.Request.URL,
						Remedy:      "Remove or minimize the Server header to avoid disclosing version details",
					})
				}
			case "x-powered-by":
				findings = append(findings, SecurityFinding{
					Severity:    "low",
					Category:    "info-disclosure",
					Title:       "X-Powered-By header present",
					Description: "X-Powered-By header '" + h.Value + "' discloses technology stack information",
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Remedy:      "Remove the X-Powered-By header",
				})
			case "x-aspnet-version":
				findings = append(findings, SecurityFinding{
					Severity:    "low",
					Category:    "info-disclosure",
					Title:       "X-AspNet-Version header present",
					Description: "X-AspNet-Version header '" + h.Value + "' discloses ASP.NET version",
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Remedy:      "Remove the X-AspNet-Version header",
				})
			case "x-generator":
				findings = append(findings, SecurityFinding{
					Severity:    "info",
					Category:    "info-disclosure",
					Title:       "X-Generator header present",
					Description: "X-Generator header '" + h.Value + "' discloses generator information",
					EntryIndex:  i,
					EntryURL:    entry.Request.URL,
					Remedy:      "Remove the X-Generator header",
				})
			}
		}
	}

	return findings
}

// headerSet converts a slice of Headers to a lowercase-keyed map.
func headerSet(headers []Headers) map[string]string {
	m := make(map[string]string, len(headers))
	for _, h := range headers {
		m[strings.ToLower(h.Name)] = h.Value
	}
	return m
}

// containsVersionInfo checks if a header value contains version information.
func containsVersionInfo(value string) bool {
	// Check for common version patterns like "Apache/2.4.1", "nginx/1.18.0"
	if strings.Contains(value, "/") {
		parts := strings.SplitN(value, "/", 2)
		if len(parts) == 2 && containsDigit(parts[1]) {
			return true
		}
	}
	// Check if value itself contains version-like patterns (e.g., "Microsoft-IIS/10.0")
	for _, sep := range []string{"/", " "} {
		if strings.Contains(value, sep) && containsDigit(value) {
			return true
		}
	}
	return false
}

// containsDigit checks if a string contains any digit.
func containsDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// intToStr converts an int to string without importing strconv.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
