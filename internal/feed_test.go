package blog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRSS(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/rss.xml", nil)
	w := httptest.NewRecorder()
	app.handleRSS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/rss+xml" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/rss+xml")
	}

	body := w.Body.String()
	if !strings.Contains(body, "<rss version=\"2.0\">") {
		t.Error("missing RSS root element")
	}
	if !strings.Contains(body, "<title>First Post</title>") {
		t.Error("missing public post")
	}
	if strings.Contains(body, "Draft Post") {
		t.Error("draft post should not appear in RSS")
	}
	if strings.Contains(body, "Private Post") {
		t.Error("private post should not appear in RSS")
	}
}
