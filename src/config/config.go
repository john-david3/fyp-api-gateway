package config

import (
	"bytes"
	"encoding/json"
	"fyp-api-gateway/src/utils"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type ConfRequest struct {
	Content string `json:"content"`
}

type TemplateData struct {
	Username    string
	Connections Connection
}

func InitUserNGINX(username string) error {
	gatewayConf, err := loadAndValidateGatewayConf(utils.DefaultConfigContent)
	if err != nil {
		slog.Error("failed to load and validate gateway config", "error", err)
		return err
	}

	nginxUserConfDir := utils.NGINXDirName + "users/" + username + "/"
	nginxUserConfFile := nginxUserConfDir + utils.NGINXConfigFileName
	nginxUserZoneFile := nginxUserConfDir + utils.NGINXZoneFileName

	// create new user nginx file
	_, err = os.Stat(nginxUserConfFile)
	if err == nil {
		slog.Error("NGINX config file already exists")
		return err
	}

	_, err = os.Stat(nginxUserZoneFile)
	if err == nil {
		slog.Error("NGINX zone file already exists")
		return err
	}

	err = os.MkdirAll(utils.NGINXDirName+"users/"+username, 0644)
	if err != nil {
		slog.Error("failed creating users directory", "error", err)
		return err
	}

	_, err = os.Create(nginxUserConfFile)
	if err != nil {
		slog.Error("failed creating users config", "error", err)
		return err
	}

	_, err = os.Create(nginxUserZoneFile)
	if err != nil {
		slog.Error("failed creating users config", "error", err)
		return err
	}

	templateData := buildTemplateData(username, gatewayConf)

	_, _, err = renderNginxTemplate(templateData)
	if err != nil {
		slog.Error("failed rendering NGINX template", "error", err)
		return err
	}

	return nil
}

func loadAndValidateGatewayConf(body string) (*GatewayConfig, error) {
	var meaningful []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			meaningful = append(meaningful, line)
		}
	}

	if len(meaningful) == 0 {
		return &GatewayConfig{}, nil
	}

	var config GatewayConfig
	decoder := yaml.NewDecoder(bytes.NewReader([]byte(body)))
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func renderNginxTemplate(data TemplateData) (string, string, error) {
	zoneTmpl, err := template.ParseFiles(utils.NGINXTemplateDirName + utils.NGINXZoneTemplateFileName)
	if err != nil {
		return "", "", err
	}

	serverTmpl, err := template.ParseFiles(utils.NGINXTemplateDirName + utils.NGINXTemplateFileName)
	if err != nil {
		return "", "", err
	}

	var zoneBuf bytes.Buffer
	if err = zoneTmpl.Execute(&zoneBuf, data); err != nil {
		slog.Error("Error executing template: ", "error", err)
		return "", "", err
	}

	var serverBuf bytes.Buffer
	if err = serverTmpl.Execute(&serverBuf, data); err != nil {
		slog.Error("Error executing template: ", "error", err)
		return "", "", err
	}

	return zoneBuf.String(), serverBuf.String(), nil
}

func buildTemplateData(username string, gw *GatewayConfig) TemplateData {
	conn := Connection{}

	for _, r := range gw.Connections.Routes {
		zoneName := strings.ReplaceAll(r.Path, "/", "_")
		if zoneName == "" {
			zoneName = "root"
		}

		conn.Routes = append(conn.Routes, Routes{
			Path:      r.Path,
			Url:       r.Url,
			Auth:      r.Auth,
			RateLimit: r.RateLimit,
			ZoneName:  zoneName,
		})
	}

	return TemplateData{Username: username, Connections: conn}
}

func LoadNewConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("loading new gateway config file")

	if r.Method != http.MethodPost {
		slog.Error("invalid method", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		slog.Error("session cookie not found in request", "error", err)
		http.Error(w, "cookie not found in request", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer func() {
		err = r.Body.Close()
		if err != nil {
			slog.Error("failed to close request body", "error", err)
			http.Error(w, "failed to close request body", http.StatusInternalServerError)
			return
		}
	}()

	var req ConfRequest
	if err = json.Unmarshal(body, &req); err != nil {
		slog.Error("failed to unmarshal request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = InsertNewConfig(cookie.Value, req.Content)
	if err != nil {
		slog.Error("failed to insert new config", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	username := RetrieveUserBySessionId(cookie.Value)
	nginxUserConfDir := utils.NGINXDirName + "users/" + username + "/"
	nginxUserConfPath := nginxUserConfDir + utils.NGINXConfigFileName

	gatewayConf, err := loadAndValidateGatewayConf(req.Content)
	if err != nil {
		slog.Error("failed to validate gateway config", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	templateData := buildTemplateData(username, gatewayConf)

	zoneStr, serverStr, err := renderNginxTemplate(templateData)
	if err != nil {
		slog.Error("failed to render NGINX template", "error", err)
		http.Error(w, "failed to render NGINX template", http.StatusInternalServerError)
		return
	}

	if err = atomicWrites(nginxUserConfDir, utils.NGINXZoneFileName, []byte(zoneStr)); err != nil {
		slog.Error("failed to write NGINX zone", "error", err)
		http.Error(w, "failed to write NGINX zone", http.StatusInternalServerError)
		return
	}

	if err = atomicWrites(nginxUserConfDir, utils.NGINXConfigFileName, []byte(serverStr)); err != nil {
		slog.Error("failed to write NGINX config", "error", err)
		http.Error(w, "failed to write NGINX zone", http.StatusInternalServerError)
		return
	}

	slog.Info("gateway config updated successfully", "path", nginxUserConfPath)
	w.WriteHeader(http.StatusOK)
}

func atomicWrites(dir, filename string, content []byte) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "nginx-*.conf")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(content); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err = tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), filepath.Join(dir, filename))
}
