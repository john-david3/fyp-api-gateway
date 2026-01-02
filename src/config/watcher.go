package config

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

// TODO: Write unit tests for this

func Watch(gatewayConfig *GatewayConfig) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("error creating watcher", "error", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				slog.Info("watcher event:", "event", event)
				if event.Has(fsnotify.Write) {
					slog.Info("watcher detected modified file:", "file", event.Name)
					err = UpdateNginxConfig(event.Name, "", gatewayConfig)
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

	err = watcher.Add("../../dataplane/nginx/")
	if err != nil {
		slog.Error("error adding watcher:", "error", err)
	}

	<-make(chan struct{})
}
