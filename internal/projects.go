package blog

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

type Project struct {
	Title       string
	Slug        string
	Description string
	Repo        string
	Status      string
	Featured    bool
	Tags        []string
	Body        template.HTML
}

type projectFrontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Repo        string   `yaml:"repo"`
	Status      string   `yaml:"status"`
	Featured    bool     `yaml:"featured"`
	Tags        []string `yaml:"tags"`
}

func loadAllProjects(dir string, md goldmark.Markdown) []Project {
	files, err := filepath.Glob(filepath.Join(dir, "projects", "*.md"))
	if err != nil || len(files) == 0 {
		return nil
	}

	var projects []Project
	for _, f := range files {
		p, err := parseProject(f, md)
		if err != nil {
			log.Printf("skipping project %s: %v", f, err)
			continue
		}
		projects = append(projects, p)
	}
	return projects
}

func parseProject(path string, md goldmark.Markdown) (Project, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return Project{}, err
	}

	ctx := parser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		return Project{}, err
	}

	var meta projectFrontmatter
	fm := frontmatter.Get(ctx)
	if fm != nil {
		if err := fm.Decode(&meta); err != nil {
			return Project{}, err
		}
	}

	slug := strings.TrimSuffix(filepath.Base(path), ".md")
	status := meta.Status
	if status == "" {
		status = "active"
	}

	return Project{
		Title:       meta.Title,
		Slug:        slug,
		Description: meta.Description,
		Repo:        meta.Repo,
		Status:      status,
		Featured:    meta.Featured,
		Tags:        meta.Tags,
		Body:        template.HTML(buf.String()),
	}, nil
}

func (app *App) handleProjectList(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	projects := app.projects
	app.mu.RUnlock()

	app.render(w, "project_list", map[string]any{
		"Projects": projects,
	})
}

func (app *App) handleProject(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	app.mu.RLock()
	project, ok := findProject(app.projects, slug)
	related := relatedPosts(app.posts, slug)
	app.mu.RUnlock()

	if !ok {
		app.renderNotFound(w, r)
		return
	}

	app.render(w, "project", map[string]any{
		"Project":      project,
		"RelatedPosts": related,
	})
}

func findProject(projects []Project, slug string) (Project, bool) {
	for _, p := range projects {
		if p.Slug == slug {
			return p, true
		}
	}
	return Project{}, false
}

func featuredProjects(projects []Project) []Project {
	var out []Project
	for _, p := range projects {
		if p.Featured {
			out = append(out, p)
		}
	}
	return out
}

func relatedPosts(posts []Post, projectSlug string) []Post {
	var out []Post
	for _, p := range posts {
		if p.Project == projectSlug && p.Status == "public" && !p.Private {
			out = append(out, p)
		}
	}
	return out
}
