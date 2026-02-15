package content

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type SyncConfig struct {
	RepoURL   string
	Branch    string
	Dir       string
	Interval  time.Duration
	AuthToken string
}

func auth(token string) *githttp.BasicAuth {
	if token == "" {
		return nil
	}
	return &githttp.BasicAuth{
		Username: "git",
		Password: token,
	}
}

func CloneOrPull(cfg SyncConfig) error {
	if _, err := os.Stat(filepath.Join(cfg.Dir, ".git")); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("checking content dir: %w", err)
		}
		slog.Info("cloning content repo", "url", cfg.RepoURL, "branch", cfg.Branch)
		_, err := git.PlainClone(cfg.Dir, false, &git.CloneOptions{
			URL:           cfg.RepoURL,
			ReferenceName: refName(cfg.Branch),
			SingleBranch:  true,
			Depth:         1,
			Auth:          auth(cfg.AuthToken),
		})
		return err
	}

	repo, err := git.PlainOpen(cfg.Dir)
	if err != nil {
		return fmt.Errorf("opening repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	slog.Info("pulling content repo")
	err = w.Pull(&git.PullOptions{
		ReferenceName: refName(cfg.Branch),
		SingleBranch:  true,
		Auth:          auth(cfg.AuthToken),
	})
	if err == git.NoErrAlreadyUpToDate {
		slog.Info("content already up to date")
		return nil
	}
	return err
}

func StartBackgroundSync(ctx context.Context, cfg SyncConfig, store *AtomicStore) {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := CloneOrPull(cfg); err != nil {
				slog.Error("content sync failed", "err", err)
				continue
			}

			cs, err := LoadFromDir(cfg.Dir)
			if err != nil {
				slog.Error("content reload failed", "err", err)
				continue
			}

			store.Store(cs)
			slog.Info("content reloaded",
				"posts", len(cs.Posts),
				"projects", len(cs.Projects),
			)
		}
	}
}

func refName(branch string) plumbing.ReferenceName {
	return plumbing.ReferenceName("refs/heads/" + branch)
}
