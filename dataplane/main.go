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

type Response struct {
	Filename string `json:"filename"`
	Body     []byte `json:"body"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/handle-config", handleNewConfig)
	err := http.ListenAndServe(":1000", mux)
	if err != nil {
		slog.Error("Error starting HTTP server", "error", err)
	}
}

func handleNewConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling config update")
	res := &Response{}

	if r.Method != "POST" {
		slog.Error("Invalid request method", "method", r.Method)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		return
	}

	dir := filepath.Dir(res.Filename)
	err = os.MkdirAll(dir, 0644)
	if err != nil {
		slog.Error("Error creating directory", "error", err)
		return
	}

	username := filepath.Base(filepath.Dir(res.Filename))
	logDir := "/var/log/nginx/users/" + username
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		slog.Error("Error creating log directory", "error", err)
		return
	}

	file, err := os.Create(res.Filename)
	if err != nil {
		slog.Error("Error creating file", "filename", res.Filename)
		return
	}

	_, err = file.Write(res.Body)
	if err != nil {
		slog.Error("Error writing to file", "filename", res.Filename)
		return
	}

	err = applyNginxConfig()
	if err != nil {
		slog.Error("Error applying nginx config", "error", err, "filename", res.Filename)
		return
	}
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
