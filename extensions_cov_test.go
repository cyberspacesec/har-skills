package har

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

// TestCovExtractCustomFieldsNumberOverflow covers lines 89-91: the fallback
// `cf[key] = string(value)` branch when json.Unmarshal(value, &v) into
// interface{} fails. A number that overflows float64 (e.g. 1e9999) parses as a
// valid json.RawMessage at the top level (so line 76 succeeds) but fails when
// unmarshaled into an interface{} (which defaults to float64).
func TestCovExtractCustomFieldsNumberOverflow(t *testing.T) {
	data := []byte(`{"_overflow": 1e9999, "_ok": "valid"}`)
	cf := extractCustomFields(data, "Har")
	if cf == nil {
		t.Fatal("expected non-nil CustomFields")
	}
	v, ok := cf["_overflow"]
	if !ok {
		t.Fatal("_overflow should be present")
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("_overflow should be stored as string fallback, got %T (%v)", v, v)
	}
	if s != "1e9999" {
		t.Errorf("_overflow string = %q, want %q", s, "1e9999")
	}
	if cf["_ok"] != "valid" {
		t.Errorf("_ok = %v, want valid", cf["_ok"])
	}
}

// TestCovExtractCustomFieldsOverflowViaEntries ensures the overflow fallback
// also works when triggered through a real Entries UnmarshalJSON path.
func TestCovExtractCustomFieldsOverflowViaEntries(t *testing.T) {
	input := `{"startedDateTime":"2024-01-01T00:00:00Z","time":0,"request":{"method":"GET","url":"u","httpVersion":"HTTP/1.1","cookies":[],"headers":[],"queryString":[],"headersSize":-1,"bodySize":-1},"response":{"status":200,"statusText":"OK","httpVersion":"HTTP/1.1","cookies":[],"headers":[],"content":{"size":0,"mimeType":""},"redirectURL":"","headersSize":-1,"bodySize":-1},"cache":{},"timings":{"blocked":-1,"dns":-1,"connect":-1,"send":-1,"wait":-1,"receive":-1,"ssl":-1},"_big":1e9999}`
	var e Entries
	if err := json.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if v := e.GetCustomField("_big"); v != "1e9999" {
		t.Errorf("Entries._big = %v (%T), want string \"1e9999\"", v, v)
	}
}

// TestCovLogMarshalJSONErrorBranch covers Log.MarshalJSON lines 226-228 by
// making a nested Entries.MarshalJSON fail (its mergeCustomFieldsIntoJSON
// fails on an unmarshallable custom field value). That failure propagates up
// through json.Marshal(Alias(l)), triggering Log's error branch.
//
// Note: we call MarshalJSON directly because the top-level json.Marshal would
// wrap any MarshalJSON error in *json.MarshalerError; the error returned by
// Log.MarshalJSON itself is the *HarError produced at line 227.
func TestCovLogMarshalJSONErrorBranch(t *testing.T) {
	l := Log{
		Version: "1.2",
		Creator: Creator{Name: "t", Version: "1"},
		Entries: []Entries{{}},
	}
	l.Entries[0].SetCustomField("_bad", func() {})
	_, err := l.MarshalJSON()
	if err == nil {
		t.Fatal("expected Log.MarshalJSON to fail when nested Entries has bad custom field")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovEntriesMarshalJSONErrorBranch covers Entries.MarshalJSON lines
// 277-279 by making a nested Response.MarshalJSON fail.
func TestCovEntriesMarshalJSONErrorBranch(t *testing.T) {
	e := Entries{}
	e.Response.SetCustomField("_bad", func() {})
	_, err := e.MarshalJSON()
	if err == nil {
		t.Fatal("expected Entries.MarshalJSON to fail when nested Response has bad custom field")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovRequestMarshalJSONErrorBranch covers Request.MarshalJSON lines
// 328-330 by making a nested Cookie.MarshalJSON fail (Request.Cookies is a
// []Cookie, and each Cookie's MarshalJSON failing aborts the slice marshal).
func TestCovRequestMarshalJSONErrorBranch(t *testing.T) {
	r := Request{
		Method:      "GET",
		URL:         "https://example.com",
		HTTPVersion: "HTTP/1.1",
		HeadersSize: -1,
		BodySize:    -1,
		Cookies:     []Cookie{{Name: "a", Value: "b"}},
	}
	r.Cookies[0].SetCustomField("_bad", func() {})
	_, err := r.MarshalJSON()
	if err == nil {
		t.Fatal("expected Request.MarshalJSON to fail when nested Cookie has bad custom field")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovResponseMarshalJSONErrorBranch covers Response.MarshalJSON lines
// 380-382 by making the nested Content.MarshalJSON fail (Content is a value
// field of Response, so its marshal error aborts Response's alias marshal).
func TestCovResponseMarshalJSONErrorBranch(t *testing.T) {
	r := Response{
		Status:      200,
		StatusText:  "OK",
		HTTPVersion: "HTTP/1.1",
		HeadersSize: -1,
		BodySize:    -1,
	}
	r.Content.SetCustomField("_bad", func() {})
	_, err := r.MarshalJSON()
	if err == nil {
		t.Fatal("expected Response.MarshalJSON to fail when nested Content has bad custom field")
	}
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovLeafMarshalJSONMergeErrorPath exercises the mergeCustomFieldsIntoJSON
// error path (line 117) for each leaf type whose MarshalJSON alias-marshal
// branch cannot otherwise fail (Content/Cookie/Pages/Timings/Cache). These
// calls confirm that a bad custom-field value yields a typed error from each
// leaf type's MarshalJSON.
func TestCovLeafMarshalJSONMergeErrorPath(t *testing.T) {
	t.Run("Content", func(t *testing.T) {
		c := Content{Size: 1, MimeType: "x"}
		c.SetCustomField("_bad", func() {})
		_, err := c.MarshalJSON()
		assertHarErrorCode(t, err, ErrCodeJSONParse)
	})
	t.Run("Cookie", func(t *testing.T) {
		c := Cookie{Name: "a", Value: "b"}
		c.SetCustomField("_bad", func() {})
		_, err := c.MarshalJSON()
		assertHarErrorCode(t, err, ErrCodeJSONParse)
	})
	t.Run("Pages", func(t *testing.T) {
		p := Pages{ID: "p1", Title: "Test", PageTimings: PageTimings{OnContentLoad: -1, OnLoad: -1}}
		p.SetCustomField("_bad", func() {})
		_, err := p.MarshalJSON()
		assertHarErrorCode(t, err, ErrCodeJSONParse)
	})
	t.Run("Timings", func(t *testing.T) {
		tt := Timings{Blocked: 5, DNS: -1, Connect: -1, Send: 1, Wait: 50, Receive: 10, Ssl: -1}
		tt.SetCustomField("_bad", func() {})
		_, err := tt.MarshalJSON()
		assertHarErrorCode(t, err, ErrCodeJSONParse)
	})
	t.Run("Cache", func(t *testing.T) {
		c := Cache{}
		c.SetCustomField("_bad", func() {})
		_, err := c.MarshalJSON()
		assertHarErrorCode(t, err, ErrCodeJSONParse)
	})
}

// TestCovMergeCustomFieldsIntoJSONFinalMarshalError attempts to cover lines
// 123-125 (the final json.Marshal(result) error branch). The branch is only
// reachable if json.Marshal(value) succeeds but the resulting RawMessage
// cannot be re-marshaled inside the result map. This is effectively
// unreachable in practice because a successful json.Marshal(value) always
// produces valid JSON bytes. We exercise the function with a channel value
// (which fails at line 117) to document the path and confirm typed errors.
func TestCovMergeCustomFieldsIntoJSONValueMarshalError(t *testing.T) {
	cf := CustomFields{"_bad": make(chan int)}
	_, err := mergeCustomFieldsIntoJSON([]byte(`{"a":1}`), cf)
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovMergeCustomFieldsIntoJSONInvalidStdData covers the json.Unmarshal
// error branch (lines 110-111) when stdData is not valid JSON.
func TestCovMergeCustomFieldsIntoJSONInvalidStdData(t *testing.T) {
	cf := CustomFields{"_k": "v"}
	_, err := mergeCustomFieldsIntoJSON([]byte(`not json`), cf)
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// badOutOfRangeTime is a time.Time whose year (10000) is outside the RFC 3339
// range [0,9999]. In Go 1.25+, time.Time.MarshalJSON uses appendStrictRFC3339
// which returns an error for such values, so any struct field of type
// time.Time holding this value causes json.Marshal to fail.
var badOutOfRangeTime = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)

// TestCovCookieMarshalJSONAliasError covers Cookie.MarshalJSON lines 482-484 by
// setting Cookie.Expires to an out-of-range time.Time, which makes
// json.Marshal(Alias(c)) fail (time.Time.MarshalJSON errors).
func TestCovCookieMarshalJSONAliasError(t *testing.T) {
	c := Cookie{Name: "a", Value: "b", Expires: badOutOfRangeTime}
	_, err := c.MarshalJSON()
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovPagesMarshalJSONAliasError covers Pages.MarshalJSON lines 533-535 by
// setting Pages.StartedDateTime to an out-of-range time.Time.
func TestCovPagesMarshalJSONAliasError(t *testing.T) {
	p := Pages{
		ID:              "p1",
		Title:           "Test",
		StartedDateTime: badOutOfRangeTime,
		PageTimings:     PageTimings{OnContentLoad: -1, OnLoad: -1},
	}
	_, err := p.MarshalJSON()
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovCacheMarshalJSONAliasError covers Cache.MarshalJSON lines 636-638 by
// setting Cache.BeforeRequest.Expires to an out-of-range time.Time. Because
// BeforeRequest has no custom MarshalJSON, json.Marshal(Alias(c)) reflects
// into its time.Time field and fails there.
func TestCovCacheMarshalJSONAliasError(t *testing.T) {
	c := Cache{
		BeforeRequest: &BeforeRequest{Expires: badOutOfRangeTime},
	}
	_, err := c.MarshalJSON()
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}

// TestCovTimingsMarshalJSONAliasError covers Timings.MarshalJSON lines
// 585-587 by setting a float64 field to NaN. json.Marshal fails on NaN
// ("unsupported value: NaN"), which aborts json.Marshal(Alias(tt)).
func TestCovTimingsMarshalJSONAliasError(t *testing.T) {
	tt := Timings{Blocked: math.NaN(), DNS: -1, Connect: -1, Send: 1, Wait: 50, Receive: 10, Ssl: -1}
	_, err := tt.MarshalJSON()
	assertHarErrorCode(t, err, ErrCodeJSONParse)
}
