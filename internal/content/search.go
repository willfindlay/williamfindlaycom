package content

import (
	"strings"
)

// SearchPosts returns posts where all whitespace-separated terms in query
// appear (case-insensitively) in the post's title, description, tags, or body.
// Returns nil if query is empty.
func SearchPosts(posts []BlogPost, query string) []BlogPost {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil
	}

	var results []BlogPost
	for _, p := range posts {
		text := searchableText(&p)
		if matchesAll(text, terms) {
			results = append(results, p)
		}
	}
	return results
}

func searchableText(p *BlogPost) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(p.Title))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(p.Description))
	for _, tag := range p.Tags {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(tag))
	}
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(p.PlainText))
	return b.String()
}

func matchesAll(text string, terms []string) bool {
	for _, t := range terms {
		if !strings.Contains(text, t) {
			return false
		}
	}
	return true
}
