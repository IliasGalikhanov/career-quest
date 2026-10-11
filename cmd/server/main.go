package main

import (
	"careerquest/internal/httpapi"
	"careerquest/internal/store"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		client := http.Client{Timeout: 3 * time.Second}
		res, err := client.Get("http://127.0.0.1:8080/api/health")
		if err != nil {
			os.Exit(1)
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			os.Exit(1)
		}
		return
	}
	dbPath := env("CQ_DATABASE", "data/career-quest.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		slog.Error("create data directory", "error", err)
		os.Exit(1)
	}
	db, err := store.Open(dbPath)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	server := &http.Server{Addr: env("CQ_ADDR", "127.0.0.1:8080"), Handler: httpapi.New(db, env("CQ_WEB_DIR", "web/dist")), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			slog.Error("shutdown", "error", err)
		}
	}()
	slog.Info("Career Quest listening", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
