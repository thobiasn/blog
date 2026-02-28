# thobiasn.dev

A personal blog, diary, and portfolio. Single Go binary, SQLite, markdown files in git. Private posts transparently encrypted with git-crypt.

## Features

- Markdown posts with YAML frontmatter, rendered server-side with syntax highlighting
- Private/diary posts encrypted at rest via git-crypt (visible locally, hidden in production)
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
blog new post <title>               create a new post (in content/private/)
blog new project <name>             create a new project
blog publish <slug>                 move post from private to public
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

New posts are created in `content/private/` and moved to `content/posts/` with `blog publish <slug>`.

## Private Posts

Private posts in `content/private/` are transparently encrypted by [git-crypt](https://github.com/AGWA/git-crypt). They're plaintext locally and encrypted in the remote repo. The server skips them if it doesn't have the key.

**Setup:**

```bash
brew install git-crypt   # or your package manager
git-crypt init
```

Add a `.gitattributes` file to define what gets encrypted:

```
content/private/** filter=git-crypt diff=git-crypt
```

Then write private posts normally — git-crypt encrypts on push and decrypts on pull.

**Back up the key:** The key lives only in your local `.git/` directory. If you lose it, encrypted posts in the remote repo are unrecoverable. Export a base64-encoded copy for your password manager:

```bash
git-crypt export-key /dev/stdout | base64
```

**On a new machine:** Decode the key and unlock:

```bash
echo "<pasted string>" | base64 -d > /tmp/git-crypt-key
git-crypt unlock /tmp/git-crypt-key
rm /tmp/git-crypt-key
```

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
