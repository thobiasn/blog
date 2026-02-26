package blog

import (
	"net/http"
	"strings"
)

func (app *App) handlePage(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/")

	app.mu.RLock()
	page, ok := findPage(app.pages, slug)
	app.mu.RUnlock()

	if !ok {
		app.renderNotFound(w, r)
		return
	}

	app.render(w, "page", map[string]any{
		"Page": page,
	})
}

func findPage(pages []Page, slug string) (Page, bool) {
	for _, p := range pages {
		if p.Slug == slug {
			return p, true
		}
	}
	return Page{}, false
}
