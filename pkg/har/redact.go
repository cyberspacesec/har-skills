package har

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

// RedactURLRule defines a rule for redacting URL path segments.
type RedactURLRule struct {
	Pattern     string // regex pattern to match URL path segments
	Replacement string // replacement for matched segments
}

// RedactOptions configures how sensitive data is redacted from a HAR file.
type RedactOptions struct {
	Headers         []string       // header names to redact (case-insensitive)
	Cookies         []string       // cookie names to redact (case-insensitive)
	QueryParams     []string       // query parameter names to redact (case-insensitive)
	PostDataFields  []string       // POST form field names to redact (case-insensitive)
	Replacement     string         // replacement text (default: "[REDACTED]")
	RedactIPs       bool           // whether to anonymize IP addresses
	RedactURLs      []RedactURLRule // URL path segment redaction rules
	CustomRedactor  func(fieldType string, name string, value string) string // custom redaction function
}

// DefaultRedactOptions returns a RedactOptions with sensible defaults for
// common sensitive fields found in HTTP traffic.
func DefaultRedactOptions() RedactOptions {
	return RedactOptions{
		Headers: []string{
			"Authorization",
			"Proxy-Authorization",
			"WWW-Authenticate",
			"Cookie",
			"Set-Cookie",
			"X-Api-Key",
			"X-Auth-Token",
			"X-CSRF-Token",
		},
		Cookies: []string{
			"session",
			"token",
			"auth",
			"password",
			"secret",
			"api_key",
			"access_token",
			"refresh_token",
		},
		QueryParams: []string{
			"password",
			"token",
			"api_key",
			"secret",
			"access_token",
			"refresh_token",
			"private_key",
			"client_secret",
		},
		PostDataFields: []string{
			"password",
			"token",
			"api_key",
			"secret",
			"access_token",
			"refresh_token",
			"private_key",
			"client_secret",
		},
		Replacement: "[REDACTED]",
		RedactIPs:   false,
	}
}

// Redact returns a new Har with sensitive data redacted.
// It deep-clones the Har first, then redacts the clone so the original is unchanged.
func (h *Har) Redact(opts RedactOptions) *Har {
	clone := h.Clone()
	clone.RedactInPlace(opts)
	return clone
}

// RedactInPlace mutates the Har in place, redacting sensitive data without cloning.
func (h *Har) RedactInPlace(opts RedactOptions) {
	replacement := opts.Replacement
	if replacement == "" {
		replacement = "[REDACTED]"
	}

	for i := range h.Log.Entries {
		entry := &h.Log.Entries[i]

		// Redact request headers
		for j := range entry.Request.Headers {
			header := &entry.Request.Headers[j]
			if matchesAny(header.Name, opts.Headers) {
				header.Value = redactHeaderValue(header.Name, header.Value, opts, replacement)
			}
		}

		// Redact response headers
		for j := range entry.Response.Headers {
			header := &entry.Response.Headers[j]
			if matchesAny(header.Name, opts.Headers) {
				header.Value = redactHeaderValue(header.Name, header.Value, opts, replacement)
			}
		}

		// Redact request cookies
		for j := range entry.Request.Cookies {
			cookie := &entry.Request.Cookies[j]
			if matchesAny(cookie.Name, opts.Cookies) {
				cookie.Value = redactCookieValue(cookie.Name, cookie.Value, opts, replacement)
			}
		}

		// Redact response cookies
		for j := range entry.Response.Cookies {
			cookie := &entry.Response.Cookies[j]
			if matchesAny(cookie.Name, opts.Cookies) {
				cookie.Value = redactCookieValue(cookie.Name, cookie.Value, opts, replacement)
			}
		}

		// Redact query string parameters
		for j := range entry.Request.QueryString {
			qs := &entry.Request.QueryString[j]
			if matchesAny(qs.Name, opts.QueryParams) {
				qs.Value = redactQueryParamValue(qs.Name, qs.Value, opts, replacement)
			}
		}

		// Redact URL (query params in URL string and path segment rules)
		if entry.Request.URL != "" {
			entry.Request.URL = redactURLString(entry.Request.URL, opts, replacement)
		}

		// Redact POST data
		if entry.Request.PostData != nil {
			pd := entry.Request.PostData
			// Redact POST params
			for j := range pd.Params {
				param := &pd.Params[j]
				if matchesAny(param.Name, opts.PostDataFields) {
					param.Value = redactPostDataFieldValue(param.Name, param.Value, opts, replacement)
				}
			}
			// Redact POST text body (key=value patterns)
			if pd.Text != "" {
				pd.Text = redactPostDataText(pd.Text, opts, replacement)
			}
		}

		// Anonymize IP addresses
		if opts.RedactIPs && entry.ServerIPAddress != "" {
			entry.ServerIPAddress = anonymizeIP(entry.ServerIPAddress)
		}
	}
}

// matchesAny checks whether name matches any of the patterns (case-insensitive).
func matchesAny(name string, patterns []string) bool {
	nameLower := strings.ToLower(name)
	for _, p := range patterns {
		if strings.ToLower(p) == nameLower {
			return true
		}
	}
	return false
}

// redactHeaderValue redacts a header value.
func redactHeaderValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("header", name, value)
	}
	return replacement
}

// redactCookieValue redacts a cookie value.
func redactCookieValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("cookie", name, value)
	}
	return replacement
}

// redactQueryParamValue redacts a query parameter value.
func redactQueryParamValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("queryparam", name, value)
	}
	return replacement
}

// redactPostDataFieldValue redacts a POST form field value.
func redactPostDataFieldValue(name string, value string, opts RedactOptions, replacement string) string {
	if opts.CustomRedactor != nil {
		return opts.CustomRedactor("postdatafield", name, value)
	}
	return replacement
}

// redactPostDataText redacts sensitive key=value patterns in POST body text.
func redactPostDataText(text string, opts RedactOptions, replacement string) string {
	// Handle URL-encoded form bodies (key=value&key=value)
	if strings.Contains(text, "=") {
		return redactKeyValuePairs(text, opts, replacement)
	}
	// Handle JSON-like bodies with key:value patterns
	if strings.Contains(text, "{") && strings.Contains(text, "}") {
		return redactJSONKeys(text, opts, replacement)
	}
	return text
}

// redactKeyValuePairs redacts sensitive fields in URL-encoded body text.
// Handles both key=value& and key=value at end of string.
func redactKeyValuePairs(text string, opts RedactOptions, replacement string) string {
	// Match key=value patterns (URL-encoded or plain)
	re := regexp.MustCompile(`([^&=]+)=([^&]*)`)
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			key := parts[1]
			value := parts[2]
			if matchesAny(key, opts.PostDataFields) {
				if opts.CustomRedactor != nil {
					return key + "=" + opts.CustomRedactor("postdatafield", key, value)
				}
				return key + "=" + replacement
			}
		}
		return match
	})
	return result
}

// redactJSONKeys redacts sensitive fields in JSON body text.
func redactJSONKeys(text string, opts RedactOptions, replacement string) string {
	// Match "key": "value" patterns in JSON
	re := regexp.MustCompile(`"([^"]+)":\s*"([^"]*)"`)
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			key := parts[1]
			value := parts[2]
			if matchesAny(key, opts.PostDataFields) {
				if opts.CustomRedactor != nil {
					return `"` + key + `": "` + opts.CustomRedactor("postdatafield", key, value) + `"`
				}
				return `"` + key + `": "` + replacement + `"`
			}
		}
		return match
	})
	return result
}

// anonymizeIP replaces the last octet of an IPv4 address with .0.
// For IPv6, it replaces the last segment with :0.
// If the string is not a valid IP, it returns the replacement text.
func anonymizeIP(ip string) string {
	// Try IPv4 first
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// Not a valid IP, just return the original
		return ip
	}

	if parsedIP.To4() != nil {
		// IPv4: replace last octet with 0
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return parts[0] + "." + parts[1] + "." + parts[2] + ".0"
		}
	}

	// IPv6: replace last hextet with :0
	str := parsedIP.String()
	lastColon := strings.LastIndex(str, ":")
	if lastColon >= 0 {
		return str[:lastColon] + ":0"
	}

	return ip
}

// redactURLString redacts sensitive query parameters in a URL string
// and applies URL path segment redaction rules.
func redactURLString(rawURL string, opts RedactOptions, replacement string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// If URL can't be parsed, try simple string-based redaction
		return redactQueryStringSimple(rawURL, opts.QueryParams, replacement, opts)
	}

	// Redact query parameters in the URL
	if parsed.RawQuery != "" {
		parsed.RawQuery = redactURLQuery(parsed.RawQuery, opts, replacement)
	}

	// Apply URL path segment redaction rules
	if len(opts.RedactURLs) > 0 {
		parsed.Path = redactURLPath(parsed.Path, opts.RedactURLs)
	}

	return parsed.String()
}

// redactURLQuery redacts sensitive query parameters in a URL query string.
func redactURLQuery(query string, opts RedactOptions, replacement string) string {
	params := strings.Split(query, "&")
	var result []string
	for _, param := range params {
		if param == "" {
			continue
		}
		parts := strings.SplitN(param, "=", 2)
		key := parts[0]
		if len(parts) == 2 {
			value := parts[1]
			if matchesAny(key, opts.QueryParams) {
				if opts.CustomRedactor != nil {
					result = append(result, key+"="+opts.CustomRedactor("queryparam", key, value))
				} else {
					result = append(result, key+"="+replacement)
				}
			} else {
				result = append(result, param)
			}
		} else {
			// No value, just a key
			if matchesAny(key, opts.QueryParams) {
				if opts.CustomRedactor != nil {
					result = append(result, opts.CustomRedactor("queryparam", key, ""))
				} else {
					result = append(result, key+"="+replacement)
				}
			} else {
				result = append(result, param)
			}
		}
	}
	return strings.Join(result, "&")
}

// redactURLPath applies URL path segment redaction rules.
func redactURLPath(path string, rules []RedactURLRule) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		for _, rule := range rules {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue
			}
			if re.MatchString(segment) {
				segments[i] = re.ReplaceAllString(segment, rule.Replacement)
			}
		}
	}
	return strings.Join(segments, "/")
}

// redactQueryStringSimple redacts query parameters in a raw URL string
// without fully parsing it. Used as a fallback when url.Parse fails.
func redactQueryStringSimple(rawURL string, paramNames []string, replacement string, opts RedactOptions) string {
	for _, name := range paramNames {
		// Match name=value pattern
		re := regexp.MustCompile(`(?i)(` + regexp.QuoteMeta(name) + `)=([^&]*)`)
		rawURL = re.ReplaceAllStringFunc(rawURL, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) >= 3 {
				if opts.CustomRedactor != nil {
					return parts[1] + "=" + opts.CustomRedactor("queryparam", parts[1], parts[2])
				}
				return parts[1] + "=" + replacement
			}
			return match
		})
	}
	return rawURL
}