package main

import (
	"fyp-api-gateway/management/handler"
	"log/slog"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/file/upload", handler.HandleNewConfig)
	mux.HandleFunc("/file/findings", handler.RecvFindings)
	mux.HandleFunc("/file/retrieve", handler.Findings)
	mux.HandleFunc("/file/accept", handler.HandleAcceptChanges)
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		slog.Error("could not start management plane", "error", err)
	}
}
