package server

import (
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

	renderer, err := render.New(williamfindlaycom.Embedded)
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
tags: [test]
---

Test content.
`
	if err := os.WriteFile(filepath.Join(blogDir, "test-post.md"), []byte(post), 0o644); err != nil {
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
		cfg:      &config.Config{},
		static:   williamfindlaycom.Embedded,
		store:    store,
		deps:     deps,
		renderer: renderer,
	}

	return httptest.NewServer(srv.routes())
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

	body, _ := io.ReadAll(resp.Body)
	if strings.TrimSpace(string(body)) != "ok" {
		t.Errorf("expected body 'ok', got %q", string(body))
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

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<feed") {
		t.Error("expected Atom feed element in body")
	}
}
