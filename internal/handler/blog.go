package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/willfindlay/williamfindlaycom/internal/config"
	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type blogListData struct {
	PageData
	Posts        []content.BlogPost
	AllPostsJSON template.JS
	AllTags      []string
	ActiveTag    string
	SearchQuery  string
}

type blogPostData struct {
	PageData
	Post   *content.BlogPost
	Giscus config.GiscusConfig
}

func (d *Deps) BlogList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := d.Store.Load()

		data := blogListData{PageData: d.basePage("blog")}
		data.PageTitle = "Blog"
		data.Description = "Blog posts by William Findlay"
		data.CanonicalURL = d.SiteURL + "/blog"
		data.JSONLD = buildCollectionPageJSONLD("Blog", data.Description, data.CanonicalURL)

		if store != nil {
			query := strings.TrimSpace(r.URL.Query().Get("q"))
			tag := r.URL.Query().Get("tag")

			if query != "" {
				data.SearchQuery = query
				data.Posts = content.SearchPosts(store.Posts, query)
			} else if tag != "" {
				data.ActiveTag = tag
				if posts, ok := store.PostsByTag[tag]; ok {
					for _, p := range posts {
						data.Posts = append(data.Posts, *p)
					}
				}
			} else {
				data.Posts = store.Posts
			}

			type postMeta struct {
				Slug        string   `json:"slug"`
				Title       string   `json:"title"`
				Date        string   `json:"date"`
				DateShort   string   `json:"dateShort"`
				Description string   `json:"description"`
				Tags        []string `json:"tags"`
			}
			meta := make([]postMeta, len(store.Posts))
			for i, p := range store.Posts {
				meta[i] = postMeta{
					Slug:        p.Slug,
					Title:       p.Title,
					Date:        p.Date.Format("January 2, 2006"),
					DateShort:   p.Date.Format("2006-01-02"),
					Description: p.Description,
					Tags:        p.Tags,
				}
			}
			if b, err := json.Marshal(meta); err == nil {
				data.AllPostsJSON = template.JS(b)
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

		data := blogPostData{PageData: d.basePage("blog"), Post: post, Giscus: d.Giscus}
		data.PageTitle = post.Title
		data.Description = post.Description
		data.CanonicalURL = d.SiteURL + "/blog/" + slug
		data.OGType = "article"
		data.JSONLD = buildBlogPostingJSONLD(post, d.SiteURL)

		d.render(w, "templates/blog/post.html", data)
	}
}
