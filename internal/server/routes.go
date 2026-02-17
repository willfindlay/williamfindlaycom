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
	mux.HandleFunc("GET /sitemap.xml", s.deps.Sitemap())

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
