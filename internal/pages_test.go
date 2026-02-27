package blog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFindPage(t *testing.T) {
	pages := []Page{
		{Title: "Uses", Slug: "uses"},
		{Title: "Now", Slug: "now"},
	}

	p, ok := findPage(pages, "uses")
	if !ok {
		t.Fatal("findPage(uses) not found")
	}
	if p.Title != "Uses" {
		t.Errorf("Title = %q, want %q", p.Title, "Uses")
	}

	_, ok = findPage(pages, "nonexistent")
	if ok {
		t.Error("findPage(nonexistent) should not be found")
	}
}

func TestHandlePage(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/uses", nil)
	w := httptest.NewRecorder()

	app.handlePage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Uses") {
		t.Error("page should contain 'Uses'")
	}
}

func TestHandlePageNotFound(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	app.handlePage(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
