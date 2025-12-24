package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Run() {
	go startProductService()
	go startOrderService()

	select {}
}

func startProductService() {
	mux := http.NewServeMux()
	mux.HandleFunc("/products", products)
	mux.HandleFunc("/health", health)
	slog.Info("server running on port 9001")
	err := http.ListenAndServe(":9001", mux)
	if err != nil {
		slog.Error("error starting product service", "error", err)
	}
}

func startOrderService() {
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", orders)
	mux.HandleFunc("/health", health)
	slog.Info("server running on port 9002")
	err := http.ListenAndServe(":9002", mux)
	if err != nil {
		slog.Error("error starting order service", "error", err)
	}
}

func products(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]string{
		"service": "products",
		"path":    r.URL.Path,
	})
	if err != nil {
		slog.Error("error encoding json", "error", err)
	}
}

func orders(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]string{
		"service": "orders",
		"path":    r.URL.Path,
	})
	if err != nil {
		slog.Error("error encoding json", "error", err)
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
