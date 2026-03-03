package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type jsonFeed struct {
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	HomePageURL string         `json:"home_page_url"`
	FeedURL     string         `json:"feed_url"`
	Items       []jsonFeedItem `json:"items"`
}

type jsonFeedItem struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	ContentHTML   string   `json:"content_html"`
	Summary       string   `json:"summary,omitempty"`
	DatePublished string   `json:"date_published"`
	Tags          []string `json:"tags,omitempty"`
}

func (d *Deps) JSONFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		feed := jsonFeed{
			Version:     "https://jsonfeed.org/version/1.1",
			Title:       d.SiteTitle,
			HomePageURL: d.SiteURL + "/",
			FeedURL:     d.SiteURL + "/feed.json",
		}

		if store != nil {
			feed.Items = make([]jsonFeedItem, len(store.Posts))
			for i, p := range store.Posts {
				feed.Items[i] = jsonFeedItem{
					ID:            d.SiteURL + "/blog/" + p.Slug,
					URL:           d.SiteURL + "/blog/" + p.Slug,
					Title:         p.Title,
					ContentHTML:   string(p.Content),
					Summary:       p.Description,
					DatePublished: p.Date.Format(time.RFC3339),
					Tags:          p.Tags,
				}
			}
		}

		w.Header().Set("Content-Type", "application/feed+json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(feed); err != nil {
			slog.Error("json feed encode error", "err", err)
		}
	}
}
