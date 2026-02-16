package handler

import (
	"bytes"
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

		var buf bytes.Buffer
		if err := d.Renderer.RenderFeed(&buf, data); err != nil {
			slog.Error("render error", "template", "feed.xml", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		if _, err := buf.WriteTo(w); err != nil {
			slog.Error("write error", "template", "feed.xml", "err", err)
		}
	}
}
