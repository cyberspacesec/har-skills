package har

import (
	"fmt"
	"strings"
)

// PerformanceCategory represents a scored performance category.
type PerformanceCategory struct {
	Name     string
	Score    float64 // 0-100
	Weight   float64
	Findings []PerformanceFinding
}

// PerformanceFinding represents a single performance issue or observation.
type PerformanceFinding struct {
	Type        string // "opportunity", "diagnostic"
	Title       string
	Description string
	Impact      string // estimated savings (e.g., "500ms", "100KB")
	EntryIndex  int
	EntryURL    string
}

// PerformanceReport contains the overall performance scoring results.
type PerformanceReport struct {
	OverallScore    float64
	Categories      []PerformanceCategory
	Recommendations []string
}

// CategoryByName returns the performance category with the given name, or nil.
func (r *PerformanceReport) CategoryByName(name string) *PerformanceCategory {
	for i := range r.Categories {
		if r.Categories[i].Name == name {
			return &r.Categories[i]
		}
	}
	return nil
}

// Grade returns a letter grade for the overall score.
// A: 90+, B: 70-89, C: 50-69, D: 0-49
func (r *PerformanceReport) Grade() string {
	switch {
	case r.OverallScore >= 90:
		return "A"
	case r.OverallScore >= 70:
		return "B"
	case r.OverallScore >= 50:
		return "C"
	default:
		return "D"
	}
}

// PerformanceScore computes a comprehensive performance score for the HAR.
func (h *Har) PerformanceScore() *PerformanceReport {
	if h == nil || len(h.Log.Entries) == 0 {
		return &PerformanceReport{
			OverallScore:    100,
			Categories:      []PerformanceCategory{},
			Recommendations: []string{},
		}
	}

	report := &PerformanceReport{}

	// Get cache report for cache efficiency
	cacheReport := h.CacheAnalysis()

	// 1. Time to First Byte (weight 0.2)
	ttfbCat := h.scoreTTFB()
	ttfbCat.Weight = 0.2

	// 2. Total Load Time (weight 0.2)
	loadTimeCat := h.scoreTotalLoadTime()
	loadTimeCat.Weight = 0.2

	// 3. Request Count (weight 0.15)
	requestCountCat := h.scoreRequestCount()
	requestCountCat.Weight = 0.15

	// 4. Transfer Size (weight 0.15)
	transferSizeCat := h.scoreTransferSize()
	transferSizeCat.Weight = 0.15

	// 5. Cache Efficiency (weight 0.15)
	cacheCat := h.scoreCacheEfficiency(cacheReport)
	cacheCat.Weight = 0.15

	// 6. Compression (weight 0.15)
	compressionCat := h.scoreCompression()
	compressionCat.Weight = 0.15

	report.Categories = []PerformanceCategory{
		ttfbCat,
		loadTimeCat,
		requestCountCat,
		transferSizeCat,
		cacheCat,
		compressionCat,
	}

	// Calculate weighted overall score
	var totalWeight, weightedSum float64
	for _, cat := range report.Categories {
		weightedSum += cat.Score * cat.Weight
		totalWeight += cat.Weight
	}
	if totalWeight > 0 {
		report.OverallScore = weightedSum / totalWeight
	}

	// Generate recommendations
	report.Recommendations = h.generateRecommendations(ttfbCat, loadTimeCat, requestCountCat, transferSizeCat, cacheCat, compressionCat)

	return report
}

// scoreTTFB scores the Time to First Byte category.
func (h *Har) scoreTTFB() PerformanceCategory {
	cat := PerformanceCategory{
		Name: "Time to First Byte",
	}

	// Find the first document entry (typically the first entry)
	var ttfb float64
	if len(h.Log.Entries) > 0 {
		// TTFB is the wait time of the first entry
		ttfb = h.Log.Entries[0].Timings.Wait
		if ttfb < 0 {
			ttfb = h.Log.Entries[0].Time
		}
	}

	// Scoring: <200ms=100, <500ms=80, <1s=50, >1s=20
	switch {
	case ttfb <= 0:
		cat.Score = 100
	case ttfb < 200:
		cat.Score = 100
	case ttfb < 500:
		cat.Score = 80
	case ttfb < 1000:
		cat.Score = 50
	default:
		cat.Score = 20
	}

	if ttfb >= 1000 {
		cat.Findings = append(cat.Findings, PerformanceFinding{
			Type:        "opportunity",
			Title:       "Slow server response time",
			Description: fmt.Sprintf("TTFB is %.0fms, which is above the 1s threshold", ttfb),
			Impact:      fmt.Sprintf("%.0fms", ttfb-1000),
			EntryIndex:  0,
			EntryURL:    h.Log.Entries[0].Request.URL,
		})
	}

	return cat
}

// scoreTotalLoadTime scores the Total Load Time category.
func (h *Har) scoreTotalLoadTime() PerformanceCategory {
	cat := PerformanceCategory{
		Name: "Total Load Time",
	}

	stats := h.Statistics()
	totalTime := stats.TotalTime // ms

	// Scoring: <1s=100, <3s=80, <5s=50, >5s=20
	switch {
	case totalTime <= 0:
		cat.Score = 100
	case totalTime < 1000:
		cat.Score = 100
	case totalTime < 3000:
		cat.Score = 80
	case totalTime < 5000:
		cat.Score = 50
	default:
		cat.Score = 20
	}

	if totalTime >= 5000 {
		cat.Findings = append(cat.Findings, PerformanceFinding{
			Type:        "diagnostic",
			Title:       "High total load time",
			Description: fmt.Sprintf("Total load time is %.0fms, exceeding 5s threshold", totalTime),
			Impact:      fmt.Sprintf("%.0fms", totalTime-5000),
		})
	}

	return cat
}

// scoreRequestCount scores the Request Count category.
func (h *Har) scoreRequestCount() PerformanceCategory {
	cat := PerformanceCategory{
		Name: "Request Count",
	}

	count := len(h.Log.Entries)

	// Scoring: <10=100, <30=80, <50=50, >50=20
	switch {
	case count < 10:
		cat.Score = 100
	case count < 30:
		cat.Score = 80
	case count < 50:
		cat.Score = 50
	default:
		cat.Score = 20
	}

	if count >= 50 {
		cat.Findings = append(cat.Findings, PerformanceFinding{
			Type:        "opportunity",
			Title:       "Too many HTTP requests",
			Description: fmt.Sprintf("Found %d HTTP requests, which is above the recommended 50", count),
			Impact:      fmt.Sprintf("%d requests", count-50),
		})
	}

	return cat
}

// scoreTransferSize scores the Transfer Size category.
func (h *Har) scoreTransferSize() PerformanceCategory {
	cat := PerformanceCategory{
		Name: "Transfer Size",
	}

	stats := h.Statistics()
	totalKB := float64(stats.TotalTransferred) / 1024.0

	// Scoring: <500KB=100, <1MB=80, <3MB=50, >3MB=20
	switch {
	case totalKB < 500:
		cat.Score = 100
	case totalKB < 1024:
		cat.Score = 80
	case totalKB < 3*1024:
		cat.Score = 50
	default:
		cat.Score = 20
	}

	if totalKB >= 3*1024 {
		cat.Findings = append(cat.Findings, PerformanceFinding{
			Type:        "opportunity",
			Title:       "Large total transfer size",
			Description: fmt.Sprintf("Total transfer size is %.0fKB, exceeding 3MB threshold", totalKB),
			Impact:      fmt.Sprintf("%.0fKB", totalKB-3*1024),
		})
	}

	return cat
}

// scoreCacheEfficiency scores the Cache Efficiency category based on a CacheReport.
func (h *Har) scoreCacheEfficiency(cacheReport *CacheReport) PerformanceCategory {
	cat := PerformanceCategory{
		Name: "Cache Efficiency",
	}

	cat.Score = cacheReport.CacheEfficiency // percentage cacheable * 100 is already 0-100

	if cat.Score < 50 {
		cat.Findings = append(cat.Findings, PerformanceFinding{
			Type:        "opportunity",
			Title:       "Low cache efficiency",
			Description: fmt.Sprintf("Only %.1f%% of resources are cacheable", cat.Score),
			Impact:      fmt.Sprintf("%.1f%% improvement possible", 100-cat.Score),
		})
	}

	return cat
}

// scoreCompression scores the Compression category.
func (h *Har) scoreCompression() PerformanceCategory {
	cat := PerformanceCategory{
		Name: "Compression",
	}

	textMimeTypes := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/xml",
		"application/x-javascript",
	}

	var textCount, compressedCount int

	for i, entry := range h.Log.Entries {
		mimeType := strings.ToLower(entry.Response.Content.MimeType)
		isText := false
		for _, textMT := range textMimeTypes {
			if strings.HasPrefix(mimeType, textMT) {
				isText = true
				break
			}
		}
		if !isText {
			continue
		}

		textCount++

		// Check Content-Encoding header
		hasCompression := false
		for _, header := range entry.Response.Headers {
			if strings.EqualFold(header.Name, "Content-Encoding") {
				val := strings.ToLower(strings.TrimSpace(header.Value))
				if strings.Contains(val, "gzip") || strings.Contains(val, "br") || strings.Contains(val, "deflate") {
					hasCompression = true
				}
			}
		}

		if hasCompression {
			compressedCount++
		} else {
			cat.Findings = append(cat.Findings, PerformanceFinding{
				Type:        "opportunity",
				Title:       "Uncompressed text resource",
				Description: fmt.Sprintf("Text resource at %s is not compressed", entry.Request.URL),
				Impact:      "potential 60-80% size reduction",
				EntryIndex:  i,
				EntryURL:    entry.Request.URL,
			})
		}
	}

	// Score: all compressed=100, none=0, proportional otherwise
	if textCount == 0 {
		cat.Score = 100
	} else {
		cat.Score = float64(compressedCount) / float64(textCount) * 100
	}

	return cat
}

// generateRecommendations generates actionable recommendations based on performance findings.
func (h *Har) generateRecommendations(ttfbCat, loadTimeCat, requestCountCat, transferSizeCat, cacheCat, compressionCat PerformanceCategory) []string {
	var recs []string

	// Check compression findings
	uncompressedCount := 0
	for _, f := range compressionCat.Findings {
		if f.Title == "Uncompressed text resource" {
			uncompressedCount++
		}
	}
	if uncompressedCount > 3 {
		recs = append(recs, "Enable compression for text resources")
	}

	// Check cache efficiency
	if cacheCat.Score < 60 {
		recs = append(recs, "Add proper cache headers")
	}

	// Check request count
	if requestCountCat.Score <= 50 {
		recs = append(recs, "Reduce the number of HTTP requests")
	}

	// Check transfer size
	if transferSizeCat.Score <= 50 {
		recs = append(recs, "Optimize resource sizes")
	}

	// Check TTFB
	if ttfbCat.Score <= 50 {
		recs = append(recs, "Optimize server response time")
	}

	return recs
}
