package main

import (
	"fyp-api-gateway/management/auth"
	"fyp-api-gateway/management/handler"
	"log/slog"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	})

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/auth.html")
	})

	mux.HandleFunc("/config", auth.RequireSession(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/config.html")
	}))

	mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	mux.HandleFunc("/file/gateway", auth.RequireSession(handler.GetGateway))
	mux.HandleFunc("/file/upload", handler.HandleNewConfig)
	mux.HandleFunc("/file/findings", handler.RecvFindings)
	mux.HandleFunc("/file/retrieve", handler.Findings)
	mux.HandleFunc("/file/accept", handler.HandleAcceptChanges)
	mux.HandleFunc("/api/login", auth.Login)
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		slog.Error("could not start management plane", "error", err)
	}
}
