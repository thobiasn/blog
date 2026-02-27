package blog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostSlug(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"2026-02-25-hello-world.md", "hello-world"},
		{"hello-world.md", "hello-world"},
		{"2026-02-25-a.md", "a"},
		{"no-date-prefix.md", "no-date-prefix"},
		{"short.md", "short"},
		{"2026-12-31-year-end-review.md", "year-end-review"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := postSlug(tt.filename)
			if got != tt.want {
				t.Errorf("postSlug(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestParsePost(t *testing.T) {
	dir := t.TempDir()
	content := `---
title: Test Post
date: 2026-02-25
tags: [go, web]
description: A test post
project: blog
---

Hello **world**.
`
	path := filepath.Join(dir, "2026-02-25-test-post.md")
	os.WriteFile(path, []byte(content), 0644)

	md := newMarkdown()
	post, err := parsePost(path, md)
	if err != nil {
		t.Fatalf("parsePost: %v", err)
	}

	if post.Title != "Test Post" {
		t.Errorf("Title = %q, want %q", post.Title, "Test Post")
	}
	if post.Slug != "test-post" {
		t.Errorf("Slug = %q, want %q", post.Slug, "test-post")
	}
	if post.Description != "A test post" {
		t.Errorf("Description = %q, want %q", post.Description, "A test post")
	}
	if post.Status != "public" {
		t.Errorf("Status = %q, want %q", post.Status, "public")
	}
	if post.Project != "blog" {
		t.Errorf("Project = %q, want %q", post.Project, "blog")
	}
	if len(post.Tags) != 2 || post.Tags[0] != "go" || post.Tags[1] != "web" {
		t.Errorf("Tags = %v, want [go web]", post.Tags)
	}
	if !strings.Contains(string(post.Body), "<strong>world</strong>") {
		t.Errorf("Body should contain rendered markdown, got %q", post.Body)
	}
}

func TestParsePostDefaultStatus(t *testing.T) {
	dir := t.TempDir()
	content := `---
title: No Status
date: 2026-01-01
---

Content.
`
	path := filepath.Join(dir, "2026-01-01-no-status.md")
	os.WriteFile(path, []byte(content), 0644)

	md := newMarkdown()
	post, err := parsePost(path, md)
	if err != nil {
		t.Fatalf("parsePost: %v", err)
	}

	if post.Status != "public" {
		t.Errorf("Status = %q, want %q (default)", post.Status, "public")
	}
}

func TestLoadAllPosts(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0755)
	os.MkdirAll(filepath.Join(dir, "private"), 0755)

	// Older post
	os.WriteFile(filepath.Join(dir, "posts", "2026-01-01-older.md"), []byte(`---
title: Older
date: 2026-01-01
---
Old content.
`), 0644)

	// Newer post
	os.WriteFile(filepath.Join(dir, "posts", "2026-02-01-newer.md"), []byte(`---
title: Newer
date: 2026-02-01
---
New content.
`), 0644)

	// Private post
	os.WriteFile(filepath.Join(dir, "private", "2026-01-15-secret.md"), []byte(`---
title: Secret
date: 2026-01-15
---
Secret content.
`), 0644)

	md := newMarkdown()
	posts, err := loadAllPosts(dir, md)
	if err != nil {
		t.Fatalf("loadAllPosts: %v", err)
	}

	if len(posts) != 3 {
		t.Fatalf("got %d posts, want 3", len(posts))
	}

	// Should be sorted newest-first
	if posts[0].Slug != "newer" {
		t.Errorf("posts[0].Slug = %q, want %q (newest first)", posts[0].Slug, "newer")
	}

	// Private post should have Private flag
	var found bool
	for _, p := range posts {
		if p.Slug == "secret" {
			found = true
			if !p.Private {
				t.Error("secret post should be Private")
			}
		}
	}
	if !found {
		t.Error("secret post not found")
	}
}

func TestLoadAllPostsSkipsUnreadable(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "posts"), 0755)
	os.MkdirAll(filepath.Join(dir, "private"), 0755)

	// Valid post
	os.WriteFile(filepath.Join(dir, "posts", "2026-01-01-valid.md"), []byte(`---
title: Valid
date: 2026-01-01
---
Content.
`), 0644)

	// Unreadable file (simulates permission error or broken file)
	path := filepath.Join(dir, "private", "2026-01-01-broken.md")
	os.WriteFile(path, []byte("data"), 0644)
	os.Chmod(path, 0000)
	t.Cleanup(func() { os.Chmod(path, 0644) })

	md := newMarkdown()
	posts, err := loadAllPosts(dir, md)
	if err != nil {
		t.Fatalf("loadAllPosts: %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("got %d posts, want 1 (unreadable should be skipped)", len(posts))
	}
	if posts[0].Slug != "valid" {
		t.Errorf("posts[0].Slug = %q, want %q", posts[0].Slug, "valid")
	}
}

func TestParsePage(t *testing.T) {
	dir := t.TempDir()
	content := `---
title: Uses
---

My **tools**.
`
	path := filepath.Join(dir, "uses.md")
	os.WriteFile(path, []byte(content), 0644)

	md := newMarkdown()
	page, err := parsePage(path, md)
	if err != nil {
		t.Fatalf("parsePage: %v", err)
	}

	if page.Title != "Uses" {
		t.Errorf("Title = %q, want %q", page.Title, "Uses")
	}
	if page.Slug != "uses" {
		t.Errorf("Slug = %q, want %q", page.Slug, "uses")
	}
	if !strings.Contains(string(page.Body), "<strong>tools</strong>") {
		t.Errorf("Body should contain rendered markdown, got %q", page.Body)
	}
}

func TestParseProject(t *testing.T) {
	dir := t.TempDir()
	content := `---
title: My Project
description: A cool project
repo: https://github.com/user/repo
featured: true
tags: [go, cli]
---

Project **description**.
`
	path := filepath.Join(dir, "my-project.md")
	os.WriteFile(path, []byte(content), 0644)

	md := newMarkdown()
	project, err := parseProject(path, md)
	if err != nil {
		t.Fatalf("parseProject: %v", err)
	}

	if project.Title != "My Project" {
		t.Errorf("Title = %q, want %q", project.Title, "My Project")
	}
	if project.Slug != "my-project" {
		t.Errorf("Slug = %q, want %q", project.Slug, "my-project")
	}
	if project.Status != "active" {
		t.Errorf("Status = %q, want %q (default)", project.Status, "active")
	}
	if !project.Featured {
		t.Error("Featured should be true")
	}
	if project.Repo != "https://github.com/user/repo" {
		t.Errorf("Repo = %q, want github URL", project.Repo)
	}
	if !strings.Contains(string(project.Body), "<strong>description</strong>") {
		t.Errorf("Body should contain rendered markdown, got %q", project.Body)
	}
}
