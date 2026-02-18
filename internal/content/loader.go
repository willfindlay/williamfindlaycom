package content

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
	"gopkg.in/yaml.v3"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		&frontmatter.Extender{},
		highlighting.NewHighlighting(
			highlighting.WithStyle("dracula"),
			highlighting.WithFormatOptions(),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

func LoadFromDir(dir string) (*ContentStore, error) {
	store := &ContentStore{
		PostsBySlug:    make(map[string]*BlogPost),
		PostsByTag:     make(map[string][]*BlogPost),
		ProjectsBySlug: make(map[string]*Project),
	}

	if err := loadBlogPosts(filepath.Join(dir, "blog"), store); err != nil {
		return nil, fmt.Errorf("loading blog posts: %w", err)
	}

	if err := loadProjects(filepath.Join(dir, "projects"), store); err != nil {
		return nil, fmt.Errorf("loading projects: %w", err)
	}

	if err := loadResume(filepath.Join(dir, "resume"), store); err != nil {
		return nil, fmt.Errorf("loading resume: %w", err)
	}

	return store, nil
}

func loadMarkdownDir[T any](dir string, decode func([]byte, string) (T, error)) ([]T, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []T
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", e.Name(), err)
		}

		slug := strings.TrimSuffix(e.Name(), ".md")
		item, err := decode(data, slug)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		items = append(items, item)
	}

	return items, nil
}

func loadBlogPosts(dir string, store *ContentStore) error {
	posts, err := loadMarkdownDir(dir, func(data []byte, slug string) (BlogPost, error) {
		var post BlogPost
		rendered, err := renderMarkdown(data, &post)
		if err != nil {
			return post, err
		}
		post.Slug = slug
		post.Content = rendered
		return post, nil
	})
	if err != nil {
		return err
	}

	store.Posts = posts
	sort.Slice(store.Posts, func(i, j int) bool {
		return store.Posts[i].Date.After(store.Posts[j].Date)
	})

	for i := range store.Posts {
		p := &store.Posts[i]
		store.PostsBySlug[p.Slug] = p
		for _, tag := range p.Tags {
			store.PostsByTag[tag] = append(store.PostsByTag[tag], p)
		}
	}

	return nil
}

func loadProjects(dir string, store *ContentStore) error {
	projects, err := loadMarkdownDir(dir, func(data []byte, slug string) (Project, error) {
		var proj Project
		rendered, err := renderMarkdown(data, &proj)
		if err != nil {
			return proj, err
		}
		proj.Slug = slug
		proj.Content = rendered
		return proj, nil
	})
	if err != nil {
		return err
	}

	store.Projects = projects
	sort.Slice(store.Projects, func(i, j int) bool {
		return store.Projects[i].Date.After(store.Projects[j].Date)
	})

	for i := range store.Projects {
		p := &store.Projects[i]
		store.ProjectsBySlug[p.Slug] = p
	}

	return nil
}

func loadResume(dir string, store *ContentStore) error {
	path := filepath.Join(dir, "resume.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var resume Resume
	if err := yaml.Unmarshal(data, &resume); err != nil {
		return fmt.Errorf("parsing resume YAML: %w", err)
	}

	// Render summary inline markdown.
	resume.Summary = renderInlineMarkdown(resume.RawSummary)

	// Compute date ranges and render bullets for entry-based sections.
	for i := range resume.Experience {
		renderEntry(&resume.Experience[i])
	}
	for i := range resume.Education {
		renderEntry(&resume.Education[i])
	}
	for i := range resume.Research {
		renderEntry(&resume.Research[i])
	}

	// Render presentation venues and dates.
	for i := range resume.Presentations {
		resume.Presentations[i].Venue = renderInlineMarkdown(resume.Presentations[i].RawVenue)
		resume.Presentations[i].DateFormatted = resume.Presentations[i].Date.FormatDate()
	}

	// Render publication items as inline markdown.
	for i := range resume.Publications {
		sec := &resume.Publications[i]
		sec.Items = make([]template.HTML, len(sec.RawItems))
		for j, raw := range sec.RawItems {
			sec.Items[j] = renderInlineMarkdown(raw)
		}
	}

	// Render open source bullets as inline markdown.
	for i := range resume.OpenSource {
		for j := range resume.OpenSource[i].Projects {
			p := &resume.OpenSource[i].Projects[j]
			p.Bullets = make([]template.HTML, len(p.RawBullets))
			for k, raw := range p.RawBullets {
				p.Bullets[k] = renderInlineMarkdown(raw)
			}
		}
	}

	store.Resume = &resume
	return nil
}

func renderEntry(e *ResumeEntry) {
	e.DateRange = FormatDateRange(e.Start, e.End)
	renderBullets(e.Bullets)
}

func renderBullets(bullets []ResumeBullet) {
	for i := range bullets {
		b := &bullets[i]
		b.Text = renderInlineMarkdown(b.RawText)
		b.Sub = make([]ResumeBullet, len(b.RawSub))
		copy(b.Sub, b.RawSub)
		renderBullets(b.Sub)
	}
}

// renderInlineMarkdown converts a markdown string to HTML, stripping all
// <p> tags that goldmark adds so the result can be used inline.
func renderInlineMarkdown(s string) template.HTML {
	if s == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(s))
	}
	out := strings.TrimSpace(buf.String())
	out = strings.ReplaceAll(out, "<p>", "")
	out = strings.ReplaceAll(out, "</p>", "")
	return template.HTML(strings.TrimSpace(out))
}

func renderMarkdown(src []byte, meta any) (template.HTML, error) {
	ctx := parser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		return "", err
	}

	d := frontmatter.Get(ctx)
	if d != nil {
		if err := d.Decode(meta); err != nil {
			return "", fmt.Errorf("decoding frontmatter: %w", err)
		}
	}

	return template.HTML(buf.String()), nil
}
