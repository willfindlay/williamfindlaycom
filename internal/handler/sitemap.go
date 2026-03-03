package handler

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type sitemapData struct {
	SiteURL  string
	Posts    []content.BlogPost
	Projects []content.Project

	BlogLastmod     time.Time
	ProjectsLastmod time.Time
	GlobalLastmod   time.Time
}

func (d *Deps) Sitemap() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		data := sitemapData{
			SiteURL: d.SiteURL,
		}

		if store != nil {
			data.Posts = store.Posts
			data.Projects = store.Projects

			if len(store.Posts) > 0 {
				data.BlogLastmod = store.Posts[0].Date
				data.GlobalLastmod = store.Posts[0].Date
			}
			if len(store.Projects) > 0 {
				data.ProjectsLastmod = store.Projects[0].Date
				if data.ProjectsLastmod.After(data.GlobalLastmod) {
					data.GlobalLastmod = data.ProjectsLastmod
				}
			}
		}

		var buf bytes.Buffer
		if err := d.Renderer.RenderSitemap(&buf, data); err != nil {
			slog.Error("render error", "template", "sitemap.xml", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		if _, err := buf.WriteTo(w); err != nil {
			slog.Error("write error", "template", "sitemap.xml", "err", err)
		}
	}
}
