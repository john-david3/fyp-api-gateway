package apis

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func CreateAPI() error {
	http.HandleFunc("/", mainRoute)
	http.HandleFunc("/products", products)
	http.HandleFunc("/health", health)

	slog.Info("server running on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("error starting http server")
		return err
	}

	return nil
}

func mainRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling mainRoute...")
}

func products(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]string{
		"service": "products",
		"path":    r.URL.Path,
	})

	if err != nil {
		slog.Error("error encoding json")
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
