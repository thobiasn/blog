package blog

import (
	"bytes"
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/yuin/goldmark"
)

type App struct {
	cfg       Config
	db        *sql.DB
	posts     []Post
	pages     []Page
	projects  []Project
	tmpls     map[string]*template.Template
	md        goldmark.Markdown
	chromaCSS string
	limiter   *rateLimiter
	mu        sync.RWMutex
}

func Serve() {
	cfg := LoadConfig()
	md := newMarkdown()

	chromaCSS, err := generateChromaCSS()
	if err != nil {
		log.Fatalf("generating chroma css: %v", err)
	}

	db, err := openDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	app := &App{
		cfg:       cfg,
		db:        db,
		md:        md,
		chromaCSS: chromaCSS,
		tmpls:     parseTemplates(),
		limiter:   newRateLimiter(),
	}

	if err := app.reload(); err != nil {
		log.Fatalf("loading content: %v", err)
	}
	app.seedNotifiedPosts()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.handleHome)
	mux.HandleFunc("GET /posts", app.handlePostList)
	mux.HandleFunc("GET /posts/{slug}", app.handlePost)
	mux.HandleFunc("POST /posts/{slug}/comments", app.handleCommentSubmit)
	mux.HandleFunc("GET /projects", app.handleProjectList)
	mux.HandleFunc("GET /projects/{slug}", app.handleProject)
	mux.HandleFunc("GET /rss.xml", app.handleRSS)
	mux.HandleFunc("GET /subscribe", app.handleSubscribeForm)
	mux.HandleFunc("POST /subscribe", app.handleSubscribe)
	mux.HandleFunc("GET /subscribe/verify", app.handleSubscribeVerify)
	mux.HandleFunc("GET /subscribe/remove", app.handleSubscribeRemove)
	mux.HandleFunc("POST /deploy", app.handleDeploy)
	mux.HandleFunc("GET /uses", app.handlePage)
	mux.HandleFunc("GET /now", app.handlePage)
	mux.HandleFunc("GET /static/chroma.css", app.handleChromaCSS)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir(filepath.Join(cfg.ContentDir, "images")))))

	// Admin API
	mux.HandleFunc("GET /api/admin/stats", app.requireAdmin(app.handleAdminStats))
	mux.HandleFunc("GET /api/admin/comments", app.requireAdmin(app.handleAdminComments))
	mux.HandleFunc("POST /api/admin/comments/{id}/toggle", app.requireAdmin(app.handleAdminCommentToggle))
	mux.HandleFunc("POST /api/admin/comments/{id}/delete", app.requireAdmin(app.handleAdminCommentDelete))
	mux.HandleFunc("GET /api/admin/subscribers", app.requireAdmin(app.handleAdminSubscribers))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// SIGHUP reloads content
	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	go func() {
		for range sighup {
			log.Println("received SIGHUP, reloading content")
			if err := app.reload(); err != nil {
				log.Printf("reload error: %v", err)
			} else {
				log.Println("content reloaded")
			}
		}
	}()

	// Graceful shutdown on SIGTERM/SIGINT
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-stop
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func parseTemplates() map[string]*template.Template {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"shortDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}

	names := []string{"home", "post", "post_list", "page", "project", "project_list", "subscribe", "404"}
	tmpls := make(map[string]*template.Template, len(names))
	for _, name := range names {
		tmpls[name] = template.Must(
			template.New("base.html").Funcs(funcMap).ParseFiles(
				"templates/base.html",
				"templates/"+name+".html",
			),
		)
	}
	return tmpls
}

func (app *App) render(w http.ResponseWriter, name string, data any) {
	tmpl, ok := app.tmpls[name]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (app *App) handleChromaCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte(app.chromaCSS))
}

func (app *App) renderNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	app.tmpls["404"].ExecuteTemplate(w, "base.html", nil)
}
