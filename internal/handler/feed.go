package handler

import (
	"log/slog"
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type feedData struct {
	SiteTitle string
	SiteURL   string
	Posts     []content.BlogPost
}

func (d *Deps) Feed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		data := feedData{
			SiteTitle: d.SiteTitle,
			SiteURL:   d.SiteURL,
		}

		if store != nil {
			data.Posts = store.Posts
		}

		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		if err := d.Renderer.RenderFeed(w, data); err != nil {
			slog.Error("render error", "template", "feed.xml", "err", err)
		}
	}
}
