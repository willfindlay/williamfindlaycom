package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/willfindlay/williamfindlaycom/internal/config"
	"github.com/willfindlay/williamfindlaycom/internal/content"
	"github.com/willfindlay/williamfindlaycom/internal/handler"
	"github.com/willfindlay/williamfindlaycom/internal/render"
)

type Server struct {
	cfg           *config.Config
	static        fs.FS
	store         *content.AtomicStore
	deps          *handler.Deps
	cssBundle     []byte
	cssBundlePath string
}

func New(cfg *config.Config, embedded fs.FS) (*Server, error) {
	bundleBytes, bundlePath, err := buildCSSBundle(embedded)
	if err != nil {
		return nil, fmt.Errorf("building CSS bundle: %w", err)
	}

	renderer, err := render.New(embedded, bundlePath)
	if err != nil {
		return nil, fmt.Errorf("initializing renderer: %w", err)
	}

	store := content.NewAtomicStore()
	deps := &handler.Deps{
		Store:     store,
		Renderer:  renderer,
		SiteTitle: cfg.SiteTitle,
		SiteURL:   cfg.SiteURL,
		Particles: cfg.Particles,
		Giscus:    cfg.Giscus,
	}

	return &Server{
		cfg:           cfg,
		static:        embedded,
		store:         store,
		deps:          deps,
		cssBundle:     bundleBytes,
		cssBundlePath: bundlePath,
	}, nil
}

func buildCSSBundle(fsys fs.FS) ([]byte, string, error) {
	cssFiles := []string{
		"static/css/reset.css",
		"static/css/typography.css",
		"static/css/layout.css",
		"static/css/components.css",
		"static/css/syntax.css",
		"static/css/main.css",
	}

	var buf bytes.Buffer
	for _, f := range cssFiles {
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, "", fmt.Errorf("reading %s: %w", f, err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}

	h := sha256.Sum256(buf.Bytes())
	hash := hex.EncodeToString(h[:8]) // 16 hex chars is plenty
	path := fmt.Sprintf("/static/css/bundle.%s.css", hash)
	return buf.Bytes(), path, nil
}

func (s *Server) Run() error {
	syncCfg := content.SyncConfig{
		RepoURL:   s.cfg.ContentRepoURL,
		Branch:    s.cfg.ContentRepoBranch,
		Dir:       s.cfg.ContentDir,
		Interval:  s.cfg.SyncInterval,
		AuthToken: s.cfg.GitAuthToken,
	}

	// Initial content load
	if err := content.CloneOrPull(syncCfg); err != nil {
		return fmt.Errorf("initial content sync: %w", err)
	}

	cs, err := content.LoadFromDir(s.cfg.ContentDir)
	if err != nil {
		return fmt.Errorf("initial content load: %w", err)
	}
	s.store.Store(cs)
	slog.Info("content loaded", "posts", len(cs.Posts), "projects", len(cs.Projects))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go content.StartBackgroundSync(ctx, syncCfg, s.store)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.Port),
		Handler:      s.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
