package har

import "testing"

func TestGeneratorHarNilReceiverMethods(t *testing.T) {
	var h *Har

	if got := h.SetBrowser("browser", "1.0"); got != nil {
		t.Fatalf("SetBrowser() = %#v, want nil", got)
	}
	if got := h.SetVersion("1.2"); got != nil {
		t.Fatalf("SetVersion() = %#v, want nil", got)
	}
	if got := h.SetCreator("creator", "1.0"); got != nil {
		t.Fatalf("SetCreator() = %#v, want nil", got)
	}
	if got := h.AddPage("page", "Page"); got != nil {
		t.Fatalf("AddPage() = %#v, want nil", got)
	}
	if got := h.AddEntry("GET", "https://example.com", "HTTP/1.1", ""); got != nil {
		t.Fatalf("AddEntry() = %#v, want nil", got)
	}
}

func TestGeneratorPageNilReceiverMethods(t *testing.T) {
	var page *Pages

	if got := page.SetPageTimings(10, 20); got != nil {
		t.Fatalf("SetPageTimings() = %#v, want nil", got)
	}
}

func TestGeneratorEntryNilReceiverMethods(t *testing.T) {
	var entry *Entries

	if got := entry.AddRequestHeader("A", "B"); got != nil {
		t.Fatalf("AddRequestHeader() = %#v, want nil", got)
	}
	if got := entry.AddResponseHeader("A", "B"); got != nil {
		t.Fatalf("AddResponseHeader() = %#v, want nil", got)
	}
	if got := entry.SetResponseStatus(200, "OK"); got != nil {
		t.Fatalf("SetResponseStatus() = %#v, want nil", got)
	}
	if got := entry.SetResponseContent(10, "text/plain"); got != nil {
		t.Fatalf("SetResponseContent() = %#v, want nil", got)
	}
	if got := entry.SetTimings(0, 0, 0, 0, 0, 0, 0); got != nil {
		t.Fatalf("SetTimings() = %#v, want nil", got)
	}
	if got := entry.AddCookie("a", "b"); got != nil {
		t.Fatalf("AddCookie() = %#v, want nil", got)
	}
	if got := entry.AddResponseCookie("a", "b"); got != nil {
		t.Fatalf("AddResponseCookie() = %#v, want nil", got)
	}
	if got := entry.AddQueryParameter("a", "b"); got != nil {
		t.Fatalf("AddQueryParameter() = %#v, want nil", got)
	}
	if got := entry.SetPostData("text/plain", "body"); got != nil {
		t.Fatalf("SetPostData() = %#v, want nil", got)
	}
	if got := entry.SetPostDataParams("application/x-www-form-urlencoded", []Param{{Name: "a", Value: "b"}}); got != nil {
		t.Fatalf("SetPostDataParams() = %#v, want nil", got)
	}
	if got := entry.SetResponseContentText("body"); got != nil {
		t.Fatalf("SetResponseContentText() = %#v, want nil", got)
	}
	if got := entry.SetServerIP("127.0.0.1"); got != nil {
		t.Fatalf("SetServerIP() = %#v, want nil", got)
	}
	if got := entry.SetConnection("conn"); got != nil {
		t.Fatalf("SetConnection() = %#v, want nil", got)
	}
	if got := entry.SetPageref("page"); got != nil {
		t.Fatalf("SetPageref() = %#v, want nil", got)
	}
}
