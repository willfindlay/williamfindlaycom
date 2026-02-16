package render

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	texttemplate "text/template"
	"time"
)

type Renderer struct {
	templates map[string]*template.Template
	feedTmpl  *texttemplate.Template
}

var funcMap = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("January 2, 2006")
	},
	"formatDateShort": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"formatRFC3339": func(t time.Time) string {
		return t.Format(time.RFC3339)
	},
	"join": strings.Join,
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"truncate": func(s string, n int) string {
		runes := []rune(s)
		if len(runes) <= n {
			return s
		}
		return string(runes[:n]) + "..."
	},
	"currentYear": func() int {
		return time.Now().Year()
	},
}

var feedFuncMap = texttemplate.FuncMap{
	"formatRFC3339": func(t time.Time) string {
		return t.Format(time.RFC3339)
	},
	"xmlEscape": xmlEscape,
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

func New(fsys fs.FS) (*Renderer, error) {
	base, err := template.New("base").Funcs(funcMap).ParseFS(fsys, "templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("parsing base template: %w", err)
	}

	pages := []string{
		"templates/home.html",
		"templates/blog/list.html",
		"templates/blog/post.html",
		"templates/projects/list.html",
		"templates/projects/project.html",
		"templates/resume.html",
		"templates/404.html",
	}

	r := &Renderer{templates: make(map[string]*template.Template)}

	for _, page := range pages {
		t, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("cloning base for %s: %w", page, err)
		}
		t, err = t.ParseFS(fsys, page)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", page, err)
		}
		r.templates[page] = t
	}

	feedTmpl, err := texttemplate.New("feed.xml").Funcs(feedFuncMap).ParseFS(fsys, "templates/feed.xml")
	if err != nil {
		return nil, fmt.Errorf("parsing feed template: %w", err)
	}
	r.feedTmpl = feedTmpl

	return r, nil
}

func (r *Renderer) Render(w io.Writer, name string, data any) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}
	return t.ExecuteTemplate(w, "base", data)
}

func (r *Renderer) RenderFeed(w io.Writer, data any) error {
	return r.feedTmpl.ExecuteTemplate(w, "feed", data)
}
