package handler

import (
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type resumeData struct {
	PageData
	Resume *content.Resume
}

func (d *Deps) Resume() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		data := resumeData{PageData: d.basePage("resume")}
		data.PageTitle = "Resume"
		data.Description = "Resume of William Findlay"
		data.CanonicalURL = d.SiteURL + "/resume"

		if store != nil {
			data.Resume = store.Resume
		}

		d.render(w, "templates/resume.html", data)
	}
}
