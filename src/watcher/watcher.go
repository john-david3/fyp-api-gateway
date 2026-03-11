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

						confPath := filepath.Join(event.Name, "nginx.conf")
						if _, err := os.Stat(confPath); err == nil {
							sendNginxToDataplane(confPath)
						}
						continue
					}
				}

				if filepath.Base(event.Name) == "nginx.conf" {
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
						slog.Info("watcher detected modified file:", "file", event.Name)

						// send the config to the dataplane!
						sendNginxToDataplane(event.Name)

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

func sendNginxToDataplane(filename string) {
	body, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("error opening file", "filename", filename, "error", err)
		return
	}

	type SendData struct {
		Filename string `json:"filename"`
		Body     []byte `json:"body"`
	}

	sendData := SendData{
		Filename: filename,
		Body:     body,
	}

	data, err := json.Marshal(sendData)
	if err != nil {
		slog.Error("error encoding file", "filename", filename, "error", err)
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

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error sending request to data plane", "error", err)
		return
	}
	defer resp.Body.Close()
}
