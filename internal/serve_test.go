package blog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	app := testApp(t)

	w := httptest.NewRecorder()

	app.render(w, "home", map[string]any{
		"Posts":    publicPosts(app.posts),
		"Projects": featuredProjects(app.projects),
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
	if w.Body.Len() == 0 {
		t.Error("body should not be empty")
	}
}

func TestRenderNotFound(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/nope", nil)
	w := httptest.NewRecorder()

	app.renderNotFound(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
}

func TestHandleChromaCSS(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/static/chroma.css", nil)
	w := httptest.NewRecorder()

	app.handleChromaCSS(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/css" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/css")
	}
	cache := w.Header().Get("Cache-Control")
	if !strings.Contains(cache, "max-age=86400") {
		t.Errorf("Cache-Control = %q, want max-age=86400", cache)
	}
	if w.Body.Len() == 0 {
		t.Error("body should contain CSS")
	}
}
