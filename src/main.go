package main

import (
	"context"
	"fyp-api-gateway/src/api"
	"fyp-api-gateway/src/config"
	"fyp-api-gateway/src/watcher"
	"log/slog"
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
		return
	}

	go watcher.Watch(gatewayConfig, store)

	err = api.HostApi(store)
	if err != nil {
		slog.Error("Error creating api", "error", err)
		cancel()
	}
}
