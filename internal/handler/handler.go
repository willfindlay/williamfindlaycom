package handler

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/config"
	"github.com/willfindlay/williamfindlaycom/internal/content"
	"github.com/willfindlay/williamfindlaycom/internal/render"
)

type Deps struct {
	Store     *content.AtomicStore
	Renderer  *render.Renderer
	SiteTitle string
	SiteURL   string
	Particles config.ParticleConfig
}

type PageData struct {
	SiteTitle    string
	SiteURL      string
	PageTitle    string
	Description  string
	CanonicalURL string
	OGType       string
	OGImage      string
	Author       string
	JSONLD       template.HTML
	ActiveNav    string
	Particles    config.ParticleConfig
}

func (d *Deps) basePage(activeNav string) PageData {
	return PageData{
		SiteTitle: d.SiteTitle,
		SiteURL:   d.SiteURL,
		OGType:    "website",
		OGImage:   d.SiteURL + "/static/og-image.png",
		Author:    "William Findlay",
		ActiveNav: activeNav,
		Particles: d.Particles,
	}
}

func (d *Deps) render(w http.ResponseWriter, tmpl string, data any) {
	var buf bytes.Buffer
	if err := d.Renderer.Render(&buf, tmpl, data); err != nil {
		slog.Error("render error", "template", tmpl, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("write error", "template", tmpl, "err", err)
	}
}

func (d *Deps) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := d.basePage("")
	data.PageTitle = "Not Found"
	data.Description = "Page not found"
	d.render(w, "templates/404.html", data)
}
