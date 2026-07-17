package har

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cover Entries.ToHTTPRequest http.NewRequest-error branch (lines 68-71):
// an invalid HTTP method (contains a space) causes http.NewRequest to
// fail, which is wrapped as ErrCodeInvalidFormat.
func TestCovToHTTPRequest_NewRequestError(t *testing.T) {
	e := &Entries{}
	e.Request.Method = "GET WITH SPACE" // illegal method token
	e.Request.URL = "https://example.com/"
	req, err := e.ToHTTPRequest()
	if err == nil {
		t.Fatal("expected error from http.NewRequest, got nil")
	}
	assert.Nil(t, req)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)
}

// Cover Entries.ToHTTPRequest URL-parse-error branch (lines 55-57):
// url.Parse fails on a control character in the URL.
func TestCovToHTTPRequest_URLParseError(t *testing.T) {
	e := &Entries{}
	e.Request.Method = "GET"
	e.Request.URL = "https://example.com\x7f"
	req, err := e.ToHTTPRequest()
	if err == nil {
		t.Fatal("expected error from url.Parse, got nil")
	}
	assert.Nil(t, req)
	assertHarErrorCode(t, err, ErrCodeInvalidValue)
}

// Cover Entries.ToHTTPRequest nil-receiver branch (lines 49-50) and
// happy path with PostData and cookies.
func TestCovToHTTPRequest_NilAndHappy(t *testing.T) {
	var e *Entries
	req, err := e.ToHTTPRequest()
	if err == nil {
		t.Fatal("expected error for nil entry, got nil")
	}
	assert.Nil(t, req)
	assertHarErrorCode(t, err, ErrCodeInvalidFormat)

	// happy path with PostData + cookies
	h := NewHar()
	ee := h.AddEntry("POST", "https://example.com/api", "HTTP/1.1", "")
	ee.SetResponseStatus(200, "OK")
	ee.SetPostData("application/json", `{"k":"v"}`)
	ee.AddRequestHeader("X-Test", "1")
	ee.AddCookie("session", "abc")
	req, err = ee.ToHTTPRequest()
	if !assert.NoError(t, err) {
		return
	}
	if assert.NotNil(t, req) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "https://example.com/api", req.URL.String())
		assert.Equal(t, "1", req.Header.Get("X-Test"))
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		// cookie present
		c, cerr := req.Cookie("session")
		if assert.NoError(t, cerr) {
			assert.Equal(t, "abc", c.Value)
		}
	}
}

// Cover normalizeReplayError nil-input branch (lines 304-306).
func TestCovNormalizeReplayError_Nil(t *testing.T) {
	assert.Nil(t, normalizeReplayError(nil))
}

// Cover normalizeReplayError HarError-as branch (lines 308-310) and the
// pass-through branch (lines 311-312).
func TestCovNormalizeReplayError_Branches(t *testing.T) {
	// HarError input -> returned as-is (errors.As matches).
	he := NewFileSystemError("fs", errors.New("x"))
	out := normalizeReplayError(he)
	assert.Same(t, he, out)

	// non-HarError input -> returned as-is (pass-through, line 312).
	plain := errors.New("plain")
	out = normalizeReplayError(plain)
	assert.Same(t, plain, out)

	// wrapped HarError -> errors.As unwraps to the *HarError.
	wrapped := errors.Join(he) // wraps he; errors.As finds *HarError
	out = normalizeReplayError(wrapped)
	// errors.As on a joined error finds the *HarError.
	assert.NotNil(t, out)
}

// createHTTPClient is exercised indirectly via Replay with a server we
// can't reach; instead call it directly to cover the FollowRedirects=false
// and MaxRedirects>0 branches (lines 287-298).
func TestCovCreateHTTPClient_Branches(t *testing.T) {
	// FollowRedirects=false -> CheckRedirect returns ErrUseLastResponse.
	c := createHTTPClient(ReplayOptions{FollowRedirects: false})
	if !assert.NotNil(t, c) || !assert.NotNil(t, c.CheckRedirect) {
		return
	}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	err := c.CheckRedirect(req, []*http.Request{req})
	assert.Error(t, err) // http.ErrUseLastResponse

	// FollowRedirects=true with MaxRedirects>0 -> CheckRedirect enforces cap.
	c = createHTTPClient(ReplayOptions{FollowRedirects: true, MaxRedirects: 2})
	via := []*http.Request{req, req} // len 2 == MaxRedirects -> error
	err = c.CheckRedirect(req, via)
	assert.Error(t, err)
	assertHarErrorCode(t, err, ErrCodeInvalidValue)

	// len < MaxRedirects -> no error.
	via2 := []*http.Request{req}
	err = c.CheckRedirect(req, via2)
	assert.NoError(t, err)
}
