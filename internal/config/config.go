package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type ParticleConfig struct {
	Count           int
	Speed           float64
	SizeMin         float64
	SizeMax         float64
	ConnectDistance int
	ConnectOpacity  float64
	PushRange       int
	PushForce       float64
	PulseSpeed      float64
	Color           string // "r,g,b"
	ColorAlt        string // "r,g,b"
}

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
	Particles         ParticleConfig
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
		Particles: ParticleConfig{
			Count:           envOrInt("PARTICLE_COUNT", 120),
			Speed:           envOrFloat("PARTICLE_SPEED", 0.3),
			SizeMin:         envOrFloat("PARTICLE_SIZE_MIN", 1),
			SizeMax:         envOrFloat("PARTICLE_SIZE_MAX", 2.5),
			ConnectDistance: envOrInt("PARTICLE_CONNECT_DISTANCE", 140),
			ConnectOpacity:  envOrFloat("PARTICLE_CONNECT_OPACITY", 0.08),
			PushRange:       envOrInt("PARTICLE_PUSH_RANGE", 180),
			PushForce:       envOrFloat("PARTICLE_PUSH_FORCE", 0.015),
			PulseSpeed:      envOrFloat("PARTICLE_PULSE_SPEED", 0.008),
			Color:           envOr("PARTICLE_COLOR", "79,209,197"),
			ColorAlt:        envOr("PARTICLE_COLOR_ALT", "128,90,213"),
		},
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envOrFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}
