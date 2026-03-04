package main

import (
	"context"
	"fyp-api-gateway/src/config"
	"fyp-api-gateway/src/semantics"
	"fyp-api-gateway/src/watcher"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-sigChan:
			slog.WarnContext(ctx, "failed to create main context")
			cancel()

			time.Sleep(5 * time.Second)
			os.Exit(1)

		case <-ctx.Done():
		}
	}()
	
	go watcher.Watch()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("Error reading database connection", "error", "DATABASE_URL is not set")
	}

	db, err := config.NewDatabase(dsn)
	if err != nil {
		slog.Error("Error initialising database", "error", err)
		return
	}
	defer db.Conn.Close()

	if err = db.StartDB("/var/lib/init.sql"); err != nil {
		slog.Error("Error running migration script", "error", err)
		return
	}
	slog.Info("Initialised database")

	server := &config.Server{DB: db}

	mux := http.NewServeMux()

	// config handler routes
	mux.HandleFunc("/analyse", semantics.RecvConfig)
	mux.HandleFunc("/config/update", config.LoadNewConfig)

	// database routes
	mux.HandleFunc("/verify-signup", server.Signup)
	mux.HandleFunc("/verify-login", server.VerifyLoginInfo)
	mux.HandleFunc("/validate-session", server.ValidateSession)
	mux.HandleFunc("/api/gateway", server.UserConfig)

	slog.Info("Control plane listening on port 10000")
	err = http.ListenAndServe(":10000", mux)
	if err != nil {
		slog.Error("Error starting server", "error", err)
		cancel()
	}

}
