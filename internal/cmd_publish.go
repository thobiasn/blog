package blog

import (
	"fmt"
	"os"
	"path/filepath"
)

func Publish(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: blog publish <slug>")
		os.Exit(1)
	}

	slug := args[0]

	files, _ := filepath.Glob(filepath.Join("content", "private", "*.md"))
	var src string
	for _, f := range files {
		if postSlug(filepath.Base(f)) == slug {
			src = f
			break
		}
	}
	if src == "" {
		fmt.Fprintf(os.Stderr, "no private post found with slug %q\n", slug)
		os.Exit(1)
	}

	dst := filepath.Join("content", "posts", filepath.Base(src))
	os.MkdirAll(filepath.Join("content", "posts"), 0o755)

	if err := os.Rename(src, dst); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("published %s → %s\n", src, dst)
}
