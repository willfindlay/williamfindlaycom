package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port              string
	ContentRepoURL    string
	ContentRepoBranch string
	ContentDir        string
	SyncInterval      time.Duration
	GitAuthToken      string
	SiteTitle         string
	SiteURL           string
	DevMode           bool
}

func Load() (*Config, error) {
	repoURL := os.Getenv("CONTENT_REPO_URL")
	if repoURL == "" {
		return nil, fmt.Errorf("CONTENT_REPO_URL is required")
	}

	syncInterval := 5 * time.Minute
	if v := os.Getenv("SYNC_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SYNC_INTERVAL %q: %w", v, err)
		}
		syncInterval = d
	}

	return &Config{
		Port:              envOr("PORT", "8080"),
		ContentRepoURL:    repoURL,
		ContentRepoBranch: envOr("CONTENT_REPO_BRANCH", "main"),
		ContentDir:        envOr("CONTENT_DIR", "/data/content"),
		SyncInterval:      syncInterval,
		GitAuthToken:      os.Getenv("GIT_AUTH_TOKEN"),
		SiteTitle:         envOr("SITE_TITLE", "William Findlay"),
		SiteURL:           envOr("SITE_URL", "https://williamfindlay.com"),
		DevMode:           os.Getenv("DEV_MODE") == "true",
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
