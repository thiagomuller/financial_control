package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/thiago/my_finances_api/internal/config"
	"github.com/thiago/my_finances_api/internal/database"
	"github.com/thiago/my_finances_api/internal/handlers"
	"github.com/thiago/my_finances_api/internal/scheduler"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	scheduler.Run(db)

	router := handlers.NewRouter(db, cfg)

	addr := ":" + cfg.Port
	slog.Info("server listening", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
