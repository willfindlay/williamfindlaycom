package content

import "testing"

func TestSearchPosts_MatchTitle(t *testing.T) {
	posts := []BlogPost{
		{Title: "Go Concurrency Patterns", PlainText: "body text"},
		{Title: "Python Basics", PlainText: "intro"},
	}
	results := SearchPosts(posts, "concurrency")
	if len(results) != 1 || results[0].Title != "Go Concurrency Patterns" {
		t.Errorf("expected 1 match on title, got %d", len(results))
	}
}

func TestSearchPosts_MatchDescription(t *testing.T) {
	posts := []BlogPost{
		{Title: "Post", Description: "deep dive into eBPF tracing"},
	}
	results := SearchPosts(posts, "ebpf")
	if len(results) != 1 {
		t.Errorf("expected 1 match on description, got %d", len(results))
	}
}

func TestSearchPosts_MatchTag(t *testing.T) {
	posts := []BlogPost{
		{Title: "Post", Tags: []string{"linux", "security"}},
	}
	results := SearchPosts(posts, "security")
	if len(results) != 1 {
		t.Errorf("expected 1 match on tag, got %d", len(results))
	}
}

func TestSearchPosts_MatchBody(t *testing.T) {
	posts := []BlogPost{
		{Title: "Post", PlainText: "This discusses kernel namespaces in detail."},
	}
	results := SearchPosts(posts, "namespaces")
	if len(results) != 1 {
		t.Errorf("expected 1 match on body, got %d", len(results))
	}
}

func TestSearchPosts_MultiWordAND(t *testing.T) {
	posts := []BlogPost{
		{Title: "Go Testing", Description: "unit tests in Go"},
		{Title: "Go Patterns", Description: "design patterns"},
	}
	results := SearchPosts(posts, "go testing")
	if len(results) != 1 || results[0].Title != "Go Testing" {
		t.Errorf("expected 1 match with AND semantics, got %d", len(results))
	}
}

func TestSearchPosts_EmptyQuery(t *testing.T) {
	posts := []BlogPost{{Title: "Post"}}
	results := SearchPosts(posts, "")
	if results != nil {
		t.Errorf("expected nil for empty query, got %d results", len(results))
	}
	results = SearchPosts(posts, "   ")
	if results != nil {
		t.Errorf("expected nil for whitespace query, got %d results", len(results))
	}
}

func TestSearchPosts_CaseInsensitive(t *testing.T) {
	posts := []BlogPost{
		{Title: "UPPERCASE Title", PlainText: "MiXeD cAsE body"},
	}
	results := SearchPosts(posts, "uppercase mixed")
	if len(results) != 1 {
		t.Errorf("expected case-insensitive match, got %d", len(results))
	}
}

func TestSearchPosts_NoMatches(t *testing.T) {
	posts := []BlogPost{
		{Title: "Go Post", Description: "about go", PlainText: "content"},
	}
	results := SearchPosts(posts, "nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
