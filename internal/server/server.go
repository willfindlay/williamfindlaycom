package server

import (
	"context"
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
	cfg      *config.Config
	static   fs.FS
	store    *content.AtomicStore
	deps     *handler.Deps
	renderer *render.Renderer
}

func New(cfg *config.Config, embedded fs.FS) (*Server, error) {
	renderer, err := render.New(embedded)
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
	}

	return &Server{
		cfg:      cfg,
		static:   embedded,
		store:    store,
		deps:     deps,
		renderer: renderer,
	}, nil
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
