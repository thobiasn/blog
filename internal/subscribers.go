package blog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (app *App) handleSubscribeForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, "subscribe", map[string]any{
		"ShowForm": true,
	})
}

func (app *App) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if app.db == nil {
		http.Error(w, "subscriptions not available", http.StatusServiceUnavailable)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4*1024)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if !emailRe.MatchString(email) || len(email) > 254 {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}

	// Rate limit by IP
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip == "" {
		ip = r.RemoteAddr
	}
	if !app.limiter.allow("sub:" + ip) {
		http.Error(w, "too many requests, try again later", http.StatusTooManyRequests)
		return
	}

	verifyToken := generateToken()
	unsubToken := generateToken()

	// INSERT OR IGNORE: don't leak whether email already exists
	_, err := app.db.Exec(
		`INSERT OR IGNORE INTO subscribers (email, verify_token, unsubscribe_token) VALUES (?, ?, ?)`,
		email, verifyToken, unsubToken,
	)
	if err != nil {
		log.Printf("subscriber insert error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Send verification email (only if newly inserted — check token matches)
	var storedToken string
	app.db.QueryRow(`SELECT verify_token FROM subscribers WHERE email = ?`, email).Scan(&storedToken)
	if storedToken == verifyToken && app.cfg.smtpConfigured() {
		link := app.cfg.BaseURL + "/subscribe/verify?token=" + verifyToken
		body := fmt.Sprintf("Verify your subscription to thobiasn.dev:\n\n%s\n\nIf you didn't subscribe, ignore this email.", link)
		if err := app.sendMail(email, "Verify your subscription", body); err != nil {
			log.Printf("verification email error: %v", err)
		}
	}

	app.render(w, "subscribe", map[string]any{
		"CheckEmail": true,
	})
}

func (app *App) handleSubscribeVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" || app.db == nil {
		app.renderNotFound(w, r)
		return
	}

	res, err := app.db.Exec(`UPDATE subscribers SET verified = 1, verify_token = '' WHERE verify_token = ?`, token)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		app.renderNotFound(w, r)
		return
	}

	app.render(w, "subscribe", map[string]any{
		"Verified": true,
	})
}

func (app *App) handleSubscribeRemove(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" || app.db == nil {
		app.renderNotFound(w, r)
		return
	}

	res, err := app.db.Exec(`DELETE FROM subscribers WHERE unsubscribe_token = ?`, token)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		app.renderNotFound(w, r)
		return
	}

	app.render(w, "subscribe", map[string]any{
		"Unsubscribed": true,
	})
}

func (app *App) handleAdminSubscribers(w http.ResponseWriter, r *http.Request) {
	var total, verified int
	if app.db != nil {
		app.db.QueryRow(`SELECT COUNT(*) FROM subscribers`).Scan(&total)
		app.db.QueryRow(`SELECT COUNT(*) FROM subscribers WHERE verified = 1`).Scan(&verified)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"total":%d,"verified":%d}`, total, verified)
}

func (app *App) notifyNewPosts(posts []Post) {
	if app.db == nil || !app.cfg.smtpConfigured() {
		return
	}

	for _, p := range posts {
		var exists bool
		app.db.QueryRow(`SELECT 1 FROM notified_posts WHERE slug = ?`, p.Slug).Scan(&exists)
		if exists {
			continue
		}

		// Mark as notified first to prevent duplicates on retry
		app.db.Exec(`INSERT OR IGNORE INTO notified_posts (slug) VALUES (?)`, p.Slug)

		app.emailSubscribers(p)
	}
}

func (app *App) emailSubscribers(p Post) {
	rows, err := app.db.Query(`SELECT email, unsubscribe_token FROM subscribers WHERE verified = 1`)
	if err != nil {
		log.Printf("querying subscribers: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var email, unsubToken string
		if err := rows.Scan(&email, &unsubToken); err != nil {
			continue
		}
		link := app.cfg.BaseURL + "/posts/" + p.Slug
		unsub := app.cfg.BaseURL + "/subscribe/remove?token=" + unsubToken
		body := fmt.Sprintf("New post on thobiasn.dev: %s\n\nRead it here: %s\n\nUnsubscribe: %s", p.Title, link, unsub)
		if err := app.sendMail(email, "New post: "+p.Title, body); err != nil {
			log.Printf("notification email to %s: %v", email, err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("iterating subscribers for %s: %v", p.Slug, err)
	}
}

func (app *App) seedNotifiedPosts() {
	if app.db == nil {
		return
	}

	app.mu.RLock()
	posts := publicPosts(app.posts)
	app.mu.RUnlock()

	for _, p := range posts {
		app.db.Exec(`INSERT OR IGNORE INTO notified_posts (slug) VALUES (?)`, p.Slug)
	}
}
