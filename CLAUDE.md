# thobiasn.dev — Blog

Personal blog, diary, and portfolio. Single Go binary, SQLite, markdown files in a public repo. Private posts transparently encrypted via git-crypt.

## Development Philosophy

Code is a liability, not an asset. Every line we write is a line we have to maintain, debug, and understand. The goal is always the least code that solves the actual problem.

- **If we're writing more than ~200 lines for a single feature, we're probably approaching it wrong.** Step back and rethink the design before writing more code.
- **If a feature requires brittle code or lots of edge case handling, the mental model is wrong.** Redesign the approach so edge cases don't exist rather than handling them. The best error handling is making errors impossible by construction.
- **Favor using what already exists.** Go's stdlib has `net/http`, `html/template`, `crypto`, and `database/sql`. Goldmark handles markdown. Chroma handles syntax highlighting. git-crypt handles encryption. SQLite handles persistence. Use them. Don't rebuild what's already there.
- **Don't over-engineer.** Build for what's needed now. No premature abstractions, no speculative generality, no "we might need this later." YAGNI. This is a personal blog, not a platform.
- **When fixing a bug, find the root cause first.** Don't patch symptoms. If a fix feels hacky, the real problem is probably somewhere else.
- **If a function needs more than 3 parameters, it needs a different design.** Same for deeply nested conditionals, long switch statements, and functions longer than ~40 lines.
- **Duplication is cheaper than the wrong abstraction.** Only extract when you see the pattern clearly after 2-3 instances, never preemptively.
- **Delete aggressively.** Dead code, commented-out code, unused imports, stale TODO comments — remove them. Git remembers.
- **One binary, one database, one repo.** Resist the urge to split things up. A single Go binary serving HTML, handling the API, and providing CLI commands is a feature, not a limitation.
- **Server-rendered HTML is the default.** Only reach for client-side JS when there's no server-side alternative. A `<form>` with a POST is almost always enough.

## Tech Stack

- **Go** — single binary, stdlib HTTP server
- **SQLite** — comments, subscribers, search (FTS5)
- **Markdown files** — content in `content/`, versioned in git
- **git-crypt** — transparent encryption for private posts (files in `content/private/`)
- **Goldmark** + **Chroma** — markdown rendering with server-side syntax highlighting
- **SMTP** (`net/smtp`) — transactional email for subscriber notifications, no SDK needed
- **Dokploy** — deployment, SSL, reverse proxy

## Project Structure

```
cmd/blog/main.go     # entrypoint, CLI dispatch
internal/            # all application code (package blog)
  config.go          # Config struct, load from env vars, validate on use
  serve.go           # HTTP server setup, routes, graceful shutdown
  content.go         # markdown/frontmatter parsing, content loading + reload
  posts.go           # post handlers (listing, single post, tag filtering)
  pages.go           # /uses, /now, static page handlers
  projects.go        # project handlers + cross-linking
  comments.go        # comment submission + display
  admin.go           # admin API endpoints (comment moderation, stats)
  subscribers.go     # email subscribe/verify/unsubscribe + auto-notify
  mail.go            # SMTP email sending via net/smtp (stdlib)
  feed.go            # RSS feed
  search.go          # FTS5 search
  deploy.go          # webhook: git pull + reload
  db.go              # SQLite init + migrations
  cmd_new.go         # CLI: blog new post/project
  cmd_publish.go     # CLI: blog publish <slug>
  cmd_admin.go       # CLI: blog comments, blog subscribers, blog (dashboard)
content/             # markdown content
  posts/             # public posts
  private/           # private/diary posts (encrypted by git-crypt)
  projects/          # project pages
  pages/             # static pages (uses, now)
  images/            # post images
templates/           # Go html/template files
static/              # CSS and static assets
```

All application code lives in `internal/` as `package blog`. The `cmd/blog/main.go` entrypoint is a thin wrapper that dispatches CLI commands.

## Key Concepts

- **Post types:** public (`content/posts/`), private (`content/private/`) — directory is the status, no frontmatter field needed
- **Encryption is transparent** — git-crypt encrypts `content/private/` in git, plaintext locally. No application-level encryption code needed
- **Server without git-crypt key** skips private posts (they're binary blobs). With the key (local dev), all posts render
- **Workflow:** new posts start in `content/private/`, publish with `blog publish <slug>` to move to `content/posts/`
- **No web auth** — admin tasks handled via CLI against the remote server's API
- **Content loaded at startup** into memory. `Reload()` called by deploy webhook and `SIGHUP`
- **Auto-notify:** on content reload, new public posts automatically trigger subscriber emails
- **Content workflow:** write markdown locally → git push → webhook triggers git pull + reload
- **blog.db** is gitignored — SQLite only lives on the server

## Commands

```bash
# Server
go run ./cmd/blog serve                 # start HTTP server

# Content creation
go run ./cmd/blog new post "Title"              # create in content/private/
go run ./cmd/blog new project "Name"            # create in content/projects/
go run ./cmd/blog publish <slug>                # move from content/private/ to content/posts/

# Admin (uses BLOG_URL + ADMIN_API_KEY env vars)
go run ./cmd/blog                       # dashboard overview
go run ./cmd/blog comments              # list recent comments
go run ./cmd/blog comments delete <id>  # delete comment
go run ./cmd/blog comments toggle <id>  # toggle visibility
go run ./cmd/blog subscribers           # subscriber stats

# subscriber emails sent automatically on content reload when new public posts detected
```

## Development

```bash
go run ./cmd/blog serve     # run locally on :8080 (shows all posts including private)
go test ./...               # run all tests
```

## Configuration

All config via environment variables, loaded into a `Config` struct in `config.go` at startup.
Features only require their config when actually used — the server boots fine with just defaults.

| Variable | Default | Required for |
|---|---|---|
| `PORT` | `8080` | — |
| `BASE_URL` | `http://localhost:8080` | — |
| `CONTENT_DIR` | `content` | — |
| `DB_PATH` | `blog.db` | — |
| `BLOG_URL` | — | remote CLI |
| `ADMIN_API_KEY` | — | remote CLI |
| `SMTP_HOST` | — | email |
| `SMTP_PORT` | `587` | email |
| `SMTP_USERNAME` | — | email |
| `SMTP_PASSWORD` | — | email |
| `FROM_EMAIL` | — | email |
| `DEPLOY_WEBHOOK_SECRET` | — | webhook |

## Conventions

- Use stdlib `net/http` (no frameworks)
- Go html/template for all rendering — no client-side JS unless explicitly needed
- Frontmatter parsed from YAML before markdown body
- CSS is minimal (~300 lines), respects `prefers-color-scheme` for dark mode
- All routes defined in `serve.go` using `http.ServeMux` (Go 1.22+ patterns)
- Graceful shutdown via `SIGTERM`/`SIGINT` handling
- Content cached in memory at startup, reloaded via `SIGHUP` or deploy webhook
- Errors returned, not panicked — handle at the handler level
- Tests use stdlib `testing` package

## Implementation Plan

See `docs/` directory (gitignored) for specs and plans:

- `docs/blog-tech-spec.md` — full technical specification
- `docs/plan/phase-1-ship-it.md` — Go server, markdown rendering, post pages, CSS, Dockerfile
- `docs/plan/phase-2-core-features.md` — git-crypt, CLI commands, comments, projects, remote admin
- `docs/plan/phase-3-distribution.md` — RSS, email subscribers, notifications, OG tags
- `docs/plan/phase-4-polish.md` — Search, pages, auto-deploy, backups, monitoring
