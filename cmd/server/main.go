package main

import (
	"log/slog"
	"os"

	williamfindlaycom "github.com/willfindlay/williamfindlaycom"
	"github.com/willfindlay/williamfindlaycom/internal/config"
	"github.com/willfindlay/williamfindlaycom/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg, williamfindlaycom.Embedded)
	if err != nil {
		slog.Error("init", "err", err)
		os.Exit(1)
	}

	if err := srv.Run(); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
