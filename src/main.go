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

	store := config.NewConfigStore()

	gatewayConfig, err := config.RegisterConfigFile(store)
	if err != nil {
		slog.Error("Error reading config file", "error", err)
		cancel()
	}

	go watcher.Watch(gatewayConfig, store)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/config/metadata/latest", store.CheckIsConfigUpdated)
	mux.HandleFunc("/v1/config/latest", store.ServeConfig)
	mux.HandleFunc("/analyse", semantics.RecvConfig)
	mux.HandleFunc("/config/update", config.LoadNewConfig)
	slog.Info("Control plane listening on port 10000")
	err = http.ListenAndServe(":10000", mux)
	if err != nil {
		slog.Error("Error starting server", "error", err)
		cancel()
	}

}
