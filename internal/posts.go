package blog

import "net/http"

func (app *App) handleHome(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	posts := publicPosts(app.posts)
	featured := featuredProjects(app.projects)
	app.mu.RUnlock()

	limit := 5
	if len(posts) < limit {
		limit = len(posts)
	}

	app.render(w, "home", map[string]any{
		"Posts":    posts[:limit],
		"Projects": featured,
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

	// In production, only show public non-private posts
	if !app.cfg.isLocal() && (post.Status != "public" || post.Private) {
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

func publicPosts(posts []Post) []Post {
	var out []Post
	for _, p := range posts {
		if p.Status == "public" && !p.Private {
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
