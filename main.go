package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	packages, err := parsePackages(os.Getenv("PKGBADGE_PACKAGES"))
	if err != nil {
		log.Error("invalid PKGBADGE_PACKAGES", "error", err)
		os.Exit(1)
	}
	if len(packages) == 0 {
		log.Error("PKGBADGE_PACKAGES is required")
		os.Exit(1)
	}

	interval := 6 * time.Hour
	if v := os.Getenv("PKGBADGE_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Error("invalid PKGBADGE_INTERVAL", "error", err)
			os.Exit(1)
		}
		interval = d
	}

	port := "8080"
	if v := os.Getenv("PKGBADGE_PORT"); v != "" {
		port = v
	}

	cache := NewCache()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("initial scrape", "packages", len(packages))
	scrapeAll(ctx, packages, cache, log)

	// Background scraper.
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Info("scheduled scrape", "packages", len(packages))
				scrapeAll(ctx, packages, cache, log)
			}
		}
	}()

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{Addr: addr, Handler: newMux(cache)}

	go func() {
		<-ctx.Done()
		log.Info("shutting down")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Info("listening", "addr", addr, "interval", interval, "packages", len(packages))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

// parsePackages parses PKGBADGE_PACKAGES env var.
// Format: "owner/package" or "owner/repo/package" (comma-separated).
// The 3-part format is needed when the GitHub repo name differs from the
// GHCR package name (e.g. "Will-Luck/Docker-Sentinel/docker-sentinel").
func parsePackages(s string) ([]PackageRef, error) {
	if s == "" {
		return nil, nil
	}

	var refs []PackageRef
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, "/")
		switch len(parts) {
		case 2:
			refs = append(refs, PackageRef{
				Owner:   parts[0],
				Repo:    parts[1],
				Package: parts[1],
			})
		case 3:
			refs = append(refs, PackageRef{
				Owner:   parts[0],
				Repo:    parts[1],
				Package: parts[2],
			})
		default:
			return nil, fmt.Errorf("invalid package ref %q: expected owner/package or owner/repo/package", entry)
		}
	}
	return refs, nil
}
