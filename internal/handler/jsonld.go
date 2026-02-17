package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"

	"github.com/willfindlay/williamfindlaycom/internal/content"
)

type jsonLDBase struct {
	Context string `json:"@context,omitempty"`
	Type    string `json:"@type"`
}

type jsonLDWebSite struct {
	jsonLDBase
	Name string `json:"name"`
	URL  string `json:"url"`
}

type jsonLDPerson struct {
	jsonLDBase
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	SameAs []string `json:"sameAs,omitempty"`
}

type jsonLDBlogPosting struct {
	jsonLDBase
	Headline      string       `json:"headline"`
	Description   string       `json:"description,omitempty"`
	DatePublished string       `json:"datePublished"`
	Author        jsonLDPerson `json:"author"`
	URL           string       `json:"url"`
}

type jsonLDCollectionPage struct {
	jsonLDBase
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

func marshalJSONLD(v any) template.HTML {
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("jsonld marshal error", "err", err)
		return ""
	}
	return template.HTML(b)
}

func buildHomeJSONLD(siteTitle, siteURL string) template.HTML {
	graph := []any{
		jsonLDWebSite{
			jsonLDBase: jsonLDBase{Context: "https://schema.org", Type: "WebSite"},
			Name:       siteTitle,
			URL:        siteURL,
		},
		jsonLDPerson{
			jsonLDBase: jsonLDBase{Context: "https://schema.org", Type: "Person"},
			Name:       "William Findlay",
			URL:        siteURL,
			SameAs: []string{
				"https://github.com/willfindlay",
				"https://linkedin.com/in/willfindlay",
			},
		},
	}
	b, err := json.Marshal(graph)
	if err != nil {
		slog.Error("jsonld marshal error", "err", err)
		return ""
	}
	return template.HTML(b)
}

func buildBlogPostingJSONLD(post *content.BlogPost, siteURL string) template.HTML {
	ld := jsonLDBlogPosting{
		jsonLDBase:    jsonLDBase{Context: "https://schema.org", Type: "BlogPosting"},
		Headline:      post.Title,
		Description:   post.Description,
		DatePublished: post.Date.Format("2006-01-02"),
		Author: jsonLDPerson{
			jsonLDBase: jsonLDBase{Type: "Person"},
			Name:       "William Findlay",
			URL:        siteURL,
		},
		URL: siteURL + "/blog/" + post.Slug,
	}
	return marshalJSONLD(ld)
}

func buildCollectionPageJSONLD(name, description, url string) template.HTML {
	ld := jsonLDCollectionPage{
		jsonLDBase:  jsonLDBase{Context: "https://schema.org", Type: "CollectionPage"},
		Name:        name,
		Description: description,
		URL:         url,
	}
	return marshalJSONLD(ld)
}
