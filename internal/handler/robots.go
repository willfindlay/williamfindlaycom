package handler

import (
	"net/http"
)

func (d *Deps) Robots() http.HandlerFunc {
	body := "User-agent: *\nAllow: /\n\nSitemap: " + d.SiteURL + "/sitemap.xml\n"

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(body)) //nolint:errcheck
	}
}
