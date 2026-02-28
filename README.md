# thobiasn.dev

A personal blog, diary, and portfolio. Single Go binary, SQLite, markdown files in git. Private posts transparently encrypted with git-crypt.

## Features

- Markdown posts with YAML frontmatter, rendered server-side with syntax highlighting
- Private/diary posts encrypted at rest via git-crypt
- Draft posts visible locally, hidden in production
- Full-text search (SQLite FTS5)
- RSS feed
- Email subscribers with auto-notify on new posts
- Comments with admin moderation (CLI-based, no web auth)
- Projects section with post cross-linking
- Dark mode (respects `prefers-color-scheme`)
- Deploy webhook (git pull + content reload)
- Litestream backups for SQLite

## Quick Start

```bash
git clone https://github.com/thobiasn/blog.git
cd blog
cp .env.example .env
go run ./cmd/blog serve
# open http://localhost:8080
```

## CLI

```
blog serve                          start HTTP server
blog new post [--private] <title>   create a new post
blog new project <name>             create a new project
blog dash                           admin dashboard
blog comments                       list recent comments
blog comments delete <id>           delete a comment
blog comments toggle <id>           toggle comment visibility
blog subscribers                    subscriber stats
```

Admin commands (`dash`, `comments`, `subscribers`) talk to a remote server. Set `BLOG_URL` and `ADMIN_API_KEY` in your environment.

## Content

Posts are markdown files with YAML frontmatter:

```markdown
---
title: Hello World
date: 2026-02-25
tags: [blog, go]
status: public
description: A short description for previews and OG tags.
---

Your post content here.
```

| Directory | Purpose |
|---|---|
| `content/posts/` | Public posts |
| `content/private/` | Private posts (encrypted by git-crypt) |
| `content/projects/` | Project pages |
| `content/pages/` | Static pages (uses, now) |

**Status values:** `public` (listed everywhere), `draft` (visible locally only).

**Private posts:** Install [git-crypt](https://github.com/AGWA/git-crypt) and unlock the repo to read/write private posts. Without the key, the server skips them.

## Configuration

All config via environment variables. See `.env.example` for the full list.

| Variable | Default | Required for |
|---|---|---|
| `PORT` | `8080` | - |
| `BASE_URL` | `http://localhost:8080` | - |
| `CONTENT_DIR` | `content` | - |
| `DB_PATH` | `blog.db` | - |
| `BLOG_URL` | - | remote CLI |
| `ADMIN_API_KEY` | - | remote CLI |
| `SMTP_HOST` | - | email |
| `SMTP_PORT` | `587` | email |
| `SMTP_USERNAME` | - | email |
| `SMTP_PASSWORD` | - | email |
| `FROM_EMAIL` | - | email |
| `DEPLOY_WEBHOOK_SECRET` | - | webhook |

Only features you configure will activate. The server runs fine with just the defaults.

## Deployment

The included Dockerfile builds a minimal image with [Litestream](https://litestream.io/) for continuous SQLite backups to S3-compatible storage.

```bash
docker build -t blog .
docker run -p 8080:8080 blog
```

To enable Litestream backups, set `LITESTREAM_REPLICA_BUCKET` and configure `litestream.yml`. The entrypoint automatically restores from the replica on startup and replicates while running.

Content reloads on `SIGHUP` or via the deploy webhook (`DEPLOY_WEBHOOK_SECRET`).

## License

MIT
