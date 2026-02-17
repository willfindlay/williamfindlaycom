package handler

import (
	"net/http"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type projectListData struct {
	PageData
	Projects []content.Project
}

type projectDetailData struct {
	PageData
	Project *content.Project
}

func (d *Deps) ProjectList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		data := projectListData{PageData: d.basePage("projects")}
		data.PageTitle = "Projects"
		data.Description = "Projects by William Findlay"
		data.CanonicalURL = d.SiteURL + "/projects"
		data.JSONLD = buildCollectionPageJSONLD("Projects", data.Description, data.CanonicalURL)

		if store != nil {
			data.Projects = store.Projects
		}

		d.render(w, "templates/projects/list.html", data)
	}
}

func (d *Deps) ProjectDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		store := d.Store.Load()

		if store == nil {
			d.notFound(w, r)
			return
		}

		proj, ok := store.ProjectsBySlug[slug]
		if !ok {
			d.notFound(w, r)
			return
		}

		data := projectDetailData{PageData: d.basePage("projects"), Project: proj}
		data.PageTitle = proj.Title
		data.Description = proj.Description
		data.CanonicalURL = d.SiteURL + "/projects/" + slug

		d.render(w, "templates/projects/project.html", data)
	}
}
