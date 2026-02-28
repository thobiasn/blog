package blog

import (
	"database/sql"
	"html/template"
	"path/filepath"
	"testing"
	"time"
)

// testApp creates a fully functional *App with in-memory SQLite, real templates,
// and pre-loaded test content.
func testApp(t *testing.T) *App {
	t.Helper()

	db := testDB(t)
	md := newMarkdown()
	chromaCSS, err := generateChromaCSS()
	if err != nil {
		t.Fatalf("generating chroma css: %v", err)
	}

	tmpls := testTemplates(t)

	app := &App{
		cfg: Config{
			BaseURL:     "http://localhost:8080",
			ContentDir:  "content",
			AdminAPIKey: "test-key",
		},
		db:        db,
		md:        md,
		chromaCSS: chromaCSS,
		tmpls:     tmpls,
		limiter:   newRateLimiter(),
		posts: []Post{
			{
				Title:       "First Post",
				Slug:        "first-post",
				Description: "A post about Go and the web",
				Date:        time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC),
				Tags:        []string{"go", "web"},
				Body:        "<p>Hello world</p>",
				Project:     "blog",
			},
			{
				Title:   "Private Post",
				Slug:    "private-post",
				Private: true,
				Date:    time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC),
				Body:    "<p>Secret stuff</p>",
			},
		},
		pages: []Page{
			{Title: "Uses", Slug: "uses", Body: "<p>My tools</p>"},
			{Title: "Now", Slug: "now", Body: "<p>What I'm doing now</p>"},
		},
		projects: []Project{
			{
				Title:       "Blog",
				Slug:        "blog",
				Description: "Personal blog",
				Featured:    true,
				Body:        "<p>A blog project</p>",
			},
			{
				Title:       "Side Project",
				Slug:        "side-project",
				Description: "Something else",
				Featured:    false,
				Body:        "<p>Side project</p>",
			},
		},
	}

	return app
}

// testTemplates parses real templates from ../templates/.
func testTemplates(t *testing.T) map[string]*template.Template {
	t.Helper()

	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string { return t.Format("January 2, 2006") },
		"shortDate":  func(t time.Time) string { return t.Format("2006-01-02") },
		"isLocal": func() bool { return true },
	}

	base := filepath.Join("..", "templates")
	names := []string{"home", "post", "post_list", "page", "project", "project_list", "subscribe", "search", "404"}
	tmpls := make(map[string]*template.Template, len(names))
	for _, name := range names {
		tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
			filepath.Join(base, "base.html"),
			filepath.Join(base, "tracking.html"),
			filepath.Join(base, name+".html"),
		)
		if err != nil {
			t.Fatalf("parsing template %s: %v", name, err)
		}
		tmpls[name] = tmpl
	}
	return tmpls
}

// testDB creates an in-memory SQLite database with tables initialized.
func testDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seedComments inserts test comments into the database.
func seedComments(t *testing.T, db *sql.DB, comments []Comment) {
	t.Helper()

	for _, c := range comments {
		visible := 1
		if !c.Visible {
			visible = 0
		}
		_, err := db.Exec(
			`INSERT INTO comments (post_slug, author, body, visible) VALUES (?, ?, ?, ?)`,
			c.PostSlug, c.Author, c.Body, visible,
		)
		if err != nil {
			t.Fatalf("seeding comment: %v", err)
		}
	}
}
