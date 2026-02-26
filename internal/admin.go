package blog

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func (app *App) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := app.cfg.AdminAPIKey
		if key == "" {
			http.Error(w, "admin not configured", http.StatusForbidden)
			return
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || subtle.ConstantTimeCompare([]byte(auth[7:]), []byte(key)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

type adminStats struct {
	PublicPosts  int `json:"public_posts"`
	DraftPosts   int `json:"draft_posts"`
	PrivatePosts int `json:"private_posts"`
	Comments     int `json:"comments"`
}

func (app *App) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	posts := app.posts
	app.mu.RUnlock()

	stats := adminStats{}
	for _, p := range posts {
		switch {
		case p.Private:
			stats.PrivatePosts++
		case p.Status == "draft":
			stats.DraftPosts++
		default:
			stats.PublicPosts++
		}
	}

	if app.db != nil {
		app.db.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&stats.Comments)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (app *App) handleAdminComments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if app.db == nil {
		json.NewEncoder(w).Encode([]Comment{})
		return
	}

	rows, err := app.db.Query(
		`SELECT id, post_slug, author, body, visible, created_at
		 FROM comments ORDER BY created_at DESC`,
	)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostSlug, &c.Author, &c.Body, &c.Visible, &c.CreatedAt); err != nil {
			continue
		}
		comments = append(comments, c)
	}

	if comments == nil {
		comments = []Comment{}
	}
	json.NewEncoder(w).Encode(comments)
}

func (app *App) handleAdminCommentToggle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if app.db == nil {
		http.Error(w, "db not available", http.StatusServiceUnavailable)
		return
	}

	_, err := app.db.Exec(`UPDATE comments SET visible = NOT visible WHERE id = ?`, id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleAdminCommentDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if app.db == nil {
		http.Error(w, "db not available", http.StatusServiceUnavailable)
		return
	}

	_, err := app.db.Exec(`DELETE FROM comments WHERE id = ?`, id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
