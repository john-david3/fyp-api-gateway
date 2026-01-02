package src

import (
	"context"
	"fyp-api-gateway/src/config"
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

	gatewayConfig, err := config.RegisterConfigFile()
	if err != nil {
		slog.Error("Error reading config file", "error", err)
		return
	}

	config.Watch(gatewayConfig)

}
