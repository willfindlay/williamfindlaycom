package handler

import (
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type homeData struct {
	PageData
	RecentPosts      []content.BlogPost
	FeaturedProjects []content.Project
}

func (d *Deps) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			d.notFound(w, r)
			return
		}

		store := d.Store.Load()

		data := homeData{PageData: d.basePage("")}
		data.CanonicalURL = d.SiteURL
		data.Description = "Personal website of William Findlay — software engineer, security researcher, and systems thinker."
		data.JSONLD = buildHomeJSONLD(d.SiteTitle, d.SiteURL)

		if store != nil {
			limit := 5
			if len(store.Posts) < limit {
				limit = len(store.Posts)
			}
			data.RecentPosts = store.Posts[:limit]

			for _, p := range store.Projects {
				if p.Featured {
					data.FeaturedProjects = append(data.FeaturedProjects, p)
				}
			}
		}

		d.render(w, "templates/home.html", data)
	}
}
