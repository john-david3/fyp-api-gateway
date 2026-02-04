package watcher

import (
	"fyp-api-gateway/src/config"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

// TODO: Write unit tests for this

func Watch(gatewayConfig *config.GatewayConfig, store *config.ConfigStore) {
	slog.Info("Starting file watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("error creating watcher", "error", err)
	}
	defer watcher.Close()

	go func() {
		slog.Info("File watcher started")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				slog.Info("watcher event:", "event", event)
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
					slog.Info("watcher detected modified file:", "file", event.Name)

					// Perform Semantics Analysis

					gatewayConfig, err = config.LoadAndValidateConfigFile(event.Name)
					if err != nil {
						slog.Error("error loading config", "error", err)
					}

					err = config.UpdateNginxConfig(event.Name, "", gatewayConfig, store)
					if err != nil {
						slog.Error("error updating config", "error", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("watcher error:", "error", err)
			}
		}
	}()

	err = watcher.Add(config.WatcherDirName)
	if err != nil {
		slog.Error("error adding watcher:", "error", err)
	}

	<-make(chan struct{})
}
