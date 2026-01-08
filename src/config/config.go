package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigStore struct {
	mu      sync.RWMutex
	latest  string
	configs map[string]ConfigPayload
	meta    map[string]ConfigMetadata
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		configs: make(map[string]ConfigPayload),
		meta:    make(map[string]ConfigMetadata),
	}
}

func (s *ConfigStore) UpdateConfig(nginxConfig string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	version := time.Now().UTC().Format("20060102-150405")
	checksum := sha256.Sum256([]byte(nginxConfig))

	meta := ConfigMetadata{
		Version:   version,
		Checksum:  hex.EncodeToString(checksum[:]),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	payload := ConfigPayload{
		Version: version,
		Config:  nginxConfig,
	}

	s.latest = version
	s.meta[version] = meta
	s.configs[version] = payload

}

func (s *ConfigStore) CheckIsConfigUpdated(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.latest == "" {
		http.Error(w, "no config available", http.StatusNotFound)
		return
	}

	meta := s.meta[s.latest]
	_ = json.NewEncoder(w).Encode(meta)

}

func (s *ConfigStore) ServeConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payload, ok := s.configs[s.latest]
	if !ok {
		http.Error(w, "no config available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(payload.Config))
	if err != nil {
		slog.Error("failed to write config response", "error", err)
		return
	}
}

func RegisterConfigFile(store *ConfigStore) (*GatewayConfig, error) {
	// Check if the file exists
	configFile, err := checkFileExists(GatewayConfigDirName + GatewayConfigFileName)
	if err != nil {
		return nil, err
	}

	// Validate the config file and load in it into a struct
	config, err := LoadAndValidateConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	nginxStr, err := renderNginxTemplate(config)
	if err != nil {
		return nil, err
	}

	store.UpdateConfig(nginxStr)

	return config, nil
}

func UpdateNginxConfig(filepath, user string, gatewayConfig *GatewayConfig, store *ConfigStore) error {
	// TODO: Make a defaults page so tests can overwrite

	// Check if NGINX config exists
	_, err := checkFileExists(filepath)
	if err != nil {
		return err
	}

	// Render the NGINX config template. Stored in /etc/nginx
	nginxString, err := renderNginxTemplate(gatewayConfig)
	if err != nil {
		return err
	}

	store.UpdateConfig(nginxString)
	slog.Info("(+) Successfully updated nginx config!")
	fmt.Println(nginxString)

	return nil
}

/*
Renders an updated NGINX config file from the users `gatewayCfg` file.
Uses templates to load the config back into `/etc/nginx`. Called by `UpdateNginxConfig`.
*/
func renderNginxTemplate(gatewayCfg *GatewayConfig) (string, error) {
	// load the template file
	tmpl, err := template.ParseFiles(NGINXTemplateDirName + NGINXTemplateFileName)
	if err != nil {
		return "", err
	}

	// Create the updated NGINX config file and save it into the file containing the current config
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, gatewayCfg); err != nil {
		slog.Error("Error executing template: ", "error", err)
		return "", err
	}
	nginxString := buf.String()

	return nginxString, nil
}

func checkFileExists(filePath string) (string, error) {
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	return "", errors.New("filepath does not contain a valid config file: " + filePath)
}

func LoadAndValidateConfigFile(filepath string) (*GatewayConfig, error) {
	var config GatewayConfig

	fileBody, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(fileBody))
	if err = decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
