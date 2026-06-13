package har

import (
	"strconv"
	"strings"
	"time"
)

// CacheControlDirectives represents the parsed directives from a Cache-Control header.
type CacheControlDirectives struct {
	MaxAge               *int
	SMaxAge              *int
	NoCache              bool
	NoStore              bool
	Private              bool
	Public               bool
	MustRevalidate       bool
	ProxyRevalidate      bool
	StaleWhileRevalidate *int
	StaleIfError         *int
}

// CacheEntryAssessment represents the cacheability assessment for a single entry.
type CacheEntryAssessment struct {
	EntryIndex      int
	URL             string
	Cacheable       bool
	CacheType       string // "public", "private", "no-cache", "no-store"
	MaxAge          time.Duration
	HasETag         bool
	HasLastModified bool
	HasVary         bool
	VaryHeaders     []string
	Age             time.Duration
}

// CacheReport contains the overall cache analysis results.
type CacheReport struct {
	Assessments       []CacheEntryAssessment
	CacheableCount    int
	NonCacheableCount int
	CacheEfficiency   float64 // percentage of cacheable resources
}

// FindByURL returns the cache assessment for the given URL, or nil if not found.
func (r *CacheReport) FindByURL(url string) *CacheEntryAssessment {
	for i := range r.Assessments {
		if r.Assessments[i].URL == url {
			return &r.Assessments[i]
		}
	}
	return nil
}

// NonCacheableEntries returns all non-cacheable entry assessments.
func (r *CacheReport) NonCacheableEntries() []CacheEntryAssessment {
	var result []CacheEntryAssessment
	for _, a := range r.Assessments {
		if !a.Cacheable {
			result = append(result, a)
		}
	}
	return result
}

// CacheAnalysis analyzes cache headers for all entries in the HAR.
func (h *Har) CacheAnalysis() *CacheReport {
	if h == nil || len(h.Log.Entries) == 0 {
		return &CacheReport{}
	}

	report := &CacheReport{
		Assessments: make([]CacheEntryAssessment, 0, len(h.Log.Entries)),
	}

	for i, entry := range h.Log.Entries {
		assessment := CacheEntryAssessment{
			EntryIndex: i,
			URL:        entry.Request.URL,
		}

		// Parse response headers
		var cc *CacheControlDirectives
		var pragmaNoCache bool
		var expiresHeader string

		for _, header := range entry.Response.Headers {
			name := strings.ToLower(header.Name)

			switch name {
			case "cache-control":
				cc = ParseCacheControl(header.Value)
			case "etag":
				assessment.HasETag = true
			case "last-modified":
				assessment.HasLastModified = true
			case "vary":
				assessment.HasVary = true
				assessment.VaryHeaders = parseVary(header.Value)
			case "age":
				if age, err := strconv.Atoi(strings.TrimSpace(header.Value)); err == nil && age > 0 {
					assessment.Age = time.Duration(age) * time.Second
				}
			case "expires":
				expiresHeader = strings.TrimSpace(header.Value)
			case "pragma":
				if strings.EqualFold(strings.TrimSpace(header.Value), "no-cache") {
					pragmaNoCache = true
				}
			}
		}

		// Determine cacheability
		cacheable, cacheType, maxAge := assessCacheability(cc, pragmaNoCache, expiresHeader)
		assessment.Cacheable = cacheable
		assessment.CacheType = cacheType
		assessment.MaxAge = maxAge

		if cacheable {
			report.CacheableCount++
		} else {
			report.NonCacheableCount++
		}

		report.Assessments = append(report.Assessments, assessment)
	}

	// Calculate cache efficiency
	total := len(report.Assessments)
	if total > 0 {
		report.CacheEfficiency = float64(report.CacheableCount) / float64(total) * 100
	}

	return report
}

// ParseCacheControl parses a Cache-Control header value into structured directives.
func ParseCacheControl(value string) *CacheControlDirectives {
	d := &CacheControlDirectives{}

	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		lower := strings.ToLower(part)

		if lower == "no-cache" {
			d.NoCache = true
		} else if lower == "no-store" {
			d.NoStore = true
		} else if lower == "private" {
			d.Private = true
		} else if lower == "public" {
			d.Public = true
		} else if lower == "must-revalidate" {
			d.MustRevalidate = true
		} else if lower == "proxy-revalidate" {
			d.ProxyRevalidate = true
		} else if strings.HasPrefix(lower, "max-age=") {
			if v, err := strconv.Atoi(part[len("max-age="):]); err == nil {
				d.MaxAge = intPtr(v)
			}
		} else if strings.HasPrefix(lower, "s-maxage=") {
			if v, err := strconv.Atoi(part[len("s-maxage="):]); err == nil {
				d.SMaxAge = intPtr(v)
			}
		} else if strings.HasPrefix(lower, "stale-while-revalidate=") {
			if v, err := strconv.Atoi(part[len("stale-while-revalidate="):]); err == nil {
				d.StaleWhileRevalidate = intPtr(v)
			}
		} else if strings.HasPrefix(lower, "stale-if-error=") {
			if v, err := strconv.Atoi(part[len("stale-if-error="):]); err == nil {
				d.StaleIfError = intPtr(v)
			}
		}
	}

	return d
}

// assessCacheability determines if a response is cacheable based on its headers.
func assessCacheability(cc *CacheControlDirectives, pragmaNoCache bool, expiresHeader string) (cacheable bool, cacheType string, maxAge time.Duration) {
	// Default: cacheable (HTTP allows caching by default for GET requests)
	cacheType = "public"
	cacheable = true

	if cc != nil {
		if cc.NoStore {
			return false, "no-store", 0
		}
		if cc.NoCache {
			// no-cache means the response can be stored but must be revalidated
			// We consider this as non-cacheable for simplicity
			return false, "no-cache", 0
		}
		if cc.Private {
			cacheType = "private"
		}
		if cc.Public {
			cacheType = "public"
		}

		// Determine max-age
		if cc.SMaxAge != nil {
			maxAge = time.Duration(*cc.SMaxAge) * time.Second
		} else if cc.MaxAge != nil {
			maxAge = time.Duration(*cc.MaxAge) * time.Second
		}
	}

	// Pragma: no-cache overrides
	if pragmaNoCache && cc == nil {
		return false, "no-cache", 0
	}

	// If no Cache-Control but has Expires header, it's cacheable
	if cc == nil && expiresHeader != "" {
		cacheable = true
		cacheType = "public"
	}

	return cacheable, cacheType, maxAge
}

// parseVary parses the Vary header value into a list of header names.
func parseVary(value string) []string {
	parts := strings.Split(value, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// intPtr returns a pointer to the given int.
func intPtr(v int) *int {
	return &v
}
