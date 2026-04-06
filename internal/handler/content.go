package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".avif": true,
}

// ContentFile serves image files from the content directory.
func (d *Deps) ContentFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("path")

		if !allowedImageExts[strings.ToLower(filepath.Ext(path))] {
			d.notFound(w, r)
			return
		}

		full := filepath.Join(d.ContentDir, path)
		full = filepath.Clean(full)

		// Guard against path traversal.
		base := filepath.Clean(d.ContentDir) + string(os.PathSeparator)
		if !strings.HasPrefix(full, base) {
			d.notFound(w, r)
			return
		}

		if _, err := os.Stat(full); err != nil {
			d.notFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=3600, s-maxage=300")
		http.ServeFile(w, r, full)
	}
}
