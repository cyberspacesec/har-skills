package har

import (
	"encoding/json"
	"testing"
)

// TestCustomFieldsTypeBasics tests the CustomFields map type operations
func TestCustomFieldsTypeBasics(t *testing.T) {
	cf := make(CustomFields)

	// Test SetCustomField / GetCustomField
	cf.SetCustomField("_myField", "hello")
	if v := cf.GetCustomField("_myField"); v != "hello" {
		t.Errorf("GetCustomField = %v, want hello", v)
	}

	// Test missing field
	if v := cf.GetCustomField("_missing"); v != nil {
		t.Errorf("GetCustomField for missing key = %v, want nil", v)
	}

	// Test HasCustomField
	if !cf.HasCustomField("_myField") {
		t.Error("HasCustomField should return true for existing field")
	}
	if cf.HasCustomField("_missing") {
		t.Error("HasCustomField should return false for missing field")
	}

	// Test DeleteCustomField
	cf.DeleteCustomField("_myField")
	if cf.HasCustomField("_myField") {
		t.Error("HasCustomField should return false after DeleteCustomField")
	}

	// Test CustomFieldsKeys
	cf.SetCustomField("_a", 1)
	cf.SetCustomField("_b", 2)
	keys := cf.CustomFieldsKeys()
	if len(keys) != 2 {
		t.Errorf("CustomFieldsKeys returned %d keys, want 2", len(keys))
	}
}

// TestCustomFieldsNilOperations tests that nil CustomFields don't panic
func TestCustomFieldsNilOperations(t *testing.T) {
	var cf CustomFields

	if v := cf.GetCustomField("_any"); v != nil {
		t.Errorf("GetCustomField on nil = %v, want nil", v)
	}
	cf.SetCustomField("_any", "value") // should not panic
	if cf.HasCustomField("_any") {
		t.Error("HasCustomField on nil should return false")
	}
	cf.DeleteCustomField("_any") // should not panic
	if keys := cf.CustomFieldsKeys(); keys != nil {
		t.Errorf("CustomFieldsKeys on nil = %v, want nil", keys)
	}
}

// TestHarUnmarshalCustomFields tests unmarshaling HAR with custom extension fields
func TestHarUnmarshalCustomFields(t *testing.T) {
	input := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "test", "version": "1.0"},
			"_customLogField": "logValue",
			"entries": [
				{
					"startedDateTime": "2024-01-01T00:00:00.000Z",
					"time": 100,
					"request": {
						"method": "GET",
						"url": "https://example.com",
						"httpVersion": "HTTP/1.1",
						"cookies": [],
						"headers": [],
						"queryString": [],
						"headersSize": 100,
						"bodySize": 0,
						"_customRequestField": 42
					},
					"response": {
						"status": 200,
						"statusText": "OK",
						"httpVersion": "HTTP/1.1",
						"cookies": [],
						"headers": [],
						"content": {
							"size": 1234,
							"mimeType": "text/html",
							"_customContentField": true
						},
						"redirectURL": "",
						"headersSize": 200,
						"bodySize": 1234,
						"_transferSize": 1500,
						"_error": null,
						"_customResponseField": "extra"
					},
					"cache": {},
					"timings": {
						"blocked": 5,
						"dns": 10,
						"connect": 15,
						"ssl": 20,
						"send": 1,
						"wait": 50,
						"receive": 10,
						"_blocked_queueing": 3,
						"_customTimingField": 7.5
					},
					"_initiator": {"type": "parser", "url": "https://example.com", "lineNumber": 10},
					"_priority": "High",
					"_resourceType": "document",
					"_customEntryField": "entryValue"
				}
			]
		},
		"_customHarField": "harValue"
	}`

	var h Har
	if err := json.Unmarshal([]byte(input), &h); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check Har custom fields
	if v := h.GetCustomField("_customHarField"); v != "harValue" {
		t.Errorf("Har._customHarField = %v, want harValue", v)
	}

	// Check Log custom fields
	if v := h.Log.GetCustomField("_customLogField"); v != "logValue" {
		t.Errorf("Log._customLogField = %v, want logValue", v)
	}

	// Check Entries custom fields
	if len(h.Log.Entries) == 0 {
		t.Fatal("No entries found")
	}
	entry := h.Log.Entries[0]

	// Known extension fields should be in struct fields, not CustomFields
	if entry.Initiator.Type != "parser" {
		t.Errorf("Entries._initiator.Type = %v, want parser", entry.Initiator.Type)
	}
	if entry.Priority != "High" {
		t.Errorf("Entries._priority = %v, want High", entry.Priority)
	}
	if entry.ResourceType != "document" {
		t.Errorf("Entries._resourceType = %v, want document", entry.ResourceType)
	}

	// Unknown extension field should be in CustomFields
	if v := entry.GetCustomField("_customEntryField"); v != "entryValue" {
		t.Errorf("Entries._customEntryField = %v, want entryValue", v)
	}

	// Known extension fields should NOT be in CustomFields (avoid duplication)
	if entry.CustomFields.HasCustomField("_initiator") {
		t.Error("Entries.CustomFields should not contain _initiator (it's a struct field)")
	}
	if entry.CustomFields.HasCustomField("_priority") {
		t.Error("Entries.CustomFields should not contain _priority (it's a struct field)")
	}
	if entry.CustomFields.HasCustomField("_resourceType") {
		t.Error("Entries.CustomFields should not contain _resourceType (it's a struct field)")
	}

	// Check Request custom fields
	if v := entry.Request.GetCustomField("_customRequestField"); v != float64(42) {
		t.Errorf("Request._customRequestField = %v, want 42", v)
	}

	// Check Response custom fields
	if entry.Response.TransferSize != 1500 {
		t.Errorf("Response._transferSize = %v, want 1500", entry.Response.TransferSize)
	}
	if v := entry.Response.GetCustomField("_customResponseField"); v != "extra" {
		t.Errorf("Response._customResponseField = %v, want extra", v)
	}
	// Known Response extension fields should NOT be in CustomFields
	if entry.Response.CustomFields.HasCustomField("_transferSize") {
		t.Error("Response.CustomFields should not contain _transferSize (it's a struct field)")
	}
	if entry.Response.CustomFields.HasCustomField("_error") {
		t.Error("Response.CustomFields should not contain _error (it's a struct field)")
	}

	// Check Content custom fields
	if v := entry.Response.Content.GetCustomField("_customContentField"); v != true {
		t.Errorf("Content._customContentField = %v, want true", v)
	}

	// Check Timings custom fields
	if entry.Timings.BlockedQueueing != 3 {
		t.Errorf("Timings._blocked_queueing = %v, want 3", entry.Timings.BlockedQueueing)
	}
	if v := entry.Timings.GetCustomField("_customTimingField"); v != 7.5 {
		t.Errorf("Timings._customTimingField = %v, want 7.5", v)
	}
	// Known Timings extension fields should NOT be in CustomFields
	if entry.Timings.CustomFields.HasCustomField("_blocked_queueing") {
		t.Error("Timings.CustomFields should not contain _blocked_queueing (it's a struct field)")
	}
}

// TestHarMarshalCustomFields tests marshaling HAR with custom extension fields
func TestHarMarshalCustomFields(t *testing.T) {
	h := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
			Entries: []Entries{
				{
					Request: Request{
						Method:      "GET",
						URL:         "https://example.com",
						HTTPVersion: "HTTP/1.1",
						HeadersSize: 100,
						BodySize:    0,
					},
					Response: Response{
						Status:       200,
						StatusText:   "OK",
						HTTPVersion:  "HTTP/1.1",
						HeadersSize:  200,
						BodySize:     1234,
						TransferSize: 1500,
					},
				},
			},
		},
	}
	h.SetCustomField("_harCustom", "harVal")
	h.Log.SetCustomField("_logCustom", 42)
	h.Log.Entries[0].SetCustomField("_entryCustom", "entryVal")
	h.Log.Entries[0].Request.SetCustomField("_reqCustom", "reqVal")
	h.Log.Entries[0].Response.SetCustomField("_respCustom", "respVal")
	h.Log.Entries[0].Response.Content.SetCustomField("_contentCustom", "contentVal")

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify custom fields are in output
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal of marshaled data failed: %v", err)
	}

	log := raw["log"].(map[string]interface{})
	if v := log["_logCustom"]; v != float64(42) {
		t.Errorf("log._logCustom = %v, want 42", v)
	}

	entries := log["entries"].([]interface{})
	entry := entries[0].(map[string]interface{})
	if v := entry["_entryCustom"]; v != "entryVal" {
		t.Errorf("entry._entryCustom = %v, want entryVal", v)
	}

	req := entry["request"].(map[string]interface{})
	if v := req["_reqCustom"]; v != "reqVal" {
		t.Errorf("request._reqCustom = %v, want reqVal", v)
	}

	resp := entry["response"].(map[string]interface{})
	if v := resp["_respCustom"]; v != "respVal" {
		t.Errorf("response._respCustom = %v, want respVal", v)
	}
	// Verify struct-level extension field is also present
	if v := resp["_transferSize"]; v != float64(1500) {
		t.Errorf("response._transferSize = %v, want 1500", v)
	}

	content := resp["content"].(map[string]interface{})
	if v := content["_contentCustom"]; v != "contentVal" {
		t.Errorf("content._contentCustom = %v, want contentVal", v)
	}
}

// TestHarRoundTripCustomFields tests that marshal -> unmarshal preserves custom fields
func TestHarRoundTripCustomFields(t *testing.T) {
	original := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
			Entries: []Entries{
				{
					Request: Request{
						Method:      "POST",
						URL:         "https://api.example.com/data",
						HTTPVersion: "HTTP/2.0",
						HeadersSize: 200,
						BodySize:    50,
					},
					Response: Response{
						Status:      201,
						StatusText:  "Created",
						HTTPVersion: "HTTP/2.0",
						HeadersSize: 150,
						BodySize:    100,
						TransferSize: 250,
					},
					Timings: Timings{
						BlockedQueueing: 5.0,
					},
				},
			},
		},
	}
	original.SetCustomField("_harExt", "harExtension")
	original.Log.SetCustomField("_logExt", "logExtension")
	original.Log.Entries[0].SetCustomField("_entryExt", "entryExtension")
	original.Log.Entries[0].Request.SetCustomField("_reqExt", "requestExtension")
	original.Log.Entries[0].Response.SetCustomField("_respExt", "responseExtension")
	original.Log.Entries[0].Response.Content.SetCustomField("_contentExt", "contentExtension")
	original.Log.Entries[0].Timings.SetCustomField("_timingExt", 3.14)
	original.Log.Entries[0].Cache.SetCustomField("_cacheExt", "cacheExtension")

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var restored Har
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify custom fields survived round trip
	if v := restored.GetCustomField("_harExt"); v != "harExtension" {
		t.Errorf("Har._harExt after round trip = %v, want harExtension", v)
	}
	if v := restored.Log.GetCustomField("_logExt"); v != "logExtension" {
		t.Errorf("Log._logExt after round trip = %v, want logExtension", v)
	}
	if len(restored.Log.Entries) == 0 {
		t.Fatal("No entries after round trip")
	}
	entry := restored.Log.Entries[0]
	if v := entry.GetCustomField("_entryExt"); v != "entryExtension" {
		t.Errorf("Entries._entryExt after round trip = %v, want entryExtension", v)
	}
	if v := entry.Request.GetCustomField("_reqExt"); v != "requestExtension" {
		t.Errorf("Request._reqExt after round trip = %v, want requestExtension", v)
	}
	if v := entry.Response.GetCustomField("_respExt"); v != "responseExtension" {
		t.Errorf("Response._respExt after round trip = %v, want responseExtension", v)
	}
	if v := entry.Response.Content.GetCustomField("_contentExt"); v != "contentExtension" {
		t.Errorf("Content._contentExt after round trip = %v, want contentExtension", v)
	}
	if v := entry.Timings.GetCustomField("_timingExt"); v != 3.14 {
		t.Errorf("Timings._timingExt after round trip = %v, want 3.14", v)
	}
	if v := entry.Cache.GetCustomField("_cacheExt"); v != "cacheExtension" {
		t.Errorf("Cache._cacheExt after round trip = %v, want cacheExtension", v)
	}

	// Verify struct-level extension fields also survived
	if entry.Response.TransferSize != 250 {
		t.Errorf("Response.TransferSize after round trip = %v, want 250", entry.Response.TransferSize)
	}
	if entry.Timings.BlockedQueueing != 5.0 {
		t.Errorf("Timings.BlockedQueueing after round trip = %v, want 5.0", entry.Timings.BlockedQueueing)
	}
}

// TestEmptyCustomFieldsDoesNotAffectOutput tests that empty CustomFields produces clean JSON
func TestEmptyCustomFieldsDoesNotAffectOutput(t *testing.T) {
	h := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
		},
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify no underscore-prefixed fields in output (besides what struct fields produce)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	log := raw["log"].(map[string]interface{})
	for key := range log {
		if key[0] == '_' {
			t.Errorf("Unexpected underscore key in output: %s", key)
		}
	}
}

// TestChromeExtensionFields tests common Chrome DevTools extension fields
func TestChromeExtensionFields(t *testing.T) {
	input := `{
		"log": {
			"version": "1.2",
			"creator": {"name": "Chrome", "version": "120.0"},
			"entries": [
				{
					"startedDateTime": "2024-01-01T00:00:00.000Z",
					"time": 200,
					"request": {
						"method": "GET",
						"url": "https://example.com/api",
						"httpVersion": "HTTP/2.0",
						"cookies": [],
						"headers": [],
						"queryString": [],
						"headersSize": 0,
						"bodySize": 0
					},
					"response": {
						"status": 200,
						"statusText": "",
						"httpVersion": "HTTP/2.0",
						"cookies": [],
						"headers": [],
						"content": {"size": 500, "mimeType": "application/json"},
						"redirectURL": "",
						"headersSize": 0,
						"bodySize": 500,
						"_transferSize": 600
					},
					"cache": {},
					"timings": {
						"blocked": 2.5,
						"dns": -1,
						"connect": -1,
						"ssl": -1,
						"send": 0.5,
						"wait": 100,
						"receive": 50,
						"_blocked_queueing": 1.2,
						"_blocked_proxy": 0.8
					},
					"_initiator": {
						"type": "script",
						"url": "https://example.com/app.js",
						"lineNumber": 42
					},
					"_priority": "High",
					"_resourceType": "xhr",
					"_error": null,
					"_sameSite": "None",
					"_wasFetchedViaServiceWorker": true,
					"_wasAlternateProtocolAvailable": false
				}
			]
		}
	}`

	var h Har
	if err := json.Unmarshal([]byte(input), &h); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	entry := h.Log.Entries[0]

	// Verify struct-level Chrome extension fields
	if entry.Initiator.Type != "script" {
		t.Errorf("Initiator.Type = %v, want script", entry.Initiator.Type)
	}
	if entry.Initiator.URL != "https://example.com/app.js" {
		t.Errorf("Initiator.URL = %v, want https://example.com/app.js", entry.Initiator.URL)
	}
	if entry.Initiator.LineNumber != 42 {
		t.Errorf("Initiator.LineNumber = %v, want 42", entry.Initiator.LineNumber)
	}
	if entry.Priority != "High" {
		t.Errorf("Priority = %v, want High", entry.Priority)
	}
	if entry.ResourceType != "xhr" {
		t.Errorf("ResourceType = %v, want xhr", entry.ResourceType)
	}
	if entry.Response.TransferSize != 600 {
		t.Errorf("TransferSize = %v, want 600", entry.Response.TransferSize)
	}
	if entry.Timings.BlockedQueueing != 1.2 {
		t.Errorf("BlockedQueueing = %v, want 1.2", entry.Timings.BlockedQueueing)
	}
	if entry.Timings.BlockedProxy != 0.8 {
		t.Errorf("BlockedProxy = %v, want 0.8", entry.Timings.BlockedProxy)
	}

	// Verify unknown Chrome extension fields are in CustomFields
	if v := entry.GetCustomField("_sameSite"); v != "None" {
		t.Errorf("Entries._sameSite = %v, want None", v)
	}
	if v := entry.GetCustomField("_wasFetchedViaServiceWorker"); v != true {
		t.Errorf("Entries._wasFetchedViaServiceWorker = %v, want true", v)
	}
	if v := entry.GetCustomField("_wasAlternateProtocolAvailable"); v != false {
		t.Errorf("Entries._wasAlternateProtocolAvailable = %v, want false", v)
	}

	// Known struct fields should not be in CustomFields
	if entry.CustomFields.HasCustomField("_initiator") {
		t.Error("_initiator should not be in CustomFields")
	}
	if entry.CustomFields.HasCustomField("_priority") {
		t.Error("_priority should not be in CustomFields")
	}
	if entry.CustomFields.HasCustomField("_resourceType") {
		t.Error("_resourceType should not be in CustomFields")
	}
}

// TestCookieCustomFields tests custom fields on Cookie
func TestCookieCustomFields(t *testing.T) {
	input := `{
		"name": "session",
		"value": "abc123",
		"_sameSite": "Strict",
		"_customCookie": "cookieExt"
	}`

	var c Cookie
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("Unmarshal Cookie failed: %v", err)
	}

	if c.Name != "session" {
		t.Errorf("Cookie.Name = %v, want session", c.Name)
	}
	if v := c.GetCustomField("_sameSite"); v != "Strict" {
		t.Errorf("Cookie._sameSite = %v, want Strict", v)
	}
	if v := c.GetCustomField("_customCookie"); v != "cookieExt" {
		t.Errorf("Cookie._customCookie = %v, want cookieExt", v)
	}

	// Round trip
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal Cookie failed: %v", err)
	}

	var c2 Cookie
	if err := json.Unmarshal(data, &c2); err != nil {
		t.Fatalf("Unmarshal round-trip Cookie failed: %v", err)
	}
	if v := c2.GetCustomField("_sameSite"); v != "Strict" {
		t.Errorf("Cookie._sameSite after round trip = %v, want Strict", v)
	}
	if v := c2.GetCustomField("_customCookie"); v != "cookieExt" {
		t.Errorf("Cookie._customCookie after round trip = %v, want cookieExt", v)
	}
}

// TestPagesCustomFields tests custom fields on Pages
func TestPagesCustomFields(t *testing.T) {
	input := `{
		"startedDateTime": "2024-01-01T00:00:00.000Z",
		"id": "page_1",
		"title": "Test Page",
		"pageTimings": {"onContentLoad": 100, "onLoad": 200},
		"_pageCustom": "pageExt"
	}`

	var p Pages
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal Pages failed: %v", err)
	}

	if p.ID != "page_1" {
		t.Errorf("Pages.ID = %v, want page_1", p.ID)
	}
	if v := p.GetCustomField("_pageCustom"); v != "pageExt" {
		t.Errorf("Pages._pageCustom = %v, want pageExt", v)
	}

	// Round trip
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal Pages failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw Pages failed: %v", err)
	}
	if v := raw["_pageCustom"]; v != "pageExt" {
		t.Errorf("Pages._pageCustom in JSON = %v, want pageExt", v)
	}
}

// TestContentCustomFields tests custom fields on Content
func TestContentCustomFields(t *testing.T) {
	input := `{
		"size": 1024,
		"mimeType": "text/html",
		"_contentHash": "abc123",
		"_contentEncoding": "br"
	}`

	var c Content
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("Unmarshal Content failed: %v", err)
	}

	if c.Size != 1024 {
		t.Errorf("Content.Size = %v, want 1024", c.Size)
	}
	if v := c.GetCustomField("_contentHash"); v != "abc123" {
		t.Errorf("Content._contentHash = %v, want abc123", v)
	}
	if v := c.GetCustomField("_contentEncoding"); v != "br" {
		t.Errorf("Content._contentEncoding = %v, want br", v)
	}
}

// TestCacheCustomFields tests custom fields on Cache
func TestCacheCustomFields(t *testing.T) {
	input := `{
		"_cacheHit": true,
		"_cacheKey": "v2/abc"
	}`

	var c Cache
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("Unmarshal Cache failed: %v", err)
	}

	if v := c.GetCustomField("_cacheHit"); v != true {
		t.Errorf("Cache._cacheHit = %v, want true", v)
	}
	if v := c.GetCustomField("_cacheKey"); v != "v2/abc" {
		t.Errorf("Cache._cacheKey = %v, want v2/abc", v)
	}
}

// TestSetCustomFieldCreatesMap tests that SetCustomField creates the map if nil
func TestSetCustomFieldCreatesMap(t *testing.T) {
	h := Har{}
	if h.CustomFields != nil {
		t.Error("CustomFields should start as nil for zero-value Har")
	}
	h.SetCustomField("_test", "value")
	if h.CustomFields == nil {
		t.Error("SetCustomField should create the CustomFields map")
	}
	if v := h.GetCustomField("_test"); v != "value" {
		t.Errorf("GetCustomField = %v, want value", v)
	}
}

// TestCustomFieldComplexValues tests custom fields with complex values
func TestCustomFieldComplexValues(t *testing.T) {
	complexValue := map[string]interface{}{
		"nested":   true,
		"count":    float64(42),
		"children": []interface{}{"a", "b", "c"},
	}

	h := Har{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "test", Version: "1.0"},
		},
	}
	h.SetCustomField("_complex", complexValue)

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var restored Har
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	v := restored.GetCustomField("_complex")
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("_complex value type = %T, want map[string]interface{}", v)
	}
	if m["nested"] != true {
		t.Errorf("_complex.nested = %v, want true", m["nested"])
	}
	if m["count"] != float64(42) {
		t.Errorf("_complex.count = %v, want 42", m["count"])
	}
}

// TestNoCustomFieldsNoExtraKeys tests that when there are no custom fields,
// the JSON output doesn't have extra keys
func TestNoCustomFieldsNoExtraKeys(t *testing.T) {
	r := Request{
		Method:      "GET",
		URL:         "https://example.com",
		HTTPVersion: "HTTP/1.1",
		HeadersSize: 100,
		BodySize:    0,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	for key := range raw {
		if key[0] == '_' {
			t.Errorf("Unexpected underscore key in Request output with no custom fields: %s", key)
		}
	}
}
