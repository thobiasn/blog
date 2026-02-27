package blog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFindProject(t *testing.T) {
	projects := []Project{
		{Title: "Blog", Slug: "blog"},
		{Title: "CLI", Slug: "cli"},
	}

	p, ok := findProject(projects, "blog")
	if !ok {
		t.Fatal("findProject(blog) not found")
	}
	if p.Title != "Blog" {
		t.Errorf("Title = %q, want %q", p.Title, "Blog")
	}

	_, ok = findProject(projects, "nope")
	if ok {
		t.Error("findProject(nope) should not be found")
	}
}

func TestFeaturedProjects(t *testing.T) {
	projects := []Project{
		{Slug: "featured", Featured: true},
		{Slug: "normal", Featured: false},
		{Slug: "also-featured", Featured: true},
	}

	got := featuredProjects(projects)
	if len(got) != 2 {
		t.Fatalf("got %d featured, want 2", len(got))
	}
	if got[0].Slug != "featured" || got[1].Slug != "also-featured" {
		t.Errorf("got %v, want featured and also-featured", got)
	}
}

func TestRelatedPosts(t *testing.T) {
	posts := []Post{
		{Slug: "related", Project: "blog", Status: "public", Private: false},
		{Slug: "draft-related", Project: "blog", Status: "draft", Private: false},
		{Slug: "private-related", Project: "blog", Status: "public", Private: true},
		{Slug: "unrelated", Project: "other", Status: "public", Private: false},
	}

	got := relatedPosts(posts, "blog")
	if len(got) != 1 {
		t.Fatalf("got %d related, want 1 (only public non-private)", len(got))
	}
	if got[0].Slug != "related" {
		t.Errorf("got %q, want %q", got[0].Slug, "related")
	}
}

func TestHandleProjectList(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/projects", nil)
	w := httptest.NewRecorder()

	app.handleProjectList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Blog") {
		t.Error("project list should contain 'Blog'")
	}
	if !strings.Contains(body, "Side Project") {
		t.Error("project list should contain 'Side Project'")
	}
}

func TestHandleProject(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/projects/blog", nil)
	req.SetPathValue("slug", "blog")
	w := httptest.NewRecorder()

	app.handleProject(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Blog") {
		t.Error("project page should contain 'Blog'")
	}
}

func TestHandleProjectNotFound(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/projects/nope", nil)
	req.SetPathValue("slug", "nope")
	w := httptest.NewRecorder()

	app.handleProject(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
