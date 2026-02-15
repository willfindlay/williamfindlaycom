package handler

import (
	"log/slog"
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/content"
	"github.com/willfindlay/williamfindlaycom/internal/render"
)

type Deps struct {
	Store     *content.AtomicStore
	Renderer  *render.Renderer
	SiteTitle string
	SiteURL   string
}

type PageData struct {
	SiteTitle    string
	SiteURL      string
	PageTitle    string
	Description  string
	CanonicalURL string
	ActiveNav    string
}

func (d *Deps) basePage(activeNav string) PageData {
	return PageData{
		SiteTitle: d.SiteTitle,
		SiteURL:   d.SiteURL,
		ActiveNav: activeNav,
	}
}

func (d *Deps) render(w http.ResponseWriter, tmpl string, data any) {
	if err := d.Renderer.Render(w, tmpl, data); err != nil {
		slog.Error("render error", "template", tmpl, "err", err)
	}
}

func (d *Deps) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := d.basePage("")
	data.PageTitle = "Not Found"
	data.Description = "Page not found"
	d.render(w, "templates/404.html", data)
}
