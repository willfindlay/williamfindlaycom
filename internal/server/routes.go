package server

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/handler"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health())
	mux.HandleFunc("GET /{$}", s.deps.Home())
	mux.HandleFunc("GET /blog", s.deps.BlogList())
	mux.HandleFunc("GET /blog/{slug}", s.deps.BlogPost())
	mux.HandleFunc("GET /projects", s.deps.ProjectList())
	mux.HandleFunc("GET /projects/{slug}", s.deps.ProjectDetail())
	mux.HandleFunc("GET /resume", s.deps.Resume())
	mux.HandleFunc("GET /feed.xml", s.deps.Feed())
	mux.HandleFunc("GET /feed.json", s.deps.JSONFeed())
	mux.HandleFunc("GET /sitemap.xml", s.deps.Sitemap())
	mux.HandleFunc("GET /robots.txt", s.deps.Robots())
	mux.HandleFunc("GET /index", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	mux.HandleFunc("GET "+s.cssBundlePath, s.serveCSSBundle)

	staticFS, err := fs.Sub(s.static, "static")
	if err != nil {
		panic(fmt.Sprintf("embedded static fs: %v", err))
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", cacheStatic(http.FileServerFS(staticFS))))

	// Catch-all for 404
	mux.HandleFunc("GET /", s.deps.Home())

	var h http.Handler = mux
	h = securityHeaders(h)
	h = logging(h)
	return h
}

func (s *Server) serveCSSBundle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(s.cssBundle) //nolint:errcheck
}
