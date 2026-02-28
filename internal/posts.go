package blog

import "net/http"

func (app *App) handleHome(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	posts := publicPosts(app.posts)
	journal := privatePosts(app.posts)
	featured := featuredProjects(app.projects)
	app.mu.RUnlock()

	limit := 5
	if len(posts) < limit {
		limit = len(posts)
	}

	data := map[string]any{
		"Posts":    posts[:limit],
		"Projects": featured,
	}

	if app.cfg.isLocal() {
		jLimit := 5
		if len(journal) < jLimit {
			jLimit = len(journal)
		}
		data["JournalPosts"] = journal[:jLimit]
	}

	app.render(w, "home", data)
}

func (app *App) handleJournal(w http.ResponseWriter, r *http.Request) {
	if !app.cfg.isLocal() {
		app.renderNotFound(w, r)
		return
	}

	app.mu.RLock()
	posts := privatePosts(app.posts)
	app.mu.RUnlock()

	app.render(w, "journal", map[string]any{
		"Posts": posts,
	})
}

func (app *App) handlePostList(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	posts := publicPosts(app.posts)
	app.mu.RUnlock()

	tag := r.URL.Query().Get("tag")
	if tag != "" {
		posts = filterByTag(posts, tag)
	}

	app.render(w, "post_list", map[string]any{
		"Posts":      posts,
		"CurrentTag": tag,
	})
}

func (app *App) handlePost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	app.mu.RLock()
	post, ok := findPost(app.posts, slug)
	app.mu.RUnlock()

	if !ok {
		app.renderNotFound(w, r)
		return
	}

	// In production, only show public posts
	if !app.cfg.isLocal() && post.Private {
		app.renderNotFound(w, r)
		return
	}

	comments := app.commentsBySlug(slug)

	app.render(w, "post", map[string]any{
		"Post":     post,
		"Comments": comments,
		"BaseURL":  app.cfg.BaseURL,
	})
}

func privatePosts(posts []Post) []Post {
	var out []Post
	for _, p := range posts {
		if p.Private {
			out = append(out, p)
		}
	}
	return out
}

func publicPosts(posts []Post) []Post {
	var out []Post
	for _, p := range posts {
		if !p.Private {
			out = append(out, p)
		}
	}
	return out
}

func filterByTag(posts []Post, tag string) []Post {
	var out []Post
	for _, p := range posts {
		for _, t := range p.Tags {
			if t == tag {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

func findPost(posts []Post, slug string) (Post, bool) {
	for _, p := range posts {
		if p.Slug == slug {
			return p, true
		}
	}
	return Post{}, false
}
