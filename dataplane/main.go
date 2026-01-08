package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fyp-api-gateway/src/config"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	controlPlaneURL = "http://control-plane:10000"
	NginxDirectory  = "/etc/nginx/"
	NginxFile       = "nginx.conf"
)

func main() {
	var checksum string
	ctx, _ := context.WithCancel(context.Background())

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			slog.Info("Polling for new NGINX config")
			meta, err := fetchMetadata()
			if err != nil {
				slog.Error("failed to fetch metadata", "error", err)
				continue
			}

			if meta.Checksum == checksum {
				continue
			}

			slog.Info("new config detected", "version", meta.Version)

			cfg, err := fetchConfig(meta.Version)
			if err != nil {
				slog.Error("failed to fetch configuration", "error", err)
				continue
			}

			if err := applyNginxConfig(cfg); err != nil {
				slog.Error("failed to apply nginx configuration", "error", err)
				continue
			}

			checksum = meta.Checksum
		}
	}

}

func fetchMetadata() (*config.ConfigMetadata, error) {
	resp, err := http.Get(controlPlaneURL + "/v1/config/metadata/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var metadata config.ConfigMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func fetchConfig(version string) (string, error) {
	resp, err := http.Get(controlPlaneURL + "/v1/config/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func applyNginxConfig(cfg string) error {
	const path = NginxDirectory + NginxFile

	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		return err
	}

	if err := exec.Command("nginx", "-t").Run(); err != nil {
		return err
	}

	if err := exec.Command("nginx", "-s", "reload").Run(); err != nil {
		return err
	}

	return nil
}
