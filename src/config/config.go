package config

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"
)

const (
	GatewayConfigFileName = "gateway.yaml"
	GatewayConfigDirName  = "/etc/config/"
	NGINXConfigDirName    = "../../dataplane/nginx/"
	NGINXConfigFileName   = "nginx.conf"
	NGINXUserDirName      = "/dataplane/nginx/users/"
)

func RegisterConfigFile() (*GatewayConfig, error) {
	// Check if the file exists
	configFile, err := checkFileExists(GatewayConfigDirName + GatewayConfigFileName)
	if err != nil {
		return nil, err
	}

	// Validate the config file and load in it into a struct
	config, err := loadAndValidateConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	err = renderNginxTemplate(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func UpdateNginxConfig(filepath, user string, gatewayConfig *GatewayConfig) error {
	/* Called when watcher finds new changes to gateway config file */
	// TODO: Make a defaults page so tests can overwrite

	// Check if NGINX config exists, and if so, open it
	_, err := checkFileExists(filepath)
	if err != nil {
		return err
	}

	// Render the NGINX config template
	err = renderNginxTemplate(gatewayConfig)
	if err != nil {
		return err
	}

	return nil
}

func renderConfigsAtomically(allConfigs map[string]string) error {
	tmpDir := "/tmp/nginx-config-new"
	os.MkdirAll(tmpDir, 0755)

	for tenant, cfg := range allConfigs {
		filename := filepath.Join(tmpDir, tenant+".conf")
		os.WriteFile(filename, []byte(cfg), 0644)
	}

	// Validate
	cmd := exec.Command("nginx", "-t", "-c", filepath.Join(tmpDir, "nginx.conf"))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Atomic swap
	oldDir := "/etc/nginx/conf.d.old"
	os.Rename("/etc/nginx/conf.d", oldDir)
	os.Rename(tmpDir, "/etc/nginx/conf.d")

	// Reload NGINX
	exec.Command("nginx", "-s", "reload").Run()
	return nil
}


func renderNginxTemplate(gatewayCfg *GatewayConfig) error {
	tmpl, err := template.ParseFiles("../templates/nginx.conf.tmpl")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(NGINXConfigDirName+NGINXConfigFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	if err = tmpl.Execute(f, gatewayCfg); err != nil {
		slog.Error("Error executing template: ", "error", err)
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func checkFileExists(filePath string) (string, error) {
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	return "", errors.New("filepath does not contain a valid config file")
}

func loadAndValidateConfigFile(filepath string) (*GatewayConfig, error) {
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

func validateNginx() error {
	cmd := exec.Command("nginx", "-t")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nginx validation failed: %v", err)
	}

	if err := exec.Command("nginx", "-s", "reload").Run(); err != nil {
		return fmt.Errorf("nginx validation failed: %v", err)
	}

	return nil
}
