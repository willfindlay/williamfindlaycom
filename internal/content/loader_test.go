package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromDir_BlogPosts(t *testing.T) {
	dir := t.TempDir()
	blogDir := filepath.Join(dir, "blog")
	if err := os.MkdirAll(blogDir, 0o755); err != nil {
		t.Fatal(err)
	}

	post1 := `---
title: Second Post
date: 2024-01-15
description: The second post
tags: [go, testing]
---

Hello **world**.
`
	post2 := `---
title: First Post
date: 2024-06-01
description: The first post
tags: [go]
---

Some content here.
`
	if err := os.WriteFile(filepath.Join(blogDir, "second-post.md"), []byte(post1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "first-post.md"), []byte(post2), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	if len(store.Posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(store.Posts))
	}

	// Should be sorted newest first
	if store.Posts[0].Title != "First Post" {
		t.Errorf("expected newest post first, got %q", store.Posts[0].Title)
	}
	if store.Posts[1].Title != "Second Post" {
		t.Errorf("expected oldest post second, got %q", store.Posts[1].Title)
	}

	// Slugs
	if store.Posts[0].Slug != "first-post" {
		t.Errorf("expected slug 'first-post', got %q", store.Posts[0].Slug)
	}

	// PostsBySlug
	if _, ok := store.PostsBySlug["second-post"]; !ok {
		t.Error("PostsBySlug missing 'second-post'")
	}

	// PostsByTag
	goPosts := store.PostsByTag["go"]
	if len(goPosts) != 2 {
		t.Errorf("expected 2 posts tagged 'go', got %d", len(goPosts))
	}
	testingPosts := store.PostsByTag["testing"]
	if len(testingPosts) != 1 {
		t.Errorf("expected 1 post tagged 'testing', got %d", len(testingPosts))
	}

	// Content rendered
	if store.Posts[1].Content == "" {
		t.Error("expected rendered content, got empty")
	}
}

func TestLoadFromDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	store, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	if len(store.Posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(store.Posts))
	}
	if len(store.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(store.Projects))
	}
	if store.Resume != nil {
		t.Error("expected nil resume")
	}
}

func TestLoadFromDir_Projects(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "projects")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	proj := `---
title: My Project
date: 2024-03-10
description: A cool project
repo: https://github.com/example/project
status: active
featured: true
---

Project description.
`
	if err := os.WriteFile(filepath.Join(projDir, "my-project.md"), []byte(proj), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	if len(store.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(store.Projects))
	}

	p := store.Projects[0]
	if p.Slug != "my-project" {
		t.Errorf("expected slug 'my-project', got %q", p.Slug)
	}
	if p.Title != "My Project" {
		t.Errorf("expected title 'My Project', got %q", p.Title)
	}
	if !p.Featured {
		t.Error("expected featured=true")
	}

	if _, ok := store.ProjectsBySlug["my-project"]; !ok {
		t.Error("ProjectsBySlug missing 'my-project'")
	}
}

func TestRenderMarkdown(t *testing.T) {
	src := `---
title: Test
date: 2024-01-01
description: A test
tags: [a, b]
---

# Heading

Paragraph with **bold**.
`
	var post BlogPost
	rendered, err := renderMarkdown([]byte(src), &post)
	if err != nil {
		t.Fatalf("renderMarkdown: %v", err)
	}

	if post.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", post.Title)
	}
	if len(post.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(post.Tags))
	}
	if rendered == "" {
		t.Error("expected rendered HTML, got empty")
	}

	html := string(rendered)
	if !strings.Contains(html, "<h1") {
		t.Error("expected <h1 in rendered output")
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Error("expected <strong>bold</strong> in rendered output")
	}
}

func TestLoadFromDir_Resume(t *testing.T) {
	dir := t.TempDir()
	resumeDir := filepath.Join(dir, "resume")
	if err := os.MkdirAll(resumeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	resume := `# William Findlay

Software engineer.
`
	if err := os.WriteFile(filepath.Join(resumeDir, "resume.md"), []byte(resume), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	if store.Resume == nil {
		t.Fatal("expected resume, got nil")
	}
	if store.Resume.Content == "" {
		t.Error("expected rendered resume content, got empty")
	}
	if !strings.Contains(string(store.Resume.Content), "<h1") {
		t.Error("expected <h1 in resume output")
	}
}

func TestLoadFromDir_SkipsNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	blogDir := filepath.Join(dir, "blog")
	if err := os.MkdirAll(blogDir, 0o755); err != nil {
		t.Fatal(err)
	}

	post := `---
title: Only Post
date: 2024-01-01
description: test
---

Content.
`
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte(post), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "readme.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(blogDir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	store, err := LoadFromDir(dir)
	if err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	if len(store.Posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(store.Posts))
	}
}
