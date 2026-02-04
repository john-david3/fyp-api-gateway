package handler

import (
	"log/slog"
	"net/http"
)

type ConfigRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	slog.Info("request successfully reached backend")

	// Semantics analysis - diff, english

	// Return the results of the semantics analysis to the frontend

}

func handleError() {

}
