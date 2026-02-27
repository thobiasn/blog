package blog

import (
	"database/sql"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type SearchResult struct {
	Slug        string
	Title       string
	ContentType string
	Snippet     template.HTML
}

// stripTags removes HTML tags from rendered content for plain-text indexing.
func stripTags(html template.HTML) string {
	s := string(html)
	var b strings.Builder
	b.Grow(len(s))
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			b.WriteByte(' ')
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func rebuildSearchIndex(db *sql.DB, posts []Post, projects []Project) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("search index: begin tx: %v", err)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM search_index`); err != nil {
		log.Printf("search index: delete: %v", err)
		return
	}

	for _, p := range posts {
		_, err := tx.Exec(
			`INSERT INTO search_index (slug, title, tags, body, content_type) VALUES (?, ?, ?, ?, 'post')`,
			p.Slug, p.Title, strings.Join(p.Tags, " "), stripTags(p.Body),
		)
		if err != nil {
			log.Printf("search index: insert post %s: %v", p.Slug, err)
			return
		}
	}

	for _, p := range projects {
		_, err := tx.Exec(
			`INSERT INTO search_index (slug, title, tags, body, content_type) VALUES (?, ?, ?, ?, 'project')`,
			p.Slug, p.Title, strings.Join(p.Tags, " "), stripTags(p.Body),
		)
		if err != nil {
			log.Printf("search index: insert project %s: %v", p.Slug, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("search index: commit: %v", err)
	}
}

func (app *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	var results []SearchResult
	if q != "" && app.db != nil {
		// Quote the query to prevent FTS5 operator injection
		quoted := `"` + strings.ReplaceAll(q, `"`, `""`) + `"`
		rows, err := app.db.Query(
			`SELECT slug, title, content_type, snippet(search_index, 3, '<mark>', '</mark>', '...', 30)
			 FROM search_index WHERE search_index MATCH ? ORDER BY rank`,
			quoted,
		)
		if err != nil {
			log.Printf("search query error: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var sr SearchResult
				var snippet string
				if err := rows.Scan(&sr.Slug, &sr.Title, &sr.ContentType, &snippet); err != nil {
					log.Printf("search scan error: %v", err)
					break
				}
					snippet = html.EscapeString(snippet)
				snippet = strings.ReplaceAll(snippet, "&lt;mark&gt;", "<mark>")
				snippet = strings.ReplaceAll(snippet, "&lt;/mark&gt;", "</mark>")
				sr.Snippet = template.HTML(snippet)
				results = append(results, sr)
			}
		}
	}

	app.render(w, "search", map[string]any{
		"Query":   q,
		"Results": results,
	})
}

func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if app.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"error"}`))
		return
	}
	if err := app.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"error"}`))
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}
