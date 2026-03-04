package main

import (
	"encoding/json"
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
		slog.Error("Error applying nginx config", "filename", res.Filename)
		return
	}
}

func applyNginxConfig() error {
	if err := exec.Command("nginx", "-t").Run(); err != nil {
		return err
	}

	if err := exec.Command("nginx", "-s", "reload").Run(); err != nil {
		return err
	}

	return nil
}
