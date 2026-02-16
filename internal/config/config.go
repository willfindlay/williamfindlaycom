package config

import (
	"fmt"
	"log/slog"
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

	cfg := &Config{
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
			Count:           clampInt(envOrInt("PARTICLE_COUNT", 120), 1, 500),
			Speed:           clampFloat(envOrFloat("PARTICLE_SPEED", 0.3), 0.01, 10),
			SizeMin:         clampFloat(envOrFloat("PARTICLE_SIZE_MIN", 1), 0.1, 20),
			SizeMax:         clampFloat(envOrFloat("PARTICLE_SIZE_MAX", 2.5), 0.1, 20),
			ConnectDistance: clampInt(envOrInt("PARTICLE_CONNECT_DISTANCE", 140), 10, 1000),
			ConnectOpacity:  clampFloat(envOrFloat("PARTICLE_CONNECT_OPACITY", 0.08), 0, 1),
			PushRange:       clampInt(envOrInt("PARTICLE_PUSH_RANGE", 180), 10, 1000),
			PushForce:       clampFloat(envOrFloat("PARTICLE_PUSH_FORCE", 0.015), 0.001, 1),
			PulseSpeed:      clampFloat(envOrFloat("PARTICLE_PULSE_SPEED", 0.008), 0.0001, 0.1),
			Color:           envOr("PARTICLE_COLOR", "79,209,197"),
			ColorAlt:        envOr("PARTICLE_COLOR_ALT", "128,90,213"),
		},
	}

	if cfg.Particles.SizeMax < cfg.Particles.SizeMin {
		cfg.Particles.SizeMax = cfg.Particles.SizeMin
	}

	return cfg, nil
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
		slog.Warn("invalid env var, using default", "key", key, "value", v, "default", fallback)
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
		slog.Warn("invalid env var, using default", "key", key, "value", v, "default", fallback)
		return fallback
	}
	return f
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
