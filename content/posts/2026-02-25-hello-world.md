---
title: Hello World
date: 2026-02-25
tags: [blog, go]
status: public
description: The first post on thobiasn.dev.
---

Welcome to my blog. This is the first post, mostly to verify that everything works — markdown rendering, syntax highlighting, dark mode, the whole thing.

## Why build a blog from scratch?

Because it's fun, and because I want full control over the stack. One Go binary, one SQLite database, markdown files in a git repo. No frameworks, no build steps, no JavaScript bundles.

## Code highlighting test

Here's a Go HTTP handler:

```go
func (app *App) handleHome(w http.ResponseWriter, r *http.Request) {
    posts := app.recentPosts(5)
    app.render(w, "home", map[string]any{
        "Posts": posts,
    })
}
```

And some shell:

```bash
go run . serve
curl localhost:8080
```

That's it for now. More to come.
