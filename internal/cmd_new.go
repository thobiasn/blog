package blog

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func New(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: blog new [post|project] <title>")
		os.Exit(1)
	}

	switch args[0] {
	case "post":
		newPost(args[1:])
	case "project":
		newProject(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown type: %s\nusage: blog new [post|project] <title>\n", args[0])
		os.Exit(1)
	}
}

func newPost(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: blog new post <title>")
		os.Exit(1)
	}

	title := args[0]
	slug := slugify(title)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)

	dir := filepath.Join("content", "private")
	os.MkdirAll(dir, 0o755)

	path := filepath.Join(dir, filename)
	content := fmt.Sprintf(`---
title: %q
date: %s
tags: []
description: ""
---

`, title, date)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", path)
	openEditor(path)
}

func newProject(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: blog new project <name>")
		os.Exit(1)
	}

	title := args[0]
	slug := slugify(title)
	dir := filepath.Join("content", "projects")
	os.MkdirAll(dir, 0o755)

	path := filepath.Join(dir, slug+".md")
	content := fmt.Sprintf(`---
title: %q
description: ""
repo: ""
status: active
featured: false
tags: []
---

`, title)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", path)
	openEditor(path)
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphaNum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func openEditor(path string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
