package content

import (
	"html/template"
	"time"
)

type BlogPost struct {
	Slug        string
	Title       string        `yaml:"title"`
	Date        time.Time     `yaml:"date"`
	Description string        `yaml:"description"`
	Tags        []string      `yaml:"tags"`
	Content     template.HTML // rendered markdown
}

type Project struct {
	Slug        string
	Title       string        `yaml:"title"`
	Date        time.Time     `yaml:"date"`
	Description string        `yaml:"description"`
	Tags        []string      `yaml:"tags"`
	Repo        string        `yaml:"repo"`
	URL         string        `yaml:"url"`
	Status      string        `yaml:"status"`
	Featured    bool          `yaml:"featured"`
	Content     template.HTML // rendered markdown
}

type Resume struct {
	Content template.HTML
}

type ContentStore struct {
	Posts       []BlogPost
	PostsBySlug map[string]*BlogPost
	PostsByTag  map[string][]*BlogPost

	Projects       []Project
	ProjectsBySlug map[string]*Project

	Resume *Resume
}
