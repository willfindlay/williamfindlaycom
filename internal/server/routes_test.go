package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/willfindlay/williamfindlaycom/internal/config"
	"github.com/willfindlay/williamfindlaycom/internal/content"
	"github.com/willfindlay/williamfindlaycom/internal/handler"
	"github.com/willfindlay/williamfindlaycom/internal/render"

	williamfindlaycom "github.com/willfindlay/williamfindlaycom"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	bundleBytes, bundlePath, err := buildCSSBundle(williamfindlaycom.Embedded)
	if err != nil {
		t.Fatalf("buildCSSBundle: %v", err)
	}

	renderer, err := render.New(williamfindlaycom.Embedded, bundlePath)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}

	// Create content dir with a blog post
	contentDir := t.TempDir()
	blogDir := filepath.Join(contentDir, "blog")
	if err := os.MkdirAll(blogDir, 0o755); err != nil {
		t.Fatal(err)
	}
	post := `---
title: Test Post
date: 2024-01-01
description: A test post
tags: [test, go]
---

Test content.
`
	if err := os.WriteFile(filepath.Join(blogDir, "test-post.md"), []byte(post), 0o644); err != nil {
		t.Fatal(err)
	}

	post2 := `---
title: Second Post
date: 2024-01-02
description: Another test post
tags: [test, rust]
---

More content.
`
	if err := os.WriteFile(filepath.Join(blogDir, "second-post.md"), []byte(post2), 0o644); err != nil {
		t.Fatal(err)
	}

	redirectsYAML := `- from: /old-post
  to: /blog/test-post
  code: 301
`
	if err := os.WriteFile(filepath.Join(contentDir, "_redirects.yaml"), []byte(redirectsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	store := content.NewAtomicStore()
	cs, err := content.LoadFromDir(contentDir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	store.Store(cs)

	deps := &handler.Deps{
		Store:     store,
		Renderer:  renderer,
		SiteTitle: "Test Site",
		SiteURL:   "http://localhost",
		Particles: config.ParticleConfig{},
	}

	srv := &Server{
		cfg:           &config.Config{},
		static:        williamfindlaycom.Embedded,
		store:         store,
		deps:          deps,
		cssBundle:     bundleBytes,
		cssBundlePath: bundlePath,
	}

	return httptest.NewServer(srv.routes())
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}
	return string(body)
}

func TestRoutes_Home(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "<html") {
		t.Error("expected HTML in response body")
	}
}

func TestRoutes_NotFound(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("GET /nonexistent: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "Not Found") {
		t.Error("expected 'Not Found' in response body")
	}
}

func TestRoutes_BlogList(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blog")
	if err != nil {
		t.Fatalf("GET /blog: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "<html") {
		t.Error("expected HTML in response body")
	}
}

func TestRoutes_BlogPost_NotFound(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blog/nonexistent")
	if err != nil {
		t.Fatalf("GET /blog/nonexistent: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "Not Found") {
		t.Error("expected 'Not Found' in response body")
	}
}

func TestRoutes_Health(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if strings.TrimSpace(body) != "ok" {
		t.Errorf("expected body 'ok', got %q", body)
	}
}

func TestRoutes_StaticFiles(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/css/main.css")
	if err != nil {
		t.Fatalf("GET /static/css/main.css: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestRoutes_SecurityHeaders(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	checks := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	}
	for header, want := range checks {
		got := resp.Header.Get(header)
		if got != want {
			t.Errorf("%s: expected %q, got %q", header, want, got)
		}
	}

	csp := resp.Header.Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header to be set")
	}
}

func TestRoutes_Robots(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/robots.txt")
	if err != nil {
		t.Fatalf("GET /robots.txt: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %q", ct)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "User-agent: *") {
		t.Error("expected User-agent directive in body")
	}
	if !strings.Contains(body, "Sitemap:") {
		t.Error("expected Sitemap directive in body")
	}
}

func TestRoutes_Sitemap(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sitemap.xml")
	if err != nil {
		t.Fatalf("GET /sitemap.xml: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "xml") {
		t.Errorf("expected xml content type, got %q", ct)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "<urlset") {
		t.Error("expected urlset element in body")
	}
	if !strings.Contains(body, "/blog/test-post") {
		t.Error("expected test post URL in sitemap")
	}
}

func TestRoutes_BlogPost(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blog/test-post")
	if err != nil {
		t.Fatalf("GET /blog/test-post: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "application/ld+json") {
		t.Error("expected JSON-LD script tag in body")
	}
	if !strings.Contains(body, "BlogPosting") {
		t.Error("expected BlogPosting in JSON-LD")
	}
	if !strings.Contains(body, `og:type" content="article"`) {
		t.Error("expected og:type article in body")
	}
}

func TestRoutes_BlogSearch(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blog?q=test")
	if err != nil {
		t.Fatalf("GET /blog?q=test: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "Test Post") {
		t.Error("expected search results to contain 'Test Post'")
	}
	if !strings.Contains(body, `value="test"`) {
		t.Error("expected search input to preserve query value")
	}
}

func TestRoutes_BlogSearchNoResults(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blog?q=nonexistent")
	if err != nil {
		t.Fatalf("GET /blog?q=nonexistent: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "No posts matching") {
		t.Error("expected 'No posts matching' empty state")
	}
}

func TestRoutes_BlogTagFilter(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blog?tag=go")
	if err != nil {
		t.Fatalf("GET /blog?tag=go: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, `data-slug="test-post"`) {
		t.Error("expected test-post card in filtered results")
	}
	if strings.Contains(body, `data-slug="second-post"`) {
		t.Error("did not expect second-post card when filtering by 'go'")
	}
}

func TestRoutes_BlogMultiTagFilter(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// Both posts have "test", only "test-post" has "go"
	resp, err := http.Get(ts.URL + "/blog?tag=test&tag=go")
	if err != nil {
		t.Fatalf("GET /blog?tag=test&tag=go: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, `data-slug="test-post"`) {
		t.Error("expected test-post card matching both tags")
	}
	if strings.Contains(body, `data-slug="second-post"`) {
		t.Error("did not expect second-post card (has test but not go)")
	}

	// No post has both "go" and "rust"
	resp2, err := http.Get(ts.URL + "/blog?tag=go&tag=rust")
	if err != nil {
		t.Fatalf("GET /blog?tag=go&tag=rust: %v", err)
	}
	defer resp2.Body.Close() //nolint:errcheck

	body2 := readBody(t, resp2)
	if strings.Contains(body2, `data-slug="test-post"`) || strings.Contains(body2, `data-slug="second-post"`) {
		t.Error("expected no post cards matching both 'go' and 'rust'")
	}
}

func TestRoutes_BlogCombinedSearchAndTag(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// "another" matches only second-post (description), "test" tag matches both
	resp, err := http.Get(ts.URL + "/blog?q=another&tag=test")
	if err != nil {
		t.Fatalf("GET /blog?q=another&tag=test: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, `data-slug="second-post"`) {
		t.Error("expected second-post matching search and tag")
	}
	if strings.Contains(body, `data-slug="test-post"`) {
		t.Error("did not expect test-post (does not match search 'another')")
	}

	// "test" matches both posts, "go" tag matches only test-post
	resp2, err := http.Get(ts.URL + "/blog?q=test&tag=go")
	if err != nil {
		t.Fatalf("GET /blog?q=test&tag=go: %v", err)
	}
	defer resp2.Body.Close() //nolint:errcheck

	body2 := readBody(t, resp2)
	if !strings.Contains(body2, `data-slug="test-post"`) {
		t.Error("expected test-post matching search and tag")
	}
	if strings.Contains(body2, `data-slug="second-post"`) {
		t.Error("did not expect second-post (has 'test' in title but no 'go' tag)")
	}
}

func TestRoutes_Feed(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/feed.xml")
	if err != nil {
		t.Fatalf("GET /feed.xml: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "atom+xml") {
		t.Errorf("expected atom+xml content type, got %q", ct)
	}

	body := readBody(t, resp)
	if !strings.Contains(body, "<feed") {
		t.Error("expected Atom feed element in body")
	}
}

func TestRoutes_Redirect(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + "/old-post")
	if err != nil {
		t.Fatalf("GET /old-post: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusMovedPermanently {
		t.Errorf("expected 301, got %d", resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if loc != "/blog/test-post" {
		t.Errorf("expected Location /blog/test-post, got %q", loc)
	}
}

func TestRoutes_JSONFeed(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/feed.json")
	if err != nil {
		t.Fatalf("GET /feed.json: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/feed+json") {
		t.Errorf("expected application/feed+json content type, got %q", ct)
	}

	body := readBody(t, resp)

	var feed struct {
		Version     string `json:"version"`
		Title       string `json:"title"`
		HomePageURL string `json:"home_page_url"`
		FeedURL     string `json:"feed_url"`
		Items       []struct {
			ID          string   `json:"id"`
			URL         string   `json:"url"`
			Title       string   `json:"title"`
			ContentHTML string   `json:"content_html"`
			Tags        []string `json:"tags"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(body), &feed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if feed.Version != "https://jsonfeed.org/version/1.1" {
		t.Errorf("expected JSON Feed 1.1 version, got %q", feed.Version)
	}
	if feed.Title != "Test Site" {
		t.Errorf("expected title 'Test Site', got %q", feed.Title)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(feed.Items))
	}

	// Posts are sorted by date descending, so Second Post (2024-01-02) comes first
	if feed.Items[0].Title != "Second Post" {
		t.Errorf("expected first item 'Second Post', got %q", feed.Items[0].Title)
	}
	if feed.Items[0].ContentHTML == "" {
		t.Error("expected non-empty content_html")
	}
}
