package blog

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type Comment struct {
	ID        int64
	PostSlug  string
	Author    string
	Body      string
	Visible   bool
	CreatedAt time.Time
}

// commentsBySlug returns visible comments for a post, oldest first.
// Returns nil if db is nil (local dev without DB).
func (app *App) commentsBySlug(slug string) []Comment {
	if app.db == nil {
		return nil
	}

	rows, err := app.db.Query(
		`SELECT id, post_slug, author, body, created_at FROM comments
		 WHERE post_slug = ? AND visible = 1 ORDER BY created_at ASC`, slug,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostSlug, &c.Author, &c.Body, &c.CreatedAt); err != nil {
			continue
		}
		c.Visible = true
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return comments
}

func (app *App) createComment(slug, author, body string) error {
	_, err := app.db.Exec(
		`INSERT INTO comments (post_slug, author, body) VALUES (?, ?, ?)`,
		slug, author, body,
	)
	return err
}

func (app *App) handleCommentSubmit(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	app.mu.RLock()
	post, postExists := findPost(app.posts, slug)
	app.mu.RUnlock()

	if !postExists || post.Private {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Honeypot — bots fill the hidden "url" field
	if r.FormValue("url") != "" {
		http.Redirect(w, r, "/posts/"+slug+"#comments", http.StatusSeeOther)
		return
	}

	author := strings.TrimSpace(r.FormValue("author"))
	body := strings.TrimSpace(r.FormValue("body"))

	if author == "" || body == "" {
		http.Error(w, "name and comment are required", http.StatusBadRequest)
		return
	}
	if len(author) > 100 || len(body) > 5000 {
		http.Error(w, "name or comment too long", http.StatusBadRequest)
		return
	}

	if app.db == nil {
		http.Error(w, "comments not available", http.StatusServiceUnavailable)
		return
	}

	if !app.limiter.allow(clientIP(r)) {
		http.Error(w, "too many comments, try again later", http.StatusTooManyRequests)
		return
	}

	if err := app.createComment(slug, author, body); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts/"+slug+"#comments", http.StatusSeeOther)
}

// rateLimiter tracks request timestamps per key. Self-cleaning on each allow() call.
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	calls    int
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    5,
		window:   time.Hour,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Periodically sweep stale keys to prevent unbounded growth
	rl.calls++
	if rl.calls%1000 == 0 {
		for k, times := range rl.requests {
			if len(times) > 0 && times[len(times)-1].Before(cutoff) {
				delete(rl.requests, k)
			}
		}
	}

	// Clean old entries for this key
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.requests[key] = recent
		return false
	}

	rl.requests[key] = append(recent, now)
	return true
}
