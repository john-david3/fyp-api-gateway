package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

type ConfigRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

var findings map[string][]string

func Findings(w http.ResponseWriter, r *http.Request) {
	// send the findings to the front end
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(findings)
	if err != nil {
		slog.Error("Error marshalling request body", "error", err)
		return
	}
}

func HandleNewConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("request successfully reached backend")
	configRequest := ConfigRequest{}

	if r.Method != "POST" {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&configRequest)
	if err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		return
	}

	_, err = submitConfig([]byte(configRequest.Content))
	if err != nil {
		slog.Error("Error submitting config", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func RecvFindings(w http.ResponseWriter, r *http.Request) {
	slog.Info("received config from control plane")

	err := json.NewDecoder(r.Body).Decode(&findings)
	if err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		return
	}
}

func HandleAcceptChanges(w http.ResponseWriter, r *http.Request) {
	slog.Info("user has accepted the changes to the config file")
	configRequest := ConfigRequest{}

	if r.Method != "POST" {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&configRequest)
	if err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		return
	}

	_, err = http.Post(
		"http://control-plane:10000/config/update",
		"application/x-yaml",
		bytes.NewBuffer([]byte(configRequest.Content)),
	)
}

func submitConfig(cfg []byte) (*http.Response, error) {
	return http.Post(
		"http://control-plane:10000/analyse",
		"application/x-yaml",
		bytes.NewBuffer(cfg),
	)
}

func GetGateway(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/gateway.yaml")
}
