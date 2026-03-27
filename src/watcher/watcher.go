package watcher

import (
	"bytes"
	"encoding/json"
	"fyp-api-gateway/src/utils"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func Watch() {
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
				if event.Has(fsnotify.Create) {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						slog.Info("New user directory detected, adding watcher", "dir", event.Name)
						if err = watcher.Add(event.Name); err != nil {
							slog.Error("error adding new user directory to watcher", "error", err)
						}

						zonePath := filepath.Join(event.Name, utils.NGINXZoneFileName)
						serverPath := filepath.Join(event.Name, utils.NGINXConfigFileName)
						_, zoneErr := os.Stat(zonePath)
						_, serverErr := os.Stat(serverPath)
						if zoneErr == nil && serverErr == nil {
							sendNginxToDataplane(zonePath, serverPath)
							continue
						}
					}
				}

				base := filepath.Base(event.Name)
				if base == utils.NGINXConfigFileName || base == utils.NGINXZoneFileName {
					if event.Has(fsnotify.Rename) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
						slog.Info("watcher detected modified file:", "file", event.Name)

						userDir := filepath.Dir(event.Name)
						zonePath := filepath.Join(userDir, utils.NGINXZoneFileName)
						serverPath := filepath.Join(userDir, utils.NGINXConfigFileName)

						sendNginxToDataplane(zonePath, serverPath)
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

	root := utils.NGINXUserDirName
	err = addUserDirs(watcher, root)
	if err != nil {
		slog.Error("error adding watcher:", "error", err)
	}

	err = watcher.Add(utils.NGINXUserDirName)
	if err != nil {
		slog.Error("error adding base watcher:", "error", err)
	}

	<-make(chan struct{})
}

func addUserDirs(w *fsnotify.Watcher, root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(root, entry.Name())
			if err = w.Add(dirPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func sendNginxToDataplane(zonePath, serverPath string) {
	files := []string{zonePath, serverPath}

	type SendData struct {
		Files map[string][]byte `json:"files"`
	}

	payload := SendData{
		Files: make(map[string][]byte),
	}

	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			slog.Error("error reading file", "file", file)
			return
		}
		payload.Files[file] = body
	}

	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("error encoding file", "error", err)
		return
	}

	req, err := http.NewRequest(
		"POST",
		"http://data-plane:1000/api/handle-config",
		bytes.NewBuffer(data),
	)
	if err != nil {
		slog.Error("error creating request to send to control plane", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error sending request to data plane", "error", err)
		return
	}
	defer resp.Body.Close()
}
