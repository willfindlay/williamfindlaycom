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

func loadBlogPosts(dir string, store *ContentStore) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("reading %s: %w", e.Name(), err)
		}

		var post BlogPost
		rendered, err := renderMarkdown(data, &post)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		post.Slug = strings.TrimSuffix(e.Name(), ".md")
		post.Content = rendered
		store.Posts = append(store.Posts, post)
	}

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
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("reading %s: %w", e.Name(), err)
		}

		var proj Project
		rendered, err := renderMarkdown(data, &proj)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		proj.Slug = strings.TrimSuffix(e.Name(), ".md")
		proj.Content = rendered
		store.Projects = append(store.Projects, proj)
	}

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
	path := filepath.Join(dir, "resume.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var buf bytes.Buffer
	if err := md.Convert(data, &buf); err != nil {
		return fmt.Errorf("rendering resume: %w", err)
	}

	store.Resume = &Resume{Content: template.HTML(buf.String())}
	return nil
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
