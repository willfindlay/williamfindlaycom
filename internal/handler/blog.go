package handler

import (
	"net/http"
	"sort"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type blogListData struct {
	PageData
	Posts     []content.BlogPost
	AllTags   []string
	ActiveTag string
}

type blogPostData struct {
	PageData
	Post *content.BlogPost
}

func (d *Deps) BlogList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		data := blogListData{PageData: d.basePage("blog")}
		data.PageTitle = "Blog"
		data.Description = "Blog posts by William Findlay"

		if store != nil {
			tag := r.URL.Query().Get("tag")
			if tag != "" {
				data.ActiveTag = tag
				if posts, ok := store.PostsByTag[tag]; ok {
					for _, p := range posts {
						data.Posts = append(data.Posts, *p)
					}
				}
			} else {
				data.Posts = store.Posts
			}

			tags := make(map[string]bool)
			for t := range store.PostsByTag {
				tags[t] = true
			}
			for t := range tags {
				data.AllTags = append(data.AllTags, t)
			}
			sort.Strings(data.AllTags)
		}

		d.render(w, "templates/blog/list.html", data)
	}
}

func (d *Deps) BlogPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		store := d.Store.Load()

		if store == nil {
			d.notFound(w, r)
			return
		}

		post, ok := store.PostsBySlug[slug]
		if !ok {
			d.notFound(w, r)
			return
		}

		data := blogPostData{PageData: d.basePage("blog"), Post: post}
		data.PageTitle = post.Title
		data.Description = post.Description
		data.CanonicalURL = d.SiteURL + "/blog/" + slug

		d.render(w, "templates/blog/post.html", data)
	}
}
