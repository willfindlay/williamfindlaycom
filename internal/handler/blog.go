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
	ActiveTags   []string
	ActiveTagSet map[string]bool
	SearchQuery  string
}

type blogPostData struct {
	PageData
	Post     *content.BlogPost
	PrevPost *content.BlogPost // older
	NextPost *content.BlogPost // newer
	Giscus   config.GiscusConfig
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
			tags := r.URL.Query()["tag"]

			data.ActiveTagSet = make(map[string]bool)

			data.SearchQuery = query
			filtered := store.Posts

			if query != "" {
				filtered = content.SearchPosts(filtered, query)
			}

			if len(tags) > 0 {
				for _, t := range tags {
					data.ActiveTags = append(data.ActiveTags, t)
					data.ActiveTagSet[t] = true
				}
				filtered = postsWithAllTags(filtered, data.ActiveTagSet)
			}

			data.Posts = filtered

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

			allTagSet := make(map[string]bool)
			for t := range store.PostsByTag {
				allTagSet[t] = true
			}
			for t := range allTagSet {
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
		for i, p := range store.Posts {
			if p.Slug == slug {
				if i+1 < len(store.Posts) {
					data.PrevPost = &store.Posts[i+1] // older
				}
				if i > 0 {
					data.NextPost = &store.Posts[i-1] // newer
				}
				break
			}
		}
		data.PageTitle = post.Title
		data.Description = post.Description
		data.CanonicalURL = d.SiteURL + "/blog/" + slug
		data.OGType = "article"
		data.JSONLD = buildBlogPostingJSONLD(post, d.SiteURL)

		d.render(w, "templates/blog/post.html", data)
	}
}

func postsWithAllTags(posts []content.BlogPost, required map[string]bool) []content.BlogPost {
	var result []content.BlogPost
	for _, p := range posts {
		tags := make(map[string]bool, len(p.Tags))
		for _, t := range p.Tags {
			tags[t] = true
		}
		match := true
		for r := range required {
			if !tags[r] {
				match = false
				break
			}
		}
		if match {
			result = append(result, p)
		}
	}
	return result
}
