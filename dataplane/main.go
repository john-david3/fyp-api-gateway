package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type ConfigRequest struct {
	Files map[string][]byte `json:"files"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/handle-config", handleNewConfig)

	slog.Info("Starting dataplane on :1000")
	if err := http.ListenAndServe(":1000", mux); err != nil {
		slog.Error("Error starting HTTP server", "error", err)
	}
}

func handleNewConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling config update")

	if r.Method != http.MethodPost {
		slog.Error("Invalid request method", "method", r.Method)
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	for path, content := range req.Files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			slog.Error("Error creating directory", "dir", dir, "error", err)
			continue
		}

		// Ensure username log dir exists
		username := filepath.Base(filepath.Dir(path))
		logDir := filepath.Join("/var/log/nginx/users", username)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			slog.Error("Error creating log directory", "dir", logDir, "error", err)
		}

		// Atomic write
		tmpFile, err := os.CreateTemp(dir, "nginx-*.conf")
		if err != nil {
			slog.Error("Error creating temp file", "dir", dir, "error", err)
			continue
		}

		if _, err := tmpFile.Write(content); err != nil {
			slog.Error("Error writing to temp file", "file", tmpFile.Name(), "error", err)
			tmpFile.Close()
			continue
		}

		if err := tmpFile.Close(); err != nil {
			slog.Error("Error closing temp file", "file", tmpFile.Name(), "error", err)
			continue
		}

		if err := os.Rename(tmpFile.Name(), path); err != nil {
			slog.Error("Error moving temp file to final location", "src", tmpFile.Name(), "dst", path, "error", err)
			continue
		}

		slog.Info("Updated file", "path", path)
	}

	// Apply NGINX config after all files updated
	if err := applyNginxConfig(); err != nil {
		slog.Error("Error applying NGINX config", "error", err)
		http.Error(w, "failed to apply nginx config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	slog.Info("Config update applied successfully")
}

func applyNginxConfig() error {
	slog.Info("reloading nginx config")
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", string(output))
	}

	cmd = exec.Command("nginx", "-s", "reload")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx reload failed: %s", string(output))
	}

	return nil
}
