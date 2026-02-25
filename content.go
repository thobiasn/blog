package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/frontmatter"
)

type Post struct {
	Title       string
	Slug        string
	Description string
	Status      string
	Project     string
	Date        time.Time
	Tags        []string
	Body        template.HTML
}

type Page struct {
	Title string
	Slug  string
	Body  template.HTML
}

type postFrontmatter struct {
	Title       string   `yaml:"title"`
	Date        string   `yaml:"date"`
	Tags        []string `yaml:"tags"`
	Status      string   `yaml:"status"`
	Description string   `yaml:"description"`
	Project     string   `yaml:"project"`
}

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
			extension.Typographer,
			extension.Table,
			extension.Strikethrough,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(html.WithClasses(true)),
			),
			&anchor.Extender{Texter: anchor.Text("#")},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(),
			goldhtml.WithUnsafe(),
		),
	)
}

func generateChromaCSS() (string, error) {
	var buf bytes.Buffer
	formatter := html.New(html.WithClasses(true))

	light := styles.Get("github")
	if err := formatter.WriteCSS(&buf, light); err != nil {
		return "", fmt.Errorf("chroma light css: %w", err)
	}

	buf.WriteString("\n@media (prefers-color-scheme: dark) {\n")
	dark := styles.Get("dracula")
	if err := formatter.WriteCSS(&buf, dark); err != nil {
		return "", fmt.Errorf("chroma dark css: %w", err)
	}
	buf.WriteString("}\n")

	return buf.String(), nil
}

func loadAllPosts(dir string, md goldmark.Markdown) ([]Post, error) {
	pattern := filepath.Join(dir, "posts", "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, f := range files {
		if strings.HasSuffix(f, ".md.age") {
			continue
		}
		p, err := parsePost(f, md)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f, err)
		}
		posts = append(posts, p)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts, nil
}

func parsePost(path string, md goldmark.Markdown) (Post, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}

	ctx := parser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		return Post{}, err
	}

	var meta postFrontmatter
	fm := frontmatter.Get(ctx)
	if fm != nil {
		if err := fm.Decode(&meta); err != nil {
			return Post{}, fmt.Errorf("frontmatter: %w", err)
		}
	}

	date, _ := time.Parse("2006-01-02", meta.Date)
	slug := postSlug(filepath.Base(path))
	status := meta.Status
	if status == "" {
		status = "public"
	}

	return Post{
		Title:       meta.Title,
		Slug:        slug,
		Description: meta.Description,
		Status:      status,
		Project:     meta.Project,
		Date:        date,
		Tags:        meta.Tags,
		Body:        template.HTML(buf.String()),
	}, nil
}

// postSlug derives slug from filename: "2026-02-25-hello-world.md" → "hello-world"
func postSlug(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	// Strip YYYY-MM-DD- prefix if present
	if len(name) > 11 && name[4] == '-' && name[7] == '-' && name[10] == '-' {
		name = name[11:]
	}
	return name
}

func loadAllPages(dir string, md goldmark.Markdown) ([]Page, error) {
	pattern := filepath.Join(dir, "pages", "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var pages []Page
	for _, f := range files {
		p, err := parsePage(f, md)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f, err)
		}
		pages = append(pages, p)
	}
	return pages, nil
}

func parsePage(path string, md goldmark.Markdown) (Page, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return Page{}, err
	}

	ctx := parser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		return Page{}, err
	}

	var meta struct {
		Title string `yaml:"title"`
	}
	fm := frontmatter.Get(ctx)
	if fm != nil {
		if err := fm.Decode(&meta); err != nil {
			return Page{}, fmt.Errorf("frontmatter: %w", err)
		}
	}

	slug := strings.TrimSuffix(filepath.Base(path), ".md")

	return Page{
		Title: meta.Title,
		Slug:  slug,
		Body:  template.HTML(buf.String()),
	}, nil
}

func (app *App) reload() error {
	posts, err := loadAllPosts(app.cfg.ContentDir, app.md)
	if err != nil {
		return err
	}
	pages, err := loadAllPages(app.cfg.ContentDir, app.md)
	if err != nil {
		return err
	}

	app.mu.Lock()
	app.posts = posts
	app.pages = pages
	app.mu.Unlock()
	return nil
}
