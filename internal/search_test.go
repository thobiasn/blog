package blog

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStripTags(t *testing.T) {
	tests := []struct {
		name string
		in   template.HTML
		want string
	}{
		{"plain text", "hello world", "hello world"},
		{"simple tag", "<p>hello</p>", " hello "},
		{"nested tags", "<div><p>hello</p></div>", "  hello  "},
		{"empty", "", ""},
		{"attributes", `<a href="url">link</a>`, " link "},
		{"self-closing", "before<br/>after", "before after"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripTags(tt.in)
			if got != tt.want {
				t.Errorf("stripTags(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRebuildSearchIndex(t *testing.T) {
	db := testDB(t)

	posts := []Post{
		{Slug: "hello", Title: "Hello World", Tags: []string{"go", "web"}, Body: "<p>content</p>"},
	}
	projects := []Project{
		{Slug: "blog", Title: "Blog", Tags: []string{"go"}, Body: "<p>project body</p>"},
	}

	rebuildSearchIndex(db, posts, projects)

	var count int
	db.QueryRow(`SELECT count(*) FROM search_index`).Scan(&count)
	if count != 2 {
		t.Fatalf("expected 2 rows, got %d", count)
	}

	// Rebuild again — should be idempotent
	rebuildSearchIndex(db, posts, projects)
	db.QueryRow(`SELECT count(*) FROM search_index`).Scan(&count)
	if count != 2 {
		t.Fatalf("after rebuild: expected 2 rows, got %d", count)
	}
}

func TestHandleSearch(t *testing.T) {
	app := testApp(t)
	rebuildSearchIndex(app.db, publicPosts(app.posts), app.projects)

	t.Run("empty query", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search", nil)
		w := httptest.NewRecorder()
		app.handleSearch(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("query match", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search?q=First", nil)
		w := httptest.NewRecorder()
		app.handleSearch(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "First Post") {
			t.Error("expected result to contain 'First Post'")
		}
	})

	t.Run("no results", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/search?q=nonexistent", nil)
		w := httptest.NewRecorder()
		app.handleSearch(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "No results") {
			t.Error("expected 'No results' message")
		}
	})
}

func TestHandleSearchExcludesPrivateAndDraft(t *testing.T) {
	app := testApp(t)
	// Only public posts are indexed
	rebuildSearchIndex(app.db, publicPosts(app.posts), app.projects)

	r := httptest.NewRequest("GET", "/search?q=Draft", nil)
	w := httptest.NewRecorder()
	app.handleSearch(w, r)
	if strings.Contains(w.Body.String(), "Draft Post") {
		t.Error("draft post should not appear in search results")
	}

	r = httptest.NewRequest("GET", "/search?q=Private", nil)
	w = httptest.NewRecorder()
	app.handleSearch(w, r)
	if strings.Contains(w.Body.String(), "Private Post") {
		t.Error("private post should not appear in search results")
	}
}

func TestHandleHealth(t *testing.T) {
	app := testApp(t)

	r := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	app.handleHealth(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"ok"`) {
		t.Errorf("expected ok status, got %s", w.Body.String())
	}
}
