package api

import (
	"fyp-api-gateway/src/config"
	"log/slog"
	"net/http"
)

func HostApi(store *config.ConfigStore) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/config/metadata/latest", store.CheckIsConfigUpdated)
	mux.HandleFunc("/v1/config/latest", store.ServeConfig)
	slog.Info("Control plane API listening on port 10000")
	err := http.ListenAndServe(":10000", mux)
	if err != nil {
		return err
	}

	return nil
}
