package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type ConfigRequest struct {
	Content string `json:"content"`
}
type FindingsResponse struct {
	Errors  []string `json:"errors"`
	Updates []string `json:"updates"`
}

var findings FindingsResponse // TODO: This is cowboy code

func Gateway(w http.ResponseWriter, r *http.Request) {
	// get the session so we know what user is requesting their config
	cookie, err := r.Cookie("session")
	if err != nil {
		slog.Error("Error getting session id", "error", err)
		return
	}

	// Create a request object, don't send yet
	req, err := http.NewRequest(
		"POST",
		"http://control-plane:10000/api/gateway",
		nil,
	)
	if err != nil {
		slog.Error("error creating request to send to control plane", "error", err)
		return
	}

	req.Header.Set("Cookie", "session="+cookie.Value)

	// send the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error sending request to control plane", "error", err)
		return
	}
	defer resp.Body.Close()

	// display the file
	// set the content-type to the same as resp from control plane
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Error("error copying response body", "error", err)
		return
	}
}

func HandleNewConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("request successfully reached file handler")
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

	_, err = submitConfig([]byte(configRequest.Content), r)
	if err != nil {
		slog.Error("Error submitting config", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func submitConfig(cfg []byte, r *http.Request) (*http.Response, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		slog.Error("Error getting session id", "error", err)
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		"http://control-plane:10000/analyse",
		bytes.NewBuffer(cfg),
	)
	if err != nil {
		slog.Error("error creating request to send to control plane", "error", err)
		return nil, err
	}

	req.Header.Set("Cookie", "session="+cookie.Value)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error sending request to control plane", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func RecvFindings(w http.ResponseWriter, r *http.Request) {
	slog.Info("received config from control plane")

	err := json.NewDecoder(r.Body).Decode(&findings)
	if err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		return
	}
}

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

func HandleAcceptChanges(w http.ResponseWriter, r *http.Request) {
	slog.Info("user has accepted the changes to the config file")
	configRequest := ConfigRequest{}

	if r.Method != "POST" {
		slog.Error("Method not allowed", "method", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		slog.Error("Error getting session id", "error", err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&configRequest)
	if err != nil {
		slog.Error("Error unmarshalling request body", "error", err)
		return
	}

	configRequestObj, err := json.Marshal(configRequest)
	if err != nil {
		slog.Error("Error marshalling request body", "error", err)
		return
	}

	req, err := http.NewRequest(
		"POST",
		"http://control-plane:10000/config/update",
		bytes.NewBuffer(configRequestObj),
	)
	if err != nil {
		slog.Error("error creating request to send to control plane", "error", err)
		return
	}

	req.Header.Set("Cookie", "session="+cookie.Value)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error sending request to control plane", "error", err)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusOK)
}
