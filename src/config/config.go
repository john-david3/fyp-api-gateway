package config

import (
	"bytes"
	"encoding/json"
	"fyp-api-gateway/src/utils"
	"io"
	"log/slog"
	"net/http"
	"os"
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
	// load the default config
	gatewayConf, err := loadAndValidateGatewayConf(utils.DefaultConfigContent)
	if err != nil {
		slog.Error("failed to load and validate gateway config", "error", err)
		return err
	}

	nginxUserConfDir := utils.NGINXDirName + "users/" + username + "/" + utils.NGINXConfigFileName

	// create new user nginx file
	_, err = os.Stat(nginxUserConfDir)
	if err == nil {
		slog.Error("NGINX config file already exists")
		return err
	}

	err = os.MkdirAll(utils.NGINXDirName+"users/"+username, 0644)
	if err != nil {
		slog.Error("failed creating users directory", "error", err)
		return err
	}

	_, err = os.Create(nginxUserConfDir)
	if err != nil {
		slog.Error("failed creating users config", "error", err)
		return err
	}

	templateData := buildTemplateData(username, gatewayConf)

	err = renderNginxTemplate(templateData, nginxUserConfDir)
	if err != nil {
		slog.Error("failed rendering NGINX template", "error", err)
		return err
	}

	return nil
}

func loadAndValidateGatewayConf(body string) (*GatewayConfig, error) {
	var config GatewayConfig

	decoder := yaml.NewDecoder(bytes.NewReader([]byte(body)))
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func renderNginxTemplate(data TemplateData, nginxUserConfDir string) error {
	// load the template file
	tmpl, err := template.ParseFiles(utils.NGINXTemplateDirName + utils.NGINXTemplateFileName)
	if err != nil {
		return err
	}

	// Create the updated NGINX config file and save it into the file containing the current config
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		slog.Error("Error executing template: ", "error", err)
		return err
	}
	nginxString := buf.String()

	// write the file
	err = os.WriteFile(nginxUserConfDir, []byte(nginxString), 0644)
	if err != nil {
		return err
	}

	return nil
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

	err = renderNginxTemplate(templateData, nginxUserConfPath)
	if err != nil {
		slog.Error("failed to render NGINX template", "error", err)
		http.Error(w, "failed to render NGINX template", http.StatusInternalServerError)
		return
	}

	// Atomic writes
	renderedConfig, err := os.ReadFile(nginxUserConfPath)
	if err != nil {
		slog.Error("failed to read rendered NGINX template", "error", err)
		http.Error(w, "failed to render NGINX template", http.StatusInternalServerError)
		return
	}

	tempFile, err := os.CreateTemp(nginxUserConfDir, "nginx-*.conf")
	if err != nil {
		slog.Error("failed creating temp file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err = tempFile.Write(renderedConfig); err != nil {
		slog.Error("failed writing temp config", "error", err)
		tempFile.Close()
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err = tempFile.Close(); err != nil {
		slog.Error("failed closing temp file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Atomic replace
	if err = os.Rename(tempFile.Name(), nginxUserConfPath); err != nil {
		slog.Error("failed replacing config file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("gateway config updated successfully", "path", nginxUserConfPath)

	w.WriteHeader(http.StatusOK)
}
